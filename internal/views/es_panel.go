package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// ESPanel is the Elasticsearch index browser, DSL query, cluster info and
// REST command terminal.
func ESPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(6, 10).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Elasticsearch · "+a.ES.Label).FontSize(12).Bold().SingleLine()
			ui.Spacer(c)
			ui.Button(c, "刷新索引").FontSize(11).OnClick(a.ESRefreshIndices)
		})
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			esIndexPane(c, a, t)
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				ui.Row(c).Padding(4, 8).Background(t.Surface).Children(func() {
					tabs := ui.Tabs(c, &a.ES.MainTab, "DSL 查询", "集群信息")
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
		esStatusBar(c, a, t)
	})
}

func esIndexPane(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(260).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("索引 (%d)", len(a.ES.Indices))).FontSize(12).Bold()
		})
		ui.Divider(c)
		switch {
		case a.ES.IndicesLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取索引…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.ES.IndicesErr != "":
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				dbErrorText(c, a.ES.IndicesErr, t)
				ui.Button(c, "重试").OnClick(a.ESRefreshIndices)
			})
		case len(a.ES.Indices) == 0:
			ui.Text(c, "没有索引").FontSize(11).TextColor(t.TextMuted).Padding(8)
		default:
			a.ES.IndexGrid.Selected = &a.ES.IndexRow
			cols := []ui.TableColumn{{Title: "索引", MinWidth: 220}}
			table := ui.Table(c, &a.ES.IndexGrid, cols, len(a.ES.Indices),
				func(row, col int) {
					if row < 0 || row >= len(a.ES.Indices) {
						return
					}
					idx := a.ES.Indices[row]
					ui.Text(c, idx).FontSize(12).SingleLine().Tooltip(idx)
				}).Grow(1).MinWidth(0)
			if table.Changed() && a.ES.IndexRow >= 0 && a.ES.IndexRow < len(a.ES.Indices) {
				a.ESSelectIndex(a.ES.Indices[a.ES.IndexRow])
			}
		}
	})
}

func esQuery(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			label := "索引：未选择"
			if a.ES.SelectedIndex != "" {
				label = "索引：" + a.ES.SelectedIndex
			}
			ui.Text(c, label).FontSize(12).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
			ui.Spacer(c)
			run := ui.PrimaryButton(c, "执行查询").FontSize(11)
			if a.ES.SelectedIndex == "" || a.ES.QueryRunning {
				run = run.Disabled(true)
			}
			run.OnClick(a.ESRunQuery)
		})
		c.Key("es-dsl")
		ui.TextArea(c, &a.ES.QueryDSL).Lines(4, 10).Font("monospace").
			FontSize(11).Label("Query DSL").Placeholder(`{"query": {"match_all": {}}}`)
		switch {
		case a.ES.QueryRunning:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "查询中…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.ES.QueryErr != "":
			dbErrorText(c, a.ES.QueryErr, t)
		case a.ES.QueryResult != nil:
			ResultGridView(c, &a.ES.QueryGrid, a.ES.QueryResult, t)
		default:
			dbHint(c, "选择索引并输入 Query DSL 后执行。", t)
		}
	})
}

func esInfo(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "集群信息").FontSize(12).Bold()
			ui.Spacer(c)
			ui.Button(c, "刷新").FontSize(11).OnClick(a.ESLoadInfo)
		})
		switch {
		case a.ES.InfoLoading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取集群信息…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.ES.InfoErr != "":
			dbErrorText(c, a.ES.InfoErr, t)
		case len(a.ES.Info) == 0:
			dbHint(c, "没有可展示的集群指标。", t)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, m := range a.ES.Info {
						ui.Row(c).Padding(2, 4).Gap(12).Children(func() {
							ui.Text(c, m.Name).Width(220).FontSize(11).
								TextColor(t.TextMuted).SingleLine()
							ui.Text(c, m.Value).FontSize(11).SingleLine().Selectable()
						})
					}
				})
			})
		}
	})
}

func esCommandTerminal(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Column(c).Height(240).Shrink(0).Background(t.Surface).
		Border(1, t.Border).Radius(4).Children(func() {
		ui.Row(c).Padding(4, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "REST 命令终端").FontSize(12).Bold()
			ui.Spacer(c)
			history := ui.Button(c, fmt.Sprintf("历史 (%d)", len(a.ES.CommandHistory))).
				FontSize(10)
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
						ui.Text(c, "执行中…").FontSize(11).TextColor(t.TextMuted)
					})
				case a.ES.CommandErr != "":
					dbErrorText(c, a.ES.CommandErr, t)
				case a.ES.CommandResult != nil:
					ResultGridView(c, &a.ES.CommandGrid, a.ES.CommandResult, t)
				default:
					dbHint(c, "输入 REST 命令；首行为路径，后续行为 JSON body。", t)
				}
			})
		})
		ui.Divider(c)
		ui.Row(c).Padding(6, 8).Gap(6).AlignItems(ui.End).Children(func() {
			c.Key("es-command")
			ui.TextArea(c, &a.ES.Command).Grow(1).Lines(2, 6).
				Font("monospace").FontSize(11).
				Placeholder("_cat/indices?format=json\n或 logs-*/_search\n{\"query\":{\"match_all\":{}}}")
			run := ui.PrimaryButton(c, "执行").FontSize(11)
			if a.ES.CommandRunning {
				run = run.Disabled(true)
			}
			run.OnClick(a.ESRunCommand)
		})
	})
}

func esHistory(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(500).MaxHeight(320).Padding(8).Gap(4).Children(func() {
		ui.Text(c, "命令历史").FontSize(12).Bold()
		if len(a.ES.CommandHistory) == 0 {
			ui.Text(c, "暂无历史").FontSize(11).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(260).Children(func() {
			ui.Column(c).Gap(2).Children(func() {
				for _, cmd := range a.ES.CommandHistory {
					command := cmd
					ui.Button(c, command).FontSize(11).SingleLine().OnClick(func() {
						a.ESPickHistory(command)
					})
				}
			})
		})
	})
}

func esStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.ES.Label).FontSize(11).TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.ES.Err != "" {
			ui.Text(c, a.ES.Err).FontSize(11).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.ES.Notice
		if text == "" {
			text = "就绪"
		}
		ui.Text(c, text).FontSize(11).TextColor(t.TextMuted).SingleLine()
	})
}
