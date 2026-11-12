package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

// DB tab kinds.
const (
	DBTabQuery = "query"
	DBTabTable = "table"
)

// dbPageSize is how many rows a table tab loads at a time.
const dbPageSize = 100

// DBTab is one query editor or table browser tab.
type DBTab struct {
	Key       string
	Kind      string
	Title     string
	TableName string

	Editor *editor.State
	Grid   ui.ListState

	Result     *dbexec.Result
	Err        string
	Running    bool
	DurationMS int64

	Page       int
	PageSize   int
	Total      int64
	HasTotal   bool
	OrderBy    string
	ResultPage int

	seq uint64
}

// DBState is the database workbench state.
type DBState struct {
	HostID   string
	HostName string
	Protocol string

	Schema        *dbexec.Schema
	SchemaLoading bool
	SchemaErr     string

	Tabs   []*DBTab
	Active int
	// Split is the height of the SQL editor above the results.
	Split float32
	// Open holds the tree's group/table expansion state.
	Open map[string]*bool
	// Filter narrows the object tree to tables/views whose name contains it.
	Filter string

	Err    string
	Notice string

	schemaSeq uint64
	tabSeq    int
}

// DBDatabaseLabel returns the selected asset's database name or file path.
func (a *App) DBDatabaseLabel() string {
	if a.Selected == nil {
		return ""
	}
	if v, ok := a.Selected.Params["database"].(string); ok && v != "" {
		return v
	}
	if v, ok := a.Selected.Params["filePath"].(string); ok && v != "" {
		return v
	}
	return ""
}

// IsDBProtocol reports whether a protocol is a database connector.
func (a *App) IsDBProtocol(protocol string) bool {
	c := a.hostConnector(protocol)
	return c != nil && c.IsDB()
}

// attachDB initializes the workbench for a database asset.
func (a *App) attachDB(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" || !a.IsDBProtocol(n.Protocol) {
		return
	}
	if a.DB.HostID == n.ID && a.DB.Protocol == n.Protocol {
		// Keep tabs and editor buffers; retry a load that failed without
		// rebuilding the whole workbench.
		if !a.DB.SchemaLoading && (a.DB.Schema == nil || a.DB.SchemaErr != "") {
			a.loadDBSchema(n.ID)
		}
		return
	}
	a.DB = DBState{
		HostID: n.ID, HostName: n.Name, Protocol: n.Protocol,
		Split: 190, Open: map[string]*bool{}, SchemaLoading: true,
	}
	a.NewDBQueryTab()
	a.loadDBSchema(n.ID)
}

func (a *App) loadDBSchema(hostID string) {
	a.DB.schemaSeq++
	seq := a.DB.schemaSeq
	if a.Backend == nil {
		a.DB.SchemaLoading = false
		a.DB.SchemaErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		schema, err := b.ListSchema(hostID)
		a.Post(func() {
			if seq != a.DB.schemaSeq || hostID != a.DB.HostID {
				return
			}
			a.DB.SchemaLoading = false
			if err != nil {
				a.DB.SchemaErr = "加载结构失败：" + err.Error()
				return
			}
			a.DB.SchemaErr = ""
			a.DB.Schema = schema
			a.initDBOpen(schema)
		})
	}()
}

// RefreshDBSchema reloads tables and columns.
func (a *App) RefreshDBSchema() {
	if a.DB.HostID == "" {
		return
	}
	a.DB.Err = ""
	a.DB.Notice = ""
	a.DB.SchemaLoading = true
	a.DB.SchemaErr = ""
	a.loadDBSchema(a.DB.HostID)
}

func (a *App) initDBOpen(schema *dbexec.Schema) {
	if a.DB.Open == nil {
		a.DB.Open = map[string]*bool{}
	}
	ensure := func(key string, open bool) {
		if _, ok := a.DB.Open[key]; !ok {
			v := open
			a.DB.Open[key] = &v
		}
	}
	ensure("table", true)
	ensure("view", true)
	if schema == nil {
		return
	}
	for _, t := range schema.Tables {
		ensure(DBTableKey(t), false)
	}
}

// DBTableKey is the tree key of a table or view.
func DBTableKey(t dbexec.Table) string {
	group := "table"
	if t.Type == "view" {
		group = "view"
	}
	return group + ":" + t.Name
}

