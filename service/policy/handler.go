package policy

import configstore "nextopsx/store/config"

func (s *Service) GetApprovalPolicy() (configstore.ApprovalPolicy, error) {
	return s.policy.GetApprovalPolicy()
}

// SaveApprovalPolicy 保存策略并触发热更新（使已构建 Graph 按新策略重建）。
func (s *Service) SaveApprovalPolicy(p configstore.ApprovalPolicy) error {
	if err := s.policy.SaveApprovalPolicy(p); err != nil {
		return err
	}
	if s.onChange != nil {
		s.onChange()
	}
	return nil
}
