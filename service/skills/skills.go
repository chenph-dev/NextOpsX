// Package skills manages installed ops skills.
package skills

import "nextopsx/domain/skill"

// Service manages installed ops skills.
type Service struct {
	manager      *skill.Manager
	invalidateAI func()
}

// New builds the skill service. invalidateAI may be nil.
func New(manager *skill.Manager, invalidateAI func()) *Service {
	return &Service{manager: manager, invalidateAI: invalidateAI}
}

// ListSkills returns all installed skills, sorted by name.
func (s *Service) ListSkills() ([]*skill.Skill, error) {
	return s.manager.List()
}

// InstallSkill installs a skill ZIP (SKILL.md + files) and invalidates AI
// session graphs so the new catalog is picked up.
func (s *Service) InstallSkill(zipBytes []byte) (*skill.Skill, error) {
	sk, err := s.manager.Install(zipBytes)
	if err != nil {
		return nil, err
	}
	s.invalidated()
	return sk, nil
}

// SetSkillEnabled enables or disables a skill.
func (s *Service) SetSkillEnabled(name string, enabled bool) error {
	if err := s.manager.SetEnabled(name, enabled); err != nil {
		return err
	}
	s.invalidated()
	return nil
}

// DeleteSkill removes a skill's metadata and files.
func (s *Service) DeleteSkill(name string) error {
	if err := s.manager.Delete(name); err != nil {
		return err
	}
	s.invalidated()
	return nil
}

// ReadSkillMarkdown returns a skill's SKILL.md.
func (s *Service) ReadSkillMarkdown(name string) (string, error) {
	sk, err := s.manager.Lookup(name)
	if err != nil {
		return "", err
	}
	return s.manager.ReadMarkdown(sk)
}

func (s *Service) invalidated() {
	if s.invalidateAI != nil {
		s.invalidateAI()
	}
}
