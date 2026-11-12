package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

// ESDefaultDSL is the initial query body in the ES query tab.
const ESDefaultDSL = `{
  "query": {
    "match_all": {}
  }
}`

// esInfoKeys lists cluster health fields shown in the server tab.
var esInfoKeys = []string{
	"cluster_name", "status", "timed_out", "number_of_nodes",
	"number_of_data_nodes", "active_primary_shards", "active_shards",
	"relocating_shards", "initializing_shards", "unassigned_shards",
	"number_of_pending_tasks",
}

// ESMetric is one cluster health metric.
type ESMetric struct {
	Name  string
	Value string
}

// ESState is the Elasticsearch workbench state.
type ESState struct {
	HostID string
	Label  string

	Indices        []string
	IndicesLoading bool
	IndicesErr     string
	SelectedIndex  string
	DefaultIndex   string
	IndexRow       int
	IndexGrid      ui.ListState

	MainTab int

	QueryDSL     string
	QueryResult  *dbexec.Result
	QueryErr     string
	QueryRunning bool
	QueryGrid    ui.ListState

	Info        []ESMetric
	InfoLoading bool
	InfoErr     string

	Command        string
	CommandResult  *dbexec.Result
	CommandErr     string
	CommandRunning bool
	CommandHistory []string
	HistoryOpen    bool
	CommandGrid    ui.ListState

	Notice string
	Err    string

	indexSeq uint64
	querySeq uint64
	infoSeq  uint64
	cmdSeq   uint64
}

// IsESProtocol reports whether a protocol is Elasticsearch.
func (a *App) IsESProtocol(protocol string) bool {
	return strings.EqualFold(protocol, "elasticsearch")
}

func esLabel(n *hoststore.TreeNode) string {
	return fmt.Sprintf("%s · %s:%d", n.Name, n.Addr, n.Port)
}

// AttachES initializes the Elasticsearch workbench.
func (a *App) AttachES(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" || !a.IsESProtocol(n.Protocol) {
		return
	}
	def := ""
	if v, ok := n.Params["index"]; ok {
		def = strings.TrimSpace(fmt.Sprint(v))
	}
	if a.ES.HostID == n.ID {
		a.ES.Label = esLabel(n)
		if def != a.ES.DefaultIndex {
			old := a.ES.DefaultIndex
			a.ES.DefaultIndex = def
			if a.ES.SelectedIndex == "" || a.ES.SelectedIndex == old {
				a.ESSelectedIndexChanged(def)
			}
		}
		return
	}
	a.ES = ESState{
		HostID: n.ID, Label: esLabel(n),
		QueryDSL: ESDefaultDSL, SelectedIndex: def, DefaultIndex: def,
		IndexRow: -1,
	}
	a.ESRefreshIndices()
}

// EnsureES refreshes the panel when it is re-entered.
func (a *App) EnsureES() {
	n := a.Selected
	if n == nil || !a.IsESProtocol(n.Protocol) {
		return
	}
	if a.ES.HostID != n.ID {
		a.AttachES(n)
		return
	}
	if !a.ES.IndicesLoading {
		a.ESRefreshIndices()
	}
	if a.ES.MainTab == 1 {
		a.ESLoadInfo()
	}
}

// ESRefreshIndices reloads the index list from _cat/indices.
func (a *App) ESRefreshIndices() {
	hostID := a.ES.HostID
	if hostID == "" {
		return
	}
	a.ES.indexSeq++
	seq := a.ES.indexSeq
	a.ES.IndicesLoading = true
	a.ES.IndicesErr = ""
	a.ES.Err = ""
	a.ES.Notice = ""
	if a.Backend == nil {
		a.ES.IndicesLoading = false
		a.ES.IndicesErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, "_cat/indices?format=json")
		a.Post(func() {
			if hostID != a.ES.HostID || seq != a.ES.indexSeq {
				return
			}
			a.ES.IndicesLoading = false
			if err != nil {
				a.ES.IndicesErr = "读取索引失败：" + err.Error()
				return
			}
			indices := esIndexNames(res)
			sort.Strings(indices)
			a.ES.Indices = indices
			if a.ES.SelectedIndex != "" && !containsString(indices, a.ES.SelectedIndex) {
				a.ESSelectedIndexChanged("")
			}
			a.syncESIndexRow()
			a.ES.Notice = fmt.Sprintf("索引已刷新（%d）", len(indices))
		})
	}()
}

func esIndexNames(res *dbexec.Result) []string {
	if res == nil {
		return nil
	}
	col := -1
	for i, name := range res.Columns {
		if strings.EqualFold(name, "index") {
			col = i
			break
		}
	}
	if col < 0 {
		return nil
	}
	out := make([]string, 0, len(res.Rows))
	for _, row := range res.Rows {
		if col < len(row) && row[col] != nil {
			out = append(out, fmt.Sprint(row[col]))
		}
	}
	return out
}

