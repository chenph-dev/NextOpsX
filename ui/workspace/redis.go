package workspace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/dbexec"
	"nextopsx/i18n"
	hoststore "nextopsx/store/hosts"
)

// redisPreviewLimit caps collection previews (list/hash/set/zset).
const redisPreviewLimit = 200

// RedisKeysPageSize is how many keys one page of the key list shows.
const RedisKeysPageSize = 20

// redisScanStepsLimit stops a page load from chasing a sparse MATCH through
// many SCAN round trips; the page then shows what it found so far.
const redisScanStepsLimit = 100

// redisPageStart is where a page's SCAN begins: a cursor plus keys that
// overflowed the previous page. done means the scan already reached its
// end, so only pending is left.
type redisPageStart struct {
	cursor  string
	pending []string
	done    bool
}

// RedisDetail is the selected key's type, TTL and value.
type RedisDetail struct {
	Key       string
	Type      string
	TTL       int64
	TTLKnown  bool
	Truncated bool
	Result    *dbexec.Result
	Loading   bool
	Err       string
}

// RedisMetric is one server INFO metric.
type RedisMetric struct {
	Name  string
	Value string
}

// RedisState is the Redis workbench state.
type RedisState struct {
	HostID   string
	HostName string
	DBLabel  string

	Keys        []string
	Cursor      string
	Pattern     string
	ScanPattern string
	KeysLoading bool
	KeysErr     string
	// Page is the zero-based key page shown; pageStarts records where
	// each loaded page's scan starts.
	Page       int
	pageStarts []redisPageStart

	SelectedKey string
	Detail      RedisDetail

	MainTab     int
	Info        []RedisMetric
	InfoLoading bool
	InfoErr     string

	Command        string
	CommandResult  *dbexec.Result
	CommandErr     string
	CommandRunning bool
	CommandHistory []string
	HistoryOpen    bool

	DetailGrid  ui.ListState
	CommandGrid ui.ListState

	EditOpen  bool
	EditKey   string
	EditValue string

	TTLOpen  bool
	TTLKey   string
	TTLValue string

	ConfirmDeleteKey string

	Notice string
	Err    string

	keySeq    uint64
	detailSeq uint64
	infoSeq   uint64
	cmdSeq    uint64
}

// redisInfoKeys lists the INFO fields shown in the server tab, in order.
var redisInfoKeys = []string{
	"redis_version", "redis_mode", "os", "uptime_in_seconds",
	"connected_clients", "used_memory", "used_memory_human",
	"used_memory_peak_human", "total_commands_processed",
	"keyspace_hits", "keyspace_misses", "expired_keys", "evicted_keys",
	"total_net_input_bytes", "total_net_output_bytes",
}

// IsRedisProtocol reports whether a protocol is Redis.
func (a *State) IsRedisProtocol(protocol string) bool {
	return strings.EqualFold(protocol, "redis")
}

func redisDBLabel(n *hoststore.TreeNode) string {
	db := "0"
	if v, ok := n.Params["db"]; ok && fmt.Sprint(v) != "" {
		db = fmt.Sprint(v)
	}
	return n.Name + " · db" + db
}

// AttachRedis initializes the Redis workbench.
func (a *State) AttachRedis(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" || !a.IsRedisProtocol(n.Protocol) {
		return
	}
	if a.Redis.HostID == n.ID {
		a.Redis.HostName = n.Name
		a.Redis.DBLabel = redisDBLabel(n)
		return
	}
	a.Redis = RedisState{
		HostID: n.ID, HostName: n.Name, DBLabel: redisDBLabel(n),
		Cursor: "0",
	}
	a.RedisRefreshKeys(true)
}

// RedisReloadKeys reloads the current page without dropping a selection.
func (a *State) RedisReloadKeys() {
	if a.Redis.KeysLoading || len(a.Redis.pageStarts) == 0 {
		return
	}
	a.loadRedisKeysPage(a.Redis.Page)
}

