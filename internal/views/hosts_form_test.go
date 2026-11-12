package views

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/backend"
	hoststore "nextopsx/internal/store/hosts"
)

type viewUpdater struct{ ch chan func() }

func (u *viewUpdater) Update(fn func()) { u.ch <- fn }

func pumpView(t *testing.T, u *viewUpdater, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		select {
		case fn := <-u.ch:
			fn()
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestHostFormDialogRenders(t *testing.T) {
	a := &app.App{}
	tt := ui.NewTester(Shell(a), 1000, 800)
	if !tt.HasText("新建主机") {
		t.Fatalf("empty state missing new-host action: %q", tt.Texts())
	}
	if err := tt.Click("新建主机"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, want := range []string{
		"位置：根目录", "名称", "协议", "地址", "端口", "用户",
		"认证", "密码", "自动放行", "测试连接", "保存", "取消",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestHostFormTypingBindsFields(t *testing.T) {
	a := &app.App{}
	a.OpenNewHostForm()
	tt := ui.NewTester(Shell(a), 1000, 800)
	tt.Frame()

	typeInto := func(label, text string) {
		t.Helper()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("label %q not found", label)
		}
		tt.ClickAt(r.X+r.W+80, r.Y+r.H/2)
		tt.Type(text)
	}
	typeInto("名称", "web-01")
	typeInto("地址", "10.0.0.5")
	typeInto("凭据", "pw")
	if a.Hosts.Form.Name != "web-01" || a.Hosts.Form.Addr != "10.0.0.5" || a.Hosts.Form.Secret != "pw" {
		t.Fatalf("bound form = %+v", a.Hosts.Form)
	}
}

func TestCreateHostThroughFormUI(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := backend.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.Close)
	a, err := app.New(b)
	if err != nil {
		t.Fatal(err)
	}
	u := &viewUpdater{ch: make(chan func(), 128)}
	a.Attach(u)

	tt := ui.NewTester(Shell(a), 1000, 900)
	if err := tt.Click("新建主机"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	typeInto := func(label, text string) {
		t.Helper()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("label %q not found", label)
		}
		tt.ClickAt(r.X+r.W+80, r.Y+r.H/2)
		tt.Type(text)
	}
	typeInto("名称", "web-01")
	typeInto("地址", "10.0.0.5")
	typeInto("凭据", "pw")
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	pumpView(t, u, func() bool {
		return len(a.Roots) == 1 && !a.Hosts.Form.Open
	})
	if a.Roots[0].Name != "web-01" || a.Roots[0].Addr != "10.0.0.5" || a.Roots[0].Protocol != "ssh" {
		t.Fatalf("created host = %+v", a.Roots[0])
	}
	secret, _, err := b.GetHostSecret(a.Roots[0].ID)
	if err != nil || secret != "pw" {
		t.Fatalf("stored secret = %q err=%v", secret, err)
	}
	pumpView(t, u, func() bool { return a.AI.SessionID != "" })
}

func TestHostFormValidationMessage(t *testing.T) {
	a := &app.App{}
	a.OpenNewHostForm()
	tt := ui.NewTester(Shell(a), 1000, 800)
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("请输入名称") {
		t.Fatalf("validation message missing: %q", tt.Texts())
	}
}

func TestHostFormProtocolParams(t *testing.T) {
	a := &app.App{}
	a.OpenNewHostForm()
	a.Hosts.Form.Protocol = "mysql"
	a.SyncHostProtocol()
	if a.Hosts.Form.Port != "3306" {
		t.Fatalf("mysql default port = %q", a.Hosts.Form.Port)
	}

	tt := ui.NewTester(Shell(a), 1000, 900)
	for _, want := range []string{"数据库", "地址"} {
		if !tt.HasText(want) {
			t.Errorf("mysql form missing %q in %q", want, tt.Texts())
		}
	}

	a.Hosts.Form.Protocol = "sqlite"
	a.SyncHostProtocol()
	tt.Frame()
	for _, want := range []string{"数据库文件路径", "选择文件"} {
		if !tt.HasText(want) {
			t.Errorf("sqlite form missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("地址") {
		t.Errorf("sqlite form should hide host fields: %q", tt.Texts())
	}

	a.Hosts.Form.Protocol = "ssh"
	a.SyncHostProtocol()
	tt.Frame()
	if !tt.HasText("地址") {
		t.Errorf("ssh form did not restore host fields: %q", tt.Texts())
	}
	if a.Hosts.Form.Port != "22" {
		t.Errorf("ssh default port = %q", a.Hosts.Form.Port)
	}
}

func TestFolderActionsAndDialogs(t *testing.T) {
	a := &app.App{}
	a.Selected = &hoststore.TreeNode{ID: "f1", Name: "prod", NodeType: "folder"}
	a.Roots = []hoststore.TreeNode{{ID: "f1", Name: "prod", NodeType: "folder"}}
	tt := ui.NewTester(Shell(a), 1000, 800)

	for _, want := range []string{"在此新建主机", "在此新建目录", "重命名", "移动", "删除"} {
		if !tt.HasText(want) {
			t.Errorf("folder action %q missing: %q", want, tt.Texts())
		}
	}
	if err := tt.Click("在此新建目录"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || !a.Hosts.Form.IsFolder || a.Hosts.Form.ParentID != "f1" {
		t.Fatalf("folder form = %+v", a.Hosts.Form)
	}
	tt.Frame()
	if !tt.HasText("新建目录") || !tt.HasText("位置：prod") {
		t.Fatalf("folder dialog missing: %q", tt.Texts())
	}
}

func TestEditHostFormAndDeleteDialog(t *testing.T) {
	a := &app.App{}
	a.Selected = &hoststore.TreeNode{
		ID: "h1", Name: "web-01", NodeType: "host", Addr: "10.0.0.5",
		Port: 22, User: "root", AuthType: "password", Protocol: "ssh",
		AutoApprove: "inherit",
	}
	tt := ui.NewTester(Shell(a), 1000, 900)
	if err := tt.Click("编辑"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || a.Hosts.Form.EditingID != "h1" {
		t.Fatalf("edit form = %+v", a.Hosts.Form)
	}
	tt.Frame()
	for _, want := range []string{"编辑主机", "web-01", "10.0.0.5", "保留原值"} {
		if !tt.HasText(want) {
			t.Errorf("edit form missing %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}

	a.OpenDeleteDialog()
	tt.Frame()
	if !tt.HasText("删除资产") || !tt.HasText("确定删除「web-01」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	if a.Hosts.DeleteOpen {
		t.Fatal("delete dialog did not close")
	}
}
