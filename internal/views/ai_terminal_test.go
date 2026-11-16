package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	convstore "nextopsx/internal/store/conversations"
)

func TestTruncateCommand(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := truncateCommand(long, 72)
	if n := len([]rune(got)); n != 72 {
		t.Fatalf("length = %d, want 72: %q", n, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("missing ellipsis: %q", got)
	}
	if short := truncateCommand("ls -la", 72); short != "ls -la" {
		t.Fatalf("short command changed: %q", short)
	}
}

func TestANSISpansInvalidExtendedColor(t *testing.T) {
	spans := ansiSpans("\x1b[38;5;999mX\x1b[0mY")
	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.Text)
		if s.Color == (ui.Color{}) {
			t.Fatalf("invalid color made text transparent: %+v", s)
		}
	}
	if text.String() != "XY" {
		t.Fatalf("text = %q, want %q", text.String(), "XY")
	}
}

func TestLiveOutputTail(t *testing.T) {
	s := "old\n" + strings.Repeat("x", 200)
	got := liveOutputTail(s, 100)
	if !strings.HasPrefix(got, "…\n") {
		t.Fatalf("missing tail marker: %q", got)
	}
	if !strings.HasSuffix(got, "x") {
		t.Fatalf("did not keep the newest output: %q", got)
	}
	if short := liveOutputTail("short", 100); short != "short" {
		t.Fatalf("short output changed: %q", short)
	}
}

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
	tt := ui.NewTester(Shell(a), 1000, 900)
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
	tt := ui.NewTester(Shell(a), 1000, 800)
	head := "▸ $ " + truncateCommand(long, 72) + " · exit 0 · 1 行"
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

func TestANSISpansParsesColors(t *testing.T) {
	spans := ansiSpans("\x1b[31mred\x1b[0m plain\x1b[2J")
	var got strings.Builder
	for _, s := range spans {
		got.WriteString(s.Text)
	}
	if got.String() != "red plain" {
		t.Fatalf("text = %q, want %q", got.String(), "red plain")
	}
	red := false
	for _, s := range spans {
		if s.Text == "red" && s.Color == ansiBasic[1] {
			red = true
		}
	}
	if !red {
		t.Fatalf("red span not colored: %+v", spans)
	}
}

func TestANSISpansExtendedColors(t *testing.T) {
	spans := ansiSpans("\x1b[38;5;196mX\x1b[0m\x1b[38;2;1;2;3mY\x1b[0m")
	var got strings.Builder
	for _, s := range spans {
		got.WriteString(s.Text)
	}
	if got.String() != "XY" {
		t.Fatalf("text = %q", got.String())
	}
	var x, y bool
	for _, s := range spans {
		switch s.Text {
		case "X":
			x = s.Color == ui.RGB(255, 0, 0)
		case "Y":
			y = s.Color == ui.RGB(1, 2, 3)
		}
	}
	if !x || !y {
		t.Fatalf("extended colors not parsed: %+v", spans)
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
	tt := ui.NewTester(Shell(a), 1000, 900)
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
	tt := ui.NewTester(Shell(a), 1000, 900)
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
	tt := ui.NewTester(Shell(a), 1000, 900)
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
