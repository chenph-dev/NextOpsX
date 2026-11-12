package app

import (
	"encoding/json"
	"testing"

	"nextopsx/internal/dbexec"
	hoststore "nextopsx/internal/store/hosts"
)

func TestESDefaultDSLIsJSON(t *testing.T) {
	var v any
	if err := json.Unmarshal([]byte(ESDefaultDSL), &v); err != nil {
		t.Fatalf("default DSL is not JSON: %v", err)
	}
}

func TestESIndexNames(t *testing.T) {
	res := &dbexec.Result{
		Columns: []string{"health", "index"},
		Rows:    [][]any{{"green", "logs-2024"}, {"yellow", "metrics-*"}},
	}
	names := esIndexNames(res)
	if len(names) != 2 || names[0] != "logs-2024" || names[1] != "metrics-*" {
		t.Fatalf("indices = %v", names)
	}
}

func TestESHealthMetrics(t *testing.T) {
	res := &dbexec.Result{
		Columns: []string{"cluster_name", "status", "number_of_nodes", "number_of_pending_tasks", "other"},
		Rows:    [][]any{{"prod", "green", int64(3), int64(0), "x"}},
	}
	metrics := esHealthMetrics(res)
	if len(metrics) != 4 {
		t.Fatalf("metrics = %+v", metrics)
	}
	if metrics[1].Name != "status" || metrics[1].Value != "green" {
		t.Fatalf("status metric = %+v", metrics[1])
	}
}

func TestAttachESWithoutBackend(t *testing.T) {
	a := &App{}
	n := &hoststore.TreeNode{
		ID: "e1", Name: "es", NodeType: "host", Protocol: "elasticsearch",
		Addr: "10.0.0.11", Port: 9200,
		Params: map[string]any{"index": "logs-*"},
	}
	a.AttachES(n)
	if a.ES.HostID != "e1" || a.ES.SelectedIndex != "logs-*" || a.ES.QueryDSL == "" {
		t.Fatalf("state = %+v", a.ES)
	}
	if a.ES.IndicesErr != "后端未初始化" {
		t.Fatalf("indices err = %q", a.ES.IndicesErr)
	}
	a.ES.MainTab = 1
	a.Selected = n
	a.EnsureES()
	if a.ES.InfoErr != "后端未初始化" {
		t.Fatalf("info err = %q", a.ES.InfoErr)
	}

	a.ES.SelectedIndex = ""
	a.ESRunQuery()
	if a.ES.Err != "请先选择索引" {
		t.Fatalf("query without index = %q", a.ES.Err)
	}
	a.ES.SelectedIndex = "logs-*"
	a.ESRunQuery()
	if a.ES.QueryErr != "后端未初始化" {
		t.Fatalf("query err = %q", a.ES.QueryErr)
	}
	a.ES.Command = "_cluster/health"
	a.ESRunCommand()
	if a.ES.CommandErr != "后端未初始化" {
		t.Fatalf("command err = %q", a.ES.CommandErr)
	}
}

func TestESHistoryPick(t *testing.T) {
	a := &App{}
	a.ES.CommandHistory = []string{"_cat/indices?format=json"}
	a.ES.HistoryOpen = true
	a.ESPickHistory("_cluster/health")
	if a.ES.Command != "_cluster/health" || a.ES.HistoryOpen {
		t.Fatalf("pick = %q open=%v", a.ES.Command, a.ES.HistoryOpen)
	}
}

func TestESSelectIndexInvalidatesQuery(t *testing.T) {
	a := &App{}
	a.ES = ESState{
		SelectedIndex: "a",
		QueryResult:   &dbexec.Result{Columns: []string{"x"}, Rows: [][]any{{1}}},
		QueryRunning:  true,
		Indices:       []string{"a", "b"},
		IndexRow:      0,
	}
	a.ESSelectIndex("b")
	if a.ES.SelectedIndex != "b" || a.ES.QueryResult != nil || a.ES.QueryRunning {
		t.Fatalf("state = %+v", a.ES)
	}
	if a.ES.IndexRow != 1 {
		t.Fatalf("row = %d, want 1", a.ES.IndexRow)
	}
}

func TestESAttachUpdatesDefaultIndex(t *testing.T) {
	a := &App{}
	first := &hoststore.TreeNode{
		ID: "e1", Name: "es", NodeType: "host", Protocol: "elasticsearch",
		Addr: "h", Port: 9200, Params: map[string]any{"index": "logs"},
	}
	a.AttachES(first)
	if a.ES.SelectedIndex != "logs" || a.ES.DefaultIndex != "logs" {
		t.Fatalf("initial = %+v", a.ES)
	}
	changed := *first
	changed.Params = map[string]any{"index": "metrics"}
	a.AttachES(&changed)
	if a.ES.DefaultIndex != "metrics" || a.ES.SelectedIndex != "metrics" {
		t.Fatalf("after default change = %+v", a.ES)
	}

	// A manual selection is kept when the asset's default is unchanged.
	a.ESSelectIndex("orders")
	a.AttachES(&changed)
	if a.ES.SelectedIndex != "orders" {
		t.Fatalf("manual selection overwritten: %q", a.ES.SelectedIndex)
	}
}
