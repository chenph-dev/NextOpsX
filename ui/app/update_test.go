package app

import "testing"

func TestLoadAbout(t *testing.T) {
	a := &App{}
	a.LoadAbout()
	if a.Settings.About.Version == "" {
		t.Fatal("version is empty")
	}
	// Development/test builds have no update feed, so updates are disabled.
	if a.Settings.About.Enabled {
		t.Fatal("updates should be disabled in a development build")
	}
}

func TestCheckForUpdatesDisabled(t *testing.T) {
	a := &App{}
	a.LoadAbout()
	a.CheckForUpdates()
	if !a.Settings.About.Checked {
		t.Fatal("check should finish immediately when updates are disabled")
	}
	if a.Settings.About.Err == "" {
		t.Fatal("expected a disabled-build message")
	}
	if a.Settings.About.Available {
		t.Fatal("no update should be available")
	}
}

func TestUpdatePercent(t *testing.T) {
	a := &App{}
	a.Settings.About.Downloaded = 50
	a.Settings.About.Total = 200
	if got := a.UpdatePercent(); got != 25 {
		t.Fatalf("percent = %d, want 25", got)
	}
	a.Settings.About.Total = 0
	if got := a.UpdatePercent(); got != -1 {
		t.Fatalf("unknown total = %d, want -1", got)
	}
	a.Settings.About.Total = 10
	a.Settings.About.Downloaded = 20
	if got := a.UpdatePercent(); got != 100 {
		t.Fatalf("overflowing percent = %d, want 100", got)
	}
}
