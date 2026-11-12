package session

import (
	"testing"

	"nextopsx/internal/einoagent/testutil"
	"nextopsx/internal/sshexec"
	configstore "nextopsx/internal/store/config"
	convstore "nextopsx/internal/store/conversations"
	hoststore "nextopsx/internal/store/hosts"
)

func TestSendMessagePreflightFailureDoesNotPersistUserMessage(t *testing.T) {
	app := testutil.OpenTempStore(t)
	cfg := configstore.NewConfigStore(app)
	convs := convstore.NewConvStore(app)
	hostID, err := hoststore.NewHostsStore(app).SaveHost(hoststore.HostInput{
		Name: "h", Addr: "1.1.1.1", Port: 22, User: "u",
		AuthType: "password", Secret: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	mgr := NewSessionManager(app, cfg, func(string) sshexec.Exec { return nil }, nil, nil)
	sid, err := mgr.CreateSession(hostID, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.SendMessage(sid, "hello"); err == nil {
		t.Fatal("expected preflight failure without an executor")
	}
	msgs, err := convs.LoadMessages(sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Fatalf("persisted %d messages after preflight failure: %+v", len(msgs), msgs)
	}
}
