package backend

import configstore "nextopsx/internal/store/config"

// GetApprovalPolicy returns the global approval policy.
func (b *Backend) GetApprovalPolicy() (configstore.ApprovalPolicy, error) {
	return b.approval.GetApprovalPolicy()
}

// SaveApprovalPolicy stores the global approval policy and invalidates
// cached agent graphs so the next turn uses it.
func (b *Backend) SaveApprovalPolicy(p configstore.ApprovalPolicy) error {
	return b.approval.SaveApprovalPolicy(p)
}
