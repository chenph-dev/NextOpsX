package workspace

import (
	"testing"

	hoststore "nextopsx/store/hosts"
)

func TestWinRMRunCommandWithoutBackend(t *testing.T) {
	a := &State{}
	a.WinRM = WinRMState{HostID: "w1", HostName: "win"}
	a.WinRM.Input = "Get-Process"
	a.WinRMRunCommand()

	if a.WinRM.Running || a.WinRM.Input != "" {
		t.Fatalf("state = %+v", a.WinRM)
	}
	if len(a.WinRM.Lines) != 2 {
		t.Fatalf("lines = %+v", a.WinRM.Lines)
	}
	if a.WinRM.Lines[0].Stream != "cmd" || a.WinRM.Lines[0].Text != "Get-Process" {
		t.Fatalf("command line = %+v", a.WinRM.Lines[0])
	}
	if a.WinRM.Lines[1].Stream != "error" {
		t.Fatalf("backend error line = %+v", a.WinRM.Lines[1])
	}
}

func TestWinRMProbeWithoutBackend(t *testing.T) {
	a := &State{}
	a.WinRM = WinRMState{HostID: "w1", HostName: "win"}
	a.WinRMProbe()
	if a.WinRM.Status != WinRMStatusFailed || a.WinRM.Info == "" {
		t.Fatalf("probe state = %+v", a.WinRM)
	}
}

func TestWinRMTabsPreserveConsole(t *testing.T) {
	a := &State{Roots: []hoststore.TreeNode{
		{ID: "wa", Name: "win-a", NodeType: "host", Protocol: "winrm"},
		{ID: "wb", Name: "win-b", NodeType: "host", Protocol: "winrm"},
	}}

	a.OpenAsset(&a.Roots[0])
	if !a.IsAssetTabOpen("wa") || a.WinRM.HostID != "wa" {
		t.Fatalf("open wa: tabs=%v state=%+v", a.Tabs, a.WinRM)
	}
	a.WinRM.Input = "dir"
	a.WinRM.Lines = append(a.WinRM.Lines, WinRMLine{Stream: "stdout", Text: "file.txt"})

	a.OpenAsset(&a.Roots[1])
	if a.WinRM.HostID != "wb" {
		t.Fatalf("open wb: %+v", a.WinRM)
	}

	a.ActivateAssetTab("wa")
	if a.WinRM.HostID != "wa" || a.WinRM.Input != "dir" || len(a.WinRM.Lines) != 1 {
		t.Fatalf("restore wa: %+v", a.WinRM)
	}
}
