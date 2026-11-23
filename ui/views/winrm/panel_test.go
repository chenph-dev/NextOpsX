package winrm_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func winRMTestApp() *workspace.State {
	a := &workspace.State{Roots: []hoststore.TreeNode{{
		ID: "w1", Name: "win-01", NodeType: "host", Protocol: "winrm",
		Addr: "10.0.0.20", Port: 5985, User: "Administrator",
	}}}
	a.Selected = &a.Roots[0]
	a.Tabs = []string{"w1"}
	a.WinRM = workspace.WinRMState{
		HostID: "w1", HostName: "win-01",
		Status: workspace.WinRMStatusReady,
		Info:   "PowerShell 5.1 · Administrator@WIN-01 · C:\\Users\\Administrator",
		Lines: []workspace.WinRMLine{
			{Stream: "cmd", Text: "Get-Process"},
			{Stream: "stdout", Text: "explorer"},
			{Stream: "stderr", Text: "warning"},
			{Stream: "error", Text: "boom"},
			{Stream: "exit", Text: "exit_code=0"},
		},
	}
	return a
}

func TestWinRMPanelRenders(t *testing.T) {
	a := winRMTestApp()
	tt := ui.NewTester(views.Shell(a), 1200, 800)
	for _, want := range []string{
		"PS> Get-Process", "explorer", "warning", "boom", "exit_code=0",
		"PS>", "远程桌面",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestWinRMPanelAutoFocus(t *testing.T) {
	a := winRMTestApp()
	a.WinRM.Lines = nil
	a.WinRM.Input = ""
	tt := ui.NewTester(views.Shell(a), 1200, 800)
	tt.Type("Get-Date")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !tt.HasText("PS> Get-Date") || !tt.HasText("后端未初始化") {
		t.Fatalf("input did not run: %q", tt.Texts())
	}
}

func TestWinRMPanelFailure(t *testing.T) {
	a := winRMTestApp()
	a.WinRM.Status = workspace.WinRMStatusFailed
	a.WinRM.Info = "拒绝访问"
	tt := ui.NewTester(views.Shell(a), 1200, 800)
	for _, want := range []string{"连接失败", "拒绝访问", "重新连接"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestWinRMPanelEmpty(t *testing.T) {
	a := winRMTestApp()
	a.WinRM.Lines = nil
	tt := ui.NewTester(views.Shell(a), 1200, 800)
	if !tt.HasText("输入 PowerShell 命令开始") {
		t.Fatalf("empty hint missing: %q", tt.Texts())
	}
}
