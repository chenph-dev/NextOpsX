package app

import (
	"testing"

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