// EnsureRedis refreshes the Redis panel when it is re-entered.
func (a *State) EnsureRedis() {
	n := a.Selected
	if n == nil || !a.IsRedisProtocol(n.Protocol) {
		return
	}
	if a.Redis.HostID != n.ID {
		a.AttachRedis(n)
		return
	}
	if !a.Redis.KeysLoading {
		a.RedisRefreshKeys(true)
	}
	if a.Redis.MainTab == 1 {
		a.RedisLoadInfo()
	}
}

// RedisRefreshKeys restarts at the first page, or moves to the next page.
func (a *State) RedisRefreshKeys(reset bool) {
	if reset {
		a.resetRedisKeys(true)
		return
	}
	a.RedisNextPage()
}

func (a *State) resetRedisKeys(clearSelection bool) {
	hostID := a.Redis.HostID
	if hostID == "" {
		return
	}
	// A SCAN cursor is only valid with the MATCH argument it was created
	// with. Keep the pattern used for this scan separately from the
	// editable search box.
	a.Redis.ScanPattern = strings.TrimSpace(a.Redis.Pattern)
	a.Redis.Page = 0
	a.Redis.pageStarts = []redisPageStart{{cursor: "0"}}
	a.Redis.Cursor = "0"
	a.Redis.Keys = nil
	a.Redis.KeysErr = ""
	a.Redis.Err = ""
	a.Redis.Notice = ""
	if clearSelection {
		a.Redis.SelectedKey = ""
		a.Redis.Detail = RedisDetail{}
	}
	a.loadRedisKeysPage(0)
}

// loadRedisKeysPage collects up to RedisKeysPageSize keys from where page
// starts, keeping any extra keys for the next page.
func (a *State) loadRedisKeysPage(page int) {
	if page < 0 || page >= len(a.Redis.pageStarts) {
		return
	}
	hostID := a.Redis.HostID
	if hostID == "" {
		return
	}
	start := a.Redis.pageStarts[page]
	pattern := a.Redis.ScanPattern
	a.Redis.keySeq++
	seq := a.Redis.keySeq
	a.Redis.KeysLoading = true
	a.Redis.KeysErr = ""
	a.Redis.Err = ""
	a.Redis.Notice = ""
	a.Redis.Page = page
	b := a.Backend
	if b == nil {
		a.Redis.KeysLoading = false
		a.Redis.KeysErr = i18n.T("redis.error.backendUninitialized")
		return
	}
	go func() {
		keys, next, pending, err := scanRedisKeysPage(
			b.DB, hostID, start, pattern, RedisKeysPageSize)
		a.Post(func() {
			if hostID != a.Redis.HostID || seq != a.Redis.keySeq {
				return
			}
			a.Redis.KeysLoading = false
			if err != nil {
				a.Redis.KeysErr = i18n.Tf("redis.error.scanFailed", err.Error())
				return
			}
			a.Redis.Keys = keys
			a.Redis.Cursor = next
			a.recordRedisPageStart(page, next, pending)
		})
	}()
}

// recordRedisPageStart remembers where the next page starts and drops any
// pages recorded beyond it.
func (a *State) recordRedisPageStart(page int, next string, pending []string) {
	if next == "0" && len(pending) == 0 {
		if page+1 < len(a.Redis.pageStarts) {
			a.Redis.pageStarts = a.Redis.pageStarts[:page+1]
		}
		return
	}
	start := redisPageStart{cursor: next, pending: pending, done: next == "0"}
	if page+1 < len(a.Redis.pageStarts) {
		a.Redis.pageStarts[page+1] = start
		a.Redis.pageStarts = a.Redis.pageStarts[:page+2]
		return
	}
	a.Redis.pageStarts = append(a.Redis.pageStarts, start)
}

// RedisPrevPage shows the page before the current one.
func (a *State) RedisPrevPage() {
	if a.Redis.KeysLoading || a.Redis.Page <= 0 {
		return
	}
	a.loadRedisKeysPage(a.Redis.Page - 1)
}

