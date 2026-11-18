// Package hosts manages assets, folders and connection tests.
package hosts

import (
	"context"
	"fmt"
	"time"

	"nextopsx/domain/connector"
	"nextopsx/service/executor"
	hoststore "nextopsx/store/hosts"
)

// Service exposes asset CRUD and connection checks over the asset store.
type Service struct {
	hosts        *hoststore.HostsStore
	resolver     *executor.Resolver
	invalidateAI func()
}

// New builds the asset service. invalidateAI may be nil.
func New(hosts *hoststore.HostsStore, resolver *executor.Resolver, invalidateAI func()) *Service {
	return &Service{hosts: hosts, resolver: resolver, invalidateAI: invalidateAI}
}

// CreateHost stores a new asset and invalidates cached session graphs.
func (s *Service) CreateHost(in hoststore.HostInput) (string, error) {
	id, err := s.hosts.SaveHost(in)
	if err != nil {
		return "", err
	}
	s.invalidated()
	return id, nil
}

// UpdateHost updates an asset; an empty Secret keeps the stored credential.
func (s *Service) UpdateHost(id string, in hoststore.HostInput) error {
	if err := s.hosts.UpdateHost(id, in); err != nil {
		return err
	}
	s.resolver.Invalidate(id)
	s.invalidated()
	return nil
}

// CreateFolder creates an asset folder.
func (s *Service) CreateFolder(name, parentID string) (string, error) {
	return s.hosts.CreateFolder(name, parentID)
}

// RenameNode renames an asset or folder.
func (s *Service) RenameNode(id, name string) error {
	if err := s.hosts.RenameNode(id, name); err != nil {
		return err
	}
	s.invalidated()
	return nil
}

// MoveNode moves an asset or folder under another folder ("" is the root).
func (s *Service) MoveNode(id, parentID string) error {
	if err := s.hosts.MoveNode(id, parentID); err != nil {
		return err
	}
	s.invalidated()
	return nil
}

// DeleteNode deletes an asset or folder, cascading to its children.
func (s *Service) DeleteNode(id string) error {
	if err := s.hosts.DeleteNode(id); err != nil {
		return err
	}
	s.resolver.InvalidateAll()
	s.invalidated()
	return nil
}

// GetHostSecret returns the decrypted credential and auth type of an asset.
func (s *Service) GetHostSecret(id string) (secret, authType string, err error) {
	return s.hosts.GetHostSecret(id)
}

// TestConnection validates a not-yet-saved asset: database drivers ping
// their target, command drivers run `echo ok`. It blocks for up to 15s and
// must be called off the UI thread.
func (s *Service) TestConnection(in hoststore.HostInput) (bool, error) {
	protocol := in.Protocol
	if protocol == "" {
		protocol = connector.CommandSSH
	}
	if d := connector.Get(protocol); d != nil && d.IsDB() {
		params := in.Params
		if params == nil {
			params = map[string]any{}
		}
		cap, err := connector.New(protocol, connector.Config{
			Addr: in.Addr, Port: in.Port, User: in.User,
			Password: in.Secret, Params: params,
		})
		if err != nil {
			return false, err
		}
		pingable, ok := cap.(connector.Pingable)
		if !ok {
			return false, fmt.Errorf("连接类型 %q 不支持连接测试", protocol)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := pingable.Ping(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	ex := executor.ExecutorForHost(protocol, in.Addr, in.Port, in.User, in.AuthType, in.Secret)
	if ex == nil {
		return false, fmt.Errorf("无法解析资产连接，请确认协议与凭据")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ch, err := ex.Exec(ctx, "echo ok")
	if err != nil {
		return false, err
	}
	for ln := range ch {
		switch ln.Stream {
		case "error":
			return false, fmt.Errorf("%s", ln.Text)
		case "exit":
			return false, fmt.Errorf("命令执行失败：%s", ln.Text)
		}
	}
	return true, nil
}

// ListHosts returns the asset tree.
func (s *Service) ListHosts() ([]hoststore.TreeNode, error) {
	return s.hosts.ListTree()
}

func (s *Service) invalidated() {
	if s.invalidateAI != nil {
		s.invalidateAI()
	}
}
