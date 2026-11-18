package terminal

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/ansi"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// SnippetPicker is the terminal's command panel: executed history and
// favorites merged by command so each command appears once, grouped by
// origin and day.
func SnippetPicker(c *ui.Context, a *workspace.State, t *ui.Theme) {
	items := a.CommandItems()
	now := time.Now()
	ui.Column(c).Width(620).MaxHeight(620).Background(t.Surface).
		Border(1, t.Border).Radius(12).
		Shadow(0, 12, 32, 0, ui.RGBA(0, 0, 0, 0.35)).
		Children(func() {
			pickerHeader(c, a, t, len(items))
			pickerSearch(c, a, t)
			pickerFilters(c, a, t)
			pickerNotices(c, a, t)
			ui.Divider(c)
			pickerList(c, a, t,
				filterCommandItems(items, a.Snippets.CommandQuery, a.Snippets.PickerTab), now)
			ui.Divider(c)
			ui.Row(c).Padding(8, 16).Children(func() {
				ui.Text(c, i18n.T("history.pickerHint")).
					FontSize(style.FontSize(10)).TextColor(t.TextMuted)
			})
		})
}

func pickerHeader(c *ui.Context, a *workspace.State, t *ui.Theme, count int) {
	ui.Row(c).Padding(14, 16, 10, 16).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.T("history.title")).FontSize(style.FontSize(14)).Bold()
		ui.Text(c, i18n.Tf("history.suggestCount", count)).
			FontSize(style.FontSize(11)).TextColor(t.TextMuted)
		ui.Spacer(c)
		manage := ui.ButtonBase(c).Padding(4, 8).Radius(6).
			OnClick(func() {
				a.Snippets.PickerOpen = false
				a.OpenSettings(workspace.SettingsSnippets)
			}).Cursor(ui.CursorPointer)
		if manage.Hovered() {
			manage = manage.Background(t.SurfaceHover)
		}
		manage.Children(func() {
			ui.Text(c, i18n.T("snippets.picker.manage")).
				FontSize(style.FontSize(11)).TextColor(t.TextMuted)
		})
		commandIconAction(c, t, "close", i18n.T("history.close"), t.TextMuted, func() {
			a.Snippets.PickerOpen = false
		})
	})
}

func pickerSearch(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Row(c).Padding(0, 16, 10, 16).Children(func() {
		field := ui.Row(c).Fill().Padding(6, 10).Gap(8).Radius(8).
			Background(t.Background).Border(1, t.Border).AlignItems(ui.Center)
		field.Children(func() {
			if s := assets.SVG("search"); s != nil {
				ui.Icon(c, s).Size(13, 13).TextColor(t.TextMuted)
			}
			c.Key("command-search")
			ui.TextInputBase(c, &a.Snippets.CommandQuery).
				Placeholder(i18n.T("history.search")).
				FontSize(style.FontSize(12)).Grow(1).MinWidth(0).AutoFocus()
			if a.Snippets.CommandQuery != "" {
				commandIconAction(c, t, "close", i18n.T("history.close"), t.TextMuted, func() {
					a.Snippets.CommandQuery = ""
				})
			}
		})
	})
}

func pickerFilters(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Row(c).Padding(0, 16, 8, 16).Gap(4).AlignItems(ui.Center).Children(func() {
		pickerChip(c, t, i18n.T("history.filterAll"),
			a.Snippets.PickerTab == workspace.PickerTabAll, func() {
				a.Snippets.PickerTab = workspace.PickerTabAll
			})
		pickerChip(c, t, i18n.T("history.filterFavorites"),
			a.Snippets.PickerTab == workspace.PickerTabFavorites, func() {
				a.Snippets.PickerTab = workspace.PickerTabFavorites
			})
		pickerChip(c, t, i18n.T("history.filterHistory"),
			a.Snippets.PickerTab == workspace.PickerTabHistory, func() {
				a.Snippets.PickerTab = workspace.PickerTabHistory
			})
		ui.Spacer(c)
		if a.Selected != nil && a.Selected.Name != "" {
			ui.Text(c, i18n.Tf("history.currentHost", a.Selected.Name)).
				FontSize(style.FontSize(10)).TextColor(t.TextMuted).SingleLine()
		}
	})
}

