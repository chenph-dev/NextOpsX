package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/einoagent/session"
	convstore "nextopsx/internal/store/conversations"
	hoststore "nextopsx/internal/store/hosts"
)

func aiTestApp() *app.App {
	a := &app.App{RightTab: 1}
	a.Selected = &hoststore.TreeNode{
		ID: "h1", Name: "web-01", NodeType: "host",
		Addr: "10.0.0.5", Port: 22, User: "ops", Protocol: "ssh",
	}
	a.AI.HostID = "h1"
	a.AI.SessionID = "s1"
	return a
}

func TestAIPanelUnconfigured(t *testing.T) {
	a := aiTestApp()
	tt := ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("尚未配置 AI") {
		t.Fatalf("unconfigured hint missing: %q", tt.Texts())
	}
	if !tt.HasText("未配置 AI") {
		t.Errorf("header badge missing: %q", tt.Texts())
	}
}

func TestAIPanelRendersMessagesAndMarkdown(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Model = "test-model"
	a.AI.Messages = []convstore.Message{
		{ID: "m1", Role: "user", Content: "检查磁盘"},
		{ID: "m2", Role: "assistant", Content: "# 磁盘\n\n**正常**，`/dev` 充足\n\n- 第一项\n- 第二项"},
	}
	tt := ui.NewTester(Shell(a), 1000, 700)
	for _, want := range []string{"检查磁盘", "磁盘", "正常", "/dev", "第一项", "第二项", "test-model"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestAIPanelCommandApproval(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.State = session.StateAwaitingApproval
	a.AI.PendingCommand = &app.CommandSuggestion{Command: "echo hi", Why: "say hi"}
	a.AI.CommandStatus = app.AIStatusPending
	a.AI.CommandDraft = "echo hi"

	tt := ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("$ echo hi") {
		t.Fatalf("command card missing: %q", tt.Texts())
	}
	if err := tt.Click("批准"); err != nil {
		t.Fatal(err)
	}
	if a.AI.CommandStatus != app.AIStatusApproved {
		t.Fatalf("status = %q, want approved", a.AI.CommandStatus)
	}
}

func TestAIPanelHighRiskNeedsConfirmation(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.State = session.StateAwaitingApproval
	a.AI.PendingCommand = &app.CommandSuggestion{Command: "rm -rf /tmp/x", AssessedRisk: "high"}
	a.AI.CommandStatus = app.AIStatusPending
	a.AI.CommandDraft = "rm -rf /tmp/x"

	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.Click("批准"); err != nil {
		t.Fatal(err)
	}
	if a.AI.CommandStatus != app.AIStatusPending || !a.AI.CommandConfirm {
		t.Fatalf("first click should only confirm: status=%q confirm=%v", a.AI.CommandStatus, a.AI.CommandConfirm)
	}
	tt.Frame()
	if err := tt.Click("再次确认批准"); err != nil {
		t.Fatal(err)
	}
	if a.AI.CommandStatus != app.AIStatusApproved {
		t.Fatalf("status = %q, want approved", a.AI.CommandStatus)
	}
}

func TestAIPanelPlanAndToolOutput(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.PendingPlan = &app.PlanSuggestion{Goal: "清理磁盘", Steps: []string{"查看用量", "删除缓存"}}
	a.AI.PlanStatus = app.AIStatusPending
	a.AI.Messages = []convstore.Message{{
		ID: "t1", Role: "tool", ToolName: "execute_command",
		Content:        "line-one\nline-two",
		ToolResult:     `{"command":"ls","exitCode":0,"status":"approved"}`,
		ApprovalStatus: "approved",
	}}
	tt := ui.NewTester(Shell(a), 1000, 900)
	for _, want := range []string{"执行计划", "清理磁盘", "查看用量", "删除缓存", "▸ $ ls · exit 0 · 2 行"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("▸ $ ls · exit 0 · 2 行"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("line-one") {
		t.Fatalf("tool output did not expand: %q", tt.Texts())
	}
	if err := tt.Click("批准"); err != nil {
		t.Fatal(err)
	}
	if a.AI.PlanStatus != app.AIStatusApproved {
		t.Fatalf("plan status = %q", a.AI.PlanStatus)
	}
}

func TestAIPanelTabsAndConfig(t *testing.T) {
	a := aiTestApp()
	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.Click("终端"); err != nil {
		t.Fatal(err)
	}
	if a.RightTab != 0 {
		t.Fatalf("tab = %d, want terminal", a.RightTab)
	}
	if !tt.HasText("连接") {
		t.Errorf("welcome pane missing: %q", tt.Texts())
	}
	if err := tt.Click("AI 对话"); err != nil {
		t.Fatal(err)
	}
	if a.RightTab != 1 {
		t.Fatalf("tab = %d, want AI", a.RightTab)
	}

	a.AI.ConfigOpen = true
	tt.Frame()
	if !tt.HasText("AI 配置") || !tt.HasText("协议") || !tt.HasText("模型") {
		t.Fatalf("config form missing: %q", tt.Texts())
	}
	a.AI.Provider = "openai"
	a.AI.Model = ""
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if a.AI.Err != "请填写协议与模型名称" {
		t.Fatalf("err = %q", a.AI.Err)
	}
}

func TestAIHistoryPairsToolCalls(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{
		{
			ID: "m1", Role: "assistant", Content: "先看一下目录",
			ToolCalls: `[{"id":"c1","name":"execute_command","arguments":"{\"command\":\"ls -la\",\"why\":\"inspect\"}"}]`,
		},
		{
			ID: "m2", Role: "tool", ToolCallID: "c1", ToolName: "execute_command",
			Content: "用户拒绝了该命令", ApprovalStatus: "rejected",
			ToolResult: `{"command":"ls -la","status":"rejected"}`,
		},
	}
	tt := ui.NewTester(Shell(a), 1000, 800)
	for _, want := range []string{"先看一下目录", "$ ls -la", "已拒绝"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("批准") {
		t.Errorf("history card must not offer approval: %q", tt.Texts())
	}
}

func TestAIHistoryPopover(t *testing.T) {
	a := aiTestApp()
	a.AI.Conversations = []convstore.Conversation{
		{ID: "s0", Title: "旧对话", HostID: "h1"},
		{ID: "s1", Title: "当前对话", HostID: "h1"},
	}
	tt := ui.NewTester(Shell(a), 1000, 700)
	if err := tt.Click("历史"); err != nil {
		t.Fatal(err)
	}
	if !a.AI.HistoryOpen {
		t.Fatal("history popover did not open")
	}
	if !tt.HasText("旧对话") || !tt.HasText("● 当前对话") {
		t.Fatalf("history list missing: %q", tt.Texts())
	}
	if err := tt.Click("旧对话"); err != nil {
		t.Fatal(err)
	}
	if a.AI.SessionID != "s0" {
		t.Fatalf("session = %q, want s0", a.AI.SessionID)
	}
}

func TestAIErrorsVisibleInEmptyAndConfigStates(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Err = "发送失败：资产凭据不可用"
	tt := ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("发送失败：资产凭据不可用") {
		t.Fatalf("error missing with empty messages: %q", tt.Texts())
	}

	a = aiTestApp()
	a.AI.Configured = false
	a.AI.Err = "打开 AI 会话失败：数据库不可用"
	tt = ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("打开 AI 会话失败：数据库不可用") {
		t.Fatalf("error missing when unconfigured: %q", tt.Texts())
	}

	a = aiTestApp()
	a.AI.Configured = false
	a.AI.ConfigOpen = true
	a.AI.Err = "请填写协议与模型名称"
	tt = ui.NewTester(Shell(a), 1000, 700)
	if !tt.HasText("请填写协议与模型名称") {
		t.Fatalf("config error missing: %q", tt.Texts())
	}
}