// RedisHasPrev reports whether a page before the current one exists.
func (a *State) RedisHasPrev() bool { return a.Redis.Page > 0 }

// RedisNextPage shows the page after the current one.
func (a *State) RedisNextPage() {
	if a.Redis.KeysLoading || !a.RedisHasNext() {
		return
	}
	a.loadRedisKeysPage(a.Redis.Page + 1)
}

// RedisHasNext reports whether a page after the current one exists.
func (a *State) RedisHasNext() bool {
	return a.Redis.Page+1 < len(a.Redis.pageStarts)
}

// RedisSearch restarts the scan with the current pattern.
func (a *State) RedisSearch() { a.RedisRefreshKeys(true) }

// RedisSelectKey loads a key's type, TTL and value.
func (a *State) RedisSelectKey(key string) {
	hostID := a.Redis.HostID
	if hostID == "" || key == "" {
		return
	}
	a.Redis.SelectedKey = key
	a.Redis.detailSeq++
	seq := a.Redis.detailSeq
	a.Redis.Detail = RedisDetail{Key: key, Loading: true}
	a.Redis.EditOpen = false
	a.Redis.TTLOpen = false
	if a.Backend == nil {
		a.Redis.Detail.Loading = false
		a.Redis.Detail.Err = i18n.T("redis.error.backendUninitialized")
		return
	}
	b := a.Backend
	go func() {
		typeRes, err := b.DB.ExecuteSQL(hostID, "TYPE "+RedisQuote(key))
		if err != nil {
			a.postRedisDetail(hostID, seq, RedisDetail{Key: key, Err: err.Error()})
			return
		}
		typ := strings.ToLower(redisFirstCell(typeRes))
		if typ == "" {
			typ = "string"
		}
		detail := RedisDetail{Key: key, Type: typ}
		if ttlRes, terr := b.DB.ExecuteSQL(hostID, "TTL "+RedisQuote(key)); terr == nil {
			if ttl, perr := strconv.ParseInt(redisFirstCell(ttlRes), 10, 64); perr == nil {
				detail.TTL = ttl
				detail.TTLKnown = true
			}
		}

		// Collections are previews, capped so a huge list/hash/set cannot
		// allocate unbounded memory. HSCAN/SSCAN also return a cursor used
		// to report truncation.
		limit := redisPreviewLimit
		var valueCmd string
		switch typ {
		case "hash":
			valueCmd = fmt.Sprintf("HSCAN %s 0 COUNT %d", RedisQuote(key), limit)
		case "list":
			valueCmd = fmt.Sprintf("LRANGE %s 0 %d", RedisQuote(key), limit-1)
		case "set":
			valueCmd = fmt.Sprintf("SSCAN %s 0 COUNT %d", RedisQuote(key), limit)
		case "zset":
			valueCmd = fmt.Sprintf("ZRANGE %s 0 %d WITHSCORES", RedisQuote(key), limit-1)
		case "string":
			valueCmd = "GET " + RedisQuote(key)
		}
		if valueCmd != "" {
			if res, verr := b.DB.ExecuteSQL(hostID, valueCmd); verr != nil {
				detail.Err = verr.Error()
			} else if typ == "hash" || typ == "set" {
				if parsed, truncated, ok := redisScanCollection(typ, res); ok {
					detail.Result = parsed
					detail.Truncated = truncated
				} else {
					detail.Result = res
				}
			} else {
				detail.Result = res
				if typ == "list" && len(res.Rows) >= limit {
					detail.Truncated = true
				}
				if typ == "zset" && len(res.Rows) >= limit*2 {
					detail.Truncated = true
				}
			}
		}
		a.postRedisDetail(hostID, seq, detail)
	}()
}

