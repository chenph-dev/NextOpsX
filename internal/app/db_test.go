package app

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"nextopsx/internal/backend"
	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func TestDBTabsLifecycle(t *testing.T) {
	a := &App{}
	a.DB = DBState{HostID: "h1", Protocol: "mysql", Open: map[string]*bool{}}
	a.NewDBQueryTab()
	if len(a.DB.Tabs) != 1 || a.DB.Active != 0 {
		t.Fatalf("tabs = %d active = %d", len(a.DB.Tabs), a.DB.Active)
	}
	first := a.DB.Tabs[0]
	if first.Kind != DBTabQuery || first.Editor == nil || first.Editor.Buffer.String() != "" {
		t.Fatalf("query tab = %+v", first)
	}

	a.NewDBQueryTab()
	if len(a.DB.Tabs) != 2 || a.DB.Active != 1 {
		t.Fatalf("tabs = %d active = %d", len(a.DB.Tabs), a.DB.Active)
	}
	second := a.DB.Tabs[1]
	if second.Title != "查询 2" {
		t.Fatalf("second title = %q", second.Title)
	}

	a.ActivateDBTab(0)
	a.CloseDBTab(second.Key)
	if len(a.DB.Tabs) != 1 || a.DB.Active != 0 || a.DB.Tabs[0] != first {
		t.Fatalf("after close: tabs=%d active=%d", len(a.DB.Tabs), a.DB.Active)
	}
	a.CloseDBTab(first.Key)
	if len(a.DB.Tabs) != 0 || a.DB.Active != 0 {
		t.Fatalf("after closing all: tabs=%d active=%d", len(a.DB.Tabs), a.DB.Active)
	}
}

func TestDBQuoteIdent(t *testing.T) {
	if got := DBQuoteIdent("mysql", "we`ird"); got != "`we``ird`" {
		t.Fatalf("mysql = %q", got)
	}
	if got := DBQuoteIdent("postgres", `we"ird`); got != `"we""ird"` {
		t.Fatalf("postgres = %q", got)
	}
	if got := DBQuoteIdent("sqlserver", "dbo.we]ird"); got != "[dbo].[we]]ird]" {
		t.Fatalf("sqlserver = %q", got)
	}
}

func TestDBTablePageSQL(t *testing.T) {
	got := DBTablePageSQL("mysql", "users", "id", 100, 100)
	want := "SELECT * FROM `users` ORDER BY `id` LIMIT 100 OFFSET 100"
	if got != want {
		t.Fatalf("mysql page = %q, want %q", got, want)
	}
	got = DBTablePageSQL("sqlserver", "dbo.users", "id", 100, 100)
	want = "SELECT * FROM [dbo].[users] ORDER BY [id] OFFSET 100 ROWS FETCH NEXT 100 ROWS ONLY"
	if got != want {
		t.Fatalf("sqlserver page = %q, want %q", got, want)
	}
	got = DBTablePageSQL("postgres", "users", "1", 100, 0)
	want = `SELECT * FROM "users" ORDER BY 1 LIMIT 100 OFFSET 0`
	if got != want {
		t.Fatalf("no-order page = %q, want %q", got, want)
	}
}

func TestAttachDBKeepsTabsAfterSchemaError(t *testing.T) {
	a := &App{}
	tab := newDBQueryTab("db-q-1", "查询")
	a.DB = DBState{
		HostID: "db1", Protocol: "mysql", SchemaErr: "boom",
		Open: map[string]*bool{}, Tabs: []*DBTab{tab}, Active: 0,
	}
	a.attachDB(&hoststore.TreeNode{ID: "db1", Name: "db", NodeType: "host", Protocol: "mysql"})
	if len(a.DB.Tabs) != 1 || a.DB.Tabs[0] != tab {
		t.Fatalf("tabs were reset after a failed schema load: %d", len(a.DB.Tabs))
	}
}

func TestDBResultCSV(t *testing.T) {
	cols := []string{"id", "name", "note"}
	rows := [][]any{
		{int64(1), "alice", nil},
		{int64(2), `say "hi", ok`, "line1\nline2"},
	}
	s, err := dbResultCSV(cols, rows)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(strings.NewReader(s)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("records = %d: %q", len(records), records)
	}
	if records[1][2] != "" {
		t.Fatalf("NULL should be empty: %q", records[1][2])
	}
	if records[2][1] != `say "hi", ok` || records[2][2] != "line1\nline2" {
		t.Fatalf("quoted row = %q", records[2])
	}
}

func TestDBTreeState(t *testing.T) {
	a := &App{}
	schema := &dbexec.Schema{Tables: []dbexec.Table{
		{Name: "users", Type: "table"},
		{Name: "v_users", Type: "view"},
	}}
	a.initDBOpen(schema)
	if !*a.DBTreeOpen("table") || !*a.DBTreeOpen("view") {
		t.Fatal("groups should start open")
	}
	if *a.DBTreeOpen("table:users") {
		t.Fatal("tables should start closed")
	}
	if key := DBTableKey(schema.Tables[1]); key != "view:v_users" {
		t.Fatalf("view key = %q", key)
	}
	a.SetDBTreeAll(true)
	if !*a.DBTreeOpen("table:users") {
		t.Fatal("expand all did not open tables")
	}
	a.SetDBTreeAll(false)
	if *a.DBTreeOpen("table:users") {
		t.Fatal("collapse all did not close tables")
	}
}

