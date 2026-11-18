package hosts

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	"nextopsx/ui/assets"
	"nextopsx/ui/style"
	"nextopsx/ui/workspace"
)

// Dialogs renders the asset form, rename/move dialogs and the delete
// confirmation. They are overlays, so they can be built after the shell.
func Dialogs(c *ui.Context, a *workspace.State) {
	hostFormDialog(c, a)
	renameDialog(c, a)
	moveDialog(c, a)
	deleteDialog(c, a)
}

func hostFormDialog(c *ui.Context, a *workspace.State) {
	f := &a.Hosts.Form
	if f.IsFolder {
		ui.Modal(c, &f.Open, func() {
			t := c.Theme()
			ui.Column(c).Width(620).Padding(20).Gap(12).Children(func() {
				hostFormHeader(c, a, t, f)
				formRow(c, t, i18n.T("hosts.form.name"), func() {
					c.Key("host-form-name")
					ui.TextInput(c, &f.Name).Grow(1).Label(i18n.T("hosts.form.name")).
						Placeholder(i18n.T("hosts.form.namePlaceholder"))
				})
				if f.Err != "" {
					ui.Text(c, f.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
				}
				ui.Divider(c)
				formFooter(c, a, t, f, false, a.CloseHostForm)
			})
		})
		return
	}

	// The asset form is a full-window sheet: the protocol grid and the form
	// drawer share the whole content area.
	ui.DialogBase(c, &f.Open, func(back, panel ui.Element) {
		t := c.Theme()
		back.Background(ui.RGBA(0, 0, 0, 0.45))
		w, h := c.Size()
		panel.Size(w, h).Background(t.Background).Children(func() {
			ui.Column(c).Fill().Padding(24).Gap(12).Children(func() {
				hostFormHeader(c, a, t, f)
				ui.Box(c).Grow(1).FillWidth().MinHeight(0).Children(func() {
					protocolPicker(c, a, t, f)
					if f.Protocol != "" {
						hostFormDrawer(c, a, t, f)
					}
				})
			})
		})
	})
}

func hostFormHeader(c *ui.Context, a *workspace.State, t *ui.Theme, f *workspace.HostForm) {
	title := i18n.T("hosts.form.title.new")
	switch {
	case f.IsFolder:
		title = i18n.T("hosts.form.title.newFolder")
	case f.EditingID != "":
		title = i18n.T("hosts.form.title.edit")
	}
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(style.FontSize(17)).Bold()
		ui.Text(c, i18n.Tf("hosts.form.location", a.HostFormParentLabel())).
			FontSize(style.FontSize(12)).TextColor(t.TextMuted)
	})
}

func protocolPicker(c *ui.Context, a *workspace.State, t *ui.Theme, f *workspace.HostForm) {
	ui.Column(c).Fill().MinWidth(0).MinHeight(0).Gap(8).Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("hosts.form.chooseProtocol")).
				FontSize(style.FontSize(13)).Bold()
			ui.Text(c, i18n.T("hosts.form.protocolHint")).
				FontSize(style.FontSize(11)).TextColor(t.TextMuted).SingleLine()
		})
		columns := 3
		if w, _ := c.Size(); w >= 1400 {
			columns = 4
		}
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Grid(c).Columns(columns).Gap(10).Children(func() {
				for _, opt := range a.HostProtocolOptions() {
					protocolCard(c, a, t, f, opt)
				}
			})
		})
		if f.Protocol == "" {
			ui.Row(c).Justify(ui.End).Children(func() {
				ui.Button(c, i18n.T("hosts.form.cancel")).Width(84).OnClick(a.CloseHostForm)
			})
		}
	})
}

func protocolCard(c *ui.Context, a *workspace.State, t *ui.Theme,
	f *workspace.HostForm, opt workspace.ProtocolOption) {
	selected := strings.EqualFold(f.Protocol, opt.Value)
	card := ui.ButtonBase(c).Padding(10).Radius(8).Gap(10).AlignItems(ui.Center).
		Border(1, t.Border).Cursor(ui.CursorPointer).
		OnClick(func() { a.SelectHostProtocol(opt.Value) })
	if selected {
		card = card.Background(t.SurfaceHover).Border(1, t.Accent)
	} else if card.Hovered() {
		card = card.Background(t.SurfaceHover)
	}
	card.Children(func() {
		if b := assets.ProtocolBitmap(opt.Value); b != nil {
			ui.Image(c, b).Size(28, 28)
		}
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, opt.Label).FontSize(style.FontSize(14)).Bold()
			if opt.Summary != "" {
				ui.Text(c, opt.Summary).FontSize(style.FontSize(11)).
					TextColor(t.TextMuted).Wrap()
			}
		})
	})
}

