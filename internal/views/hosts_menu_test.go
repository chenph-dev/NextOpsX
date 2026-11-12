package views

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	hoststore "nextopsx/internal/store/hosts"
)

func TestHostTreeNestsChildren(t *testing.T) {
	a := &app.App{}
	rootOpen, childOpen := true, true
	a.Open = map[string]*bool{"root": &rootOpen, "child": &childOpen}
	a.Roots = []hoststore.TreeNode{{
		ID: "root", Name: "prod", NodeType: "folder",
		Children: []hoststore.TreeNode{{
			ID: "child", Name: "web-01", NodeType: "folder",
			ParentID: "root",
		}},
	}}
	tt := ui.NewTester(Shell(a), 1000, 700)
	rootRect, ok := tt.Find("prod")
	if !ok {
		t.Fatal("root not found")
	}
	childRect, ok := tt.Find("web-01")
	if !ok {
		t.Fatal("child not found")
	}
	if childRect.X <= rootRect.X {
		t.Fatalf("child is not indented: root x=%.1f child x=%.1f", rootRect.X, childRect.X)
	}
}

func TestFolderContextMenu(t *testing.T) {
	a := &app.App{}
	open := true
	a.Open = map[string]*bool{"f1": &open}
	a.Roots = []hoststore.TreeNode{{ID: "f1", Name: "prod", NodeType: "folder"}}
	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.RightClick("prod"); err != nil {
		t.Fatal(err)
	}
	want := []string{"在此新建主机", "在此新建目录", "-", "重命名", "移动", "-", "删除"}
	if got := tt.Menu(); !slices.Equal(got, want) {
		t.Fatalf("menu = %q, want %q", got, want)
	}
	if err := tt.ChooseMenuItem("重命名"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.RenameOpen || a.Hosts.RenameID != "f1" {
		t.Fatalf("rename dialog = %+v", a.Hosts)
	}
}

func TestHostContextMenu(t *testing.T) {
	a := &app.App{}
	a.Roots = []hoststore.TreeNode{{
		ID: "h1", Name: "web-01", NodeType: "host",
		Addr: "10.0.0.5", Port: 22, User: "root", Protocol: "ssh",
	}}
	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.RightClick("web-01"); err != nil {
		t.Fatal(err)
	}
	want := []string{"打开终端", "打开 AI 对话", "-", "编辑", "移动", "-", "删除"}
	if got := tt.Menu(); !slices.Equal(got, want) {
		t.Fatalf("menu = %q, want %q", got, want)
	}
	if err := tt.ChooseMenuItem("编辑"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || a.Hosts.Form.EditingID != "h1" {
		t.Fatalf("edit form = %+v", a.Hosts.Form)
	}
}

func TestSidebarContextMenu(t *testing.T) {
	a := &app.App{}
	tt := ui.NewTester(Shell(a), 1000, 700)
	tt.RightClickAt(100, 400)
	want := []string{"新建主机", "新建目录", "-", "刷新"}
	if got := tt.Menu(); !slices.Equal(got, want) {
		t.Fatalf("menu = %q, want %q", got, want)
	}
	if err := tt.ChooseMenuItem("新建主机"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || a.Hosts.Form.IsFolder {
		t.Fatalf("new host form = %+v", a.Hosts.Form)
	}
}
