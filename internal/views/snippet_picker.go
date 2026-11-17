package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
)

// snippetPicker is the terminal's command bubble: executed history and
// favorites merged by command text so each command appears once.
func snippetPicker(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(520).MaxHeight(560).Padding(10).Gap(8).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("history.title")).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Button(c, i18n.T("snippets.picker.manage")).FontSize(fontSize(12)).
				OnClick(func() {
					a.Snippets.PickerOpen = false
					a.OpenSettings(app.SettingsSnippets)
				})
		})
		c.Key("command-search")
		ui.TextInput(c, &a.Snippets.CommandQuery).Placeholder(i18n.T("history.search"))
		if !a.CanUseSnippet() {
			ui.Text(c, i18n.T("snippets.picker.connectFirst")).FontSize(fontSize(12)).
				TextColor(t.Warning).Wrap()
		}
		if a.Snippets.Err != "" {
			ui.Text(c, a.Snippets.Err).FontSize(fontSize(12)).TextColor(t.Danger).Wrap()
		}
		if a.Snippets.HistoryErr != "" {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.Snippets.HistoryErr).FontSize(fontSize(12)).
					TextColor(t.Danger).Wrap().Grow(1)
				ui.Button(c, i18n.T("history.retry")).FontSize(fontSize(11)).
					OnClick(a.LoadCommandHistory)
			})
		}
		items := filterCommandItems(a.CommandItems(), a.Snippets.CommandQuery)
		switch {
		case a.Snippets.HistoryLoading && len(items) == 0:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("history.title")).FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case len(items) == 0:
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				if len(a.Snippets.History) == 0 && len(a.Snippets.Items) == 0 {
					ui.Text(c, i18n.T("history.emptyTitle")).FontSize(fontSize(13)).TextColor(t.TextMuted)
					ui.Text(c, i18n.T("history.emptyHint")).FontSize(fontSize(12)).
						TextColor(t.TextMuted).Wrap()
					return
				}
				ui.Text(c, i18n.T("history.noMatch")).FontSize(fontSize(13)).TextColor(t.TextMuted)
			})
		default:
			ui.Scroll(c).MaxHeight(440).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, item := range items {
						commandItemRow(c, a, item, t)
					}
				})
			})
		}
	})
}

func commandItemRow(c *ui.Context, a *app.App, item app.CommandItem, t *ui.Theme) {
	display := a.CommandItemText(item)
	ui.Column(c).Padding(6, 8).Gap(3).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			cmd := ui.Text(c, display).Font("monospace").FontSize(fontSize(12)).
				SingleLine().Grow(1).MinWidth(0)
			cmd.Tooltip(display)
			ui.Text(c, commandItemSource(item.Source)).FontSize(fontSize(10)).TextColor(t.TextMuted)
		})
		if item.Title != "" {
			line := item.Title
			if item.Description != "" {
				line += " · " + item.Description
			}
			ui.Text(c, line).FontSize(fontSize(10)).TextColor(t.TextMuted).SingleLine()
		}
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			if item.Ts > 0 {
				ui.Text(c, time.Unix(item.Ts, 0).Format("01-02 15:04:05")).
					FontSize(fontSize(10)).TextColor(t.TextMuted).SingleLine()
			}
			if item.ExitCode != nil {
				color := t.Success
				if *item.ExitCode != 0 {
					color = t.Danger
				}
				ui.Text(c, fmt.Sprintf("exit %d", *item.ExitCode)).
					FontSize(fontSize(10)).TextColor(color)
			}
			if item.Output != "" {
				label := i18n.T("history.output")
				if a.Snippets.HistoryOpen[item.HistoryID] {
					label = i18n.T("history.hideOutput")
				}
				ui.Button(c, label).FontSize(fontSize(10)).
					OnClick(func() { a.ToggleHistoryOutput(item.HistoryID) })
			}
			ui.Spacer(c)
			insert := ui.Button(c, i18n.T("history.insert")).FontSize(fontSize(11)).
				OnClick(func() { a.InsertCommandItem(item) })
			run := ui.Button(c, i18n.T("history.run")).FontSize(fontSize(11)).
				OnClick(func() { a.RunCommandItem(item) })
			if !a.CanUseSnippet() {
				insert = insert.Disabled(true)
				run = run.Disabled(true)
			}
			if item.FavoriteID == "" {
				ui.Button(c, i18n.T("history.favorite")).FontSize(fontSize(11)).
					OnClick(func() { a.FavoriteCommandItem(item) })
			}
		})
		if item.Output != "" && a.Snippets.HistoryOpen[item.HistoryID] {
			ui.Column(c).Background(termBg).Border(1, termBorder).
				Radius(4).Padding(6).Children(func() {
				aiTerminalOutput(c, item.Output, 140)
			})
		}
	})
}

func commandItemSource(source string) string {
	switch source {
	case "both":
		return i18n.T("history.source.both")
	case "favorite":
		return i18n.T("history.source.favorite")
	default:
		return i18n.T("history.source.history")
	}
}

func filterCommandItems(items []app.CommandItem, query string) []app.CommandItem {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return items
	}
	out := make([]app.CommandItem, 0, len(items))
	for _, item := range items {
		hay := strings.ToLower(item.Command + "\n" + item.Title + "\n" + item.Description)
		if strings.Contains(hay, q) {
			out = append(out, item)
		}
	}
	return out
}
