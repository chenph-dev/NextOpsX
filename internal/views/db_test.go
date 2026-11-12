package views

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func dbViewTestApp() *app.App {
	a := &app.App{RightTab: 0}
	a.Selected = &hoststore.TreeNode{
		ID: "db1", Name: "mysql-01", NodeType: "host", Protocol: "mysql",
		Addr: "10.0.0.9", Port: 3306, User: "root",
	}
	schema := &dbexec.Schema{Tables: []dbexec.Table{
		{Name: "users", Type: "table", Columns: []dbexec.Column{
			{Name: "id", DataType: "int"},
			{Name: "name", DataType: "text"},
		}},
		{Name: "v_users", Type: "view", Columns: []dbexec.Column{
			{Name: "id", DataType: "int"},
		}},
	}}
	a.DB = app.DBState{
		HostID: "db1", HostName: "mysql-01", Protocol: "mysql",
		Schema: schema, Open: map[string]*bool{}, Split: 190,
	}
	*a.DBTreeOpen("table") = true
	*a.DBTreeOpen("view") = true
	for _, tb := range schema.Tables {
		*a.DBTreeOpen(app.DBTableKey(tb)) = true
	}
	a.NewDBQueryTab()
	tab := a.DB.Tabs[0]
	tab.Result = &dbexec.Result{
		Columns: []string{"id", "name"},
		Rows:    [][]any{{int64(1), "alice"}, {int64(2), "bob"}},
	}
	tab.DurationMS = 12
	return a
}