func pickerNotices(c *ui.Context, a *workspace.State, t *ui.Theme) {
	if a.CanUseSnippet() && a.Snippets.Err == "" && a.Snippets.HistoryErr == "" {
		return
	}
	ui.Column(c).Padding(0, 16, 8, 16).Gap(4).Children(func() {
		if !a.CanUseSnippet() {
			ui.Text(c, i18n.T("snippets.picker.connectFirst")).
				FontSize(style.FontSize(11)).TextColor(t.Warning)
		}
		if a.Snippets.Err != "" {
			ui.Text(c, a.Snippets.Err).FontSize(style.FontSize(11)).
				TextColor(t.Danger).Wrap()
		}
		if a.Snippets.HistoryErr != "" {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.Snippets.HistoryErr).FontSize(style.FontSize(11)).
					TextColor(t.Danger).Wrap().Grow(1)
				commandTextAction(c, t, i18n.T("history.retry"), t.Accent, a.LoadCommandHistory)
			})
		}
	})
}

func pickerChip(c *ui.Context, t *ui.Theme, label string, active bool, onClick func()) {
	chip := ui.ButtonBase(c).Padding(3, 9).Radius(6).
		OnClick(onClick).Cursor(ui.CursorPointer)
	if active {
		chip = chip.Background(t.SurfaceHover)
	} else if chip.Hovered() {
		chip = chip.Background(t.Surface)
	}
	chip.Children(func() {
		color := t.TextMuted
		if active {
			color = t.Text
		}
		ui.Text(c, label).FontSize(style.FontSize(11)).TextColor(color)
	})
}

func pickerList(c *ui.Context, a *workspace.State, t *ui.Theme, items []workspace.CommandItem, now time.Time) {
	switch {
	case a.Snippets.HistoryLoading && len(items) == 0:
		ui.Row(c).Padding(24).Center().Children(func() {
			ui.Spinner(c)
			ui.Text(c, i18n.T("history.title")).
				FontSize(style.FontSize(12)).TextColor(t.TextMuted)
		})
	case len(items) == 0:
		ui.Column(c).Padding(16, 16, 20, 16).Gap(6).Children(func() {
			if len(a.Snippets.History) == 0 && len(a.Snippets.Items) == 0 {
				ui.Text(c, i18n.T("history.emptyTitle")).
					FontSize(style.FontSize(13)).TextColor(t.TextMuted)
				ui.Text(c, i18n.T("history.emptyHint")).
					FontSize(style.FontSize(12)).TextColor(t.TextMuted).Wrap()
				return
			}
			ui.Text(c, i18n.T("history.noMatch")).
				FontSize(style.FontSize(13)).TextColor(t.TextMuted)
		})
	default:
		ui.Scroll(c).MaxHeight(410).Children(func() {
			ui.Column(c).Gap(1).Padding(4, 6).Children(func() {
				group := ""
				for _, item := range items {
					g := commandGroup(item, now)
					if g != group {
						group = g
						ui.Text(c, g).FontSize(style.FontSize(10)).
							TextColor(t.TextMuted).Padding(8, 8, 3, 10)
					}
					commandItemRow(c, a, item, t, now)
				}
			})
		})
	}
}

