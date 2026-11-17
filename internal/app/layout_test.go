package app

import (
	"testing"

	"nextopsx/internal/backend"
	configstore "nextopsx/internal/store/config"
	hoststore "nextopsx/internal/store/hosts"
)

func TestDrawerSizesClamp(t *testing.T) {
	a := &App{}
	if got := a.AIDrawerHeight(); got != defaultAIHeight {
		t.Fatalf("default AI height = %v", got)
	}
	if got := a.SFTPDrawerWidth(); got != defaultSFTPWidth {
		t.Fatalf("default SFTP width = %v", got)
	}

	a.SetAIDrawerHeight(1000)
	if a.AIDrawerHeight() != maxAIHeight {
		t.Fatalf("AI height = %v, want %v", a.AIDrawerHeight(), maxAIHeight)
	}
	a.SetAIDrawerHeight(10)
	if a.AIDrawerHeight() != minAIHeight {
		t.Fatalf("AI height = %v, want %v", a.AIDrawerHeight(), minAIHeight)
	}
	a.SetSFTPDrawerWidth(1000)
	if a.SFTPDrawerWidth() != maxSFTPWidth {
		t.Fatalf("SFTP width = %v, want %v", a.SFTPDrawerWidth(), maxSFTPWidth)
	}
	a.SetSFTPDrawerWidth(10)
	if a.SFTPDrawerWidth() != minSFTPWidth {
		t.Fatalf("SFTP width = %v, want %v", a.SFTPDrawerWidth(), minSFTPWidth)
	}
}

func TestDrawerTogglesFollowProtocol(t *testing.T) {
	a := &App{}
	a.ToggleAIDrawer()
	if a.Layout.AIOpen {
		t.Fatal("no selected host should not open the ops drawer")
	}

	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "mysql"}
	a.ToggleAIDrawer()
	if !a.Layout.AIOpen {
		t.Fatal("ops drawer should open for a database host")
	}
	a.ToggleSFTPDrawer()
	if a.Layout.SFTPOpen {
		t.Fatal("SFTP should stay closed for a non-SSH host")
	}

	a.Selected = &hoststore.TreeNode{ID: "h2", NodeType: "host", Protocol: "ssh"}
	a.SFTP = SFTPState{HostID: "h2"}
	a.ToggleSFTPDrawer()
	if !a.Layout.SFTPOpen {
		t.Fatal("SFTP should open for an SSH host")
	}
	a.ToggleSFTPDrawer()
	if a.Layout.SFTPOpen {
		t.Fatal("SFTP should close")
	}
}

func TestNewLoadsWorkspaceLayout(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := backend.New()
	if err != nil {
		t.Fatalf("backend.New: %v", err)
	}
	t.Cleanup(b.Close)

	if err := b.SaveUISettings(configstore.UISettings{
		Theme: "dark", Scale: 110, Language: "zh-CN",
		AIOpen: true, AIHeight: 400, SFTPOpen: true, SFTPWidth: 380,
	}); err != nil {
		t.Fatalf("SaveUISettings: %v", err)
	}
	a, err := New(b)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !a.Layout.AIOpen || a.Layout.AIHeight != 400 ||
		!a.Layout.SFTPOpen || a.Layout.SFTPWidth != 380 {
		t.Fatalf("layout = %+v", a.Layout)
	}
	if a.Settings.Theme != "dark" || a.Settings.Scale != 110 {
		t.Fatalf("settings = %+v", a.Settings)
	}
}

func TestLayoutChangesPersist(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := backend.New()
	if err != nil {
		t.Fatalf("backend.New: %v", err)
	}
	t.Cleanup(b.Close)
	a, err := New(b)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u := &chanUpdater{ch: make(chan func(), 16)}
	a.Attach(u)
	drain := func() {
		pumpUntil(t, u, func() bool { return !a.Settings.UISaving })
	}
	a.Selected = &hoststore.TreeNode{ID: "h1", NodeType: "host", Protocol: "ssh"}
	a.AI.HostID = "h1"
	a.SFTP.HostID = "h1"

	a.OpenAIDrawer()
	a.writes.Wait()
	drain()
	got, _ := b.GetUISettings()
	if !got.AIOpen {
		t.Fatalf("AI open state was not saved: %+v", got)
	}

	a.SetAIDrawerHeight(444)
	a.SaveLayout()
	a.writes.Wait()
	drain()
	got, _ = b.GetUISettings()
	if got.AIHeight != 444 {
		t.Fatalf("AI height was not saved: %+v", got)
	}

	a.CloseAIDrawer()
	a.writes.Wait()
	drain()
	got, _ = b.GetUISettings()
	if got.AIOpen {
		t.Fatalf("AI close state was not saved: %+v", got)
	}

	a.OpenSFTPDrawer()
	a.writes.Wait()
	drain()
	got, _ = b.GetUISettings()
	if !got.SFTPOpen {
		t.Fatalf("SFTP open state was not saved: %+v", got)
	}

	a.SetSFTPDrawerWidth(399)
	a.SaveLayout()
	a.writes.Wait()
	drain()
	got, _ = b.GetUISettings()
	if got.SFTPWidth != 399 {
		t.Fatalf("SFTP width was not saved: %+v", got)
	}
}

func TestSelectionRestoresOnStartup(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := backend.New()
	if err != nil {
		t.Fatalf("backend.New: %v", err)
	}
	t.Cleanup(b.Close)
	folderID, err := b.CreateFolder("ops", "")
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}

	a, err := New(b)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a.Select(&hoststore.TreeNode{ID: folderID, Name: "ops", NodeType: "folder"})
	a.writes.Wait()
	got, _ := b.GetUISettings()
	if got.LastAssetID != folderID {
		t.Fatalf("last asset = %q, want %q", got.LastAssetID, folderID)
	}

	restored, err := New(b)
	if err != nil {
		t.Fatalf("New restore: %v", err)
	}
	restored.RestoreWorkspace()
	if restored.Selected == nil || restored.Selected.ID != folderID {
		t.Fatalf("restored selection = %+v", restored.Selected)
	}
}