func (a *State) postRedisDetail(hostID string, seq uint64, d RedisDetail) {
	a.Post(func() {
		if hostID != a.Redis.HostID || seq != a.Redis.detailSeq {
			return
		}
		a.Redis.Detail = d
	})
}

// RedisOpenDelete confirms deleting the selected key.
func (a *State) RedisOpenDelete() {
	if a.Redis.SelectedKey == "" {
		return
	}
	a.Redis.ConfirmDeleteKey = a.Redis.SelectedKey
}

// RedisConfirmDelete deletes the confirmed key.
func (a *State) RedisConfirmDelete() {
	key := a.Redis.ConfirmDeleteKey
	hostID := a.Redis.HostID
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.DB.ExecuteSQL(hostID, "DEL "+RedisQuote(key))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = i18n.Tf("redis.error.deleteFailed", err.Error())
				return
			}
			a.Redis.ConfirmDeleteKey = ""
			if a.Redis.SelectedKey == key {
				a.Redis.SelectedKey = ""
				a.Redis.Detail = RedisDetail{}
			}
			a.RedisReloadKeys()
			a.Redis.Notice = i18n.Tf("redis.notice.deleted", key)
		})
	}()
}

// RedisOpenEditValue opens the string value editor.
func (a *State) RedisOpenEditValue() {
	d := a.Redis.Detail
	if d.Type != "string" || d.Key == "" || d.Loading {
		return
	}
	a.Redis.EditKey = d.Key
	a.Redis.EditValue = ""
	if d.Result != nil && len(d.Result.Rows) > 0 && len(d.Result.Rows[0]) > 0 {
		a.Redis.EditValue = fmt.Sprint(d.Result.Rows[0][0])
	}
	a.Redis.EditOpen = true
}

// RedisSaveValue applies the string editor.
func (a *State) RedisSaveValue() {
	key, value := a.Redis.EditKey, a.Redis.EditValue
	hostID := a.Redis.HostID
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.DB.ExecuteSQL(hostID, redisSetValueCommand(key, value))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = i18n.Tf("redis.error.saveFailed", err.Error())
				return
			}
			a.Redis.EditOpen = false
			a.Redis.Notice = i18n.Tf("redis.notice.saved", key)
			if a.Redis.SelectedKey == key {
				a.RedisSelectKey(key)
			}
		})
	}()
}

// RedisOpenTTL opens the TTL editor.
func (a *State) RedisOpenTTL() {
	d := a.Redis.Detail
	if d.Key == "" || d.Loading || d.Type == "" {
		a.Redis.Err = i18n.T("redis.error.keyNotReady")
		return
	}
	a.Redis.TTLKey = d.Key
	a.Redis.TTLValue = ""
	if d.TTLKnown {
		a.Redis.TTLValue = strconv.FormatInt(d.TTL, 10)
	}
	a.Redis.TTLOpen = true
}

// RedisSaveTTL applies the TTL editor.
func (a *State) RedisSaveTTL() {
	key := a.Redis.TTLKey
	hostID := a.Redis.HostID
	seconds, err := strconv.ParseInt(strings.TrimSpace(a.Redis.TTLValue), 10, 64)
	if err != nil {
		a.Redis.Err = i18n.T("redis.error.ttlNotInteger")
		return
	}
	if seconds == 0 {
		a.Redis.Err = i18n.T("redis.error.ttlZero")
		return
	}
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.DB.ExecuteSQL(hostID, redisTTLCommand(key, seconds))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = i18n.Tf("redis.error.ttlFailed", err.Error())
				return
			}
			a.Redis.TTLOpen = false
			a.Redis.Notice = i18n.Tf("redis.notice.ttlUpdated", key)
			if a.Redis.SelectedKey == key {
				a.RedisSelectKey(key)
			}
		})
	}()
}

