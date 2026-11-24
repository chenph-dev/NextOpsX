package session

import (
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/schema"

	"nextopsx/agent/testutil"
)

// planToolCallResponse 构造一条带 create_plan tool_call 的 assistant 回复。
func planToolCallResponse(goal string, steps []string) *schema.Message {
	args, _ := json.Marshal(map[string]any{"goal": goal, "steps": steps})
	return schema.AssistantMessage("", []schema.ToolCall{{
		ID: "call_plan", Type: "function",
		Function: schema.FunctionCall{Name: "create_plan", Arguments: string(args)},
	}})
}

func TestSessionManager_PlanApprovalAutoRunsCommands(t *testing.T) {
	f := newSessionFixture(t, []*schema.Message{
		planToolCallResponse("重启服务", []string{"重启 nginx"}),
		testutil.ToolCallResponse("systemctl restart nginx"),
		schema.AssistantMessage("已完成", nil),
	})
	sid, _ := f.mgr.EnsureSession(f.hostID)
	s, _ := f.mgr.sessionFor(sid)

	if err := f.mgr.SendMessage(sid, "重启 nginx"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	f.waitState(sid, StateAwaitingApproval)
	s.mu.Lock()
	approvalType := s.approvalType
	s.mu.Unlock()
	if approvalType != "plan" {
		t.Fatalf("审批类型 = %q, want plan", approvalType)
	}

	if err := f.mgr.ApprovePlan(sid); err != nil {
		t.Fatalf("ApprovePlan: %v", err)
	}
	f.waitState(sid, StateIdle)

	if cmds := f.ex.Commands(); len(cmds) != 1 || cmds[0] != "systemctl restart nginx" {
		t.Fatalf("计划批准后应自动执行命令，得到 %v", cmds)
	}
	if s.planGrant.Active() {
		t.Error("本轮结束后计划授权应撤销")
	}

	msgs, _ := f.convs.LoadMessages(sid)
	statuses := map[string]string{}
	for _, m := range msgs {
		if m.Role == "tool" {
			statuses[m.ToolName] = m.ApprovalStatus
		}
	}
	if statuses["create_plan"] != "approved" {
		t.Errorf("计划消息状态 = %q, want approved", statuses["create_plan"])
	}
	if statuses["execute_command"] != "plan" {
		t.Errorf("命令消息状态 = %q, want plan", statuses["execute_command"])
	}
}

func TestSessionManager_PlanGrantSurvivesHighRiskApproval(t *testing.T) {
	f := newSessionFixture(t, []*schema.Message{
		planToolCallResponse("重启并恢复服务", []string{"重启主机", "启动 nginx"}),
		testutil.ToolCallResponse("reboot"),
		testutil.ToolCallResponse("systemctl restart nginx"),
		schema.AssistantMessage("已完成", nil),
	})
	sid, _ := f.mgr.EnsureSession(f.hostID)
	s, _ := f.mgr.sessionFor(sid)

	if err := f.mgr.SendMessage(sid, "重启并恢复"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	f.waitState(sid, StateAwaitingApproval)
	if err := f.mgr.ApprovePlan(sid); err != nil {
		t.Fatalf("ApprovePlan: %v", err)
	}
	f.waitState(sid, StateAwaitingApproval)

	s.mu.Lock()
	approvalType := s.approvalType
	s.mu.Unlock()
	if approvalType != "command" {
		t.Fatalf("高危命令应进入命令审批，得到 %q", approvalType)
	}
	if !s.planGrant.Active() {
		t.Fatal("高危命令中断不应撤销计划授权")
	}

	if err := f.mgr.ApproveCommand(sid, "reboot"); err != nil {
		t.Fatalf("ApproveCommand: %v", err)
	}
	f.waitState(sid, StateIdle)

	cmds := f.ex.Commands()
	if len(cmds) != 2 || cmds[0] != "reboot" || cmds[1] != "systemctl restart nginx" {
		t.Fatalf("命令执行顺序错误: %v", cmds)
	}
	if s.planGrant.Active() {
		t.Error("本轮结束后计划授权应撤销")
	}
}

func TestSessionManager_RejectPlanRevokesGrant(t *testing.T) {
	f := newSessionFixture(t, []*schema.Message{
		planToolCallResponse("清理", []string{"删除缓存"}),
		schema.AssistantMessage("换个方案", nil),
	})
	sid, _ := f.mgr.EnsureSession(f.hostID)
	s, _ := f.mgr.sessionFor(sid)

	if err := f.mgr.SendMessage(sid, "清理缓存"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	f.waitState(sid, StateAwaitingApproval)
	s.planGrant.Activate()

	if err := f.mgr.RejectPlan(sid); err != nil {
		t.Fatalf("RejectPlan: %v", err)
	}
	f.waitState(sid, StateIdle)

	if s.planGrant.Active() {
		t.Error("拒绝计划后授权应撤销")
	}
	if cmds := f.ex.Commands(); len(cmds) != 0 {
		t.Errorf("拒绝计划后不应执行命令: %v", cmds)
	}
}

func TestSessionManager_PlanGrantClearedOnLifecycle(t *testing.T) {
	f := newSessionFixture(t, []*schema.Message{schema.AssistantMessage("hi", nil)})
	sid, _ := f.mgr.EnsureSession(f.hostID)
	s, _ := f.mgr.sessionFor(sid)

	s.planGrant.Activate()
	if err := f.mgr.SendMessage(sid, "hi"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if s.planGrant.Active() {
		t.Error("新消息应撤销上一轮的计划授权")
	}
	f.waitState(sid, StateIdle)

	s.planGrant.Activate()
	if err := f.mgr.ClearMessages(sid); err != nil {
		t.Fatalf("ClearMessages: %v", err)
	}
	if s.planGrant.Active() {
		t.Error("清空会话应撤销计划授权")
	}

	s.planGrant.Activate()
	if err := f.mgr.DeleteSession(sid); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if s.planGrant.Active() {
		t.Error("删除会话应撤销计划授权")
	}
}
