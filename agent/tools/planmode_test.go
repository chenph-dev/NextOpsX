package tools

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/cloudwego/eino/schema"

	"nextopsx/agent/testutil"
	"nextopsx/domain/sshexec"
	convstore "nextopsx/store/conversations"
	hoststore "nextopsx/store/hosts"
)

func TestPlanGrant_NilSafe(t *testing.T) {
	var g *PlanGrant
	if g.Active() {
		t.Fatal("nil grant 不应处于激活状态")
	}
	g.Activate()
	g.Revoke()
}

func TestPlanGrant_Concurrent(t *testing.T) {
	g := NewPlanGrant()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if i%2 == 0 {
					g.Activate()
				} else {
					g.Revoke()
				}
				_ = g.Active()
			}
		}(i)
	}
	wg.Wait()
}

func TestSSHTool_PlanGrantAutoRunsOrdinaryCommand(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	ex := &testutil.FakeExec{Lines: []sshexec.Line{{Stream: "stdout", Text: "ok"}}}
	tool := NewSSHTool("s1", ex, rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	_, err := tool.InvokableRun(context.Background(),
		`{"command":"systemctl restart nginx","why":"重启服务"}`)
	if err != nil {
		t.Fatalf("计划授权下普通命令不应中断: %v", err)
	}
	if cmds := ex.Commands(); len(cmds) != 1 || cmds[0] != "systemctl restart nginx" {
		t.Fatalf("应执行一次命令，得到 %v", cmds)
	}
	events := rec.SnapshotEvents()
	if testutil.ContainsEvent(events, "ai:command") {
		t.Error("计划自动执行不应推送 ai:command")
	}
	if !testutil.ContainsEvent(events, "run:auto") {
		t.Error("计划自动执行应推送 run:auto")
	}
}

func TestSSHTool_PlanGrantStillInterruptsHighRisk(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	ex := &testutil.FakeExec{}
	tool := NewSSHTool("s1", ex, rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"command":"reboot","why":"重启"}`); err == nil {
		t.Fatal("高危命令即使有计划授权也应中断")
	}
	if len(ex.Commands()) != 0 {
		t.Error("高危命令在审批前不应执行")
	}
	if !testutil.ContainsEvent(rec.SnapshotEvents(), "ai:command") {
		t.Error("高危命令应推送 ai:command")
	}
}

func TestSSHTool_PlanGrantStillInterruptsSensitiveRead(t *testing.T) {
	var gotInfo commandInfo
	rec := &testutil.EmitRecorder{}
	emit := func(sid, event string, data any) {
		rec.Emit(sid, event, data)
		if event == "ai:command" {
			gotInfo, _ = data.(commandInfo)
		}
	}
	ex := &testutil.FakeExec{}
	tool := NewSSHTool("s1", ex, emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"command":"cat /etc/shadow","why":"检查"}`); err == nil {
		t.Fatal("敏感文件读取即使有计划授权也应中断")
	}
	if len(ex.Commands()) != 0 {
		t.Error("敏感命令在审批前不应执行")
	}
	if gotInfo.Risk != "high" {
		t.Errorf("敏感读取的展示风险应为 high，得到 %q", gotInfo.Risk)
	}
}

func TestSSHTool_PlanGrantRevokedInterrupts(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	ex := &testutil.FakeExec{}
	tool := NewSSHTool("s1", ex, rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	grant.Revoke()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"command":"systemctl restart nginx","why":"重启"}`); err == nil {
		t.Fatal("授权撤销后普通命令应恢复逐条审批")
	}
	if len(ex.Commands()) != 0 {
		t.Error("授权撤销后命令不应执行")
	}
}

func TestSSHTool_PlanGrantPersistsStatusPlan(t *testing.T) {
	app := testutil.OpenTempStore(t)
	hosts := hoststore.NewHostsStore(app)
	convs := convstore.NewConvStore(app)
	hostID, _ := hosts.SaveHost(hoststore.HostInput{
		Name: "h", Addr: "1.1.1.1", Port: 22, User: "u",
		AuthType: "password", Secret: "x",
	})
	sid, _ := convs.NewConversation(hostID, "t")

	ex := &testutil.FakeExec{Lines: []sshexec.Line{{Stream: "stdout", Text: "ok"}}}
	holder := NewToolCallHolder()
	holder.Add(&schema.ToolCall{ID: "call_plan", Function: schema.FunctionCall{Name: "execute_command"}})
	tool := NewSSHTool(sid, ex, (&testutil.EmitRecorder{}).Emit, convs, holder)
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"command":"ls -la","why":"查看"}`); err != nil {
		t.Fatalf("计划自动执行不应中断: %v", err)
	}
	msgs, _ := convs.LoadMessages(sid)
	if len(msgs) != 1 || msgs[0].ApprovalStatus != "plan" {
		t.Errorf("计划自动执行应落库 ApprovalStatus=plan，得到 %+v", msgs)
	}
	if !strings.Contains(msgs[0].ToolResult, `"status":"plan"`) {
		t.Errorf("toolMeta.Status 应为 plan: %s", msgs[0].ToolResult)
	}
}

func TestSQLTool_PlanGrantReadOnlyAutoRuns(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	tool := NewSQLTool("s1", nil, "sql", rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(), `{"sql":"SELECT 1"}`); err != nil {
		t.Fatalf("计划授权下只读查询不应中断: %v", err)
	}
	events := rec.SnapshotEvents()
	if testutil.ContainsEvent(events, "ai:command") {
		t.Error("只读查询不应推送 ai:command")
	}
	if !testutil.ContainsEvent(events, "run:auto") {
		t.Error("只读查询应推送 run:auto")
	}
}

func TestSQLTool_PlanGrantWriteStillInterrupts(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	tool := NewSQLTool("s1", nil, "sql", rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"sql":"UPDATE users SET name='x'"}`); err == nil {
		t.Fatal("计划授权下写 SQL 仍应中断")
	}
	if !testutil.ContainsEvent(rec.SnapshotEvents(), "ai:command") {
		t.Error("写 SQL 应推送 ai:command")
	}
}

func TestSQLTool_PlanGrantHighRiskInterrupts(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	tool := NewSQLTool("s1", nil, "sql", rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(),
		`{"sql":"DROP TABLE users"}`); err == nil {
		t.Fatal("高危 SQL 即使有计划授权也应中断")
	}
}

func TestSQLTool_PlanGrantRedisReadOnly(t *testing.T) {
	rec := &testutil.EmitRecorder{}
	tool := NewSQLTool("s1", nil, "redis", rec.Emit, nil, NewToolCallHolder())
	grant := NewPlanGrant()
	grant.Activate()
	tool.SetPlanGrant(grant)

	if _, err := tool.InvokableRun(context.Background(), `{"sql":"GET k"}`); err != nil {
		t.Fatalf("Redis 只读命令不应中断: %v", err)
	}
	if _, err := tool.InvokableRun(context.Background(), `{"sql":"SET k v"}`); err == nil {
		t.Fatal("Redis 写命令仍应中断")
	}
}