// DBTreeOpen returns the expansion state pointer for a key, creating it
// closed when needed (views build only known groups/tables).
func (a *App) DBTreeOpen(key string) *bool {
	if a.DB.Open == nil {
		a.DB.Open = map[string]*bool{}
	}
	if p, ok := a.DB.Open[key]; ok {
		return p
	}
	v := false
	a.DB.Open[key] = &v
	return &v
}

// SetDBTreeAll expands or collapses every group and table.
func (a *App) SetDBTreeAll(open bool) {
	if a.DB.Open == nil {
		a.DB.Open = map[string]*bool{}
	}
	for k, p := range a.DB.Open {
		*p = open
		_ = k
	}
}

func newDBQueryTab(key, title string) *DBTab {
	buf := editor.NewBuffer("")
	buf.SetLanguage(editor.LanguageNamed("SQL"))
	return &DBTab{
		Key: key, Kind: DBTabQuery, Title: title,
		Editor: &editor.State{
			Buffer:      buf,
			Label:       "SQL 查询",
			Placeholder: "输入 SQL，点击「执行」运行",
		},
	}
}

// NewDBQueryTab adds and activates a query tab.
func (a *App) NewDBQueryTab() {
	a.DB.tabSeq++
	title := "查询"
	if n := a.DB.tabSeq; n > 1 {
		title = fmt.Sprintf("查询 %d", n)
	}
	tab := newDBQueryTab(fmt.Sprintf("db-q-%d", a.DB.tabSeq), title)
	a.DB.Tabs = append(a.DB.Tabs, tab)
	a.DB.Active = len(a.DB.Tabs) - 1
}

// ActivateDBTab selects a tab.
func (a *App) ActivateDBTab(i int) {
	if i >= 0 && i < len(a.DB.Tabs) {
		a.DB.Active = i
		a.DB.Err = ""
		a.DB.Notice = ""
	}
}

// CloseDBTab closes a tab and adjusts the active one.
func (a *App) CloseDBTab(key string) {
	for i, tab := range a.DB.Tabs {
		if tab.Key != key {
			continue
		}
		a.DB.Tabs = append(a.DB.Tabs[:i], a.DB.Tabs[i+1:]...)
		if i < a.DB.Active {
			a.DB.Active--
		}
		if a.DB.Active >= len(a.DB.Tabs) {
			a.DB.Active = len(a.DB.Tabs) - 1
		}
		if a.DB.Active < 0 {
			a.DB.Active = 0
		}
		return
	}
}

// ActiveDBTab returns the visible tab, or nil.
func (a *App) ActiveDBTab() *DBTab {
	if a.DB.Active < 0 || a.DB.Active >= len(a.DB.Tabs) {
		return nil
	}
	return a.DB.Tabs[a.DB.Active]
}

func (a *App) dbTabByKey(key string) *DBTab {
	for _, t := range a.DB.Tabs {
		if t.Key == key {
			return t
		}
	}
	return nil
}

// RunActiveDBTab runs the active query tab.
func (a *App) RunActiveDBTab() {
	tab := a.ActiveDBTab()
	if tab == nil || tab.Kind != DBTabQuery {
		return
	}
	a.ExecuteDBQuery(tab.Key)
}

// ExecuteDBQuery runs the SQL in a query tab.
func (a *App) ExecuteDBQuery(key string) {
	tab := a.dbTabByKey(key)
	if tab == nil || tab.Running || tab.Kind != DBTabQuery {
		return
	}
	sqlText := strings.TrimSpace(tab.Editor.Buffer.String())
	if sqlText == "" {
		return
	}
	a.runDBCommand(tab, sqlText)
}

func (a *App) runDBCommand(tab *DBTab, sqlText string) {
	hostID := a.DB.HostID
	a.DB.Err = ""
	a.DB.Notice = ""
	tab.Running = true
	tab.Err = ""
	tab.Result = nil
	tab.seq++
	seq := tab.seq
	if a.Backend == nil {
		tab.Running = false
		tab.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		start := time.Now()
		res, err := b.ExecuteSQL(hostID, sqlText)
		elapsed := time.Since(start).Milliseconds()
		a.Post(func() {
			if hostID != a.DB.HostID || a.dbTabByKey(tab.Key) != tab || seq != tab.seq {
				return
			}
			tab.Running = false
			tab.DurationMS = elapsed
			if err != nil {
				tab.Err = err.Error()
				return
			}
			tab.Result = res
		})
	}()
}

