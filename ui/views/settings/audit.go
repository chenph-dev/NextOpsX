package settings

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// auditSettings is the AI call audit section of the settings page.
func auditSettings(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Grow(1).MinHeight(0).MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("settings.audit.title")).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("settings.audit.refresh")).OnClick(a.LoadAudit)
		})
		ui.Divider(c)
		switch {
		case a.Audit.Loading && len(a.Audit.Logs) == 0:
			ui.Row(c).Grow(1).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("settings.audit.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.Audit.Err != "" && len(a.Audit.Logs) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				widgets.ErrorText(c, a.Audit.Err, t)
				ui.Button(c, i18n.T("settings.audit.retry")).OnClick(a.LoadAudit)
			})
		default:
			if a.Audit.Err != "" {
				ui.Text(c, a.Audit.Err).FontSize(style.FontSize(12)).TextColor(t.Danger).
					Padding(4, 16).Wrap()
			}
			auditSummary(c, a, t)
			ui.Divider(c)
			if len(a.Audit.Logs) == 0 {
				ui.Column(c).Padding(16).Gap(6).Children(func() {
					ui.Text(c, i18n.T("settings.audit.emptyTitle")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
					ui.Text(c, i18n.T("settings.audit.emptyHint")).
						FontSize(style.FontSize(12)).TextColor(t.TextMuted)
				})
				return
			}
			auditTable(c, a, t)
			ui.Divider(c)
			auditPager(c, a, t)
			auditDetail(c, a, t)
		}
	})
}

func auditSummary(c *ui.Context, a *workspace.State, t *ui.Theme) {
	s := a.Audit.Summary
	ui.Row(c).Padding(10, 16).Gap(20).AlignItems(ui.Center).Children(func() {
		auditStat(c, t, i18n.T("settings.audit.calls"), fmt.Sprintf("%d", s.TotalCalls))
		auditStat(c, t, i18n.T("settings.audit.models"), fmt.Sprintf("%d", s.ModelCalls))
		auditStat(c, t, i18n.T("settings.audit.tools"), fmt.Sprintf("%d", s.ToolCalls))
		auditStat(c, t, i18n.T("settings.audit.tokensIn"), fmt.Sprintf("%d", s.TokensIn))
		auditStat(c, t, i18n.T("settings.audit.tokensOut"), fmt.Sprintf("%d", s.TokensOut))
		auditStat(c, t, i18n.T("settings.audit.tokensTotal"), fmt.Sprintf("%d", s.TokensTotal))
	})
}

func auditStat(c *ui.Context, t *ui.Theme, label, value string) {
	ui.Column(c).Gap(1).Children(func() {
		ui.Text(c, value).FontSize(style.FontSize(15)).Bold().SingleLine()
		ui.Text(c, label).FontSize(style.FontSize(11)).TextColor(t.TextMuted).SingleLine()
	})
}

func AuditColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Title: i18n.T("settings.audit.col.time"), Width: 130},
		{Title: i18n.T("settings.audit.col.component"), Width: 80},
		{Title: i18n.T("settings.audit.col.name"), Width: 160},
		{Title: i18n.T("settings.audit.col.provider"), Width: 130},
		{Title: i18n.T("settings.audit.col.in"), Width: 80, Align: ui.End},
		{Title: i18n.T("settings.audit.col.out"), Width: 80, Align: ui.End},
		{Title: i18n.T("settings.audit.col.total"), Width: 90, Align: ui.End},
		{Title: i18n.T("settings.audit.col.duration"), Width: 88, Align: ui.End},
		{Title: i18n.T("settings.audit.col.status"), Width: 74},
	}
}

// auditPageBounds returns the clamped page and the half-open range of logs
// shown on it.
func auditPageBounds(a *workspace.State) (page, start, end, pages int) {
	pages = a.AuditPageCount()
	page = a.Audit.Page
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start = page * workspace.AuditPageSize
	end = start + workspace.AuditPageSize
	if end > len(a.Audit.Logs) {
		end = len(a.Audit.Logs)
	}
	if start > end {
		start = end
	}
	return page, start, end, pages
}

