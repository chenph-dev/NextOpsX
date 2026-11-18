package audit

import (
	"testing"

	"nextopsx/store"
	logsstore "nextopsx/store/logs"
)

func TestAuditFacade(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	data, err := store.Open()
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := data.GORM().DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	logs := logsstore.NewLogsStore(data)
	s := New(logs)

	if list, err := s.ListAICallLogs(10); err != nil || len(list) != 0 {
		t.Fatalf("初始审计列表 = %+v err=%v", list, err)
	}
	if err := logs.SaveLog(logsstore.CallLog{
		SessionID: "s1", Component: "model", Name: "llm", Provider: "OpenAI",
		TokensIn: 10, TokensOut: 20, TokensTotal: 30, DurationMS: 15, OK: true,
	}); err != nil {
		t.Fatalf("SaveLog: %v", err)
	}
	if err := logs.SaveLog(logsstore.CallLog{
		SessionID: "s1", Component: "tool", Name: "execute_command", Provider: "SSH",
		OK: false, Error: "boom",
	}); err != nil {
		t.Fatalf("SaveLog: %v", err)
	}

	list, err := s.ListAICallLogs(10)
	if err != nil || len(list) != 2 {
		t.Fatalf("审计列表 = %+v err=%v", list, err)
	}
	sum, err := s.AITokenSummary()
	if err != nil {
		t.Fatalf("AITokenSummary: %v", err)
	}
	if sum.TotalCalls != 2 || sum.ModelCalls != 1 || sum.ToolCalls != 1 || sum.TokensTotal != 30 {
		t.Fatalf("汇总 = %+v", sum)
	}
}
