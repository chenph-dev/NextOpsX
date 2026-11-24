package callback

import (
	"testing"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	logsstore "nextopsx/store/logs"
)

func TestNameOf(t *testing.T) {
	cases := []struct {
		name string
		info *callbacks.RunInfo
		want string
	}{
		{"node display name", &callbacks.RunInfo{Name: "llm", Type: "StreamingChatModel"}, "llm"},
		{"fallback to implementation type", &callbacks.RunInfo{Type: "StreamingChatModel"}, "StreamingChatModel"},
		{"empty info", &callbacks.RunInfo{}, "unknown"},
		{"nil info", nil, "unknown"},
	}
	for _, c := range cases {
		if got := nameOf(c.info); got != c.want {
			t.Errorf("%s: nameOf() = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestApplyTokenUsage(t *testing.T) {
	t.Run("component callback output", func(t *testing.T) {
		log := logsstore.CallLog{}
		applyTokenUsage(&log, &model.CallbackOutput{
			TokenUsage: &model.TokenUsage{PromptTokens: 11, CompletionTokens: 7, TotalTokens: 18},
		})
		if log.TokensIn != 11 || log.TokensOut != 7 || log.TokensTotal != 18 {
			t.Fatalf("tokens = %d/%d/%d, want 11/7/18", log.TokensIn, log.TokensOut, log.TokensTotal)
		}
	})

	t.Run("graph node message output", func(t *testing.T) {
		log := logsstore.CallLog{}
		applyTokenUsage(&log, &schema.Message{
			ResponseMeta: &schema.ResponseMeta{
				Usage: &schema.TokenUsage{PromptTokens: 3, CompletionTokens: 5, TotalTokens: 8},
			},
		})
		if log.TokensIn != 3 || log.TokensOut != 5 || log.TokensTotal != 8 {
			t.Fatalf("tokens = %d/%d/%d, want 3/5/8", log.TokensIn, log.TokensOut, log.TokensTotal)
		}
	})

	t.Run("non-model output leaves tokens empty", func(t *testing.T) {
		log := logsstore.CallLog{}
		applyTokenUsage(&log, "ok")
		if log.TokensIn != 0 || log.TokensOut != 0 || log.TokensTotal != 0 {
			t.Fatalf("tokens = %d/%d/%d, want 0/0/0", log.TokensIn, log.TokensOut, log.TokensTotal)
		}
	})
}
