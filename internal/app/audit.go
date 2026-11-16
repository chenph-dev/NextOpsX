package app

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/i18n"
	logsstore "nextopsx/internal/store/logs"
)

// auditLogLimit 审计页一次加载的记录条数（store 上限 500）。
const auditLogLimit = 200

// AuditPageSize is the number of audit rows shown on one page.
const AuditPageSize = 50

// AuditState is the audit section of the settings page.
type AuditState struct {
	Logs     []logsstore.CallLog
	Summary  logsstore.TokenSummary
	Loading  bool
	Err      string
	Selected int // 当前页选中的日志行，-1 表示未选择
	Page     int // 当前页，从 0 开始
	List     ui.ListState

	seq uint64
}

// AuditPageCount returns the number of pages for the loaded logs.
func (a *App) AuditPageCount() int {
	pages := (len(a.Audit.Logs) + AuditPageSize - 1) / AuditPageSize
	if pages == 0 {
		return 1
	}
	return pages
}

// AuditPrevPage moves to the previous audit page.
func (a *App) AuditPrevPage() { a.setAuditPage(a.Audit.Page - 1) }

// AuditNextPage moves to the next audit page.
func (a *App) AuditNextPage() { a.setAuditPage(a.Audit.Page + 1) }

func (a *App) setAuditPage(page int) {
	pages := a.AuditPageCount()
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	if page == a.Audit.Page {
		return
	}
	a.Audit.Page = page
	a.Audit.Selected = -1
	a.Audit.List.ScrollTo(0, ui.Start)
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
			if a.Audit.Page < 0 {
				a.Audit.Page = 0
			}
			if pages := a.AuditPageCount(); a.Audit.Page >= pages {
				a.Audit.Page = pages - 1
			}
			a.Audit.Selected = -1
			a.Audit.List.ScrollTo(0, ui.Start)
		})
	}()
}