func auditTable(c *ui.Context, a *workspace.State, t *ui.Theme) {
	cols := AuditColumns()
	_, start, end, _ := auditPageBounds(a)
	pageLogs := a.Audit.Logs[start:end]

	a.Audit.List.Selected = &a.Audit.Selected
	// 以记录 ID 作为行标识，刷新后选中项跟随记录而不是下标。
	a.Audit.List.Key = func(row int) any {
		if row >= 0 && row < len(pageLogs) {
			return pageLogs[row].ID
		}
		return row
	}
	ui.Table(c, &a.Audit.List, cols, len(pageLogs), func(row, col int) {
		if row < 0 || row >= len(pageLogs) {
			return
		}
		l := pageLogs[row]
		switch col {
		case 0:
			ui.Text(c, time.Unix(l.Ts, 0).Format("01-02 15:04:05")).
				FontSize(style.FontSize(12)).SingleLine()
		case 1:
			component := i18n.T("settings.audit.componentOther")
			switch l.Component {
			case "model":
				component = i18n.T("settings.audit.componentModel")
			case "tool":
				component = i18n.T("settings.audit.componentTool")
			}
			ui.Text(c, component).FontSize(style.FontSize(12)).SingleLine()
		case 2:
			ui.Text(c, l.Name).FontSize(style.FontSize(12)).SingleLine()
		case 3:
			ui.Text(c, l.Provider).FontSize(style.FontSize(12)).SingleLine()
		case 4:
			ui.Text(c, fmt.Sprintf("%d", l.TokensIn)).FontSize(style.FontSize(12)).SingleLine()
		case 5:
			ui.Text(c, fmt.Sprintf("%d", l.TokensOut)).FontSize(style.FontSize(12)).SingleLine()
		case 6:
			ui.Text(c, fmt.Sprintf("%d", l.TokensTotal)).FontSize(style.FontSize(12)).SingleLine()
		case 7:
			ui.Text(c, fmt.Sprintf("%dms", l.DurationMS)).FontSize(style.FontSize(12)).SingleLine()
		case 8:
			if l.OK {
				ui.Text(c, i18n.T("settings.audit.success")).FontSize(style.FontSize(12)).TextColor(t.Success).SingleLine()
			} else {
				ui.Text(c, i18n.T("settings.audit.failure")).FontSize(style.FontSize(12)).TextColor(t.Danger).SingleLine()
			}
		}
	}).Grow(1).MinHeight(0).MinWidth(0)
}

func auditPager(c *ui.Context, a *workspace.State, t *ui.Theme) {
	page, start, end, pages := auditPageBounds(a)
	ui.Row(c).Padding(6, 14).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.Tf("settings.audit.total", len(a.Audit.Logs))).
			FontSize(style.FontSize(12)).TextColor(t.TextMuted)
		if end > start {
			ui.Text(c, fmt.Sprintf("%d–%d", start+1, end)).
				FontSize(style.FontSize(12)).TextColor(t.TextMuted)
		}
		ui.Spacer(c)
		prev := ui.Button(c, i18n.T("settings.audit.prev")).FontSize(style.FontSize(12))
		if page <= 0 {
			prev = prev.Disabled(true)
		}
		prev.OnClick(a.AuditPrevPage)
		ui.Text(c, i18n.Tf("settings.audit.page", page+1, pages)).FontSize(style.FontSize(12))
		next := ui.Button(c, i18n.T("settings.audit.next")).FontSize(style.FontSize(12))
		if page >= pages-1 {
			next = next.Disabled(true)
		}
		next.OnClick(a.AuditNextPage)
	})
}

func auditDetail(c *ui.Context, a *workspace.State, t *ui.Theme) {
	page, _, _, _ := auditPageBounds(a)
	i := page*workspace.AuditPageSize + a.Audit.Selected
	if a.Audit.Selected < 0 || i < 0 || i >= len(a.Audit.Logs) {
		return
	}
	l := a.Audit.Logs[i]
	ui.Divider(c)
	ui.Column(c).Padding(6, 16).Gap(2).Children(func() {
		line := i18n.Tf("settings.audit.sessionLine",
			l.SessionID, time.Unix(l.Ts, 0).Format("2006-01-02 15:04:05"))
		if l.SessionID == "" {
			line = time.Unix(l.Ts, 0).Format("2006-01-02 15:04:05")
		}
		ui.Text(c, line).FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
		if l.Error != "" {
			ui.Text(c, i18n.Tf("settings.audit.error", l.Error)).FontSize(style.FontSize(12)).TextColor(t.Danger).Wrap()
		}
	})
}
