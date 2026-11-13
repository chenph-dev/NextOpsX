package views

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
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
				title := "新建资产"
				switch {
				case f.IsFolder:
					title = "新建目录"
				case f.EditingID != "":
					title = "编辑资产"
				}
				ui.Text(c, title).FontSize(fontSize(17)).Bold()
				ui.Text(c, "位置："+a.HostFormParentLabel()).
					FontSize(fontSize(12)).TextColor(t.TextMuted)
			})

			nameInput := func() {
				c.Key("host-form-name")
				ui.TextInput(c, &f.Name).Grow(1).Label("名称").Placeholder("如 web-01")
			}
			protocolInput := func() {
				options := a.HostProtocolOptions()
				values := make([]string, 0, len(options))
				for _, o := range options {
					values = append(values, o.Value)
				}
				c.Key("host-form-protocol")
				sel := ui.Select(c, &f.Protocol, values).Label("协议").Grow(1)
				if sel.Changed() {
					a.SyncHostProtocol()
				}
			}

			if f.IsFolder {
				formRow(c, t, "名称", nameInput)
				if f.Err != "" {
					ui.Text(c, f.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
				}
				ui.Divider(c)
				formFooter(c, a, t, f, false)
				return
			}

			formPair(c, func() { formField(c, t, "名称", nameInput) },
				func() { formField(c, t, "协议", protocolInput) })

			if a.HostFormNeedsHost() {
				formPair(c,
					func() {
						formField(c, t, "地址", func() {
							c.Key("host-form-addr")
							ui.TextInput(c, &f.Addr).Grow(1).Label("地址").Placeholder("10.0.0.5")
						})
					},
					func() {
						formField(c, t, "端口", func() {
							c.Key("host-form-port")
							ui.TextInput(c, &f.Port).Grow(1).Label("端口").Placeholder("22")
						})
					})

				userInput := func() {
					c.Key("host-form-user")
					ui.TextInput(c, &f.User).Grow(1).Label("用户").Placeholder("root")
				}
				switch {
				case a.HostFormIsSSH():
					formPair(c,
						func() { formField(c, t, "用户", userInput) },
						func() {
							formField(c, t, "认证", func() {
								c.Key("host-form-auth")
								sel := ui.Select(c, &f.AuthLabel, []string{"密码", "私钥"}).
									Label("认证方式").Grow(1)
								if sel.Changed() {
									a.SyncHostAuth()
								}
							})
						})
				case a.HostFormIsWinRM():
					formPair(c,
						func() { formField(c, t, "用户", userInput) },
						func() {
							formField(c, t, "RDP 端口", func() {
								c.Key("host-form-rdp")
								ui.TextInput(c, &f.RdpPort).Grow(1).Label("RDP 端口").Placeholder("3389")
							})
						})
				default:
					formRow(c, t, "用户", userInput)
				}

				secretLabel := "凭据"
				secretPlaceholder := ""
				if a.HostFormIsSSH() && f.AuthType == "privatekey" {
					secretPlaceholder = "-----BEGIN ... PRIVATE KEY-----"
				}
				if f.EditingID != "" {
					if a.HostFormIsSSH() && f.AuthType == "privatekey" {
						secretPlaceholder = "留空表示保留原私钥"
					} else {
						secretPlaceholder = "留空表示保留原密码"
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
					ui.Text(c, "编辑时凭据留空表示保留原值。").
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
							ui.Button(c, "选择文件").OnClick(func() {
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

			formRow(c, t, "自动放行", func() {
				c.Key("host-form-auto")
				sel := ui.Select(c, &f.AutoApproveLabel,
					[]string{"继承全局", "自动放行", "人工审批"}).Width(220)
				if sel.Changed() {
					a.SyncHostAutoApprove()
				}
			})
			ui.Text(c, "继承全局：按审批策略；自动放行：只读命令免审批；人工审批：全部命令先确认。").
				FontSize(fontSize(12)).TextColor(t.TextMuted).Wrap().Padding(0, 0, 0, 96)

			if f.TestMsg != "" {
				color := t.TextMuted
				if f.TestOK {
					color = t.Success
				} else if !strings.HasPrefix(f.TestMsg, "正在") {
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
			test := ui.Button(c, "测试连接").Width(100)
			if f.Busy {
				test = test.Disabled(true)
			}
			test.OnClick(a.TestHostConnection)
		}
		ui.Spacer(c)
		cancel := ui.Button(c, "取消").Width(84)
		cancel.OnClick(a.CloseHostForm)
		save := ui.PrimaryButton(c, "保存").Width(84)
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
			ui.Text(c, "重命名").FontSize(fontSize(16)).Bold()
			c.Key("host-rename-name")
			ui.TextInput(c, &a.Hosts.RenameName).Label("新名称").Grow(1)
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.Hosts.RenameOpen = false })
				ui.PrimaryButton(c, "保存").OnClick(a.SaveRename)
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
			ui.Text(c, "移动到目录").FontSize(fontSize(16)).Bold()
			ui.Text(c, "选择目标目录；根目录表示移到最外层。").
				FontSize(fontSize(12)).TextColor(t.TextMuted)
			ui.Scroll(c).MaxHeight(300).Children(func() {
				ui.Column(c).Gap(4).Children(func() {
					ui.Button(c, "根目录").OnClick(func() { a.MoveNodeTo("") })
					for _, n := range a.MoveTargets() {
						id := n.ID
						ui.Button(c, n.Name).OnClick(func() { a.MoveNodeTo(id) })
					}
				})
			})
			ui.Row(c).Justify(ui.End).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.Hosts.MoveOpen = false })
			})
			if a.Hosts.Err != "" {
				ui.Text(c, a.Hosts.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
		})
	})
}

func deleteDialog(c *ui.Context, a *app.App) {
	title := "删除资产"
	message := fmt.Sprintf("确定删除「%s」吗？此操作无法撤销。", a.Hosts.DeleteName)
	if a.Hosts.DeleteIsFolder {
		title = "删除目录"
		message = fmt.Sprintf("确定删除目录「%s」吗？其中的资产和子目录也会一并删除，无法撤销。",
			a.Hosts.DeleteName)
	}
	if ui.AlertDialog(c, &a.Hosts.DeleteOpen, title, message, "取消", "确认删除") == 1 {
		a.ConfirmDeleteNode()
	}
}
