package app

import (
	"fmt"
	"os"

	"github.com/egoist/mygo"

	"nextopsx/internal/skill"
)

// maxSkillZipBytes 单个技能包文件大小上限；解压后的总量由 skill.MaxZipBytes 限制。
const maxSkillZipBytes = 20 << 20

// SkillsState is the skill management section of the settings page.
type SkillsState struct {
	Items    []*skill.Skill
	Loading  bool
	Err      string
	Notice   string
	Selected string

	Markdown        string
	MarkdownLoading bool
	MarkdownErr     string

	ConfirmDelete string

	seq   uint64
	mdSeq uint64
}

// LoadSkills refreshes the installed skills.
func (a *App) LoadSkills() {
	a.Skills.seq++
	seq := a.Skills.seq
	a.Skills.Loading = true
	a.Skills.Err = ""
	a.Skills.Notice = ""
	if a.Backend == nil {
		a.Skills.Loading = false
		a.Skills.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		items, err := b.ListSkills()
		a.Post(func() {
			if seq != a.Skills.seq {
				return
			}
			a.Skills.Loading = false
			if err != nil {
				a.Skills.Err = "读取技能失败：" + err.Error()
				return
			}
			a.Skills.Items = items
			if a.Skills.Selected != "" && !skillExists(items, a.Skills.Selected) {
				a.Skills.Selected = ""
				a.Skills.Markdown = ""
			}
		})
	}()
}

func skillExists(items []*skill.Skill, name string) bool {
	for _, s := range items {
		if s.Name == name {
			return true
		}
	}
	return false
}

// SelectSkill previews a skill's SKILL.md.
func (a *App) SelectSkill(name string) {
	a.Skills.Selected = name
	a.Skills.Markdown = ""
	a.Skills.MarkdownErr = ""
	a.Skills.mdSeq++
	seq := a.Skills.mdSeq
	if a.Backend == nil {
		a.Skills.MarkdownErr = "后端未初始化"
		return
	}
	a.Skills.MarkdownLoading = true
	b := a.Backend
	go func() {
		md, err := b.ReadSkillMarkdown(name)
		a.Post(func() {
			if seq != a.Skills.mdSeq || name != a.Skills.Selected {
				return
			}
			a.Skills.MarkdownLoading = false
			if err != nil {
				a.Skills.MarkdownErr = "读取 SKILL.md 失败：" + err.Error()
				return
			}
			a.Skills.Markdown = md
		})
	}()
}

// InstallSkillZip asks for a ZIP file and installs it.
func (a *App) InstallSkillZip() {
	if a.Backend == nil {
		a.Skills.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "选择技能 ZIP",
			ButtonLabel: "安装",
			Filters:     []mygo.FileFilter{{Name: "技能包", Extensions: []string{"zip"}}},
		})
		if err != nil {
			a.Post(func() { a.Skills.Err = "选择文件失败：" + err.Error() })
			return
		}
		if len(paths) == 0 {
			return
		}
		info, err := os.Stat(paths[0])
		if err != nil {
			a.Post(func() { a.Skills.Err = "读取文件失败：" + err.Error() })
			return
		}
		if info.Size() > maxSkillZipBytes {
			a.Post(func() {
				a.Skills.Err = fmt.Sprintf("ZIP 文件过大（上限 %d MB）", maxSkillZipBytes>>20)
			})
			return
		}
		data, err := os.ReadFile(paths[0])
		if err != nil {
			a.Post(func() { a.Skills.Err = "读取文件失败：" + err.Error() })
			return
		}
		s, err := b.InstallSkill(data)
		a.Post(func() {
			if err != nil {
				a.Skills.Err = "安装失败：" + err.Error()
				return
			}
			a.LoadSkills()
			a.SelectSkill(s.Name)
			a.Skills.Notice = "已安装技能 " + s.Name
		})
	}()
}

// SetSkillEnabled toggles a skill. The switch writes the new value into
// Items optimistically; the list is reloaded only on failure to revert it.
func (a *App) SetSkillEnabled(name string, enabled bool) {
	if a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.SetSkillEnabled(name, enabled)
		a.Post(func() {
			if err != nil {
				a.LoadSkills()
				a.Skills.Err = "更新技能失败：" + err.Error()
				return
			}
			for _, s := range a.Skills.Items {
				if s.Name == name {
					s.Enabled = enabled
				}
			}
			state := "已停用 "
			if enabled {
				state = "已启用 "
			}
			a.Skills.Err = ""
			a.Skills.Notice = state + name
		})
	}()
}

// OpenDeleteSkill asks for confirmation.
func (a *App) OpenDeleteSkill(name string) {
	if name == "" {
		return
	}
	a.Skills.ConfirmDelete = name
}

// ConfirmDeleteSkill deletes the confirmed skill.
func (a *App) ConfirmDeleteSkill() {
	name := a.Skills.ConfirmDelete
	if name == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.DeleteSkill(name)
		a.Post(func() {
			a.Skills.ConfirmDelete = ""
			if err != nil {
				a.Skills.Err = "删除技能失败：" + err.Error()
				return
			}
			if a.Skills.Selected == name {
				a.Skills.Selected = ""
				a.Skills.Markdown = ""
			}
			a.LoadSkills()
			a.Skills.Notice = "已删除技能 " + name
		})
	}()
}