// RedisLoadInfo refreshes the server INFO tab.
func (a *State) RedisLoadInfo() {
	hostID := a.Redis.HostID
	if hostID == "" {
		return
	}
	a.Redis.infoSeq++
	seq := a.Redis.infoSeq
	a.Redis.InfoLoading = true
	a.Redis.InfoErr = ""
	if a.Backend == nil {
		a.Redis.InfoLoading = false
		a.Redis.InfoErr = i18n.T("redis.error.backendUninitialized")
		return
	}
	b := a.Backend
	go func() {
		res, err := b.DB.ExecuteSQL(hostID, "INFO")
		a.Post(func() {
			if hostID != a.Redis.HostID || seq != a.Redis.infoSeq {
				return
			}
			a.Redis.InfoLoading = false
			if err != nil {
				a.Redis.InfoErr = i18n.Tf("redis.error.infoFailed", err.Error())
				return
			}
			a.Redis.Info = redisParseInfo(redisFirstCell(res))
		})
	}()
}

// RedisRunCommand executes the command terminal input.
func (a *State) RedisRunCommand() {
	cmd := strings.TrimSpace(a.Redis.Command)
	hostID := a.Redis.HostID
	if cmd == "" || hostID == "" || a.Redis.CommandRunning {
		return
	}
	a.Redis.Command = ""
	a.Redis.cmdSeq++
	seq := a.Redis.cmdSeq
	a.Redis.CommandRunning = true
	a.Redis.CommandErr = ""
	a.Redis.CommandResult = nil
	if a.Backend == nil {
		a.Redis.CommandRunning = false
		a.Redis.CommandErr = i18n.T("redis.error.backendUninitialized")
		return
	}
	b := a.Backend
	go func() {
		res, err := b.DB.ExecuteSQL(hostID, cmd)
		a.Post(func() {
			if hostID != a.Redis.HostID || seq != a.Redis.cmdSeq {
				return
			}
			a.Redis.CommandRunning = false
			if err != nil {
				a.Redis.CommandErr = err.Error()
			} else {
				a.Redis.CommandResult = res
			}
			a.Redis.CommandHistory = append([]string{cmd},
				firstNStrings(a.Redis.CommandHistory, 49)...)
		})
	}()
}

// RedisPickHistory puts a past command back in the input.
func (a *State) RedisPickHistory(cmd string) {
	a.Redis.Command = cmd
	a.Redis.HistoryOpen = false
}

// redisSetValueCommand preserves an existing TTL; plain SET would remove it.
func redisSetValueCommand(key, value string) string {
	return "SET " + RedisQuote(key) + " " + RedisQuote(value) + " KEEPTTL"
}

// redisScanCollection turns HSCAN/SSCAN replies into a key/value or
// index/value table and reports whether the returned cursor has more.
func redisScanCollection(typ string, res *dbexec.Result) (*dbexec.Result, bool, bool) {
	if res == nil || len(res.Rows) < 2 || len(res.Rows[0]) < 2 || len(res.Rows[1]) < 2 {
		return res, false, false
	}
	cursor := fmt.Sprint(res.Rows[0][1])
	var flat []any
	switch v := res.Rows[1][1].(type) {
	case []any:
		flat = v
	case []string:
		flat = make([]any, len(v))
		for i := range v {
			flat[i] = v[i]
		}
	default:
		return res, false, false
	}
	out := &dbexec.Result{}
	if typ == "hash" {
		out.Columns = []string{"key", "value"}
		for i := 0; i+1 < len(flat); i += 2 {
			out.Rows = append(out.Rows, []any{flat[i], flat[i+1]})
		}
	} else {
		out.Columns = []string{"index", "value"}
		for i, item := range flat {
			out.Rows = append(out.Rows, []any{i + 1, item})
		}
	}
	return out, cursor != "0", true
}

