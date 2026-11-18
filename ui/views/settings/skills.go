package settings

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/skill"
	"nextopsx/i18n"
	"nextopsx/ui/markdown"
	"nextopsx/ui/style"
	"nextopsx/ui/widgets"
	"nextopsx/ui/workspace"
)

// skillsSettings is the skill management section of the settings page:
// install, enable/disable, delete and preview SKILL.md.
func skillsSettings(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		ui.Row(c).Padding(8, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("settings.skills.installed")).FontSize(style.FontSize(13)).Bold()
			ui.Spacer(c)
			ui.PrimaryButton(c, i18n.T("settings.skills.installZip")).OnClick(a.InstallSkillZip)
		})
		ui.Divider(c)
		ui.Row(c).Grow(1).FillHeight().MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			skillsList(c, a, t)
			skillsDetail(c, a, t)
		})
		skillsStatusBar(c, a, t)
	})
	skillsDeleteDialog(c, a)
}

func skillsList(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Column(c).Width(360).FillHeight().Background(t.Surface).
		Border(1, t.Border).Children(func() {
		ui.Row(c).Padding(6, 10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.Tf("settings.skills.installedCount", len(a.Skills.Items))).
				FontSize(style.FontSize(13)).Bold()
		})
		ui.Divider(c)
		switch {
		case a.Skills.Loading:
			ui.Row(c).Padding(16).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("settings.skills.loading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.Skills.Err != "":
			ui.Column(c).Padding(8).Gap(6).Children(func() {
				widgets.ErrorText(c, a.Skills.Err, t)
				ui.Button(c, i18n.T("settings.skills.retry")).OnClick(a.LoadSkills)
			})
		case len(a.Skills.Items) == 0:
			ui.Column(c).Padding(16).Gap(6).Children(func() {
				ui.Text(c, i18n.T("settings.skills.emptyTitle")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
				ui.Text(c, i18n.T("settings.skills.emptyHint")).
					FontSize(style.FontSize(12)).TextColor(t.TextMuted)
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

func skillRow(c *ui.Context, a *workspace.State, s *skill.Skill, t *ui.Theme) {
	selected := s.Name == a.Skills.Selected
	row := ui.Row(c).Padding(6, 10).Radius(6).AlignItems(ui.Center).Gap(8)
	if selected {
		row = row.Background(t.SurfaceHover)
	}
	row.OnClick(func() { a.SelectSkill(s.Name) })
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, s.Name).FontSize(style.FontSize(13)).Bold().SingleLine()
			desc := s.Description
			if desc == "" {
				desc = i18n.T("settings.skills.noDescription")
			}
			ui.Text(c, desc).FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
		})
		c.Key("skill-toggle-" + s.Name)
		sw := ui.Switch(c, &s.Enabled)
		if sw.Changed() {
			a.SetSkillEnabled(s.Name, s.Enabled)
		}
	})
}

func skillsDetail(c *ui.Context, a *workspace.State, t *ui.Theme) {
	if a.Skills.Selected == "" {
		ui.Column(c).Grow(1).FillHeight().Padding(24).Center().Children(func() {
			ui.Text(c, i18n.T("settings.skills.selectHint")).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
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
			ui.Text(c, a.Skills.Selected).FontSize(style.FontSize(15)).Bold().SingleLine().MaxWidth(420)
			if current != nil {
				state := i18n.T("settings.skills.disabled")
				if current.Enabled {
					state = i18n.T("settings.skills.enabled")
				}
				ui.Text(c, state).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			}
			ui.Spacer(c)
			if current != nil {
				toggle := i18n.T("settings.skills.enable")
				if current.Enabled {
					toggle = i18n.T("settings.skills.disable")
				}
				ui.Button(c, toggle).FontSize(style.FontSize(12)).OnClick(func() {
					a.SetSkillEnabled(current.Name, !current.Enabled)
				})
			}
			ui.Button(c, i18n.T("settings.skills.delete")).FontSize(style.FontSize(12)).OnClick(func() {
				a.OpenDeleteSkill(a.Skills.Selected)
			})
		})
		if current != nil && current.Description != "" {
			ui.Text(c, current.Description).FontSize(style.FontSize(13)).
				TextColor(t.TextMuted).Wrap()
		}
		ui.Divider(c)
		switch {
		case a.Skills.MarkdownLoading:
			ui.Row(c).Padding(20).Center().Children(func() {
				ui.Spinner(c)
				ui.Text(c, i18n.T("settings.skills.markdownLoading")).FontSize(style.FontSize(12)).TextColor(t.TextMuted)
			})
		case a.Skills.MarkdownErr != "":
			widgets.ErrorText(c, a.Skills.MarkdownErr, t)
		case strings.TrimSpace(a.Skills.Markdown) == "":
			widgets.Hint(c, i18n.T("settings.skills.markdownEmpty"), t)
		default:
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Padding(2, 4).Children(func() {
					markdown.View(c, SkillMarkdownBody(a.Skills.Markdown))
				})
			})
		}
	})
}

// skillMarkdownBody strips YAML frontmatter for the preview.
func SkillMarkdownBody(md string) string {
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

func skillsStatusBar(c *ui.Context, a *workspace.State, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "skills").FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.Skills.Err != "" {
			ui.Text(c, a.Skills.Err).FontSize(style.FontSize(12)).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.Skills.Notice
		if text == "" {
			text = i18n.T("settings.skills.ready")
		}
		ui.Text(c, text).FontSize(style.FontSize(12)).TextColor(t.TextMuted).SingleLine()
	})
}

func skillsDeleteDialog(c *ui.Context, a *workspace.State) {
	if a.Skills.ConfirmDelete == "" {
		return
	}
	open := true
	if ui.AlertDialog(c, &open, i18n.T("settings.skills.deleteTitle"),
		i18n.Tf("settings.skills.deleteConfirm", a.Skills.ConfirmDelete),
		i18n.T("settings.skills.cancel"), i18n.T("settings.skills.confirmDelete")) == 1 {
		a.ConfirmDeleteSkill()
	}
	if !open {
		a.Skills.ConfirmDelete = ""
	}
}
