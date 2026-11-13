package backend

import "nextopsx/internal/skill"

// ListSkills returns all installed skills, sorted by name.
func (b *Backend) ListSkills() ([]*skill.Skill, error) {
	return b.skills.List()
}

// InstallSkill installs a skill ZIP (SKILL.md + files) and invalidates AI
// session graphs so the new catalog is picked up.
func (b *Backend) InstallSkill(zipBytes []byte) (*skill.Skill, error) {
	s, err := b.skills.Install(zipBytes)
	if err != nil {
		return nil, err
	}
	b.sm.InvalidateConfig()
	return s, nil
}

// SetSkillEnabled enables or disables a skill.
func (b *Backend) SetSkillEnabled(name string, enabled bool) error {
	if err := b.skills.SetEnabled(name, enabled); err != nil {
		return err
	}
	b.sm.InvalidateConfig()
	return nil
}

// DeleteSkill removes a skill's metadata and files.
func (b *Backend) DeleteSkill(name string) error {
	if err := b.skills.Delete(name); err != nil {
		return err
	}
	b.sm.InvalidateConfig()
	return nil
}

// ReadSkillMarkdown returns a skill's SKILL.md.
func (b *Backend) ReadSkillMarkdown(name string) (string, error) {
	s, err := b.skills.Lookup(name)
	if err != nil {
		return "", err
	}
	return b.skills.ReadMarkdown(s)
}
