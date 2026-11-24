package tools

import "sync/atomic"

// PlanGrant 是一次已批准计划的执行授权：会话在计划批准后激活，
// 命令工具据此跳过逐条审批（高危/敏感命令除外）。
//
// 由 agentSession 持有并注入工具，而不是放在工具实例里——Graph 会因配置
// 热更新重建工具，授权必须跨重建存活。nil 安全：未注入时一律视为未授权。
type PlanGrant struct {
	active atomic.Bool
}

// NewPlanGrant 构造一份未激活的授权。
func NewPlanGrant() *PlanGrant { return &PlanGrant{} }

// Active 报告当前是否有已批准计划的执行授权。
func (g *PlanGrant) Active() bool {
	return g != nil && g.active.Load()
}

// Activate 激活授权（计划批准）。
func (g *PlanGrant) Activate() {
	if g != nil {
		g.active.Store(true)
	}
}

// Revoke 撤销授权（本轮结束、新消息、清空/删除会话、拒绝计划）。
func (g *PlanGrant) Revoke() {
	if g != nil {
		g.active.Store(false)
	}
}