// OpenDBTable opens (or activates) a table browsing tab.
func (a *App) OpenDBTable(name string) {
	a.DB.Err = ""
	a.DB.Notice = ""
	for i, tab := range a.DB.Tabs {
		if tab.Kind == DBTabTable && tab.TableName == name {
			a.DB.Active = i
			return
		}
	}
	a.DB.tabSeq++
	tab := &DBTab{
		Key: fmt.Sprintf("db-t-%d", a.DB.tabSeq), Kind: DBTabTable,
		Title: name, TableName: name, PageSize: dbPageSize,
		OrderBy: a.dbOrderColumn(name),
	}
	a.DB.Tabs = append(a.DB.Tabs, tab)
	a.DB.Active = len(a.DB.Tabs) - 1
	a.loadDBTablePage(tab, 0, true)
}

// DBTablePage moves the active table tab by delta pages.
func (a *App) DBTablePage(delta int) {
	tab := a.ActiveDBTab()
	if tab == nil || tab.Kind != DBTabTable || tab.Running {
		return
	}
	page := tab.Page + delta
	if page < 0 {
		return
	}
	if delta > 0 && tab.HasTotal && int64(page*tab.PageSize) >= tab.Total {
		return
	}
	a.loadDBTablePage(tab, page, false)
}

// RefreshActiveDBTable reloads the active table tab's page.
func (a *App) RefreshActiveDBTable() {
	tab := a.ActiveDBTab()
	if tab != nil && tab.Kind == DBTabTable && !tab.Running {
		a.loadDBTablePage(tab, tab.Page, true)
	}
}

func (a *App) loadDBTablePage(tab *DBTab, page int, recount bool) {
	hostID := a.DB.HostID
	sqlText := DBTablePageSQL(a.DB.Protocol, tab.TableName, tab.OrderBy,
		tab.PageSize, page*tab.PageSize)
	countSQL := "SELECT COUNT(*) FROM " + DBQuoteIdent(a.DB.Protocol, tab.TableName)

	a.DB.Err = ""
	a.DB.Notice = ""
	tab.Running = true
	tab.Err = ""
	tab.Page = page
	tab.seq++
	seq := tab.seq
	if a.Backend == nil {
		tab.Running = false
		tab.Err = "后端未初始化"
		return
	}
	needCount := recount || !tab.HasTotal
	b := a.Backend
	go func() {
		start := time.Now()
		res, err := b.ExecuteSQL(hostID, sqlText)
		elapsed := time.Since(start).Milliseconds()
		a.Post(func() {
			if hostID != a.DB.HostID || a.dbTabByKey(tab.Key) != tab || seq != tab.seq {
				return
			}
			tab.Running = false
			tab.DurationMS = elapsed
			if err != nil {
				tab.Err = err.Error()
				return
			}
			tab.Result = res
			tab.ResultPage = page
		})
		if err != nil || !needCount {
			return
		}
		// Count separately so the page shows immediately instead of
		// waiting on a potentially slow full-table COUNT.
		cnt, cerr := b.ExecuteSQL(hostID, countSQL)
		if cerr != nil || cnt == nil || len(cnt.Rows) == 0 || len(cnt.Rows[0]) == 0 {
			return
		}
		total := dbCountValue(cnt.Rows[0][0])
		a.Post(func() {
			if hostID != a.DB.HostID || a.dbTabByKey(tab.Key) != tab || seq != tab.seq {
				return
			}
			tab.HasTotal = true
			tab.Total = total
		})
	}()
}

// DBTablePageSQL builds a deterministic, protocol-appropriate page query.
func DBTablePageSQL(protocol, table, orderBy string, limit, offset int) string {
	quoted := DBQuoteIdent(protocol, table)
	order := "1"
	if orderBy != "" && orderBy != "1" {
		order = DBQuoteIdent(protocol, orderBy)
	}
	if strings.EqualFold(protocol, "sqlserver") {
		return fmt.Sprintf("SELECT * FROM %s ORDER BY %s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
			quoted, order, offset, limit)
	}
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s LIMIT %d OFFSET %d",
		quoted, order, limit, offset)
}

// DBPreviewQuery is the SQL the tree's "生成查询" action inserts: the
// first page, ordered and capped, valid for every supported protocol.
func (a *App) DBPreviewQuery(table string) string {
	return DBTablePageSQL(a.DB.Protocol, table, a.dbOrderColumn(table), dbPageSize, 0)
}

