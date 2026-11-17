package views

import (
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
)

// terminalWithCompletion draws the terminal with the typed-prefix command
// bubble in the pane's top-right corner.
func terminalWithCompletion(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Box(c).Fill().Children(func() {
		term := terminal.View(c, a.Term).Fill()
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
	ui.Column(c).Absolute().Right(14).Top(14).Width(520).MaxHeight(320).
		Background(t.Surface).Border(1, t.Border).Radius(8).
		Padding(8).Gap(6).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("history.suggestTitle")).FontSize(fontSize(12)).Bold()
			ui.Spacer(c)
			ui.Text(c, i18n.Tf("history.suggestCount", len(a.Completion.Items))).
				FontSize(fontSize(10)).TextColor(t.TextMuted)
		})
		ui.Scroll(c).MaxHeight(240).Children(func() {
			ui.Column(c).Gap(1).Children(func() {
				for i, item := range a.Completion.Items {
					row := ui.Row(c).Padding(5, 8).Radius(5).Gap(8).AlignItems(ui.Center)
					if i == a.Completion.Selected {
						row = row.Background(t.SurfaceHover)
					}
					row.OnClick(func() { a.AcceptCompletion(i) })
					row.Children(func() {
						ui.Column(c).Grow(1).Children(func() {
							ui.Text(c, a.CommandItemText(item)).Font("monospace").
								FontSize(fontSize(12)).SingleLine()
							if item.Title != "" {
								ui.Text(c, item.Title).FontSize(fontSize(10)).
									TextColor(t.TextMuted).SingleLine()
							}
						})
						ui.Text(c, commandItemSource(item.Source)).
							FontSize(fontSize(10)).TextColor(t.TextMuted)
					})
				}
			})
		})
		ui.Text(c, i18n.T("history.bubbleHint")).FontSize(fontSize(10)).
			TextColor(t.TextMuted).SingleLine()
	})
}
