package backend

import (
	"context"
	"fmt"
	"time"

	"nextopsx/internal/connector"
	"nextopsx/internal/handler/base"
	hoststore "nextopsx/internal/store/hosts"
)

// ListConnectors returns every registered connection type (ssh/winrm plus
// database drivers), for the asset form's protocol picker.
func (b *Backend) ListConnectors() []*connector.Connector {
	return connector.List()
}

// CreateHost stores a new asset and invalidates cached session graphs.
func (b *Backend) CreateHost(in hoststore.HostInput) (string, error) {
	id, err := b.hosts.SaveHost(in)
	if err != nil {
		return "", err
	}
	b.sm.InvalidateConfig()
	return id, nil
}

// UpdateHost updates an asset; an empty Secret keeps the stored credential.
func (b *Backend) UpdateHost(id string, in hoststore.HostInput) error {
	if err := b.hosts.UpdateHost(id, in); err != nil {
		return err
	}
	b.resolver.Invalidate(id)
	b.sm.InvalidateConfig()
	return nil
}

// CreateFolder creates an asset folder.
func (b *Backend) CreateFolder(name, parentID string) (string, error) {
	return b.hosts.CreateFolder(name, parentID)
}

// RenameNode renames an asset or folder.
func (b *Backend) RenameNode(id, name string) error {
	if err := b.hosts.RenameNode(id, name); err != nil {
		return err
	}
	b.sm.InvalidateConfig()
	return nil
}

// MoveNode moves an asset or folder under another folder ("" is the root).
func (b *Backend) MoveNode(id, parentID string) error {
	if err := b.hosts.MoveNode(id, parentID); err != nil {
		return err
	}
	b.sm.InvalidateConfig()
	return nil
}

// DeleteNode deletes an asset or folder, cascading to its children.
func (b *Backend) DeleteNode(id string) error {
	if err := b.hosts.DeleteNode(id); err != nil {
		return err
	}
	b.resolver.InvalidateAll()
	b.sm.InvalidateConfig()
	return nil
}

// GetHostSecret returns the decrypted credential and auth type of an asset.
func (b *Backend) GetHostSecret(id string) (secret, authType string, err error) {
	return b.hosts.GetHostSecret(id)
}

// TestConnection validates a not-yet-saved asset: database drivers ping
// their target, command drivers run `echo ok`. It blocks for up to 15s and
// must be called off the UI thread.
func (b *Backend) TestConnection(in hoststore.HostInput) (bool, error) {
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

	ex := base.ExecutorForHost(protocol, in.Addr, in.Port, in.User, in.AuthType, in.Secret)
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
