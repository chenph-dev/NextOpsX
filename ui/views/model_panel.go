package views

import (
	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	configstore "nextopsx/store/config"
	"nextopsx/ui/app"
)

// modelSettings manages the AI provider list and the selected provider.
func modelSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("settings.model.providers")).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.PrimaryButton(c, i18n.T("settings.model.newProvider")).OnClick(a.NewAIProvider)
		})
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			providerList(c, a, t)
			providerEditor(c, a, t)
		})
		providerStatusBar(c, a, t)
	})
	providerDeleteDialog(c, a, t)
}

func providerList(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(280).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.Tf("settings.model.providerCount", len(a.Settings.Providers))).
				FontSize(fontSize(13)).Bold()
		})
		ui.Divider(c)
		switch {
		case a.Settings.ProvidersLoading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("settings.model.loading")).FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		case a.Settings.ProvidersErr != "" && len(a.Settings.Providers) == 0:
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				dbErrorText(c, a.Settings.ProvidersErr, t)
				ui.Button(c, i18n.T("settings.model.retry")).OnClick(a.LoadAIProviders)
			})
		case len(a.Settings.Providers) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				ui.Text(c, i18n.T("settings.model.emptyTitle")).FontSize(fontSize(13)).TextColor(t.TextMuted)
				ui.Text(c, i18n.T("settings.model.emptyHint")).
					FontSize(fontSize(12)).TextColor(t.TextMuted)
			})
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, p := range a.Settings.Providers {
						providerRow(c, a, t, p)
					}
				})
			})
		}
	})
}

func providerRow(c *ui.Context, a *app.App, t *ui.Theme, p configstore.AIProvider) {
	selected := a.Settings.SelectedProvider == p.ID
	row := ui.Row(c).Padding(6, 10).Radius(6).AlignItems(ui.Center).Gap(8)
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.OnClick(func() { a.SelectAIProvider(p.ID) })
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, p.Name).FontSize(fontSize(13)).Bold().SingleLine()
			ui.Text(c, p.Protocol+" · "+p.Model).FontSize(fontSize(12)).
				TextColor(t.TextMuted).SingleLine()
		})
		if p.IsDefault {
			ui.Text(c, i18n.T("settings.model.defaultBadge")).FontSize(fontSize(11)).TextColor(t.Accent)
		}
	})
}

func providerEditor(c *ui.Context, a *app.App, t *ui.Theme) {
	if !a.Settings.DraftIsNew && a.Settings.SelectedProvider == "" {
		ui.Column(c).Grow(1).FillHeight().Padding(24).Center().Children(func() {
			ui.Text(c, i18n.T("settings.model.selectHint")).
				FontSize(fontSize(13)).TextColor(t.TextMuted)
		})
		return
	}
	if a.Settings.DraftLoading {
		ui.Row(c).Grow(1).FillHeight().Center().Children(func() {
			ui.Spinner(c)
			ui.Text(c, i18n.T("settings.model.loading")).FontSize(fontSize(12)).TextColor(t.TextMuted)
		})
		return
	}
	prefix := "provider-new"
	if a.Settings.SelectedProvider != "" {
		prefix = "provider-" + a.Settings.SelectedProvider
	}
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Padding(16, 20).Gap(10).MaxWidth(720).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				title := a.Settings.Draft.Name
				if title == "" {
					title = i18n.T("settings.model.newProviderTitle")
				}
				ui.Text(c, title).FontSize(fontSize(13)).Bold()
				if a.Settings.Draft.IsDefault {
					ui.Text(c, i18n.T("settings.model.defaultBadge")).FontSize(fontSize(11)).TextColor(t.Accent)
				}
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, i18n.T("settings.model.name")).Width(72).FontSize(fontSize(13)).TextColor(t.TextMuted)
				c.Key(prefix + "-name")
				ui.TextInput(c, &a.Settings.Draft.Name).Grow(1).
					Placeholder(i18n.T("settings.model.namePlaceholder"))
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, i18n.T("settings.model.protocol")).Width(72).FontSize(fontSize(13)).TextColor(t.TextMuted)
				c.Key(prefix + "-protocol")
				ui.Select(c, &a.Settings.Draft.Protocol, []string{"openai", "claude"})
				ui.Text(c, i18n.T("settings.model.model")).Width(36).FontSize(fontSize(13)).TextColor(t.TextMuted)
				c.Key(prefix + "-model")
				ui.TextInput(c, &a.Settings.Draft.Model).Grow(1).
					Placeholder(i18n.T("settings.model.modelPlaceholder"))
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Base URL").Width(72).FontSize(fontSize(13)).TextColor(t.TextMuted)
				c.Key(prefix + "-baseurl")
				ui.TextInput(c, &a.Settings.Draft.BaseURL).Grow(1).
					Placeholder(i18n.T("settings.model.baseURLPlaceholder"))
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "API Key").Width(72).FontSize(fontSize(13)).TextColor(t.TextMuted)
				c.Key(prefix + "-key")
				ui.TextInput(c, &a.Settings.Draft.APIKey).Grow(1).Password()
			})
			ui.Text(c, i18n.T("settings.model.protocolHint")).
				FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap()
			if a.Settings.ProvidersErr != "" {
				ui.Text(c, a.Settings.ProvidersErr).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Row(c).Justify(ui.End).Gap(6).AlignItems(ui.Center).Children(func() {
				if a.Settings.DraftIsNew {
					ui.Button(c, i18n.T("settings.model.cancel")).OnClick(a.CancelNewAIProvider)
				} else {
					if !a.Settings.Draft.IsDefault {
						ui.Button(c, i18n.T("settings.model.setDefault")).OnClick(func() {
							a.SetDefaultAIProvider(a.Settings.SelectedProvider)
						})
					}
					ui.Button(c, i18n.T("settings.model.delete")).OnClick(func() {
						a.OpenDeleteAIProvider(a.Settings.SelectedProvider)
					})
				}
				save := ui.PrimaryButton(c, i18n.T("settings.model.save")).OnClick(a.SaveAIProvider)
				if a.Settings.DraftSaving {
					save = save.Disabled(true)
				}
			})
		})
	})
}

func providerStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.T("settings.model.model")).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.Settings.ProvidersErr != "" {
			ui.Text(c, a.Settings.ProvidersErr).FontSize(fontSize(12)).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.Settings.Notice
		if text == "" {
			text = i18n.T("settings.model.ready")
		}
		ui.Text(c, text).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
	})
}

func providerDeleteDialog(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Settings.ConfirmDelete == "" {
		return
	}
	name := a.Settings.ConfirmDelete
	for _, p := range a.Settings.Providers {
		if p.ID == a.Settings.ConfirmDelete {
			name = p.Name
			break
		}
	}
	open := true
	if ui.AlertDialog(c, &open, i18n.T("settings.model.deleteTitle"),
		i18n.Tf("settings.model.deleteConfirm", name),
		i18n.T("settings.model.cancel"), i18n.T("settings.model.confirmDelete")) == 1 {
		a.ConfirmDeleteAIProvider()
	}
	if !open {
		a.Settings.ConfirmDelete = ""
	}
}
