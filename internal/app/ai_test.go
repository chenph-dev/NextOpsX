package app

import (
	"testing"

	"nextopsx/internal/einoagent/session"
)

func TestAIEventsUpdateState(t *testing.T) {
	a := &App{}
	a.AI.SessionID = "s1"
	a.AI.Configured = true
	a.AI.State = session.StateThinking

	// Events for other sessions are ignored.
	a.HandleAIEvent("other", "ai:text", map[string]any{"delta": "no"})
	if a.AI.Streaming != "" {
		t.Fatalf("event leaked across sessions: %q", a.AI.Streaming)
	}

	a.HandleAIEvent("s1", "ai:text", map[string]any{"delta": "Hel"})
	a.HandleAIEvent("s1", "ai:text", map[string]any{"delta": "lo"})
	if a.AI.Streaming != "Hello" {
		t.Fatalf("streaming = %q, want Hello", a.AI.Streaming)
	}

	a.HandleAIEvent("s1", "ai:command", map[string]any{
		"command": "df -h", "why": "check disk", "assessedRisk": "low",
	})
	if a.AI.PendingCommand == nil || a.AI.PendingCommand.Command != "df -h" {
		t.Fatalf("pending command = %+v", a.AI.PendingCommand)
	}
	if a.AI.CommandStatus != AIStatusPending || a.AI.CommandDraft != "df -h" {
		t.Fatalf("command status/draft = %q/%q", a.AI.CommandStatus, a.AI.CommandDraft)
	}
	if a.AI.Streaming != "" {
		t.Fatalf("streaming should reset when a command is proposed: %q", a.AI.Streaming)
	}

	a.HandleAIEvent("s1", "session:state", session.StateAwaitingApproval)
	if a.AI.State != session.StateAwaitingApproval {
		t.Fatalf("state = %q", a.AI.State)
	}

	a.HandleAIEvent("s1", "run:start", map[string]any{"command": "df -h"})
	a.HandleAIEvent("s1", "run:output", map[string]any{"delta": "/dev 10G"})
	if a.AI.RunningCommand != "df -h" || a.AI.RunOutput != "/dev 10G" {
		t.Fatalf("running = %q output = %q", a.AI.RunningCommand, a.AI.RunOutput)
	}

	a.HandleAIEvent("s1", "ai:plan", PlanSuggestion{Goal: "inspect", Steps: []string{"a", "b"}})
	if a.AI.PendingPlan == nil || a.AI.PendingPlan.Goal != "inspect" || a.AI.PlanStatus != AIStatusPending {
		t.Fatalf("pending plan = %+v status = %q", a.AI.PendingPlan, a.AI.PlanStatus)
	}

	a.HandleAIEvent("s1", "ai:error", map[string]any{"message": "cancelled", "cancelled": true})
	if a.AI.Err != "cancelled" || !a.AI.ErrCancelled {
		t.Fatalf("error = %q cancelled = %v", a.AI.Err, a.AI.ErrCancelled)
	}

	a.HandleAIEvent("s1", "session:state", session.StateIdle)
	if a.AI.State != session.StateIdle {
		t.Fatalf("state = %q, want Idle", a.AI.State)
	}
}

func TestAIApprovalActions(t *testing.T) {
	a := &App{}
	a.AI.SessionID = "s1"
	a.AI.PendingCommand = &CommandSuggestion{Command: "ls"}
	a.AI.CommandStatus = AIStatusPending
	a.AI.CommandDraft = "ls -la"
	a.ApproveAICommand()
	if a.AI.CommandStatus != AIStatusApproved {
		t.Fatalf("command status = %q", a.AI.CommandStatus)
	}

	a.AI.PendingCommand = &CommandSuggestion{Command: "rm -rf /tmp/x"}
	a.AI.CommandStatus = AIStatusPending
	a.RejectAICommand()
	if a.AI.CommandStatus != AIStatusRejected {
		t.Fatalf("command status = %q", a.AI.CommandStatus)
	}

	a.AI.PendingPlan = &PlanSuggestion{Goal: "g", Steps: []string{"s"}}
	a.AI.PlanStatus = AIStatusPending
	a.ApproveAIPlan()
	if a.AI.PlanStatus != AIStatusApproved {
		t.Fatalf("plan status = %q", a.AI.PlanStatus)
	}

	a.AI.PendingPlan = &PlanSuggestion{Goal: "g2"}
	a.AI.PlanStatus = AIStatusPending
	a.RejectAIPlan()
	if a.AI.PlanStatus != AIStatusRejected {
		t.Fatalf("plan status = %q", a.AI.PlanStatus)
	}
}

