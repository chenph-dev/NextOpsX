package hosts_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestHostTreeNestsChildren(t *testing.T) {
	a := &workspace.State{}
	rootOpen, childOpen := true, true
	a.Open = map[string]*bool{"root": &rootOpen, "child": &childOpen}
	a.Roots = []hoststore.TreeNode{{
		ID: "root", Name: "prod", NodeType: "folder",
		Children: []hoststore.TreeNode{{
			ID: "child", Name: "web-01", NodeType: "folder",
			ParentID: "root",
		}},
	}}
	tt := ui.NewTester(views.Shell(a), 1000, 700)
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
	a := &workspace.State{}
	open := true
	a.Open = map[string]*bool{"f1": &open}
	a.Roots = []hoststore.TreeNode{{ID: "f1", Name: "prod", NodeType: "folder"}}
	tt := ui.NewTester(views.Shell(a), 1000, 700)
	if err := tt.RightClick("prod"); err != nil {
		t.Fatal(err)
	}
	want := []string{"在此新建资产", "在此新建目录", "-", "重命名", "移动", "-", "删除"}
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
	a := &workspace.State{}
	a.Roots = []hoststore.TreeNode{{
		ID: "h1", Name: "web-01", NodeType: "host",
		Addr: "10.0.0.5", Port: 22, User: "root", Protocol: "ssh",
	}}
	tt := ui.NewTester(views.Shell(a), 1000, 700)
	if err := tt.RightClick("web-01"); err != nil {
		t.Fatal(err)
	}
	want := []string{"打开终端", "打开智能运维", "-", "编辑", "移动", "-", "删除"}
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

func TestWorkbenchHostContextMenu(t *testing.T) {
	cases := []struct {
		protocol string
		item     string
		attached func(*workspace.State) string
	}{
		{"mysql", "连接数据库", func(a *workspace.State) string { return a.DB.HostID }},
		{"redis", "连接", func(a *workspace.State) string { return a.Redis.HostID }},
		{"elasticsearch", "连接", func(a *workspace.State) string { return a.ES.HostID }},
		{"winrm", "打开 WinRM 终端", func(a *workspace.State) string { return a.WinRM.HostID }},
	}
	for _, tc := range cases {
		t.Run(tc.protocol, func(t *testing.T) {
			a := &workspace.State{}
			a.Roots = []hoststore.TreeNode{{
				ID: "h1", Name: "svc", NodeType: "host",
				Addr: "10.0.0.9", Port: 3306, User: "root", Protocol: tc.protocol,
			}}
			tt := ui.NewTester(views.Shell(a), 1000, 700)
			if err := tt.RightClick("svc"); err != nil {
				t.Fatal(err)
			}
			want := []string{tc.item, "打开智能运维", "-", "编辑", "移动", "-", "删除"}
			if tc.protocol == "winrm" {
				want = []string{tc.item, "远程桌面", "打开智能运维", "-", "编辑", "移动", "-", "删除"}
			}
			if got := tt.Menu(); !slices.Equal(got, want) {
				t.Fatalf("menu = %q, want %q", got, want)
			}
			if err := tt.ChooseMenuItem(tc.item); err != nil {
				t.Fatal(err)
			}
			if got := tc.attached(a); got != "h1" {
				t.Fatalf("workbench not attached: %q", got)
			}
		})
	}
}

func TestSidebarContextMenu(t *testing.T) {
	a := &workspace.State{}
	tt := ui.NewTester(views.Shell(a), 1000, 700)
	tt.RightClickAt(100, 400)
	want := []string{"新建资产", "新建目录", "-", "刷新"}
	if got := tt.Menu(); !slices.Equal(got, want) {
		t.Fatalf("menu = %q, want %q", got, want)
	}
	if err := tt.ChooseMenuItem("新建资产"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || a.Hosts.Form.IsFolder {
		t.Fatalf("new host form = %+v", a.Hosts.Form)
	}
}
