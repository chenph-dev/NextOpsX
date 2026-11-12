package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

// redisPreviewLimit caps collection previews (list/hash/set/zset).
const redisPreviewLimit = 200

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
func (a *App) IsRedisProtocol(protocol string) bool {
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
func (a *App) AttachRedis(n *hoststore.TreeNode) {
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

// RedisReloadKeys refreshes the current scan without dropping a selection.
func (a *App) RedisReloadKeys() { a.refreshRedisKeys(true, false) }

// EnsureRedis refreshes the Redis panel when it is re-entered.
func (a *App) EnsureRedis() {
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

// RedisRefreshKeys scans the next page, or restarts with cursor 0.
func (a *App) RedisRefreshKeys(reset bool) { a.refreshRedisKeys(reset, reset) }

func (a *App) refreshRedisKeys(reset, clearSelection bool) {
	hostID := a.Redis.HostID
	if hostID == "" {
		return
	}
	cursor := "0"
	pattern := a.Redis.ScanPattern
	if !reset {
		cursor = a.Redis.Cursor
		if cursor == "0" {
			return
		}
	} else {
		// A SCAN cursor is only valid with the MATCH argument it was
		// created with. Keep the pattern used for this scan separately
		// from the editable search box.
		pattern = strings.TrimSpace(a.Redis.Pattern)
		a.Redis.ScanPattern = pattern
	}
	a.Redis.keySeq++
	seq := a.Redis.keySeq
	a.Redis.KeysLoading = true
	a.Redis.KeysErr = ""
	a.Redis.Err = ""
	a.Redis.Notice = ""
	if reset {
		a.Redis.Cursor = "0"
		a.Redis.Keys = nil
		if clearSelection {
			a.Redis.SelectedKey = ""
			a.Redis.Detail = RedisDetail{}
		}
	}
	cmd := "SCAN " + cursor
	if pattern != "" {
		cmd += " MATCH " + RedisQuote(pattern)
	}
	cmd += " COUNT 50"

	if a.Backend == nil {
		a.Redis.KeysLoading = false
		a.Redis.KeysErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, cmd)
		a.Post(func() {
			if hostID != a.Redis.HostID || seq != a.Redis.keySeq {
				return
			}
			a.Redis.KeysLoading = false
			if err != nil {
				a.Redis.KeysErr = "扫描键失败：" + err.Error()
				return
			}
			next, keys := redisScanResult(res)
			a.Redis.Keys = keys
			a.Redis.Cursor = next
		})
	}()
}

// RedisSearch restarts the scan with the current pattern.
func (a *App) RedisSearch() { a.RedisRefreshKeys(true) }

// RedisNextPage scans the next cursor page.
func (a *App) RedisNextPage() { a.RedisRefreshKeys(false) }

// RedisSelectKey loads a key's type, TTL and value.
func (a *App) RedisSelectKey(key string) {
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
		a.Redis.Detail.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		typeRes, err := b.ExecuteSQL(hostID, "TYPE "+RedisQuote(key))
		if err != nil {
			a.postRedisDetail(hostID, seq, RedisDetail{Key: key, Err: err.Error()})
			return
		}
		typ := strings.ToLower(redisFirstCell(typeRes))
		if typ == "" {
			typ = "string"
		}
		detail := RedisDetail{Key: key, Type: typ}
		if ttlRes, terr := b.ExecuteSQL(hostID, "TTL "+RedisQuote(key)); terr == nil {
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
			if res, verr := b.ExecuteSQL(hostID, valueCmd); verr != nil {
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

func (a *App) postRedisDetail(hostID string, seq uint64, d RedisDetail) {
	a.Post(func() {
		if hostID != a.Redis.HostID || seq != a.Redis.detailSeq {
			return
		}
		a.Redis.Detail = d
	})
}

// RedisOpenDelete confirms deleting the selected key.
func (a *App) RedisOpenDelete() {
	if a.Redis.SelectedKey == "" {
		return
	}
	a.Redis.ConfirmDeleteKey = a.Redis.SelectedKey
}

// RedisConfirmDelete deletes the confirmed key.
func (a *App) RedisConfirmDelete() {
	key := a.Redis.ConfirmDeleteKey
	hostID := a.Redis.HostID
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.ExecuteSQL(hostID, "DEL "+RedisQuote(key))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = "删除失败：" + err.Error()
				return
			}
			a.Redis.ConfirmDeleteKey = ""
			if a.Redis.SelectedKey == key {
				a.Redis.SelectedKey = ""
				a.Redis.Detail = RedisDetail{}
			}
			a.refreshRedisKeys(true, false)
			a.Redis.Notice = "已删除 " + key
		})
	}()
}

// RedisOpenEditValue opens the string value editor.
func (a *App) RedisOpenEditValue() {
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
func (a *App) RedisSaveValue() {
	key, value := a.Redis.EditKey, a.Redis.EditValue
	hostID := a.Redis.HostID
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.ExecuteSQL(hostID, redisSetValueCommand(key, value))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = "保存失败：" + err.Error()
				return
			}
			a.Redis.EditOpen = false
			a.Redis.Notice = "已保存 " + key
			if a.Redis.SelectedKey == key {
				a.RedisSelectKey(key)
			}
		})
	}()
}

// RedisOpenTTL opens the TTL editor.
func (a *App) RedisOpenTTL() {
	d := a.Redis.Detail
	if d.Key == "" || d.Loading || d.Type == "" {
		a.Redis.Err = "键信息尚未加载完成"
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
func (a *App) RedisSaveTTL() {
	key := a.Redis.TTLKey
	hostID := a.Redis.HostID
	seconds, err := strconv.ParseInt(strings.TrimSpace(a.Redis.TTLValue), 10, 64)
	if err != nil {
		a.Redis.Err = "TTL 需要是整数秒"
		return
	}
	if seconds == 0 {
		a.Redis.Err = "TTL 0 会立即删除键；请使用删除按钮，或输入正数/-1"
		return
	}
	if key == "" || hostID == "" || a.Backend == nil {
		return
	}
	a.Redis.Err = ""
	b := a.Backend
	go func() {
		_, err := b.ExecuteSQL(hostID, redisTTLCommand(key, seconds))
		a.Post(func() {
			if hostID != a.Redis.HostID {
				return
			}
			if err != nil {
				a.Redis.Err = "设置 TTL 失败：" + err.Error()
				return
			}
			a.Redis.TTLOpen = false
			a.Redis.Notice = "已更新 TTL：" + key
			if a.Redis.SelectedKey == key {
				a.RedisSelectKey(key)
			}
		})
	}()
}

// RedisLoadInfo refreshes the server INFO tab.
func (a *App) RedisLoadInfo() {
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
		a.Redis.InfoErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, "INFO")
		a.Post(func() {
			if hostID != a.Redis.HostID || seq != a.Redis.infoSeq {
				return
			}
			a.Redis.InfoLoading = false
			if err != nil {
				a.Redis.InfoErr = "读取 INFO 失败：" + err.Error()
				return
			}
			a.Redis.Info = redisParseInfo(redisFirstCell(res))
		})
	}()
}

// RedisRunCommand executes the command terminal input.
func (a *App) RedisRunCommand() {
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
		a.Redis.CommandErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, cmd)
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
func (a *App) RedisPickHistory(cmd string) {
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
		return "永久"
	case -2:
		return "不存在"
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
