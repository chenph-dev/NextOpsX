package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/internal/assets"
	"nextopsx/internal/i18n"
)

// aiBubbleAction is one icon in a message's shared action toolbar.
type aiBubbleAction struct {
	Icon    *ui.SVG
	Label   string
	OnClick func(*ui.Context)
}

// aiBubbleToolbar renders the actions shared by every message bubble at the
// bottom of the bubble. Icons keep the copy action from taking a text button
// next to each message.
func aiBubbleToolbar(c *ui.Context, color ui.Color, actions ...aiBubbleAction) {
	if len(actions) == 0 {
		return
	}
	ui.Row(c).Justify(ui.End).Gap(4).Children(func() {
		for _, a := range actions {
			if a.Icon == nil {
				continue
			}
			action := a
			ui.Icon(c, action.Icon).Size(18, 18).TextColor(color).
				Padding(3).Cursor(ui.CursorPointer).Tooltip(action.Label).
				Label(action.Label).OnClick(func() { action.OnClick(c) })
		}
	})
}

func aiCopyAction(text string) aiBubbleAction {
	return aiBubbleAction{
		Icon:  assets.SVG("copy"),
		Label: i18n.T("ai.action.copy"),
		OnClick: func(c *ui.Context) {
			c.Services().WriteClipboard(text)
		},
	}
}