func TestDBWorkbenchRenders(t *testing.T) {
	a := dbViewTestApp()
	tt := ui.NewTester(Shell(a), 1200, 800)
	for _, want := range []string{
		"数据库", "mysql-01", "对象", "表 (1)", "视图 (1)",
		"users", "v_users", "id : int", "执行", "+",
		"刷新结构", "导出 CSV", "alice", "bob", "2 行 · 12ms",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestDBWorkbenchNewQuery(t *testing.T) {
	a := dbViewTestApp()
	tt := ui.NewTester(Shell(a), 1200, 800)
	if err := tt.Click("+"); err != nil {
		t.Fatal(err)
	}
	if len(a.DB.Tabs) != 2 || a.DB.Active != 1 {
		t.Fatalf("tabs = %d active = %d", len(a.DB.Tabs), a.DB.Active)
	}
	tt.Frame()
	if !tt.HasText("查询 2") {
		t.Fatalf("new query tab missing: %q", tt.Texts())
	}
}

func TestDBTruncatedIndicator(t *testing.T) {
	a := dbViewTestApp()
	a.DB.Tabs[0].Result.Truncated = true
	tt := ui.NewTester(Shell(a), 1200, 800)
	if !tt.HasText("结果已截断") {
		t.Fatalf("truncated indicator missing: %q", tt.Texts())
	}
}

func TestDBTableTabRenders(t *testing.T) {
	a := dbViewTestApp()
	a.OpenDBTable("users")
	tab := a.ActiveDBTab()
	tab.Err = ""
	tab.HasTotal = true
	tab.Total = 3
	tab.Page = 0
	tab.PageSize = 100
	tab.DurationMS = 5
	tab.Result = &dbexec.Result{
		Columns: []string{"id", "name"},
		Rows:    [][]any{{int64(1), "alice"}},
	}
	tt := ui.NewTester(Shell(a), 1200, 800)
	for _, want := range []string{
		"users", "共 3 行", "第 1 / 1 页", "上一页", "下一页", "刷新", "alice",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if a.DB.Tabs[len(a.DB.Tabs)-1].Running {
		t.Fatal("table tab should not be running in this fixture")
	}
}

func TestDBEmptySchemaHint(t *testing.T) {
	a := &app.App{RightTab: 0}
	a.Selected = &hoststore.TreeNode{
		ID: "r1", Name: "redis-01", NodeType: "host", Protocol: "redis",
		Addr: "10.0.0.10", Port: 6379, User: "",
	}
	a.DB = app.DBState{
		HostID: "r1", HostName: "redis-01", Protocol: "redis",
		Schema: &dbexec.Schema{}, Open: map[string]*bool{}, Split: 190,
	}
	a.NewDBQueryTab()
	tt := ui.NewTester(Shell(a), 1200, 800)
	if !tt.HasText("该连接类型没有关系表结构") {
		t.Fatalf("empty schema hint missing: %q", tt.Texts())
	}
}

func TestDBObjectTreeFilter(t *testing.T) {
	a := dbViewTestApp()
	a.DB.Filter = "v_"
	tt := ui.NewTester(Shell(a), 1200, 800)
	if !tt.HasText("v_users") {
		t.Fatalf("filtered view missing: %q", tt.Texts())
	}
	for _, text := range tt.Texts() {
		if text == "users" {
			t.Fatalf("filtered-out table is shown: %q", tt.Texts())
		}
	}
	if !tt.HasText("表 (0)") || !tt.HasText("视图 (1)") {
		t.Fatalf("filtered counts missing: %q", tt.Texts())
	}

	a.DB.Filter = "zzz"
	tt.Frame()
	if !tt.HasText("没有匹配「zzz」的表或视图。") {
		t.Fatalf("empty filter hint missing: %q", tt.Texts())
	}
}

func TestDBTablePagerInsideWindow(t *testing.T) {
	a := dbViewTestApp()
	a.OpenDBTable("users")
	tab := a.ActiveDBTab()
	tab.Err = ""
	tab.HasTotal = true
	tab.Total = 500
	tab.Page = 0
	tab.PageSize = 100
	a.DB.Filter = ""
	cols := make([]string, 30)
	row := make([]any, 30)
	for i := range cols {
		cols[i] = fmt.Sprintf("column_%02d", i)
		row[i] = i
	}
	tab.Result = &dbexec.Result{Columns: cols, Rows: [][]any{row}}

	const width = 1100.0
	tt := ui.NewTester(Shell(a), int(width), 800)
	for _, label := range []string{"上一页", "下一页", "刷新"} {
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("%q not found", label)
		}
		if r.X+r.W > width+1 {
			t.Fatalf("%q clipped: x=%.1f w=%.1f window=%.0f", label, r.X, r.W, width)
		}
	}
}

func TestDBTabCloseButton(t *testing.T) {
	a := dbViewTestApp()
	tt := ui.NewTester(Shell(a), 1200, 800)
	if err := tt.Click("+"); err != nil {
		t.Fatal(err)
	}
	if len(a.DB.Tabs) != 2 {
		t.Fatalf("tabs = %d, want 2", len(a.DB.Tabs))
	}
	if err := tt.Click("×"); err != nil {
		t.Fatal(err)
	}
	if len(a.DB.Tabs) != 1 {
		t.Fatalf("tabs after close = %d, want 1", len(a.DB.Tabs))
	}
}

func TestDBTableHorizontalScrollWithNoRows(t *testing.T) {
	a := dbViewTestApp()
	a.OpenDBTable("users")
	tab := a.ActiveDBTab()
	tab.Err = ""
	tab.HasTotal = true
	tab.Total = 0
	tab.PageSize = 100
	cols := make([]string, 20)
	for i := range cols {
		cols[i] = fmt.Sprintf("column_%02d", i)
	}
	tab.Result = &dbexec.Result{Columns: cols} // no rows
	tt := ui.NewTester(Shell(a), 1100, 800)
	before, ok := tt.Find("column_00")
	if !ok {
		t.Fatal("first column not found")
	}
	tt.Scroll(700, 300, 80, 0)
	tt.Frame()
	after, ok := tt.Find("column_00")
	if !ok {
		t.Fatal("first column not found after scrolling")
	}
	if after.X >= before.X {
		t.Fatalf("table did not scroll horizontally: x %.1f -> %.1f", before.X, after.X)
	}
}
