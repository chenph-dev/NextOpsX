package app

import (
	"strings"
	"testing"
	"time"

	"nextopsx/internal/backend"
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
	a := &App{}
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
	a = &App{}
	a.Hosts.Form = HostForm{Protocol: "mysql", Name: "db", Addr: "h", Port: "3306", User: "u", AuthType: "password", Secret: "pw"}
	a.SyncHostProtocol()
	if _, err := a.hostInputFromForm(); err == nil || !strings.Contains(err.Error(), "数据库") {
		t.Fatalf("mysql param error = %v", err)
	}
	a.Hosts.Form = HostForm{Protocol: "sqlite", Name: "db"}
	a.SyncHostProtocol()
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
	if !a.Hosts.Form.Open || a.Hosts.Form.Protocol != "ssh" || a.Hosts.Form.Port != "22" {
		t.Fatalf("form defaults = %+v", a.Hosts.Form)
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
	secret, authType, err := b.GetHostSecret(id)
	if err != nil || secret != "pw" || authType != "password" {
		t.Fatalf("stored secret = %q/%q err=%v", secret, authType, err)
	}
	if a.Selected == nil || a.Selected.Name != "web-02" {
		t.Fatalf("selection after refresh = %+v", a.Selected)
	}
}

func TestHostFormSeqIsMonotonic(t *testing.T) {
	a := &App{}
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
