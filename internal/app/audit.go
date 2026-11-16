package app

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/i18n"
	logsstore "nextopsx/internal/store/logs"
)

// auditLogLimit 审计页一次展示的记录条数（store 上限 500）。
const auditLogLimit = 200

// AuditState is the audit section of the settings page.
type AuditState struct {
	Logs     []logsstore.CallLog
	Summary  logsstore.TokenSummary
	Loading  bool
	Err      string
	Selected int // 选中的日志行，-1 表示未选择
	List     ui.ListState

	seq uint64
}

// LoadAudit refreshes the audit log list and token summary.
func (a *App) LoadAudit() {
	a.Audit.seq++
	seq := a.Audit.seq
	a.Audit.Loading = true
	a.Audit.Err = ""
	if a.Backend == nil {
		a.Audit.Loading = false
		a.Audit.Err = i18n.T("settings.err.backendNotInitialized")
		return
	}
	b := a.Backend
	go func() {
		list, err := b.ListAICallLogs(auditLogLimit)
		summary, sumErr := b.AITokenSummary()
		a.Post(func() {
			if seq != a.Audit.seq {
				return
			}
			a.Audit.Loading = false
			if err != nil {
				a.Audit.Err = i18n.Tf("settings.err.loadAuditFailed", err.Error())
				return
			}
			a.Audit.Logs = list
			if sumErr != nil {
				a.Audit.Err = i18n.Tf("settings.err.loadTokenStatsFailed", sumErr.Error())
			} else {
				a.Audit.Summary = summary
			}
			if a.Audit.Selected >= len(list) {
				a.Audit.Selected = -1
			}
		})
	}()
}
