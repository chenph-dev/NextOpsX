package backend

import "testing"

func TestAIEventSinkForwards(t *testing.T) {
	b := &Backend{}
	var (
		gotSession string
		gotEvent   string
		gotData    any
	)
	b.SetAIEventSink(func(sessionID, event string, data any) {
		gotSession, gotEvent, gotData = sessionID, event, data
	})

	b.emitAI("s1", "ai:text", map[string]any{"delta": "hi"})
	if gotSession != "s1" || gotEvent != "ai:text" {
		t.Fatalf("event = %q/%q, want s1/ai:text", gotSession, gotEvent)
	}
	if m, ok := gotData.(map[string]any); !ok || m["delta"] != "hi" {
		t.Fatalf("data = %#v", gotData)
	}

	b.SetAIEventSink(nil)
	b.emitAI("s1", "ai:text", nil) // must not panic
}
