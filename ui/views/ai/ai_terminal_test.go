package ai_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	convstore "nextopsx/store/conversations"
	"nextopsx/ui/ansi"
	"nextopsx/ui/views"
)

func TestAINonCommandToolResultNotHidden(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{
		{
			ID: "m1", Role: "assistant",
			ToolCalls: `[{"id":"c1","name":"load_skill","arguments":"{\"skillName\":\"disk\"}"}]`,
		},
		{
			ID: "m2", Role: "tool", ToolCallID: "c1", ToolName: "load_skill",
			Content: "# 磁盘技能\n\n检查磁盘", ApprovalStatus: "load",
		},
	}
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	for _, want := range []string{"load_skill", "磁盘技能", "检查磁盘"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestAITerminalHeaderFitsAndToolbarBelow(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	long := strings.Repeat("x", 200)
	a.AI.Messages = []convstore.Message{{
		ID: "t1", Role: "tool", ToolName: "execute_command",
		Content:        "ok",
		ToolResult:     fmt.Sprintf(`{"command":%q,"exitCode":0,"status":"approved"}`, long),
		ApprovalStatus: "approved",
	}}
	tt := ui.NewTester(views.Shell(a), 1000, 800)
	head := "▸ $ " + ansi.Truncate(long, 72) + " · exit 0 · 1 行"
	hr, ok := tt.Find(head)
	if !ok {
		t.Fatalf("header missing: %q", tt.Texts())
	}
	if hr.X+hr.W > 1000 {
		t.Fatalf("header overflows the window: %+v", hr)
	}
	cr, ok := tt.Find("复制")
	if !ok {
		t.Fatalf("copy toolbar missing: %q", tt.Texts())
	}
	if cr.Y <= hr.Y {
		t.Fatalf("toolbar is not below the bubble: header=%+v copy=%+v", hr, cr)
	}
}

func TestAIPanelTerminalBubbleParsesOutput(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{{
		ID: "t1", Role: "tool", ToolName: "execute_command",
		Content:        "\x1b[32mok\x1b[0m\nsecond",
		ToolResult:     `{"command":"echo ok","exitCode":0,"status":"approved"}`,
		ApprovalStatus: "approved",
	}}
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	const head = "▸ $ echo ok · exit 0 · 2 行"
	if !tt.HasText(head) {
		t.Fatalf("terminal header missing: %q", tt.Texts())
	}
	if err := tt.Click(head); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("ok") || !tt.HasText("second") {
		t.Fatalf("terminal output missing: %q", tt.Texts())
	}
}

func TestAIPanelToolTextUsesMarkdown(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{{
		ID: "t1", Role: "tool", ToolName: "read_file",
		Content: "# 报告\n\n**正常**",
	}}
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	for _, want := range []string{"read_file", "报告", "正常"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestAITerminalBubblePairsWithAssistantCall(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{
		{
			ID: "m1", Role: "assistant", Content: "检查目录",
			ToolCalls: `[{"id":"c1","name":"execute_command","arguments":"{\"command\":\"ls\"}"}]`,
		},
		{
			ID: "m2", Role: "tool", ToolCallID: "c1", ToolName: "execute_command",
			Content:        "a.txt",
			ToolResult:     `{"command":"ls","exitCode":0,"status":"approved"}`,
			ApprovalStatus: "approved",
		},
	}
	tt := ui.NewTester(views.Shell(a), 1000, 900)
	for _, want := range []string{"检查目录", "$ ls"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("批准") {
		t.Fatalf("executed command still shows an approval card: %q", tt.Texts())
	}
	const head = "▸ $ ls · exit 0 · 1 行"
	if err := tt.Click(head); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	n := 0
	for _, text := range tt.Texts() {
		if text == "a.txt" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("output rendered %d times, want once: %q", n, tt.Texts())
	}
}
