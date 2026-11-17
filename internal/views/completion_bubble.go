package views

import (
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/assets"
	"nextopsx/internal/i18n"
)

// terminalWithCompletion draws the terminal with the typed-prefix command
// bubble in the pane's top-right corner.
func terminalWithCompletion(c *ui.Context, a *app.App, t *ui.Theme) {
	// Return the keyboard to the terminal when the command picker closes, so
	// typing continues there instead of on the toolbar button.
	if a.Snippets.PickerOpen {
		a.Snippets.PickerWasOpen = true
	} else if a.Snippets.PickerWasOpen {
		a.Snippets.PickerWasOpen = false
		a.Snippets.FocusTerminal = true
	}
	ui.Box(c).Fill().Children(func() {
		term := terminal.View(c, a.Term).Fill()
		if a.Snippets.FocusTerminal {
			term = term.Focus()
			a.Snippets.FocusTerminal = false
		}
		if a.Completion.Open && !term.Focused() {
			// The bubble appearing must not take keyboard focus from the
			// terminal: OnKey needs it for Tab and arrow keys.
			term = term.Focus()
		}
		if a.Completion.Open && len(a.Completion.Items) > 0 {
			completionBubble(c, a, t)
		}
	})
}

func completionBubble(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Absolute().Right(14).Top(14).Width(560).MaxHeight(340).
		Background(t.Surface).Border(1, t.Border).Radius(12).
		Shadow(0, 10, 28, 0, ui.RGBA(0, 0, 0, 0.35)).
		Padding(8).Gap(4).Children(func() {
		ui.Row(c).Padding(4, 6, 2, 6).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("history.suggestTitle")).FontSize(fontSize(12)).Bold()
			ui.Text(c, i18n.Tf("history.suggestCount", len(a.Completion.Items))).
				FontSize(fontSize(10)).TextColor(t.TextMuted)
		})
		ui.Scroll(c).MaxHeight(250).Children(func() {
			ui.Column(c).Gap(1).Children(func() {
				for i, item := range a.Completion.Items {
					completionRow(c, a, t, item, i == a.Completion.Selected, i)
				}
			})
		})
		ui.Text(c, i18n.T("history.bubbleHint")).FontSize(fontSize(10)).
			TextColor(t.TextMuted).SingleLine().Padding(2, 6, 0, 6)
	})
}

func completionRow(c *ui.Context, a *app.App, t *ui.Theme, item app.CommandItem, selected bool, index int) {
	row := ui.Row(c).Padding(5, 8).Radius(6).Gap(8).AlignItems(ui.Center).
		Cursor(ui.CursorPointer)
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.OnClick(func() { a.AcceptCompletion(index) })
	row.Children(func() {
		icon, color := "clock", t.TextMuted
		if item.FavoriteID != "" {
			icon, color = "star-fill", t.Accent
		}
		if s := assets.SVG(icon); s != nil {
			ui.Icon(c, s).Size(11, 11).TextColor(color)
		}
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, a.CommandItemText(item)).Font("monospace").
				FontSize(fontSize(12)).SingleLine()
			if item.Title != "" {
				ui.Text(c, item.Title).FontSize(fontSize(10)).
					TextColor(t.TextMuted).SingleLine()
			}
		})
		ui.Text(c, commandItemSource(item.Source)).
			FontSize(fontSize(10)).TextColor(t.TextMuted).SingleLine()
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
