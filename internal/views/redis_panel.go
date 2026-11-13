package views

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
)

// RedisPanel is the Redis key browser, key detail/info and command terminal.
func RedisPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		redisToolbar(c, a, t)
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
			redisKeysPane(c, a, t)
			ui.Column(c).Grow(1).FillHeight().MinWidth(0).Children(func() {
				ui.Row(c).Padding(4, 8).Background(t.Surface).Children(func() {
					tabs := ui.Tabs(c, &a.Redis.MainTab, "键详情", "服务器信息")
					if tabs.Changed() && a.Redis.MainTab == 1 {
						a.RedisLoadInfo()
					}
				})
				ui.Divider(c)
				redisMain(c, a, t)
				redisCommandTerminal(c, a, t)
			})
		})
		redisStatusBar(c, a, t)
	})
	redisDialogs(c, a, t)
}

func redisToolbar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Padding(6, 10).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "Redis · "+a.Redis.DBLabel).FontSize(fontSize(13)).Bold().SingleLine()
		ui.Spacer(c)
		ui.Button(c, "刷新键").FontSize(fontSize(12)).OnClick(a.RedisReloadKeys)
	})
}

func redisKeysPane(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(280).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 8).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("键 (%d)", len(a.Redis.Keys))).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Text(c, "游标 "+a.Redis.Cursor).FontSize(fontSize(11)).TextColor(t.TextMuted)
		})
		ui.Row(c).Padding(0, 6).Gap(4).Children(func() {
			c.Key("redis-pattern")
			ui.TextInput(c, &a.Redis.Pattern).Grow(1).FontSize(fontSize(12)).
				Placeholder("匹配模式，如 user:*").OnSubmit(a.RedisSearch)
			ui.Button(c, "搜索").FontSize(fontSize(12)).OnClick(a.RedisSearch)
		})
		ui.Row(c).Padding(0, 6, 6, 6).Children(func() {
			next := ui.Button(c, "下一页").FontSize(fontSize(12))
			if a.Redis.Cursor == "0" || a.Redis.KeysLoading {
				next = next.Disabled(true)
			}
			next.OnClick(a.RedisNextPage)
		})
		ui.Divider(c)
		switch {
		case a.Redis.KeysLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "扫描中…").FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case a.Redis.KeysErr != "":
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				dbErrorText(c, a.Redis.KeysErr, t)
				ui.Button(c, "重试").OnClick(func() { a.RedisRefreshKeys(true) })
			})
		case len(a.Redis.Keys) == 0:
			ui.Text(c, "没有匹配的键").FontSize(fontSize(12)).TextColor(t.TextMuted).Padding(8)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(1).Children(func() {
					for _, key := range a.Redis.Keys {
						redisKeyRow(c, a, key, t)
					}
				})
			})
		}
	})
}

func redisKeyRow(c *ui.Context, a *app.App, key string, t *ui.Theme) {
	selected := key == a.Redis.SelectedKey
	row := ui.Row(c).Padding(4, 8).Radius(4).AlignItems(ui.Center)
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.Tooltip(key).OnClick(func() { a.RedisSelectKey(key) })
	row.Children(func() {
		text := ui.Text(c, key).FontSize(fontSize(13)).SingleLine()
		if selected {
			text.Bold()
		}
	})
}

func redisMain(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Redis.MainTab == 1 {
		redisInfo(c, a, t)
		return
	}
	redisDetail(c, a, t)
}

func redisDetail(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Redis.SelectedKey == "" {
		ui.Column(c).Grow(1).Padding(24).Center().Children(func() {
			ui.Text(c, "选择左侧的键查看详情").FontSize(fontSize(13)).TextColor(t.TextMuted)
		})
		return
	}
	d := a.Redis.Detail
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, d.Key).FontSize(fontSize(13)).Bold().SingleLine().MaxWidth(420)
			if d.Type != "" {
				ui.Text(c, "类型 "+d.Type).FontSize(fontSize(12)).TextColor(t.TextMuted)
			}
			switch {
			case d.Loading:
			case d.TTLKnown:
				ui.Text(c, "TTL "+app.RedisTTLText(d.TTL)).
					FontSize(fontSize(12)).TextColor(t.TextMuted)
			default:
				ui.Text(c, "TTL ?").FontSize(fontSize(12)).TextColor(t.TextMuted)
			}
			if d.Truncated {
				ui.Text(c, "仅显示前 200 项").FontSize(fontSize(12)).TextColor(t.Warning)
			}
			ui.Spacer(c)
			if d.Type == "string" && !d.Loading {
				ui.Button(c, "编辑值").FontSize(fontSize(12)).OnClick(a.RedisOpenEditValue)
			}
			ttl := ui.Button(c, "设置 TTL").FontSize(fontSize(12))
			if d.Loading || d.Type == "" {
				ttl = ttl.Disabled(true)
			}
			ttl.OnClick(a.RedisOpenTTL)
			ui.Button(c, "删除").FontSize(fontSize(12)).OnClick(a.RedisOpenDelete)
		})
		switch {
		case d.Loading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取键…").FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case d.Err != "":
			dbErrorText(c, d.Err, t)
		case d.Result != nil:
			ResultGridView(c, &a.Redis.DetailGrid, d.Result, t)
		default:
			dbHint(c, "该类型暂不支持预览。", t)
		}
	})
}