func TestOpenDBTableWithoutBackend(t *testing.T) {
	a := &App{}
	a.DB = DBState{HostID: "h1", Protocol: "mysql", Open: map[string]*bool{}}
	a.OpenDBTable("users")
	tab := a.ActiveDBTab()
	if tab == nil || tab.Kind != DBTabTable || tab.TableName != "users" {
		t.Fatalf("tab = %+v", tab)
	}
	if tab.Running || tab.Err != "后端未初始化" {
		t.Fatalf("load = running %v err %q", tab.Running, tab.Err)
	}
	a.OpenDBTable("users")
	if len(a.DB.Tabs) != 1 {
		t.Fatalf("duplicate table tab: %d", len(a.DB.Tabs))
	}
}

func TestDBWorkbenchSQLiteFlow(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	path := filepath.Join(t.TempDir(), "flow.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 150; i++ {
		if _, err := tx.Exec(`INSERT INTO users (name) VALUES (?)`, fmt.Sprintf("user-%03d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	db.Close()

	b, err := backend.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.Close)
	a, err := New(b)
	if err != nil {
		t.Fatal(err)
	}
	u := &chanUpdater{ch: make(chan func(), 256)}
	a.Attach(u)

	hostID, err := b.CreateHost(hoststore.HostInput{
		Name: "sqlite", Protocol: "sqlite",
		Params: map[string]any{"filePath": path},
	})
	if err != nil {
		t.Fatal(err)
	}
	a.Selected = &hoststore.TreeNode{
		ID: hostID, Name: "sqlite", NodeType: "host", Protocol: "sqlite",
	}
	a.attachDB(a.Selected)
	pumpUntil(t, u, func() bool {
		return a.DB.Schema != nil && !a.DB.SchemaLoading
	})
	if len(a.DB.Schema.Tables) != 1 || a.DB.Schema.Tables[0].Name != "users" {
		t.Fatalf("schema = %+v", a.DB.Schema)
	}

	tab := a.ActiveDBTab()
	tab.Editor.Buffer.Reset(`SELECT name FROM users ORDER BY id LIMIT 2`)
	a.ExecuteDBQuery(tab.Key)
	pumpUntil(t, u, func() bool { return !tab.Running })
	if tab.Err != "" || tab.Result == nil || len(tab.Result.Rows) != 2 {
		t.Fatalf("query err=%q result=%+v", tab.Err, tab.Result)
	}
	if tab.Result.Rows[0][0] != "user-001" || tab.Result.Rows[1][0] != "user-002" {
		t.Fatalf("rows = %+v", tab.Result.Rows)
	}

	a.OpenDBTable("users")
	table := a.ActiveDBTab()
	pumpUntil(t, u, func() bool {
		return !table.Running && table.HasTotal && table.Result != nil
	})
	if table.Total != 150 || len(table.Result.Rows) != 100 {
		t.Fatalf("page 0: total=%d rows=%d", table.Total, len(table.Result.Rows))
	}
	a.DBTablePage(1)
	pumpUntil(t, u, func() bool { return !table.Running && table.Page == 1 })
	if len(table.Result.Rows) != 50 {
		t.Fatalf("page 1 rows = %d", len(table.Result.Rows))
	}

	// Refresh must recount so newly inserted rows become reachable.
	if _, err := b.ExecuteSQL(hostID, `INSERT INTO users (name) VALUES ('user-151')`); err != nil {
		t.Fatal(err)
	}
	a.DBTablePage(-1)
	pumpUntil(t, u, func() bool { return !table.Running && table.Page == 0 })
	a.RefreshActiveDBTable()
	pumpUntil(t, u, func() bool { return !table.Running && table.Total == 151 })
}

func TestRefreshHostsAttachesDatabaseWorkbench(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	path := filepath.Join(t.TempDir(), "refresh.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	b, err := backend.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.Close)
	a, err := New(b)
	if err != nil {
		t.Fatal(err)
	}
	u := &chanUpdater{ch: make(chan func(), 64)}
	a.Attach(u)

	hostID, err := b.CreateHost(hoststore.HostInput{
		Name: "db", Protocol: "sqlite",
		Params: map[string]any{"filePath": path},
	})
	if err != nil {
		t.Fatal(err)
	}
	// This is the path used after creating/editing an asset: Selected is
	// set directly, without going through Select.
	a.refreshHosts(hostID)
	pumpUntil(t, u, func() bool {
		return a.Selected != nil && a.Selected.ID == hostID &&
			a.DB.HostID == hostID && !a.DB.SchemaLoading
	})
	if a.DB.Schema == nil || len(a.DB.Schema.Tables) != 1 {
		t.Fatalf("workbench not attached after refresh: %+v", a.DB)
	}
	if a.DB.Schema.Tables[0].Name != "items" {
		t.Fatalf("tables = %+v", a.DB.Schema.Tables)
	}
}
