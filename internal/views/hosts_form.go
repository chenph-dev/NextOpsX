package views

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
)

// HostsDialogs renders the asset form, rename/move dialogs and the delete
// confirmation. They are overlays, so they can be built after the shell.
func HostsDialogs(c *ui.Context, a *app.App) {
	hostFormDialog(c, a)
	renameDialog(c, a)
	moveDialog(c, a)
	deleteDialog(c, a)
}

func hostFormDialog(c *ui.Context, a *app.App) {
	f := &a.Hosts.Form
	ui.Modal(c, &f.Open, func() {
		t := c.Theme()
		ui.Column(c).Width(620).Padding(20).Gap(12).Children(func() {
			ui.Column(c).Gap(2).Children(func() {
				title := i18n.T("hosts.form.title.new")
				switch {
				case f.IsFolder:
					title = i18n.T("hosts.form.title.newFolder")
				case f.EditingID != "":
					title = i18n.T("hosts.form.title.edit")
				}
				ui.Text(c, title).FontSize(fontSize(17)).Bold()
				ui.Text(c, i18n.Tf("hosts.form.location", a.HostFormParentLabel())).
					FontSize(fontSize(12)).TextColor(t.TextMuted)
			})

			nameInput := func() {
				c.Key("host-form-name")
				ui.TextInput(c, &f.Name).Grow(1).Label(i18n.T("hosts.form.name")).
					Placeholder(i18n.T("hosts.form.namePlaceholder"))
			}
			protocolInput := func() {
				options := a.HostProtocolOptions()
				values := make([]string, 0, len(options))
				for _, o := range options {
					values = append(values, o.Value)
				}
				c.Key("host-form-protocol")
				sel := ui.Select(c, &f.Protocol, values).Label(i18n.T("hosts.form.protocol")).Grow(1)
				if sel.Changed() {
					a.SyncHostProtocol()
				}
			}

			if f.IsFolder {
				formRow(c, t, i18n.T("hosts.form.name"), nameInput)
				if f.Err != "" {
					ui.Text(c, f.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
				}
				ui.Divider(c)
				formFooter(c, a, t, f, false)
				return
			}

			formPair(c, func() { formField(c, t, i18n.T("hosts.form.name"), nameInput) },
				func() { formField(c, t, i18n.T("hosts.form.protocol"), protocolInput) })

			if a.HostFormNeedsHost() {
				formPair(c,
					func() {
						formField(c, t, i18n.T("hosts.form.address"), func() {
							c.Key("host-form-addr")
							ui.TextInput(c, &f.Addr).Grow(1).Label(i18n.T("hosts.form.address")).
								Placeholder("10.0.0.5")
						})
					},
					func() {
						formField(c, t, i18n.T("hosts.form.port"), func() {
							c.Key("host-form-port")
							ui.TextInput(c, &f.Port).Grow(1).Label(i18n.T("hosts.form.port")).
								Placeholder("22")
						})
					})

				userInput := func() {
					c.Key("host-form-user")
					ui.TextInput(c, &f.User).Grow(1).Label(i18n.T("hosts.form.user")).Placeholder("root")
				}
				switch {
				case a.HostFormIsSSH():
					formPair(c,
						func() { formField(c, t, i18n.T("hosts.form.user"), userInput) },
						func() {
							formField(c, t, i18n.T("hosts.form.auth"), func() {
								c.Key("host-form-auth")
								// Picker labels are display-only; SyncHostAuth maps either
								// language back to the canonical AuthType.
								sel := ui.Select(c, &f.AuthLabel, []string{
									i18n.T("hosts.auth.password"), i18n.T("hosts.auth.privateKey"),
								}).
									Label(i18n.T("hosts.form.authMethod")).Grow(1)
								if sel.Changed() {
									a.SyncHostAuth()
								}
							})
						})
				case a.HostFormIsWinRM():
					formPair(c,
						func() { formField(c, t, i18n.T("hosts.form.user"), userInput) },
						func() {
							formField(c, t, i18n.T("hosts.form.rdpPort"), func() {
								c.Key("host-form-rdp")
								ui.TextInput(c, &f.RdpPort).Grow(1).Label(i18n.T("hosts.form.rdpPort")).
									Placeholder("3389")
							})
						})
				default:
					formRow(c, t, i18n.T("hosts.form.user"), userInput)
				}

				secretLabel := i18n.T("hosts.form.credential")
				secretPlaceholder := ""
				if a.HostFormIsSSH() && f.AuthType == "privatekey" {
					secretPlaceholder = "-----BEGIN ... PRIVATE KEY-----"
				}
				if f.EditingID != "" {
					if a.HostFormIsSSH() && f.AuthType == "privatekey" {
						secretPlaceholder = i18n.T("hosts.form.keepPrivateKey")
					} else {
						secretPlaceholder = i18n.T("hosts.form.keepPassword")
					}
				}
				formRow(c, t, secretLabel, func() {
					c.Key("host-form-secret")
					if a.HostFormIsSSH() && f.AuthType == "privatekey" {
						ui.TextArea(c, &f.Secret).Grow(1).Lines(2, 6).Font("monospace").
							Label(secretLabel).Placeholder(secretPlaceholder)
					} else {
						ui.TextInput(c, &f.Secret).Grow(1).Password().Label(secretLabel).
							Placeholder(secretPlaceholder)
					}
				})
				if f.EditingID != "" {
					ui.Text(c, i18n.T("hosts.form.editKeepHint")).
						FontSize(fontSize(12)).TextColor(t.TextMuted).Padding(0, 0, 0, 96)
				}
			}

			for i := range f.Params {
				p := &f.Params[i]
				label := p.Label
				if p.Required {
					label += " *"
				}
				index := i
				formRow(c, t, label, func() {
					if p.File {
						ui.Row(c).Gap(8).Children(func() {
							c.Key("host-param-" + p.Key)
							ui.TextInput(c, &p.Value).Grow(1).Label(p.Label).
								Placeholder(p.Placeholder)
							ui.Button(c, i18n.T("hosts.form.chooseFile")).OnClick(func() {
								a.PickHostParamFile(index)
							})
						})
						return
					}
					c.Key("host-param-" + p.Key)
					ui.TextInput(c, &p.Value).Grow(1).Label(p.Label).
						Placeholder(p.Placeholder)
				})
			}

			formRow(c, t, i18n.T("hosts.form.autoApprove"), func() {
				c.Key("host-form-auto")
				// See the auth picker: labels are display-only and mapped
				// back to the canonical value by SyncHostAutoApprove.
				sel := ui.Select(c, &f.AutoApproveLabel,
					[]string{
						i18n.T("hosts.auto.inherit"), i18n.T("hosts.auto.allow"), i18n.T("hosts.auto.manual"),
					}).Width(220)
				if sel.Changed() {
					a.SyncHostAutoApprove()
				}
			})
			ui.Text(c, i18n.T("hosts.form.autoApproveHint")).
				FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap().Padding(0, 0, 0, 96)

			if f.TestMsg != "" {
				color := t.TextMuted
				if f.TestOK {
					color = t.Success
				} else if !strings.HasPrefix(f.TestMsg, i18n.T("hosts.test.testingPrefix")) {
					color = t.Danger
				}
				ui.Text(c, f.TestMsg).FontSize(fontSize(13)).TextColor(color).Wrap().
					Padding(0, 0, 0, 96)
			}
			if f.Err != "" {
				ui.Text(c, f.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Divider(c)
			formFooter(c, a, t, f, true)
		})
	})
}

func formRow(c *ui.Context, t *ui.Theme, label string, build func()) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Width(88).FontSize(fontSize(13)).TextColor(t.TextMuted).SingleLine()
		build()
	})
}

