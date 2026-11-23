package hosts_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestWorkbenchSelectionPanel(t *testing.T) {
	cases := []struct {
		protocol string
		panel    string
	}{
		{"redis", "命令终端"},
		{"elasticsearch", "REST 命令终端"},
	}
	for _, tc := range cases {
		t.Run(tc.protocol, func(t *testing.T) {
			a := &workspace.State{}
			a.Selected = &hoststore.TreeNode{
				ID: "h1", Name: "svc", NodeType: "host", Protocol: tc.protocol,
				Addr: "10.0.0.9", Port: 6379, User: "root",
			}
			tt := ui.NewTester(views.Shell(a), 1300, 900)
			if !tt.HasText("双击资产连接") {
				t.Fatalf("selection hint missing: %q", tt.Texts())
			}
			if tt.HasText(tc.panel) {
				t.Fatalf("panel shown before connecting: %q", tt.Texts())
			}
			a.OpenAsset(a.Selected)
			tt.Frame()
			if !tt.HasText(tc.panel) {
				t.Fatalf("panel missing after connecting: %q", tt.Texts())
			}
		})
	}
}
