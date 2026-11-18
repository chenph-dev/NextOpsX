package backend

import "context"

// ctx 是应用启动时注入的上下文，供各 handler 子包内部使用。
// 作为包级变量而非 handler 字段，避免为 SetCtx 方法暴露 context.Context 参数。
var ctx context.Context

// SetCtx 由组合根在启动时调用，注入应用上下文。
func SetCtx(c context.Context) {
	ctx = c
}

// Ctx 返回当前应用上下文，供 handler 内部使用。
// 未注入时回退到 Background，避免 context.WithTimeout(nil, …) panic。
func Ctx() context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
