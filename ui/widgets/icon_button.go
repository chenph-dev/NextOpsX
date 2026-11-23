package widgets

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/ui/assets"
)

// IconButton draws one compact icon-only action with a tooltip and an
// accessibility label. accent colors the icon, disabled dims it and stops
// it taking clicks.
func IconButton(c *ui.Context, t *ui.Theme, icon, label string, accent, disabled bool, onClick func()) {
	color := t.TextMuted
	if accent {
		color = t.Accent
	}
	btn := ui.ButtonBase(c).Size(26, 26).Radius(6).
		Tooltip(label).Label(label).Cursor(ui.CursorPointer).OnClick(onClick)
	if disabled {
		btn = btn.Disabled(true).Cursor(ui.CursorDefault)
		color = color.Alpha(0.35)
	} else if btn.Hovered() {
		btn = btn.Background(t.SurfaceHover)
	}
	btn.Children(func() {
		if s := assets.SVG(icon); s != nil {
			ui.Icon(c, s).Size(18, 18).TextColor(color)
		}
	})
}
