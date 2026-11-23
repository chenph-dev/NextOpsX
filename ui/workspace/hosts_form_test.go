package workspace

import (
	"strings"
	"testing"
	"time"

	"nextopsx/i18n"
	"nextopsx/service/backend"
	hoststore "nextopsx/store/hosts"
)

type chanUpdater struct{ ch chan func() }

func (u *chanUpdater) Update(fn func()) { u.ch <- fn }

func pumpUntil(t *testing.T, u *chanUpdater, cond func() bool) {
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

func TestHostInputValidation(t *testing.T) {
	a := &State{}
	a.Hosts.Form = HostForm{
		Protocol: "ssh", Name: "web", Addr: "10.0.0.5", Port: "22",
		User: "root", AuthType: "password", Secret: "pw",
	}
	in, err := a.hostInputFromForm()
	if err != nil {
		t.Fatalf("valid form: %v", err)
	}
	if in.Name != "web" || in.Port != 22 || in.AuthType != "password" || in.Secret != "pw" {
		t.Fatalf("input = %+v", in)
	}

	a.Hosts.Form.Name = " "
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "名称") {
		t.Fatalf("name error = %v", err)
	}
	a.Hosts.Form.Name = "web"

	a.Hosts.Form.Port = "70000"
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "端口") {
		t.Fatalf("port error = %v", err)
	}
	a.Hosts.Form.Port = "22"

	a.Hosts.Form.Secret = ""
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "密码") {
		t.Fatalf("secret error = %v", err)
	}

	// Editing may omit the secret, but switching auth type may not.
	a.Hosts.Form.EditingID = "id1"
	a.Hosts.Form.OriginalAuthType = "password"
	if _, err := a.hostInputFromForm(); err != nil {
		t.Fatalf("edit with kept secret: %v", err)
	}
	a.Hosts.Form.AuthType = "privatekey"
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "认证方式") {
		t.Fatalf("auth change error = %v", err)
	}

	// Driver-specific required parameters.
	a = &State{}
	a.Hosts.Form = HostForm{Name: "db", Addr: "h", Port: "3306", User: "u", AuthType: "password", Secret: "pw"}
	a.SelectHostProtocol("mysql")
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "数据库") {
		t.Fatalf("mysql param error = %v", err)
	}
	a.Hosts.Form = HostForm{Name: "db"}
	a.SelectHostProtocol("sqlite")
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "数据库文件路径") {
		t.Fatalf("sqlite param error = %v", err)
	}
}

func TestCreateAndEditHostFromForm(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := backend.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.Close)
	a, err := New(b)
	if err != nil {
		t.Fatal(err)
	}
	u := &chanUpdater{ch: make(chan func(), 128)}
	a.Attach(u)

	a.OpenNewHostForm()
	if !a.Hosts.Form.Open || a.Hosts.Form.Protocol != "" || a.Hosts.Form.Port != "" {
		t.Fatalf("form defaults = %+v", a.Hosts.Form)
	}
	a.SelectHostProtocol("ssh")
	if a.Hosts.Form.Protocol != "ssh" || a.Hosts.Form.Port != "22" || a.Hosts.Form.User != "root" {
		t.Fatalf("ssh defaults = %+v", a.Hosts.Form)
	}
	f := &a.Hosts.Form
	f.Name = "web-01"
	f.Addr = "10.0.0.5"
	f.User = "root"
	f.Secret = "pw"
	a.SaveHostForm()
	pumpUntil(t, u, func() bool { return len(a.Roots) == 1 && !a.Hosts.Form.Open })
	if a.Roots[0].Name != "web-01" || a.Roots[0].Protocol != "ssh" || a.Roots[0].Addr != "10.0.0.5" {
		t.Fatalf("created host = %+v", a.Roots[0])
	}
	pumpUntil(t, u, func() bool { return a.AI.SessionID != "" })

	id := a.Roots[0].ID
	a.Selected = &a.Roots[0]
	a.OpenEditHostForm()
	if a.Hosts.Form.Secret != "" || a.Hosts.Form.Name != "web-01" {
		t.Fatalf("edit form = %+v", a.Hosts.Form)
	}
	if a.Hosts.Form.AuthLabel != "密码" || a.Hosts.Form.AutoApproveLabel != "继承全局" {
		t.Fatalf("edit pickers = %q/%q", a.Hosts.Form.AuthLabel, a.Hosts.Form.AutoApproveLabel)
	}
	a.Hosts.Form.Name = "web-02"
	a.SaveHostForm()
	pumpUntil(t, u, func() bool {
		return len(a.Roots) == 1 && a.Roots[0].Name == "web-02" && !a.Hosts.Form.Open
	})
	secret, authType, err := b.Hosts.GetHostSecret(id)
	if err != nil || secret != "pw" || authType != "password" {
		t.Fatalf("stored secret = %q/%q err=%v", secret, authType, err)
	}
	if a.Selected == nil || a.Selected.Name != "web-02" {
		t.Fatalf("selection after refresh = %+v", a.Selected)
	}
}

