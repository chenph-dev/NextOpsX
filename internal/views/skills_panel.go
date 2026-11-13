package views

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/skill"
)

// skillsSettings is the skill management section of the settings page:
// install, enable/disable, delete and preview SKILL.md.
func skillsSettings(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		settingsHeader(c, t, "技能", "安装 ZIP 技能包；启用后 AI 可使用其说明与脚本。", func() {
			ui.PrimaryButton(c, "安装技能 ZIP").OnClick(a.InstallSkillZip)
		})
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			skillsList(c, a, t)
			skillsDetail(c, a, t)
		})
		skillsStatusBar(c, a, t)
	})
	skillsDeleteDialog(c, a)
}

func skillsList(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(360).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("已安装 (%d)", len(a.Skills.Items))).
				FontSize(12).Bold()
		})
		ui.Divider(c)
		switch {
		case a.Skills.Loading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取技能…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.Skills.Err != "":
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				dbErrorText(c, a.Skills.Err, t)
				ui.Button(c, "重试").OnClick(a.LoadSkills)
			})
		case len(a.Skills.Items) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				ui.Text(c, "还没有安装技能").FontSize(12).TextColor(t.TextMuted)
				ui.Text(c, "技能包需要包含 SKILL.md 与可选 scripts/ 目录。").
					FontSize(11).TextColor(t.TextMuted)
			})
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, s := range a.Skills.Items {
						skillRow(c, a, s, t)
					}
				})
			})
		}
	})
}

func skillRow(c *ui.Context, a *app.App, s *skill.Skill, t *ui.Theme) {
	selected := s.Name == a.Skills.Selected
	row := ui.Row(c).Padding(6, 10).Radius(6).AlignItems(ui.Center).Gap(8)
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.OnClick(func() { a.SelectSkill(s.Name) })
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, s.Name).FontSize(12).Bold().SingleLine()
			desc := s.Description
			if desc == "" {
				desc = "无描述"
			}
			ui.Text(c, desc).FontSize(11).TextColor(t.TextMuted).SingleLine()
		})
		c.Key("skill-toggle-" + s.Name)
		sw := ui.Switch(c, &s.Enabled)
		if sw.Changed() {
			a.SetSkillEnabled(s.Name, s.Enabled)
		}
	})
}

func skillsDetail(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.Skills.Selected == "" {
		ui.Column(c).Grow(1).FillHeight().Padding(24).Center().Children(func() {
			ui.Text(c, "选择左侧技能查看 SKILL.md").FontSize(12).TextColor(t.TextMuted)
		})
		return
	}
	var current *skill.Skill
	for _, s := range a.Skills.Items {
		if s.Name == a.Skills.Selected {
			current = s
			break
		}
	}
	ui.Column(c).Grow(1).FillHeight().MinWidth(0).Padding(10, 14).Gap(8).Children(func() {
		ui.Row(c).MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.Skills.Selected).FontSize(15).Bold().SingleLine().MaxWidth(420)
			if current != nil {
				state := "已停用"
				if current.Enabled {
					state = "已启用"
				}
				ui.Text(c, state).FontSize(11).TextColor(t.TextMuted)
			}
			ui.Spacer(c)
			if current != nil {
				toggle := "启用"
				if current.Enabled {
					toggle = "停用"
				}
				ui.Button(c, toggle).FontSize(11).OnClick(func() {
					a.SetSkillEnabled(current.Name, !current.Enabled)
				})
			}
			ui.Button(c, "删除").FontSize(11).OnClick(func() {
				a.OpenDeleteSkill(a.Skills.Selected)
			})
		})
		if current != nil && current.Description != "" {
			ui.Text(c, current.Description).FontSize(12).
				TextColor(t.TextMuted).Wrap()
		}
		ui.Divider(c)
		switch {
		case a.Skills.MarkdownLoading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, "读取 SKILL.md…").FontSize(11).TextColor(t.TextMuted)
			})
		case a.Skills.MarkdownErr != "":
			dbErrorText(c, a.Skills.MarkdownErr, t)
		case strings.TrimSpace(a.Skills.Markdown) == "":
			dbHint(c, "SKILL.md 为空。", t)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Padding(2, 4).Children(func() {
					MarkdownView(c, skillMarkdownBody(a.Skills.Markdown))
				})
			})
		}
	})
}

// skillMarkdownBody strips YAML frontmatter for the preview.
func skillMarkdownBody(md string) string {
	md = strings.TrimPrefix(md, "\uFEFF")
	lines := strings.Split(md, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				return strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\r\n")
			}
		}
	}
	return md
}

func skillsStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "skills").FontSize(11).TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.Skills.Err != "" {
			ui.Text(c, a.Skills.Err).FontSize(11).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.Skills.Notice
		if text == "" {
			text = "就绪"
		}
		ui.Text(c, text).FontSize(11).TextColor(t.TextMuted).SingleLine()
	})
}

func skillsDeleteDialog(c *ui.Context, a *app.App) {
	if a.Skills.ConfirmDelete == "" {
		return
	}
	open := true
	if ui.AlertDialog(c, &open, "删除技能",
		fmt.Sprintf("确定删除技能「%s」吗？技能文件也会一并删除。", a.Skills.ConfirmDelete),
		"取消", "确认删除") == 1 {
		a.ConfirmDeleteSkill()
	}
	if !open {
		a.Skills.ConfirmDelete = ""
	}
}