// dbOrderColumn picks a stable ordering column: a primary key, else id,
// else the first column. Without ORDER BY, LIMIT/OFFSET pages are not
// deterministic.
func (a *App) dbOrderColumn(table string) string {
	if a.DB.Schema == nil {
		return "1"
	}
	for _, tb := range a.DB.Schema.Tables {
		if tb.Name != table {
			continue
		}
		for _, c := range tb.Columns {
			if strings.EqualFold(c.Key, "PRI") {
				return c.Name
			}
		}
		for _, c := range tb.Columns {
			if strings.EqualFold(c.Name, "id") {
				return c.Name
			}
		}
		if len(tb.Columns) > 0 {
			return tb.Columns[0].Name
		}
	}
	return "1"
}

// DBQuoteIdent quotes a table or view identifier for the asset's protocol.
// SQL Server schema-qualified names are quoted part by part.
func DBQuoteIdent(protocol, name string) string {
	switch strings.ToLower(protocol) {
	case "mysql":
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "sqlserver":
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = "[" + strings.ReplaceAll(part, "]", "]]") + "]"
		}
		return strings.Join(parts, ".")
	default:
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	}
}

func dbCountValue(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		x, _ := strconv.ParseInt(string(n), 10, 64)
		return x
	default:
		x, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		return x
	}
}

// DBCellText renders a result cell for the grid and CSV. NULL is empty in
// CSV; the grid shows "NULL" when it gets nil.
func DBCellText(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprint(x)
	}
}

// dbResultCSV serializes a result as CSV (header + rows), using standard
// quoting for commas, quotes and newlines.
func dbResultCSV(columns []string, rows [][]any) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(columns); err != nil {
		return "", err
	}
	for _, row := range rows {
		rec := make([]string, len(columns))
		for i := range columns {
			if i < len(row) {
				rec[i] = DBCellText(row[i])
			}
		}
		if err := w.Write(rec); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ExportActiveDBResult writes the active result to a CSV file chosen by
// the user. Table tabs export the current page only; the notice names the
// page and total so a partial export cannot be mistaken for the whole
// table. It blocks in a dialog and file write, so it runs off the UI
// thread.
func (a *App) ExportActiveDBResult() {
	tab := a.ActiveDBTab()
	if tab == nil || tab.Result == nil || len(tab.Result.Columns) == 0 {
		a.DB.Err = "当前没有可导出的查询结果"
		return
	}
	if tab.Running || (tab.Kind == DBTabTable && tab.ResultPage != tab.Page) {
		a.DB.Err = "数据正在加载，请稍后再导出"
		return
	}
	columns := tab.Result.Columns
	rows := tab.Result.Rows
	kind := tab.Kind
	page, total, hasTotal := tab.Page, tab.Total, tab.HasTotal
	truncated := tab.Result.Truncated

	name := sanitizeFileName(tab.Title)
	if kind == DBTabTable {
		name = fmt.Sprintf("%s-page-%d", name, page+1)
	}
	a.DB.Err = ""
	a.DB.Notice = ""
	go func() {
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Title:       "导出 CSV",
			DefaultPath: name + ".csv",
			ButtonLabel: "导出",
			Filters:     []mygo.FileFilter{{Name: "CSV", Extensions: []string{"csv"}}},
		})
		if err != nil {
			a.Post(func() { a.DB.Err = "导出失败：" + err.Error() })
			return
		}
		if strings.TrimSpace(path) == "" {
			return
		}
		if !strings.HasSuffix(strings.ToLower(path), ".csv") {
			path += ".csv"
		}
		content, err := dbResultCSV(columns, rows)
		if err != nil {
			a.Post(func() { a.DB.Err = "生成 CSV 失败：" + err.Error() })
			return
		}
		if err := os.WriteFile(path, append([]byte("\uFEFF"), []byte(content)...), 0o644); err != nil {
			a.Post(func() { a.DB.Err = "写入文件失败：" + err.Error() })
			return
		}
		a.Post(func() {
			a.DB.Err = ""
			note := fmt.Sprintf("已导出 %d 行", len(rows))
			if kind == DBTabTable {
				note = fmt.Sprintf("已导出第 %d 页 %d 行", page+1, len(rows))
				if hasTotal {
					note += fmt.Sprintf("（共 %d 行，仅当前页）", total)
				}
			}
			if truncated {
				note += "；结果已截断"
			}
			a.DB.Notice = note + "：" + path
		})
	}()
}

func sanitizeFileName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "db-result"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", `"`, "-", "<", "-", ">", "-", "|", "-")
	return replacer.Replace(s)
}
