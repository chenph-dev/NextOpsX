package app

import (
	"testing"

	hoststore "nextopsx/internal/store/hosts"
)

func TestAuditWithoutBackend(t *testing.T) {
	a := &App{}
	a.LoadAudit()
	if a.Audit.Loading || a.Audit.Err != "后端未初始化" {
		t.Fatalf("audit state = %+v", a.Audit)
	}
}

func TestApprovalPolicyWithoutBackend(t *testing.T) {
	a := &App{}
	a.LoadApprovalPolicy()
	if a.Policy.Loading || a.Policy.Err != "后端未初始化" {
		t.Fatalf("policy state = %+v", a.Policy)
	}
	a.Policy.Err = ""
	a.SaveApprovalPolicy()
	if a.Policy.Err != "后端未初始化" {
		t.Fatalf("save err = %q", a.Policy.Err)
	}
}

func TestOpenSettingsLoadsNewSections(t *testing.T) {
	a := &App{}
	a.OpenSettings(SettingsAudit)
	if a.View != ViewSettings || a.Settings.Tab != SettingsAudit {
		t.Fatalf("view = %v tab = %v", a.View, a.Settings.Tab)
	}
	if a.Audit.Err != "后端未初始化" {
		t.Fatalf("audit not loaded: %+v", a.Audit)
	}
	a.OpenSettings(SettingsPolicy)
	if a.Policy.Err != "后端未初始化" {
		t.Fatalf("policy not loaded: %+v", a.Policy)
	}
}

func TestNewAssetFormsLeaveSettings(t *testing.T) {
	a := &App{View: ViewSettings}
	a.OpenNewHostForm()
	if a.View != ViewAssets || !a.Hosts.Form.Open {
		t.Fatalf("host form: view=%v open=%v", a.View, a.Hosts.Form.Open)
	}
	a.View = ViewSettings
	a.OpenNewFolderForm()
	if a.View != ViewAssets || !a.Hosts.Form.Open || !a.Hosts.Form.IsFolder {
		t.Fatalf("folder form: view=%v form=%+v", a.View, a.Hosts.Form)
	}
}

func TestSetThemeNormalizes(t *testing.T) {
	a := &App{}
	a.SetTheme("light")
	if a.Settings.Theme != "light" {
		t.Fatalf("theme = %q", a.Settings.Theme)
	}
	a.SetTheme("neon")
	if a.Settings.Theme != "system" {
		t.Fatalf("invalid theme = %q", a.Settings.Theme)
	}
}

func TestParseWhitelistKeepsSpaces(t *testing.T) {
	got := parseWhitelist("df -h\nfree -m,uptime\n\ndf -h\r\n")
	want := []string{"df -h", "free -m", "uptime"}
	if len(got) != len(want) {
		t.Fatalf("whitelist = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("whitelist = %q, want %q", got, want)
		}
	}
}

func TestOpenRDPWithoutBackend(t *testing.T) {
	a := &App{}
	a.Selected = &hoststore.TreeNode{NodeType: "host", Protocol: "winrm"}
	a.OpenRDP() // must not panic or touch state without a backend
	if a.ErrMsg != "" {
		t.Fatalf("err = %q", a.ErrMsg)
	}
}
