package es

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/style"
	"nextopsx/ui/views/result"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// Panel is the Elasticsearch index browser, DSL query, cluster info and
// REST command terminal.
func Panel(c *ui.Context, a *workspace.State) {
	t := c.Theme()
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			esIndexPane(c, a, t)
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				ui.Row(c).Padding(4, 8).Background(t.Surface).Children(func() {
					tabs := ui.Tabs(c, &a.ES.MainTab, i18n.T("es.tab.query"), i18n.T("es.tab.info"))
					if tabs.Changed() && a.ES.MainTab == 1 {
						a.ESLoadInfo()
					}
				})
				ui.Divider(c)
				if a.ES.MainTab == 1 {
					esInfo(c, a, t)
				} else {
					esQuery(c, a, t)
				}
				esCommandTerminal(c, a, t)
			})
		})
	})
}

func esIndexPane(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Width(260).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.Tf("es.indices.title", len(a.ES.Indices))).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			widgets.IconButton(c, t, "refresh", i18n.T("es.action.refreshIndices"),
				false, false, a.ESRefreshIndices)
		})
		ui.Divider(c)
		switch {
		case a.ES.IndicesLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("es.indices.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.ES.IndicesErr != "":
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				widgets.ErrorText(c, a.ES.IndicesErr, t)
				ui.Button(c, i18n.T("es.action.retry")).OnClick(a.ESRefreshIndices)
			})
		case len(a.ES.Indices) == 0:
			ui.Text(c, i18n.T("es.indices.empty")).FontSize(style.FontSize(12)).TextColor(t.TextMuted).Padding(8)
		default:
			a.ES.IndexGrid.Selected = &a.ES.IndexRow
			list := ui.List(c, &a.ES.IndexGrid, len(a.ES.Indices), func(row int) {
				if row < 0 || row >= len(a.ES.Indices) {
					return
				}
				idx := a.ES.Indices[row]
				ui.Text(c, idx).FontSize(style.FontSize(13)).SingleLine().
					Padding(5, 10).Tooltip(idx)
			}).Grow(1).MinWidth(0)
			if list.Changed() && a.ES.IndexRow >= 0 && a.ES.IndexRow < len(a.ES.Indices) {
				a.ESSelectIndex(a.ES.Indices[a.ES.IndexRow])
			}
		}
	})
}

func esQuery(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			label := i18n.T("es.query.noIndex")
			if a.ES.SelectedIndex != "" {
				label = i18n.Tf("es.query.index", a.ES.SelectedIndex)
			}
			ui.Text(c, label).FontSize(style.FontSize(13)).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
			ui.Spacer(c)
			run := ui.PrimaryButton(c, i18n.T("es.action.runQuery")).FontSize(style.FontSize(12))
			if a.ES.SelectedIndex == "" || a.ES.QueryRunning {
				run = run.Disabled(true)
			}
			run.OnClick(a.ESRunQuery)
		})
		c.Key("es-dsl")
		ui.TextArea(c, &a.ES.QueryDSL).Lines(4, 10).Font("monospace").
			FontSize(style.FontSize(12)).Label("Query DSL").Placeholder(`{"query": {"match_all": {}}}`)
		switch {
		case a.ES.QueryRunning:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("es.query.running")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.ES.QueryErr != "":
			widgets.ErrorText(c, a.ES.QueryErr, t)
		case a.ES.QueryResult != nil:
			result.GridView(c, &a.ES.QueryGrid, a.ES.QueryResult, t)
		default:
			widgets.Hint(c, i18n.T("es.query.hint"), t)
		}
	})
}

func esInfo(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("es.info.title")).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("es.action.refresh")).FontSize(style.FontSize(12)).OnClick(a.ESLoadInfo)
		})
		switch {
		case a.ES.InfoLoading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("es.info.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.ES.InfoErr != "":
			widgets.ErrorText(c, a.ES.InfoErr, t)
		case len(a.ES.Info) == 0:
			widgets.Hint(c, i18n.T("es.info.empty"), t)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, m := range a.ES.Info {
						ui.Row(c).Padding(2, 4).Gap(12).Children(func() {
							ui.Text(c, m.Name).Width(220).FontSize(style.FontSize(12)).
								TextColor(t.TextMuted).SingleLine()
							ui.Text(c, m.Value).FontSize(style.FontSize(12)).SingleLine().Selectable()
						})
					}
				})
			})
		}
	})
}

func esCommandTerminal(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Divider(c)
	ui.Column(c).Height(240).Shrink(0).Background(t.Surface).
		Border(1, t.Border).Radius(4).Children(func() {
		ui.Row(c).Padding(4, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("es.command.title")).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			history := ui.Button(c, i18n.Tf("es.command.historyCount", len(a.ES.CommandHistory))).
				FontSize(style.FontSize(11))
			history = history.OnClick(func() {
				a.ES.HistoryOpen = !a.ES.HistoryOpen
			})
			ui.Popover(c, history, &a.ES.HistoryOpen, func() {
				esHistory(c, a, t)
			})
		})
		ui.Divider(c)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(6, 8).Children(func() {
				switch {
				case a.ES.CommandRunning:
					ui.Row(c).Padding(8).Children(func() {
						ui.Spinner(c)
						ui.Text(c, i18n.T("es.command.running")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
					})
				case a.ES.CommandErr != "":
					widgets.ErrorText(c, a.ES.CommandErr, t)
				case a.ES.CommandResult != nil:
					result.GridView(c, &a.ES.CommandGrid, a.ES.CommandResult, t)
				default:
					widgets.Hint(c, i18n.T("es.command.hint"), t)
				}
			})
		})
		ui.Divider(c)
		ui.Row(c).Padding(6, 8).Gap(6).AlignItems(ui.End).Children(func() {
			c.Key("es-command")
			ui.TextArea(c, &a.ES.Command).Grow(1).Lines(2, 6).
				Font("monospace").FontSize(style.FontSize(12)).
				Placeholder(i18n.T("es.command.placeholder"))
			run := ui.PrimaryButton(c, i18n.T("es.action.execute")).FontSize(style.FontSize(12))
			if a.ES.CommandRunning {
				run = run.Disabled(true)
			}
			run.OnClick(a.ESRunCommand)
		})
	})
}

func esHistory(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Width(500).MaxHeight(320).Padding(8).Gap(4).Children(func() {
		ui.Text(c, i18n.T("es.command.historyTitle")).FontSize(style.FontSize(13)).Bold()
		if len(a.ES.CommandHistory) == 0 {
			ui.Text(c, i18n.T("es.command.historyEmpty")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(260).Children(func() {
			ui.Column(c).Gap(2).Children(func() {
				for _, cmd := range a.ES.CommandHistory {
					command := cmd
					ui.Button(c, command).FontSize(style.FontSize(12)).SingleLine().OnClick(func() {
						a.ESPickHistory(command)
					})
				}
			})
		})
	})
}
