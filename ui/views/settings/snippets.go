package settings

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	snippetsstore "nextopsx/store/snippets"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// snippetsSettings is the command snippet library section of settings.
func snippetsSettings(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.Tf("snippets.count", len(a.Snippets.Items))).
				FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			ui.PrimaryButton(c, i18n.T("snippets.new")).OnClick(a.OpenNewSnippet)
		})
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			snippetsList(c, a, t)
			snippetsEditor(c, a, t)
		})
	})
	snippetsDeleteDialog(c, a)
}

func snippetsList(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Width(340).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 10).Children(func() {
			c.Key("snippet-search")
			ui.TextInput(c, &a.Snippets.Query).
				Placeholder(i18n.T("snippets.search")).Grow(1)
		})
		ui.Divider(c)
		if a.Snippets.Err != "" {
			ui.Text(c, a.Snippets.Err).FontSize(style.FontSize(12)).TextColor(t.Danger).
				Padding(6, 10).Wrap()
		}
		items := filterSnippets(a.Snippets.Items, a.Snippets.Query)
		switch {
		case a.Snippets.Loading && len(a.Snippets.Items) == 0:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("snippets.title")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case len(items) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				if len(a.Snippets.Items) == 0 {
					ui.Text(c, i18n.T("snippets.emptyTitle")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
					ui.Text(c, i18n.T("snippets.emptyHint")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
					return
				}
				ui.Text(c, i18n.T("snippets.picker.empty")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
			})
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Padding(4).Children(func() {
					group := "\x00"
					for _, s := range items {
						g := s.Group
						if g == "" {
							g = i18n.T("snippets.picker.noGroup")
						}
						if g != group {
							group = g
							ui.Text(c, g).FontSize(style.FontSize(11)).TextColor(t.TextMuted).
								Padding(6, 8, 2, 10)
						}
						snippetListRow(c, a, s, t)
					}
				})
			})
		}
	})
}

func snippetListRow(c *ui.Context, a *workspace.State, s snippetsstore.Snippet, t *ui.Theme) {
	row := ui.Row(c).Padding(6, 10).Radius(6).Gap(8).AlignItems(ui.Center)
	if s.ID == a.Snippets.Selected {
		row = row.Background(t.SurfaceHover)
	}
	row.OnClick(func() { a.SelectSnippet(s.ID) })
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, s.Name).FontSize(style.FontSize(13)).Bold().SingleLine()
			ui.Text(c, s.Command).Font("monospace").FontSize(style.FontSize(11)).
				TextColor(t.TextMuted).SingleLine()
		})
	})
}

func snippetsEditor(c *ui.Context, a *workspace.State, t *ui.Theme) {
	if !a.Snippets.Editing {
		ui.Column(c).Grow(1).FillHeight().Padding(24).Center().Children(func() {
			ui.Text(c, i18n.T("snippets.selectHint")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
		})
		return
	}
	draft := &a.Snippets.Draft
	title := i18n.T("snippets.newTitle")
	if draft.ID != "" {
		title = i18n.T("snippets.editTitle")
	}
	ui.Column(c).Grow(1).FillHeight().MinWidth(0).Padding(10, 14).Gap(8).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, title).FontSize(style.FontSize(15)).Bold()
			ui.Spacer(c)
			if draft.ID != "" {
				ui.Button(c, i18n.T("snippets.delete")).FontSize(style.FontSize(12)).
					OnClick(func() { a.OpenDeleteSnippet(draft.ID) })
			}
			ui.Button(c, i18n.T("snippets.cancel")).FontSize(style.FontSize(12)).
				OnClick(a.CancelSnippetEdit)
			ui.PrimaryButton(c, i18n.T("snippets.save")).FontSize(style.FontSize(12)).
				OnClick(a.SaveSnippet)
		})
		ui.Divider(c)
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(10).Children(func() {
				snippetField(c, i18n.T("snippets.field.name"), func() {
					c.Key("snippet-name")
					ui.TextInput(c, &draft.Name).
						Placeholder(i18n.T("snippets.field.namePlaceholder"))
				})
				snippetField(c, i18n.T("snippets.field.group"), func() {
					c.Key("snippet-group")
					ui.TextInput(c, &draft.Group).
						Placeholder(i18n.T("snippets.field.groupPlaceholder"))
				})
				snippetField(c, i18n.T("snippets.field.command"), func() {
					c.Key("snippet-command")
					ui.TextArea(c, &draft.Command).Lines(4, 12).Font("monospace").
						FontSize(style.FontSize(12))
				})
				snippetField(c, i18n.T("snippets.field.description"), func() {
					c.Key("snippet-description")
					ui.TextInput(c, &draft.Description).
						Placeholder(i18n.T("snippets.field.descriptionPlaceholder"))
				})
				ui.Text(c, i18n.T("snippets.variables")).FontSize(style.FontSize(11)).
					TextColor(t.TextMuted).Wrap()
				if a.Selected != nil && a.Selected.NodeType == "host" &&
					strings.TrimSpace(draft.Command) != "" {
					ui.Text(c, i18n.Tf("snippets.preview",
						a.ResolveSnippetCommand(draft.Command))).
						Font("monospace").FontSize(style.FontSize(11)).
						TextColor(t.TextMuted).Wrap()
				}
				if a.Snippets.DraftErr != "" {
					ui.Text(c, a.Snippets.DraftErr).FontSize(style.FontSize(12)).TextColor(t.Danger).Wrap()
				}
			})
		})
	})
}

func snippetField(c *ui.Context, label string, control func()) {
	ui.Column(c).Gap(3).Children(func() {
		ui.Text(c, label).FontSize(style.FontSize(12)).TextColor(c.Theme().TextMuted)
		control()
	})
}

func snippetsDeleteDialog(c *ui.Context, a *workspace.State) {
	if a.Snippets.ConfirmDelete == "" {
		return
	}
	name := a.Snippets.ConfirmDelete
	for _, s := range a.Snippets.Items {
		if s.ID == a.Snippets.ConfirmDelete {
			name = s.Name
			break
		}
	}
	open := true
	if ui.AlertDialog(c, &open, i18n.T("snippets.deleteTitle"),
		i18n.Tf("snippets.deleteConfirm", name),
		i18n.T("snippets.cancel"), i18n.T("snippets.confirmDelete")) == 1 {
		a.ConfirmDeleteSnippet()
	}
	if !open {
		a.CancelDeleteSnippet()
	}
}

func filterSnippets(items []snippetsstore.Snippet, query string) []snippetsstore.Snippet {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return items
	}
	out := make([]snippetsstore.Snippet, 0, len(items))
	for _, s := range items {
		hay := strings.ToLower(s.Name + "\n" + s.Group + "\n" + s.Command + "\n" + s.Description)
		if strings.Contains(hay, q) {
			out = append(out, s)
		}
	}
	return out
}
