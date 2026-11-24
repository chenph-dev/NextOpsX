// Package callback 提供智能体调用的审计日志（基于 eino callbacks，落库 ai_call_logs）。
package callback

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	logsstore "nextopsx/store/logs"
)

// startedAtKey 在 OnStart 时把开始时间写入回调 ctx，供 OnEnd 计算耗时。
// 每个组件调用（含并发工具）都有独立的 ctx，不会互相覆盖。
type startedAtKey struct{}

// NewLogHandler 构造智能体调用审计 handler：
// 每次模型/工具调用结束落库一条记录（含 token 用量、耗时、成败）。
// sessionID 由调用方传入（每次 run 独立创建，避免共享实例无法区分会话）。
// 通过 compose.WithCallbacks 挂到 Graph 节点，由组件内部触发。
func NewLogHandler(logs *logsstore.LogsStore, sessionID string) callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, _ *callbacks.RunInfo, _ callbacks.CallbackInput) context.Context {
		return context.WithValue(ctx, startedAtKey{}, time.Now())
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		log := logsstore.CallLog{
			SessionID: sessionID, Ts: time.Now().Unix(), OK: true,
			Component: componentOf(info), Name: nameOf(info), Provider: typeOf(info),
		}
		if started, ok := ctx.Value(startedAtKey{}).(time.Time); ok {
			log.DurationMS = time.Since(started).Milliseconds()
		}
		applyTokenUsage(&log, output)
		if logs != nil {
			_ = logs.SaveLog(log)
		}
		return ctx
	})
	builder.OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
		log := logsstore.CallLog{
			SessionID: sessionID, Ts: time.Now().Unix(), OK: false,
			Component: componentOf(info), Name: nameOf(info), Provider: typeOf(info),
			Error: err.Error(),
		}
		if logs != nil {
			_ = logs.SaveLog(log)
		}
		return ctx
	})
	return builder.Build()
}

func nameOf(info *callbacks.RunInfo) string {
	if info == nil {
		return "unknown"
	}
	if info.Name != "" {
		return info.Name
	}
	// 未显式设置显示名的节点退回实现类型，避免审计页出现无意义的 unknown。
	if info.Type != "" {
		return info.Type
	}
	return "unknown"
}

// applyTokenUsage 填充 token 用量。模型调用有两种回调形态：
//   - 组件自身发回调时为 *model.CallbackOutput；
//   - 经 Graph 节点包装时（如 StreamingChatModel）为节点的原始输出 *schema.Message。
func applyTokenUsage(log *logsstore.CallLog, output callbacks.CallbackOutput) {
	switch o := output.(type) {
	case *model.CallbackOutput:
		if o != nil && o.TokenUsage != nil {
			log.TokensIn = o.TokenUsage.PromptTokens
			log.TokensOut = o.TokenUsage.CompletionTokens
			log.TokensTotal = o.TokenUsage.TotalTokens
		}
	case *schema.Message:
		if o != nil && o.ResponseMeta != nil && o.ResponseMeta.Usage != nil {
			log.TokensIn = o.ResponseMeta.Usage.PromptTokens
			log.TokensOut = o.ResponseMeta.Usage.CompletionTokens
			log.TokensTotal = o.ResponseMeta.Usage.TotalTokens
		}
	}
}

func typeOf(info *callbacks.RunInfo) string {
	if info != nil {
		return info.Type
	}
	return ""
}

func componentOf(info *callbacks.RunInfo) string {
	if info != nil {
		// 归一为 "model" / "tool"（其余组件归为 "other"，实际主要是模型/工具节点）。
		switch info.Component {
		case components.ComponentOfChatModel:
			return "model"
		case components.ComponentOfTool:
			return "tool"
		}
	}
	return "other"
}
