// Package terminal adapts interactive SSH sessions to the native terminal.
package terminal

import (
	"context"
	"fmt"
	"io"

	"nextopsx/domain/sshexec"
	hoststore "nextopsx/store/hosts"
)

// Service opens interactive channels for SSH assets.
type Service struct {
	hosts *hoststore.HostsStore
}

// New builds the terminal service from the asset store.
func New(hosts *hoststore.HostsStore) *Service {
	return &Service{hosts: hosts}
}

// OpenTerminalConn opens an interactive SSH session for hostID and adapts
// it to the terminal plugin's io.ReadWriteCloser. Only SSH assets support
// an interactive terminal.
func (s *Service) OpenTerminalConn(hostID string, cols, rows int) (io.ReadWriteCloser, error) {
	meta, err := s.hosts.HostMetaByID(hostID)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, fmt.Errorf("资产不存在")
	}
	proto := meta.Protocol
	if proto != "" && proto != "ssh" {
		return nil, fmt.Errorf("仅 SSH 资产支持交互式终端（当前：%s）", proto)
	}
	secret, authType, err := s.hosts.GetHostSecret(hostID)
	if err != nil {
		return nil, fmt.Errorf("读取凭据失败: %w", err)
	}
	host := sshexec.Host{
		Addr: meta.Addr, Port: meta.Port, User: meta.User,
		AuthType: authType, Secret: secret,
	}
	// TOFU: pin the host key on first connect, refuse a changed key.
	host.TrustHostKey = func(fp string) error {
		cur, err := s.hosts.HostKeyFingerprint(hostID)
		if err != nil {
			return err
		}
		if cur == "" {
			return s.hosts.SaveHostKeyFingerprint(hostID, fp)
		}
		if cur != fp {
			return fmt.Errorf("主机密钥已变更（%s → %s），可能遭中间人替换", cur, fp)
		}
		return nil
	}

	sess, err := sshexec.OpenSession(context.Background(), host, cols, rows)
	if err != nil {
		return nil, err
	}
	return newSSHConn(sess), nil
}