// ESSelectIndex selects the index used by the DSL query tab.
func (a *App) ESSelectIndex(index string) {
	if index == a.ES.SelectedIndex {
		return
	}
	a.ESSelectedIndexChanged(index)
}

// ESSelectedIndexChanged changes the target index and drops any result or
// in-flight response from the previous index.
func (a *App) ESSelectedIndexChanged(index string) {
	a.ES.SelectedIndex = index
	a.ES.querySeq++
	a.ES.QueryRunning = false
	a.ES.QueryResult = nil
	a.ES.QueryErr = ""
	a.ES.Err = ""
	a.syncESIndexRow()
}

func (a *App) syncESIndexRow() {
	row := -1
	for i, idx := range a.ES.Indices {
		if idx == a.ES.SelectedIndex {
			row = i
			break
		}
	}
	a.ES.IndexRow = row
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// ESRunQuery runs the DSL against the selected index.
func (a *App) ESRunQuery() {
	index := strings.TrimSpace(a.ES.SelectedIndex)
	body := strings.TrimSpace(a.ES.QueryDSL)
	if index == "" {
		a.ES.Err = "请先选择索引"
		return
	}
	if body == "" {
		a.ES.Err = "请输入 Query DSL"
		return
	}
	hostID := a.ES.HostID
	a.ES.querySeq++
	seq := a.ES.querySeq
	a.ES.QueryRunning = true
	a.ES.QueryErr = ""
	a.ES.QueryResult = nil
	a.ES.Err = ""
	if a.Backend == nil {
		a.ES.QueryRunning = false
		a.ES.QueryErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, index+"/_search\n"+body)
		a.Post(func() {
			if hostID != a.ES.HostID || seq != a.ES.querySeq {
				return
			}
			a.ES.QueryRunning = false
			if err != nil {
				a.ES.QueryErr = err.Error()
				return
			}
			a.ES.QueryResult = res
			if res != nil {
				a.ES.Notice = fmt.Sprintf("查询完成（%d 行）", len(res.Rows))
			}
		})
	}()
}

// ESLoadInfo reloads cluster health.
func (a *App) ESLoadInfo() {
	hostID := a.ES.HostID
	if hostID == "" {
		return
	}
	a.ES.infoSeq++
	seq := a.ES.infoSeq
	a.ES.InfoLoading = true
	a.ES.InfoErr = ""
	if a.Backend == nil {
		a.ES.InfoLoading = false
		a.ES.InfoErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, "_cluster/health")
		a.Post(func() {
			if hostID != a.ES.HostID || seq != a.ES.infoSeq {
				return
			}
			a.ES.InfoLoading = false
			if err != nil {
				a.ES.InfoErr = "读取集群信息失败：" + err.Error()
				return
			}
			a.ES.Info = esHealthMetrics(res)
		})
	}()
}

func esHealthMetrics(res *dbexec.Result) []ESMetric {
	if res == nil || len(res.Rows) == 0 {
		return nil
	}
	values := map[string]string{}
	for i, name := range res.Columns {
		if i < len(res.Rows[0]) && res.Rows[0][i] != nil {
			values[name] = fmt.Sprint(res.Rows[0][i])
		}
	}
	out := make([]ESMetric, 0, len(esInfoKeys))
	for _, k := range esInfoKeys {
		if v, ok := values[k]; ok {
			out = append(out, ESMetric{Name: k, Value: v})
		}
	}
	return out
}

// ESRunCommand executes a raw REST command from the terminal.
func (a *App) ESRunCommand() {
	cmd := strings.TrimSpace(a.ES.Command)
	hostID := a.ES.HostID
	if cmd == "" || hostID == "" || a.ES.CommandRunning {
		return
	}
	a.ES.Command = ""
	a.ES.cmdSeq++
	seq := a.ES.cmdSeq
	a.ES.CommandRunning = true
	a.ES.CommandErr = ""
	a.ES.CommandResult = nil
	a.ES.Err = ""
	if a.Backend == nil {
		a.ES.CommandRunning = false
		a.ES.CommandErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		res, err := b.ExecuteSQL(hostID, cmd)
		a.Post(func() {
			if hostID != a.ES.HostID || seq != a.ES.cmdSeq {
				return
			}
			a.ES.CommandRunning = false
			if err != nil {
				a.ES.CommandErr = err.Error()
			} else {
				a.ES.CommandResult = res
				a.ES.Notice = "命令执行完成"
			}
			a.ES.CommandHistory = append([]string{cmd},
				firstNStrings(a.ES.CommandHistory, 49)...)
		})
	}()
}

// ESPickHistory puts a past command back in the terminal input.
func (a *App) ESPickHistory(cmd string) {
	a.ES.Command = cmd
	a.ES.HistoryOpen = false
}
