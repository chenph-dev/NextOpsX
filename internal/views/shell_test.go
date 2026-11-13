package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/backend"
	"nextopsx/internal/store"
	hoststore "nextopsx/internal/store/hosts"
)

// The view runs without a window in tests, which click and type as a user
// would. Each test gets its own data directory, so no real app data is
// touched and tests do not share a database.
func TestHostTreeSelects(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")

	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	hosts := hoststore.NewHostsStore(db)
	folder, err := hosts.CreateFolder("prod", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hosts.SaveHost(hoststore.HostInput{
		Name: "web-01", ParentID: folder, Addr: "10.0.0.5", Port: 22,
		User: "ops", AuthType: "password", Secret: "x", Protocol: "ssh",
	}); err != nil {
		t.Fatal(err)
	}
	closeSQL(t, db)

	b, err := backend.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(b.Close)
	a, err := app.New(b)
	if err != nil {
		t.Fatal(err)
	}

	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.Click("web-01"); err != nil {
		t.Fatal(err)
	}
	if a.Selected == nil || a.Selected.Name != "web-01" {
		t.Fatalf("selected = %+v, texts = %q", a.Selected, tt.Texts())
	}
	if !tt.HasText("ops@10.0.0.5:22") {
		t.Errorf("welcome pane did not show the asset: %q", tt.Texts())
	}
	if !tt.HasText("1") {
		t.Errorf("host count missing: %q", tt.Texts())
	}
}

func TestWinRMHostShowsRDPEntry(t *testing.T) {
	a := &app.App{}
	a.Selected = &hoststore.TreeNode{
		ID: "h1", Name: "win-01", NodeType: "host",
		Addr: "10.0.0.9", Port: 5985, User: "Administrator", Protocol: "winrm",
	}
	tt := ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("远程桌面") {
		t.Fatalf("RDP entry missing for WinRM asset: %q", tt.Texts())
	}

	a.Selected.Protocol = "ssh"
	tt = ui.NewTester(Shell(a), 1000, 700)
	if tt.HasText("远程桌面") {
		t.Fatalf("RDP entry should be hidden for SSH asset: %q", tt.Texts())
	}
}

func closeSQL(t *testing.T, db *store.DB) {
	t.Helper()
	sqlDB, err := db.GORM().DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}
