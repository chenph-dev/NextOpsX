package db_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestDBSelectionPanelBeforeConnect(t *testing.T) {
	a := &workspace.State{}
	a.Selected = &hoststore.TreeNode{
		ID: "db1", Name: "mysql-01", NodeType: "host", Protocol: "mysql",
		Addr: "10.0.0.9", Port: 3306, User: "root",
		Params: map[string]any{"database": "app"},
	}
	tt := ui.NewTester(views.Shell(a), 1200, 800)
	if !tt.HasText("双击资产连接数据库") {
		t.Fatalf("selection hint missing: %q", tt.Texts())
	}
	if !tt.HasText("MySQL · app") {
		t.Fatalf("selection meta missing: %q", tt.Texts())
	}
	if tt.HasText("对象") {
		t.Fatalf("workbench shown before connecting: %q", tt.Texts())
	}

	a.OpenAsset(a.Selected)
	tt.Frame()
	if !tt.HasText("对象") {
		t.Fatalf("workbench missing after connect: %q", tt.Texts())
	}
}
