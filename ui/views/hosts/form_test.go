package hosts_test

import (
	"testing"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/service/backend"
	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
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
	a := &workspace.State{}
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	if !tt.HasText("新建资产") {
		t.Fatalf("empty state missing new-host action: %q", tt.Texts())
	}
	if err := tt.Click("新建资产"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, want := range []string{"选择协议", "Linux", "Windows", "MySQL", "取消"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("保存") {
		t.Fatalf("form drawer is visible before a protocol is chosen: %q", tt.Texts())
	}
	if err := tt.Click("Linux"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, want := range []string{
		"位置：根目录", "名称", "地址", "端口", "用户",
		"认证", "密码", "自动放行", "测试连接", "保存", "取消",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestHostFormTypingBindsFields(t *testing.T) {
	a := &workspace.State{}
	a.OpenNewHostForm()
	a.SelectHostProtocol("ssh")
	tt := ui.NewTester(views.Shell(a), 1000, 800)
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

func TestHostFormSecretIsMasked(t *testing.T) {
	a := &workspace.State{}
	a.OpenNewHostForm()
	a.SelectHostProtocol("ssh")
	a.Hosts.Form.Secret = "s3cret!"
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	tt.Frame()

	label, ok := tt.Find("凭据")
	if !ok {
		t.Fatalf("credential field missing: %q", tt.Texts())
	}
	// The field's ink, from the end of its label across the text: the eye
	// sits at the far right, outside this window.
	ink := func() uint64 {
		img := tt.Image()
		b := img.Bounds()
		x0 := min(max(int(label.X+label.W)+2, b.Min.X), b.Max.X)
		y0 := min(max(int(label.Y)-6, b.Min.Y), b.Max.Y)
		x1 := min(x0+140, b.Max.X)
		y1 := min(y0+32, b.Max.Y)
		var sum uint64
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				r, g, bl, _ := img.At(x, y).RGBA()
				sum += uint64(r) + uint64(g) + uint64(bl)
			}
		}
		return sum
	}

	hidden := ink()
	a.Hosts.Form.SecretVisible = true
	tt.Frame()
	if ink() == hidden {
		t.Fatal("the credential is drawn the same hidden and shown")
	}
}

func TestHostFormSecretVisibilityToggle(t *testing.T) {
	a := &workspace.State{}
	a.OpenNewHostForm()
	a.SelectHostProtocol("ssh")
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	tt.Frame()

	r, ok := tt.Find("凭据")
	if !ok {
		t.Fatalf("credential field missing: %q", tt.Texts())
	}
	tt.ClickAt(r.X+r.W+80, r.Y+r.H/2)
	tt.Type("s3cret")

	if err := tt.Click("显示密码"); err != nil {
		t.Fatalf("eye button missing: %v (%q)", err, tt.Texts())
	}
	if !a.Hosts.Form.SecretVisible {
		t.Fatal("clicking the eye did not reveal the secret")
	}
	if a.Hosts.Form.Secret != "s3cret" {
		t.Fatalf("secret changed while revealing: %q", a.Hosts.Form.Secret)
	}

	// The eye must not take the focus: typing carries on in the field.
	tt.Type("!")
	if a.Hosts.Form.Secret != "s3cret!" {
		t.Fatalf("typing after revealing = %q", a.Hosts.Form.Secret)
	}

	tt.Frame()
	if err := tt.Click("隐藏密码"); err != nil {
		t.Fatalf("eye button missing after revealing: %v (%q)", err, tt.Texts())
	}
	if a.Hosts.Form.SecretVisible {
		t.Fatal("clicking the eye again did not hide the secret")
	}
	if a.Hosts.Form.Secret != "s3cret!" {
		t.Fatalf("secret changed while hiding: %q", a.Hosts.Form.Secret)
	}

	// A private key is a text area, which cannot hide its text.
	a.Hosts.Form.AuthType = "privatekey"
	tt.Frame()
	if tt.HasText("显示密码") || tt.HasText("隐藏密码") {
		t.Errorf("private key text area should not offer the eye: %q", tt.Texts())
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
	a, err := workspace.New(b)
	if err != nil {
		t.Fatal(err)
	}
	u := &viewUpdater{ch: make(chan func(), 128)}
	a.Attach(u)

	tt := ui.NewTester(views.Shell(a), 1000, 900)
	if err := tt.Click("新建资产"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Linux"); err != nil {
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
	secret, _, err := b.Hosts.GetHostSecret(a.Roots[0].ID)
	if err != nil || secret != "pw" {
		t.Fatalf("stored secret = %q err=%v", secret, err)
	}
	pumpView(t, u, func() bool { return a.AI.SessionID != "" })
}

func TestHostFormValidationMessage(t *testing.T) {
	a := &workspace.State{}
	a.OpenNewHostForm()
	a.SelectHostProtocol("ssh")
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("请输入名称") {
		t.Fatalf("validation message missing: %q", tt.Texts())
	}
}

func TestHostFormProtocolParams(t *testing.T) {
	a := &workspace.State{}
	a.OpenNewHostForm()
	a.SelectHostProtocol("mysql")
	if a.Hosts.Form.Port != "3306" {
		t.Fatalf("mysql default port = %q", a.Hosts.Form.Port)
	}

	tt := ui.NewTester(views.Shell(a), 1000, 900)
	for _, want := range []string{"数据库", "地址"} {
		if !tt.HasText(want) {
			t.Errorf("mysql form missing %q in %q", want, tt.Texts())
		}
	}

	a.SelectHostProtocol("sqlite")
	tt.Frame()
	for _, want := range []string{"数据库文件路径", "选择文件"} {
		if !tt.HasText(want) {
			t.Errorf("sqlite form missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("地址") {
		t.Errorf("sqlite form should hide host fields: %q", tt.Texts())
	}

	a.SelectHostProtocol("ssh")
	tt.Frame()
	if !tt.HasText("地址") {
		t.Errorf("ssh form did not restore host fields: %q", tt.Texts())
	}
	if a.Hosts.Form.Port != "22" {
		t.Errorf("ssh default port = %q", a.Hosts.Form.Port)
	}
}

func TestFolderActionsAndDialogs(t *testing.T) {
	a := &workspace.State{}
	a.Selected = &hoststore.TreeNode{ID: "f1", Name: "prod", NodeType: "folder"}
	a.Roots = []hoststore.TreeNode{{ID: "f1", Name: "prod", NodeType: "folder"}}
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	if !tt.HasText("从右键菜单新建资产") {
		t.Fatalf("folder hint missing: %q", tt.Texts())
	}
	if err := tt.RightClick("prod"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("在此新建目录"); err != nil {
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
	a := &workspace.State{}
	a.Roots = []hoststore.TreeNode{{
		ID: "h1", Name: "web-01", NodeType: "host", Addr: "10.0.0.5",
		Port: 22, User: "root", AuthType: "password", Protocol: "ssh",
		AutoApprove: "inherit",
	}}
	a.Selected = &a.Roots[0]
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	if err := tt.RightClick("web-01"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("编辑"); err != nil {
		t.Fatal(err)
	}
	if !a.Hosts.Form.Open || a.Hosts.Form.EditingID != "h1" {
		t.Fatalf("edit form = %+v", a.Hosts.Form)
	}
	tt.Frame()
	for _, want := range []string{"编辑资产", "web-01", "10.0.0.5", "保留原值"} {
		if !tt.HasText(want) {
			t.Errorf("edit form missing %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	if a.Hosts.Form.Open {
		t.Fatalf("asset sheet did not close: %+v", a.Hosts.Form)
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

func TestHostFormEnglishLabels(t *testing.T) {
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })
	i18n.SetLanguage(i18n.EnUS)
	a := &workspace.State{}
	a.Settings.Language = "en-US"
	a.Settings.LanguageIndex = 1
	a.OpenNewHostForm()
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	tt.Frame()
	for _, want := range []string{"Choose a protocol", "Linux", "Windows", "Cancel"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("Linux"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, want := range []string{
		"New asset", "Name", "Address", "Port", "User",
		"Auth", "Password", "Auto-approve", "Test connection", "Save", "Cancel",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	for _, text := range tt.Texts() {
		for _, r := range text {
			if unicode.Is(unicode.Han, r) {
				t.Fatalf("untranslated Chinese in English host form: %q", text)
			}
		}
	}
}

func TestProtocolPickerOpensDefaults(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })
	a := &workspace.State{}
	a.OpenNewHostForm()
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	tt.Frame()
	if err := tt.Click("MySQL"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Hosts.Form.Protocol != "mysql" || a.Hosts.Form.Port != "3306" || a.Hosts.Form.User != "root" {
		t.Fatalf("mysql picker defaults = %+v", a.Hosts.Form)
	}
	for _, want := range []string{"MySQL 数据库，支持查询、结构浏览与结果导出。", "数据库", "地址", "测试连接", "保存"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Hosts.Form.Protocol != "" || !a.Hosts.Form.Open {
		t.Fatalf("drawer cancel: open=%v protocol=%q", a.Hosts.Form.Open, a.Hosts.Form.Protocol)
	}
	if tt.HasText("自动放行") || tt.HasText("保存") {
		t.Fatalf("drawer still visible after cancel: %q", tt.Texts())
	}
}