// RedisQuote quotes a key, pattern or value for parseCommand.
func RedisQuote(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

// redisTTLCommand expires a key, or removes its expiry with PERSIST when
// seconds is negative (-1 in the UI means "permanent"; EXPIRE -1 would
// delete the key instead).
func redisTTLCommand(key string, seconds int64) string {
	if seconds < 0 {
		return "PERSIST " + RedisQuote(key)
	}
	return fmt.Sprintf("EXPIRE %s %d", RedisQuote(key), seconds)
}

// RedisTTLText renders a TTL in seconds.
func RedisTTLText(ttl int64) string {
	switch ttl {
	case -1:
		return i18n.T("redis.ttl.permanent")
	case -2:
		return i18n.T("redis.ttl.missing")
	default:
		return strconv.FormatInt(ttl, 10) + "s"
	}
}

func redisFirstCell(res *dbexec.Result) string {
	if res == nil || len(res.Rows) == 0 || len(res.Rows[0]) == 0 {
		return ""
	}
	return fmt.Sprint(res.Rows[0][0])
}

// redisScanResult parses SCAN's [[index,cursor],[index,[keys]]] shape.
// redisScanExecutor is what scanRedisKeysPage needs from the backend.
type redisScanExecutor interface {
	ExecuteSQL(hostID, sqlText string) (*dbexec.Result, error)
}

// scanRedisKeysPage collects up to limit keys from start, following SCAN
// cursors. Keys beyond limit come back as pending, with next as the cursor
// that follows the whole batch, so the next page starts exactly there.
func scanRedisKeysPage(exec redisScanExecutor, hostID string, start redisPageStart, pattern string, limit int) (keys []string, next string, pending []string, err error) {
	if limit <= 0 {
		return nil, "0", nil, nil
	}
	seen := make(map[string]bool, limit)
	take := func(items []string) {
		for _, key := range items {
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			if len(keys) < limit {
				keys = append(keys, key)
				continue
			}
			pending = append(pending, key)
		}
	}
	take(start.pending)
	if len(keys) >= limit || start.done {
		return keys, start.cursor, pending, nil
	}
	cursor := start.cursor
	if cursor == "" {
		cursor = "0"
	}
	// The first step may start from cursor "0"; after that, "0" means the
	// scan has ended.
	for step := 0; len(keys) < limit; step++ {
		if step >= redisScanStepsLimit || (step > 0 && cursor == "0") {
			break
		}
		res, serr := exec.ExecuteSQL(hostID,
			redisScanCommand(cursor, pattern, limit-len(keys)))
		if serr != nil {
			return keys, cursor, pending, serr
		}
		var batch []string
		cursor, batch = redisScanResult(res)
		take(batch)
	}
	return keys, cursor, pending, nil
}

func redisScanCommand(cursor, pattern string, count int) string {
	if count < 1 {
		count = 1
	}
	cmd := "SCAN " + cursor
	if pattern != "" {
		cmd += " MATCH " + RedisQuote(pattern)
	}
	cmd += " COUNT " + strconv.Itoa(count)
	return cmd
}

func redisScanResult(res *dbexec.Result) (cursor string, keys []string) {
	cursor = "0"
	if res == nil {
		return cursor, nil
	}
	if len(res.Rows) > 0 && len(res.Rows[0]) > 1 {
		cursor = fmt.Sprint(res.Rows[0][1])
	}
	if len(res.Rows) > 1 && len(res.Rows[1]) > 1 {
		switch v := res.Rows[1][1].(type) {
		case []string:
			keys = v
		case []any:
			for _, item := range v {
				keys = append(keys, fmt.Sprint(item))
			}
		default:
			if v != nil {
				keys = []string{fmt.Sprint(v)}
			}
		}
	}
	return cursor, keys
}

func redisParseInfo(text string) []RedisMetric {
	values := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, ':'); i > 0 {
			values[line[:i]] = line[i+1:]
		}
	}
	out := make([]RedisMetric, 0, len(redisInfoKeys))
	for _, k := range redisInfoKeys {
		if v, ok := values[k]; ok {
			out = append(out, RedisMetric{Name: k, Value: v})
		}
	}
	return out
}

func firstNStrings(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