func TestSendMessageAndSlashCommands(t *testing.T) {
	a := &App{}
	a.AI.HostID = "h1"
	a.AI.SessionID = "s1"
	a.AI.Configured = true
	a.AI.Input = "hello"
	a.SendAIMessage()
	if a.AI.Input != "" {
		t.Fatalf("input = %q, want cleared", a.AI.Input)
	}
	if len(a.AI.Messages) != 1 || a.AI.Messages[0].Role != "user" || a.AI.Messages[0].Content != "hello" {
		t.Fatalf("messages = %+v", a.AI.Messages)
	}

	a.AI.Input = "/clear"
	a.SendAIMessage()
	if a.AI.Input != "" {
		t.Fatalf("slash command left input = %q", a.AI.Input)
	}
	a.AI.Input = "/new"
	a.SendAIMessage()
	if a.AI.Input != "" {
		t.Fatalf("slash command left input = %q", a.AI.Input)
	}
}

func TestAttachAIWithoutBackend(t *testing.T) {
	a := &App{}
	a.attachAI("h1")
	if a.AI.HostID != "h1" || a.AI.Loading {
		t.Fatalf("host = %q loading = %v", a.AI.HostID, a.AI.Loading)
	}
	a.Selected = nil
	if got := a.SelectedProtocol(); got != "ssh" {
		t.Fatalf("protocol = %q, want ssh", got)
	}
}

func TestAIConversationSwitchKeepsPendingApproval(t *testing.T) {
	a := &App{}
	a.AI.SessionID = "s1"
	a.AI.Configured = true
	a.AI.State = session.StateAwaitingApproval
	a.AI.Streaming = "partial"
	a.AI.PendingCommand = &CommandSuggestion{Command: "rm -rf /tmp/x", Why: "cleanup"}
	a.AI.CommandStatus = AIStatusPending
	a.AI.CommandDraft = "rm -rf /tmp/x"

	a.SwitchAIConversation("s2")
	if a.AI.PendingCommand != nil || a.AI.State != "" {
		t.Fatalf("new conversation inherited state: %+v", a.AI)
	}

	a.SwitchAIConversation("s1")
	if a.AI.PendingCommand == nil || a.AI.PendingCommand.Command != "rm -rf /tmp/x" {
		t.Fatalf("pending command lost: %+v", a.AI.PendingCommand)
	}
	if a.AI.CommandStatus != AIStatusPending || a.AI.CommandDraft != "rm -rf /tmp/x" {
		t.Fatalf("command card state lost: %q/%q", a.AI.CommandStatus, a.AI.CommandDraft)
	}
	if a.AI.State != session.StateAwaitingApproval {
		t.Fatalf("state = %q, want AwaitingApproval", a.AI.State)
	}
}

func TestAIConversationSwitchKeepsPendingPlan(t *testing.T) {
	a := &App{}
	a.AI.SessionID = "s1"
	a.AI.State = session.StateAwaitingApproval
	a.AI.PendingPlan = &PlanSuggestion{Goal: "cleanup", Steps: []string{"a", "b"}}
	a.AI.PlanStatus = AIStatusPending

	a.SwitchAIConversation("s2")
	a.SwitchAIConversation("s1")
	if a.AI.PendingPlan == nil || a.AI.PendingPlan.Goal != "cleanup" || a.AI.PlanStatus != AIStatusPending {
		t.Fatalf("pending plan lost: %+v %q", a.AI.PendingPlan, a.AI.PlanStatus)
	}
}
