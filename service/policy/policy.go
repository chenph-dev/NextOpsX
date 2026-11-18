// Package policy reads and writes the approval policy.
package policy

import configstore "nextopsx/store/config"

// Service reads and writes the global approval policy.
type Service struct {
	policy   *configstore.PolicyStore
	onChange func()
}

// New builds the policy service. onChange may be nil.
func New(policy *configstore.PolicyStore, onChange func()) *Service {
	return &Service{policy: policy, onChange: onChange}
}
