package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func esTestApp() *app.App {
	a := &app.App{RightTab: 0}
	a.Selected = &hoststore.TreeNode{
		ID: "e1", Name: "es-01", NodeType: "host", Protocol: "elasticsearch",
		Addr: "10.0.0.11", Port: 9200, User: "elastic",
	}
	a.ES = app.ESState{
		HostID: "e1", Label: "es-01 · 10.0.0.11:9200", IndexRow: 0,
		Indices:       []string{"logs-2024", "metrics-*"},
		SelectedIndex: "logs-2024",
		QueryDSL:      app.ESDefaultDSL,
		QueryResult: &dbexec.Result{
			Columns: []string{"_id", "message"},
			Rows:    [][]any{{"1", "hello"}, {"2", "world"}},
		},
		Info: []app.ESMetric{
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
	tt := ui.NewTester(Shell(a), 1300, 900)
	for _, want := range []string{
		"ES", "Elasticsearch · es-01 · 10.0.0.11:9200", "索引 (2)",
		"logs-2024", "metrics-*", "DSL 查询", "集群信息",
		"索引：logs-2024", "执行查询", "hello", "world",
		"REST 命令终端", "执行", "green",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestESInfoAndEmptySelection(t *testing.T) {
	a := esTestApp()
	a.ES.MainTab = 1
	tt := ui.NewTester(Shell(a), 1300, 900)
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