func redisInfo(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Grow(1).MinWidth(0).Padding(8).Gap(6).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "服务器信息").FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, "刷新").FontSize(fontSize(12)).OnClick(a.RedisLoadInfo)
		})
		switch {
		case a.Redis.InfoLoading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取 INFO…").FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case a.Redis.InfoErr != "":
			dbErrorText(c, a.Redis.InfoErr, t)
		case len(a.Redis.Info) == 0:
			dbHint(c, "没有可展示的 INFO 指标。", t)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, m := range a.Redis.Info {
						ui.Row(c).Padding(2, 4).Gap(12).Children(func() {
							ui.Text(c, m.Name).Width(220).FontSize(fontSize(12)).
								TextColor(t.TextMuted).SingleLine()
							ui.Text(c, m.Value).FontSize(fontSize(12)).SingleLine().
								Selectable()
						})
					}
				})
			})
		}
	})
}

func redisCommandTerminal(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Column(c).Height(220).Shrink(0).Background(t.Surface).
		Border(1, t.Border).Radius(4).Children(func() {
		ui.Row(c).Padding(4, 8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "命令终端").FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			history := ui.Button(c, fmt.Sprintf("历史 (%d)", len(a.Redis.CommandHistory))).
				FontSize(fontSize(11))
			history = history.OnClick(func() {
				a.Redis.HistoryOpen = !a.Redis.HistoryOpen
			})
			ui.Popover(c, history, &a.Redis.HistoryOpen, func() {
				redisHistory(c, a, t)
			})
		})
		ui.Divider(c)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(6, 8).Children(func() {
				switch {
				case a.Redis.CommandRunning:
					ui.Row(c).Padding(8).Children(func() {
						ui.Spinner(c)
						ui.Text(c, "执行中…").FontSize(fontSize(12)).TextColor(t.TextMuted)
					})
				case a.Redis.CommandErr != "":
					dbErrorText(c, a.Redis.CommandErr, t)
				case a.Redis.CommandResult != nil:
					ResultGridView(c, &a.Redis.CommandGrid, a.Redis.CommandResult, t)
				default:
					dbHint(c, "输入 Redis 命令，Enter 执行。", t)
				}
			})
		})
		ui.Row(c).Padding(6, 8).Gap(6).AlignItems(ui.Center).Children(func() {
			c.Key("redis-command")
			ui.TextInput(c, &a.Redis.Command).Grow(1).Font("monospace").
				FontSize(fontSize(12)).Placeholder("GET key / HGETALL key / INFO …").
				OnSubmit(a.RedisRunCommand)
			run := ui.PrimaryButton(c, "执行").FontSize(fontSize(12))
			if a.Redis.CommandRunning {
				run = run.Disabled(true)
			}
			run.OnClick(a.RedisRunCommand)
		})
	})
}

func redisHistory(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(460).MaxHeight(320).Padding(8).Gap(4).Children(func() {
		ui.Text(c, "命令历史").FontSize(fontSize(13)).Bold()
		if len(a.Redis.CommandHistory) == 0 {
			ui.Text(c, "暂无历史").FontSize(fontSize(12)).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(260).Children(func() {
			ui.Column(c).Gap(2).Children(func() {
				for _, cmd := range a.Redis.CommandHistory {
					command := cmd
					ui.Button(c, command).FontSize(fontSize(12)).SingleLine().OnClick(func() {
						a.RedisPickHistory(command)
					})
				}
			})
		})
	})
}

func redisStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.Redis.DBLabel).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.Redis.Err != "" {
			ui.Text(c, a.Redis.Err).FontSize(fontSize(12)).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.Redis.Notice
		if text == "" {
			text = "就绪"
		}
		ui.Text(c, text).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
	})
}

func redisDialogs(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Modal(c, &a.Redis.EditOpen, func() {
		ui.Column(c).Width(540).Padding(18).Gap(10).Children(func() {
			ui.Text(c, "编辑字符串值").FontSize(fontSize(16)).Bold()
			ui.Text(c, a.Redis.EditKey).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
			c.Key("redis-edit-value")
			ui.TextArea(c, &a.Redis.EditValue).Lines(3, 8).Grow(1).Font("monospace")
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.Redis.EditOpen = false })
				ui.PrimaryButton(c, "保存").OnClick(a.RedisSaveValue)
			})
		})
	})
	ui.Modal(c, &a.Redis.TTLOpen, func() {
		ui.Column(c).Width(380).Padding(18).Gap(10).Children(func() {
			ui.Text(c, "设置 TTL").FontSize(fontSize(16)).Bold()
			ui.Text(c, a.Redis.TTLKey).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				c.Key("redis-ttl")
				ui.TextInput(c, &a.Redis.TTLValue).Width(160).Label("秒数")
				ui.Text(c, "-1 表示永久").FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.Redis.TTLOpen = false })
				ui.PrimaryButton(c, "保存").OnClick(a.RedisSaveTTL)
			})
		})
	})
	redisDeleteDialog(c, a)
}

func redisDeleteDialog(c *ui.Context, a *app.App) {
	if a.Redis.ConfirmDeleteKey == "" {
		return
	}
	open := true
	if ui.AlertDialog(c, &open, "删除键",
		fmt.Sprintf("确定删除「%s」吗？此操作无法撤销。", a.Redis.ConfirmDeleteKey),
		"取消", "确认删除") == 1 {
		a.RedisConfirmDelete()
	}
	if !open {
		a.Redis.ConfirmDeleteKey = ""
	}
}