func TestHostFormSeqIsMonotonic(t *testing.T) {
	a := &State{}
	a.OpenNewHostForm()
	first := a.Hosts.Form.seq
	a.CloseHostForm()
	closed := a.Hosts.Form.seq
	a.OpenNewHostForm()
	second := a.Hosts.Form.seq
	if first == 0 || second <= closed || closed <= first {
		t.Fatalf("seqs: first=%d closed=%d second=%d", first, closed, second)
	}
	if second == first {
		t.Fatal("reopened form reused a sequence number")
	}
}

func TestProtocolPickerDefaults(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })
	a := &State{}
	a.OpenNewHostForm()
	if a.HostProtocolLabel("ssh") != "Linux" {
		t.Fatalf("ssh label = %q", a.HostProtocolLabel("ssh"))
	}
	if a.HostProtocolSummary("mysql") == "" {
		t.Fatal("mysql summary is empty")
	}

	a.SelectHostProtocol("mysql")
	if a.Hosts.Form.Port != "3306" || a.Hosts.Form.User != "root" {
		t.Fatalf("mysql defaults = %+v", a.Hosts.Form)
	}
	a.SelectHostProtocol("postgres")
	if a.Hosts.Form.Port != "5432" || a.Hosts.Form.User != "postgres" {
		t.Fatalf("postgres defaults = %+v", a.Hosts.Form)
	}
	a.SelectHostProtocol("postgres")
	if a.Hosts.Form.Port != "5432" {
		t.Fatalf("re-select postgres port = %q", a.Hosts.Form.Port)
	}
	a.SelectHostProtocol("sqlite")
	if a.Hosts.Form.Port != "" || a.Hosts.Form.User != "" || a.HostFormNeedsHost() {
		t.Fatalf("sqlite defaults = %+v", a.Hosts.Form)
	}
	foundPath := false
	for _, p := range a.Hosts.Form.Params {
		if p.Key == "filePath" {
			foundPath = true
		}
	}
	if !foundPath {
		t.Fatalf("sqlite params = %+v", a.Hosts.Form.Params)
	}

	a.SelectHostProtocol("postgres")
	a.ClearHostProtocol()
	a.SelectHostProtocol("ssh")
	if a.Hosts.Form.User != "root" || a.Hosts.Form.Port != "22" {
		t.Fatalf("reselect ssh after cancel = %+v", a.Hosts.Form)
	}
}

func TestProtocolMetadataTranslated(t *testing.T) {
	for _, lang := range i18n.Languages {
		for _, key := range []string{"hosts.protocol.ssh.name", "hosts.protocol.winrm.name",
			"hosts.form.chooseProtocol", "hosts.err.protocolRequired"} {
			if !i18n.Has(lang, key) {
				t.Errorf("%s missing %q", lang, key)
			}
		}
		for proto, key := range hostProtocolSummaryKeys {
			if !i18n.Has(lang, key) {
				t.Errorf("%s missing %q for %s", lang, key, proto)
			}
		}
	}
}

func TestTerminalSessionTabs(t *testing.T) {
	a := &State{}
	a.Roots = []hoststore.TreeNode{
		{ID: "h1", Name: "first", NodeType: "host", Protocol: "ssh"},
		{ID: "h2", Name: "second", NodeType: "host", Protocol: "ssh"},
	}
	a.Sessions = []*TerminalSession{
		{HostID: "h1", Protocol: "ssh", Status: TerminalConnecting},
		{HostID: "h2", Protocol: "ssh", Status: TerminalConnecting},
	}
	a.Tabs = []string{"h1", "h2"}
	a.Selected = &a.Roots[1]
	a.ActivateSession("h1")
	if a.ActiveSessionID != "h1" || a.Selected == nil || a.Selected.ID != "h1" || a.TermHostID != "h1" {
		t.Fatalf("activate h1: %+v active=%q term=%q", a.Selected, a.ActiveSessionID, a.TermHostID)
	}
	a.NextSession()
	if a.ActiveSessionID != "h2" || a.Selected == nil || a.Selected.ID != "h2" {
		t.Fatalf("next session: %+v active=%q", a.Selected, a.ActiveSessionID)
	}
	a.PrevSession()
	if a.ActiveSessionID != "h1" {
		t.Fatalf("previous session: active=%q", a.ActiveSessionID)
	}
	a.CloseSession("h1")
	if len(a.Sessions) != 1 || a.ActiveSessionID != "h2" || a.Selected == nil || a.Selected.ID != "h2" {
		t.Fatalf("close active session: sessions=%d active=%q selected=%v", len(a.Sessions), a.ActiveSessionID, a.Selected)
	}
	if a.IsAssetTabOpen("h1") {
		t.Fatalf("closed session kept its tab: %v", a.Tabs)
	}
	a.CloseSession("h2")
	if len(a.Sessions) != 0 || a.ActiveSessionID != "" || a.TermHostID != "" {
		t.Fatalf("close last session: sessions=%d active=%q term=%q", len(a.Sessions), a.ActiveSessionID, a.TermHostID)
	}
	if len(a.Tabs) != 0 {
		t.Fatalf("closed sessions kept tabs: %v", a.Tabs)
	}
}
