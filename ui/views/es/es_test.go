package es_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/dbexec"
	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func esTestApp() *workspace.State {
	a := &workspace.State{}
	a.Selected = &hoststore.TreeNode{
		ID: "e1", Name: "es-01", NodeType: "host", Protocol: "elasticsearch",
		Addr: "10.0.0.11", Port: 9200, User: "elastic",
	}
	a.ES = workspace.ESState{
		HostID: "e1", Label: "es-01 · 10.0.0.11:9200", IndexRow: 0,
		Indices:       []string{"logs-2024", "metrics-*"},
		SelectedIndex: "logs-2024",
		QueryDSL:      workspace.ESDefaultDSL,
		QueryResult: &dbexec.Result{
			Columns: []string{"_id", "message"},
			Rows:    [][]any{{"1", "hello"}, {"2", "world"}},
		},
		Info: []workspace.ESMetric{
			{Name: "cluster_name", Value: "prod"},
			{Name: "status", Value: "green"},
		},
		CommandResult:  &dbexec.Result{Columns: []string{"status"}, Rows: [][]any{{"green"}}},
		CommandHistory: []string{"_cluster/health"},
	}
	return a
}

func TestESPanelRenders(t *testing.T) {
	a := esTestApp()
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	for _, want := range []string{
		"ES", "索引 (2)", "刷新索引",
		"logs-2024", "metrics-*", "DSL 查询", "集群信息",
		"索引：logs-2024", "执行查询", "hello", "world",
		"REST 命令终端", "执行", "green",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestESRefreshIconInIndexHeader(t *testing.T) {
	a := esTestApp()
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	title, ok := tt.Find("索引 (2)")
	if !ok {
		t.Fatalf("index header not found: %q", tt.Texts())
	}
	r, ok := tt.Find("刷新索引")
	if !ok {
		t.Fatalf("refresh icon not found: %q", tt.Texts())
	}
	if r.X < title.X || r.X+r.W > title.X+260 {
		t.Fatalf("refresh icon outside the index pane: x=%.1f w=%.1f pane=%.1f", r.X, r.W, title.X)
	}
	if cy := r.Y + r.H/2; cy < title.Y || cy > title.Y+title.H {
		t.Fatalf("refresh icon not on the index header row: y=%.1f header=%.1f..%.1f",
			cy, title.Y, title.Y+title.H)
	}
	if tt.HasText("Elasticsearch · es-01 · 10.0.0.11:9200") {
		t.Fatal("top header row should be gone")
	}
	if tt.HasText("就绪") {
		t.Fatal("empty status bar should be gone")
	}
	for _, text := range tt.Texts() {
		if text == "索引" {
			t.Fatalf("duplicate column header remains: %q", tt.Texts())
		}
	}
}

func TestESIndexListSelection(t *testing.T) {
	a := esTestApp()
	a.ES.SelectedIndex = ""
	a.ES.IndexRow = -1
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	if err := tt.Click("metrics-*"); err != nil {
		t.Fatal(err)
	}
	if a.ES.SelectedIndex != "metrics-*" || a.ES.IndexRow != 1 {
		t.Fatalf("selected index = %q row = %d", a.ES.SelectedIndex, a.ES.IndexRow)
	}
}

func TestESInfoAndEmptySelection(t *testing.T) {
	a := esTestApp()
	a.ES.MainTab = 1
	tt := ui.NewTester(views.Shell(a), 1300, 900)
	for _, want := range []string{"集群信息", "刷新", "cluster_name", "prod", "status"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}

	a.ES.MainTab = 0
	a.ES.SelectedIndex = ""
	tt.Frame()
	if !tt.HasText("索引：未选择") {
		t.Fatalf("unselected index hint missing: %q", tt.Texts())
	}
}
