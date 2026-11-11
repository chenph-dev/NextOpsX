package backend

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/sshexec"
)

// TestSSHTerminal is the Phase 1 end-to-end check: a real SSH session is
// adapted to terminal.Conn and rendered by the Ghostty-backed terminal in
// a headless view. It is opt-in because it needs a reachable host and key.
//
//	NEXTOPSX_SSH_TEST_HOST=192.0.2.10 \
//	NEXTOPSX_SSH_TEST_USER=ops \
//	NEXTOPSX_SSH_TEST_KEY=$HOME/.ssh/id_ed25519 \
//	go test -run TestSSHTerminal -v .
func TestSSHTerminal(t *testing.T) {
	host := os.Getenv("NEXTOPSX_SSH_TEST_HOST")
	user := os.Getenv("NEXTOPSX_SSH_TEST_USER")
	keyPath := os.Getenv("NEXTOPSX_SSH_TEST_KEY")
	if host == "" || user == "" || keyPath == "" {
		t.Skip("set NEXTOPSX_SSH_TEST_HOST, NEXTOPSX_SSH_TEST_USER and NEXTOPSX_SSH_TEST_KEY")
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	sess, err := sshexec.OpenSession(ctx, sshexec.Host{
		Addr: host, Port: 22, User: user,
		AuthType: "privatekey", Secret: string(key),
	}, 80, 24)
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	defer sess.Close()

	conn := newSSHConn(sess)
	term, err := terminal.New(terminal.Options{Conn: conn})
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	defer term.Close()

	tt := ui.NewTester(func(c *ui.Context) {
		terminal.View(c, term).Fill().AutoFocus()
	}, 800, 500)
	tt.Type("echo opsmate-ok")
	tt.Key(0, ui.KeyEnter)

	deadline := time.Now().Add(10 * time.Second)
	for !strings.Contains(term.Text(), "opsmate-ok") {
		if time.Now().After(deadline) {
			t.Fatalf("terminal did not show the SSH output: %q", term.Text())
		}
		time.Sleep(20 * time.Millisecond)
		tt.Frame()
	}
}