// formField is one half of a two-column row. Two grow equally, so the
// controls line up between rows.
func formField(c *ui.Context, t *ui.Theme, label string, build func()) {
	ui.Row(c).Grow(1).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Width(88).FontSize(fontSize(13)).TextColor(t.TextMuted).SingleLine()
		build()
	})
}

func formPair(c *ui.Context, left, right func()) {
	ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
		left()
		right()
	})
}

func formFooter(c *ui.Context, a *app.App, t *ui.Theme, f *app.HostForm, canTest bool) {
	ui.Row(c).AlignItems(ui.Center).Children(func() {
		if canTest {
			test := ui.Button(c, i18n.T("hosts.form.test")).Width(100)
			if f.Busy {
				test = test.Disabled(true)
			}
			test.OnClick(a.TestHostConnection)
		}
		ui.Spacer(c)
		cancel := ui.Button(c, i18n.T("hosts.form.cancel")).Width(84)
		cancel.OnClick(a.CloseHostForm)
		save := ui.PrimaryButton(c, i18n.T("hosts.form.save")).Width(84)
		if f.Busy {
			save = save.Disabled(true)
		}
		save.OnClick(a.SaveHostForm)
	})
}

func renameDialog(c *ui.Context, a *app.App) {
	ui.Modal(c, &a.Hosts.RenameOpen, func() {
		t := c.Theme()
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, i18n.T("hosts.form.renameTitle")).FontSize(fontSize(16)).Bold()
			c.Key("host-rename-name")
			ui.TextInput(c, &a.Hosts.RenameName).Label(i18n.T("hosts.form.newName")).Grow(1)
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, i18n.T("hosts.form.cancel")).OnClick(func() { a.Hosts.RenameOpen = false })
				ui.PrimaryButton(c, i18n.T("hosts.form.save")).OnClick(a.SaveRename)
			})
			if a.Hosts.Err != "" {
				ui.Text(c, a.Hosts.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
		})
	})
}