func hostFormDrawer(c *ui.Context, a *workspace.State, t *ui.Theme, f *workspace.HostForm) {
	ui.Column(c).Absolute().Right(0).Top(0).Bottom(0).Width(600).
		Background(t.Background).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).
		Shadow(-10, 0, 30, 0, ui.RGBA(0, 0, 0, 0.28)).
		Padding(16, 16, 16, 20).Gap(10).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			if b := assets.ProtocolBitmap(f.Protocol); b != nil {
				ui.Image(c, b).Size(24, 24)
			}
			ui.Text(c, a.HostProtocolLabel(f.Protocol)).
				FontSize(style.FontSize(15)).Bold()
		})
		ui.Divider(c)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Gap(12).Padding(0, 0, 8, 8).Children(func() {
				hostFormFields(c, a, t, f)
			})
		})
		ui.Divider(c)
		formFooter(c, a, t, f, true, a.ClearHostProtocol)
	})
}

func hostFormFields(c *ui.Context, a *workspace.State, t *ui.Theme, f *workspace.HostForm) {
	nameInput := func() {
		c.Key("host-form-name")
		ui.TextInput(c, &f.Name).Grow(1).Label(i18n.T("hosts.form.name")).
			Placeholder(i18n.T("hosts.form.namePlaceholder"))
	}

	formRow(c, t, i18n.T("hosts.form.name"), nameInput)

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
				FontSize(style.FontSize(12)).TextColor(t.TextMuted).Padding(0, 0, 0, 100)
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
		FontSize(style.FontSize(12)).TextColor(t.TextMuted).Wrap().Padding(0, 0, 0, 100)

	if f.TestMsg != "" {
		color := t.TextMuted
		if f.TestOK {
			color = t.Success
		} else if !strings.HasPrefix(f.TestMsg, i18n.T("hosts.test.testingPrefix")) {
			color = t.Danger
		}
		ui.Text(c, f.TestMsg).FontSize(style.FontSize(13)).TextColor(color).Wrap().
			Padding(0, 0, 0, 100)
	}
	if f.Err != "" {
		ui.Text(c, f.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
	}
}

func formRow(c *ui.Context, t *ui.Theme, label string, build func()) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Width(100).FontSize(style.FontSize(13)).TextColor(t.TextMuted).SingleLine()
		build()
	})
}

// formField is one half of a two-column row. Two grow equally, so the
// controls line up between rows.
func formField(c *ui.Context, t *ui.Theme, label string, build func()) {
	ui.Row(c).Grow(1).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Width(100).FontSize(style.FontSize(13)).TextColor(t.TextMuted).SingleLine()
		build()
	})
}

func formPair(c *ui.Context, left, right func()) {
	ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
		left()
		right()
	})
}

func formFooter(c *ui.Context, a *workspace.State, t *ui.Theme, f *workspace.HostForm,
	canTest bool, onCancel func()) {
	ui.Row(c).AlignItems(ui.Center).Children(func() {
		if canTest {
			test := ui.Button(c, i18n.T("hosts.form.test")).Width(100)
			if f.Busy {
				test = test.Disabled(true)
			}
			test.OnClick(a.TestHostConnection)
		}
		ui.Spacer(c)
		ui.Row(c).Gap(8).Children(func() {
			cancel := ui.Button(c, i18n.T("hosts.form.cancel")).Width(84)
			cancel.OnClick(onCancel)
			save := ui.PrimaryButton(c, i18n.T("hosts.form.save")).Width(84)
			if f.Busy {
				save = save.Disabled(true)
			}
			save.OnClick(a.SaveHostForm)
		})
	})
}

func renameDialog(c *ui.Context, a *workspace.State) {
	ui.Modal(c, &a.Hosts.RenameOpen, func() {
		t := c.Theme()
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, i18n.T("hosts.form.renameTitle")).FontSize(style.FontSize(16)).Bold()
			c.Key("host-rename-name")
			ui.TextInput(c, &a.Hosts.RenameName).Label(i18n.T("hosts.form.newName")).Grow(1)
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, i18n.T("hosts.form.cancel")).OnClick(func() { a.Hosts.RenameOpen = false })
				ui.PrimaryButton(c, i18n.T("hosts.form.save")).OnClick(a.SaveRename)
			})
			if a.Hosts.Err != "" {
				ui.Text(c, a.Hosts.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
			}
		})
	})
}

func moveDialog(c *ui.Context, a *workspace.State) {
	ui.Modal(c, &a.Hosts.MoveOpen, func() {
		t := c.Theme()
		ui.Column(c).Width(460).Padding(18).Gap(8).Children(func() {
			ui.Text(c, i18n.T("hosts.form.moveTitle")).FontSize(style.FontSize(16)).Bold()
			ui.Text(c, i18n.T("hosts.form.moveHint")).
				FontSize(style.FontSize(12)).TextColor(t.TextMuted)
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
				ui.Text(c, a.Hosts.Err).FontSize(style.FontSize(13)).TextColor(t.Danger).Wrap()
			}
		})
	})
}

func deleteDialog(c *ui.Context, a *workspace.State) {
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