func commandItemRow(c *ui.Context, a *workspace.State, item workspace.CommandItem, t *ui.Theme, now time.Time) {
	display := a.CommandItemText(item)
	row := ui.Column(c).Padding(7, 10).Gap(4).Radius(8)
	hovered := row.Hovered()
	if hovered {
		row = row.Background(t.SurfaceHover)
	}
	row.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			icon, color := "clock", t.TextMuted
			if item.FavoriteID != "" {
				icon, color = "star-fill", t.Accent
			}
			if s := assets.SVG(icon); s != nil {
				ui.Icon(c, s).Size(12, 12).TextColor(color)
			}
			cmd := ui.Text(c, display).Font("monospace").FontSize(style.FontSize(13)).
				SingleLine().Grow(1).MinWidth(0).Cursor(ui.CursorPointer)
			cmd.Tooltip(display)
			cmd.OnClick(func() { a.InsertCommandItem(item) })
			if item.ExitCode != nil && *item.ExitCode != 0 {
				ui.Text(c, fmt.Sprintf("exit %d", *item.ExitCode)).
					FontSize(style.FontSize(11)).TextColor(t.Danger)
			}
			if when := commandTime(item.Ts, now); when != "" {
				ui.Text(c, when).FontSize(style.FontSize(11)).
					TextColor(t.TextMuted).SingleLine()
			}
		})
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			meta := item.Title
			if item.Description != "" {
				if meta != "" {
					meta += " · " + item.Description
				} else {
					meta = item.Description
				}
			}
			if meta != "" {
				ui.Text(c, meta).FontSize(style.FontSize(11)).TextColor(t.TextMuted).
					SingleLine().Grow(1).MinWidth(0)
			} else {
				ui.Spacer(c)
			}
			actionColor := t.TextMuted
			if hovered {
				actionColor = t.Text
			}
			if item.Output != "" {
				label := i18n.T("history.output")
				if a.Snippets.HistoryOpen[item.HistoryID] {
					label = i18n.T("history.hideOutput")
				}
				commandTextAction(c, t, label, actionColor, func() {
					a.ToggleHistoryOutput(item.HistoryID)
				})
			}
			commandTextAction(c, t, i18n.T("history.insert"), actionColor, func() {
				a.InsertCommandItem(item)
			})
			commandTextAction(c, t, i18n.T("history.run"), t.Accent, func() {
				a.RunCommandItem(item)
			})
			if item.FavoriteID == "" {
				commandIconAction(c, t, "star", i18n.T("history.favorite"),
					actionColor, func() { a.FavoriteCommandItem(item) })
			}
		})
		if item.Output != "" && a.Snippets.HistoryOpen[item.HistoryID] {
			ui.Column(c).Background(ansi.Bg).Border(1, ansi.Border).Radius(6).
				Padding(8).Children(func() {
				widgets.TerminalOutput(c, item.Output, 140)
			})
		}
	})
}

func commandTextAction(c *ui.Context, t *ui.Theme, label string, color ui.Color, onClick func()) {
	action := ui.ButtonBase(c).Padding(2, 7).Radius(5).
		OnClick(onClick).Cursor(ui.CursorPointer)
	if action.Hovered() {
		action = action.Background(t.SurfacePressed)
	}
	action.Children(func() {
		ui.Text(c, label).FontSize(style.FontSize(11)).TextColor(color)
	})
}

func commandIconAction(c *ui.Context, t *ui.Theme, icon, label string, color ui.Color, onClick func()) {
	action := ui.ButtonBase(c).Padding(3).Radius(5).Label(label).
		OnClick(onClick).Cursor(ui.CursorPointer).Tooltip(label)
	if action.Hovered() {
		action = action.Background(t.SurfacePressed)
	}
	action.Children(func() {
		if s := assets.SVG(icon); s != nil {
			ui.Icon(c, s).Size(12, 12).TextColor(color)
		}
	})
}

func filterCommandItems(items []workspace.CommandItem, query string, tab int) []workspace.CommandItem {
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]workspace.CommandItem, 0, len(items))
	for _, item := range items {
		switch tab {
		case workspace.PickerTabFavorites:
			if item.Source == "history" {
				continue
			}
		case workspace.PickerTabHistory:
			if item.Source == "favorite" {
				continue
			}
		}
		if q != "" {
			hay := strings.ToLower(item.Text + "\n" + item.Command + "\n" +
				item.Title + "\n" + item.Description)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func commandGroup(item workspace.CommandItem, now time.Time) string {
	if item.FavoriteID != "" {
		return i18n.T("history.groupFavorites")
	}
	if item.Ts <= 0 {
		return i18n.T("history.groupEarlier")
	}
	when := time.Unix(item.Ts, 0)
	switch {
	case sameDay(when, now):
		return i18n.T("history.groupToday")
	case sameDay(when, now.AddDate(0, 0, -1)):
		return i18n.T("history.groupYesterday")
	default:
		return i18n.T("history.groupEarlier")
	}
}

func commandTime(ts int64, now time.Time) string {
	if ts <= 0 {
		return ""
	}
	when := time.Unix(ts, 0)
	if d := now.Sub(when); d >= 0 && d < time.Minute {
		return i18n.T("history.justNow")
	} else if d >= time.Minute && d < time.Hour {
		return i18n.Tf("history.minutesAgo", int(d.Minutes()))
	}
	switch {
	case sameDay(when, now):
		return i18n.Tf("history.todayAt", when.Format("15:04"))
	case sameDay(when, now.AddDate(0, 0, -1)):
		return i18n.Tf("history.yesterdayAt", when.Format("15:04"))
	default:
		return when.Format("01-02 15:04")
	}
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