func moveDialog(c *ui.Context, a *app.App) {
	ui.Modal(c, &a.Hosts.MoveOpen, func() {
		t := c.Theme()
		ui.Column(c).Width(460).Padding(18).Gap(8).Children(func() {
			ui.Text(c, i18n.T("hosts.form.moveTitle")).FontSize(fontSize(16)).Bold()
			ui.Text(c, i18n.T("hosts.form.moveHint")).
				FontSize(fontSize(12)).TextColor(t.TextMuted)
			ui.Scroll(c).MaxHeight(300).Children(func() {
				ui.Column(c).Gap(4).Children(func() {
					ui.Button(c, i18n.T("hosts.root")).OnClick(func() { a.MoveNodeTo("") })
					for _, n := range a.MoveTargets() {
						id := n.ID
						ui.Button(c, n.Name).OnClick(func() { a.MoveNodeTo(id) })
					}
				})
			})
			ui.Row(c).Justify(ui.End).Children(func() {
				ui.Button(c, i18n.T("hosts.form.cancel")).OnClick(func() { a.Hosts.MoveOpen = false })
			})
			if a.Hosts.Err != "" {
				ui.Text(c, a.Hosts.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
		})
	})
}

func deleteDialog(c *ui.Context, a *app.App) {
	title := i18n.T("hosts.form.deleteHostTitle")
	message := i18n.Tf("hosts.form.deleteHostMessage", a.Hosts.DeleteName)
	if a.Hosts.DeleteIsFolder {
		title = i18n.T("hosts.form.deleteFolderTitle")
		message = i18n.Tf("hosts.form.deleteFolderMessage", a.Hosts.DeleteName)
	}
	if ui.AlertDialog(c, &a.Hosts.DeleteOpen, title, message,
		i18n.T("hosts.form.cancel"), i18n.T("hosts.form.deleteConfirm")) == 1 {
		a.ConfirmDeleteNode()
	}
}
