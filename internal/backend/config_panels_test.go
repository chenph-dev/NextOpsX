package backend

import (
	"testing"

	configstore "nextopsx/internal/store/config"
	logsstore "nextopsx/internal/store/logs"
)

func TestApprovalPolicyFacade(t *testing.T) {
	b := openTestBackend(t)

	p, err := b.GetApprovalPolicy()
	if err != nil {
		t.Fatalf("GetApprovalPolicy: %v", err)
	}
	if !p.EnableAuto {
		t.Fatalf("默认应开启自动放行: %+v", p)
	}

	if err := b.SaveApprovalPolicy(configstore.ApprovalPolicy{
		EnableAuto:   false,
		ReadOnlyList: []string{"df -h", "free -m"},
	}); err != nil {
		t.Fatalf("SaveApprovalPolicy: %v", err)
	}
	got, err := b.GetApprovalPolicy()
	if err != nil {
		t.Fatalf("GetApprovalPolicy: %v", err)
	}
	if got.EnableAuto || len(got.ReadOnlyList) != 2 {
		t.Fatalf("保存后不匹配: %+v", got)
	}
}

func TestAuditFacade(t *testing.T) {
	b := openTestBackend(t)

	if list, err := b.ListAICallLogs(10); err != nil || len(list) != 0 {
		t.Fatalf("初始审计列表 = %+v err=%v", list, err)
	}
	if err := b.logs.SaveLog(logsstore.CallLog{
		SessionID: "s1", Component: "model", Name: "llm", Provider: "OpenAI",
		TokensIn: 10, TokensOut: 20, TokensTotal: 30, DurationMS: 15, OK: true,
	}); err != nil {
		t.Fatalf("SaveLog: %v", err)
	}
	if err := b.logs.SaveLog(logsstore.CallLog{
		SessionID: "s1", Component: "tool", Name: "execute_command", Provider: "SSH",
		OK: false, Error: "boom",
	}); err != nil {
		t.Fatalf("SaveLog: %v", err)
	}

	list, err := b.ListAICallLogs(10)
	if err != nil || len(list) != 2 {
		t.Fatalf("审计列表 = %+v err=%v", list, err)
	}
	sum, err := b.AITokenSummary()
	if err != nil {
		t.Fatalf("AITokenSummary: %v", err)
	}
	if sum.TotalCalls != 2 || sum.ModelCalls != 1 || sum.ToolCalls != 1 || sum.TokensTotal != 30 {
		t.Fatalf("汇总 = %+v", sum)
	}
}
