// Package sftp exposes remote file browsing and transfer tasks.
package sftp

import (
	"context"

	sftppkg "nextopsx/domain/sftp"
)

// Service exposes remote file browsing and transfer tasks.
type Service struct {
	manager *sftppkg.Manager
}

// New builds the SFTP service.
func New(manager *sftppkg.Manager) *Service {
	return &Service{manager: manager}
}

// ListSFTP lists a remote directory.
func (s *Service) ListSFTP(hostID, dir string) ([]sftppkg.Entry, error) {
	return s.manager.List(context.Background(), hostID, dir)
}

// SFTPMkdir creates a remote directory.
func (s *Service) SFTPMkdir(hostID, dir string) error {
	return s.manager.Mkdir(context.Background(), hostID, dir)
}

// SFTPDelete removes a remote file or empty directory.
func (s *Service) SFTPDelete(hostID, p string) error {
	return s.manager.Remove(context.Background(), hostID, p)
}

// SFTPRename renames or moves a remote entry.
func (s *Service) SFTPRename(hostID, oldPath, newPath string) error {
	return s.manager.Rename(context.Background(), hostID, oldPath, newPath)
}

// StartSFTPUpload queues a local file upload into remoteDir.
func (s *Service) StartSFTPUpload(hostID, localPath, remoteDir string) (string, error) {
	return s.manager.StartUpload(context.Background(), hostID, localPath, remoteDir)
}

// StartSFTPDownload queues a remote file download into localPath.
func (s *Service) StartSFTPDownload(hostID, remotePath, localPath string) (string, error) {
	return s.manager.StartDownload(context.Background(), hostID, remotePath, localPath)
}

// SFTPPause pauses a transfer task.
func (s *Service) SFTPPause(id string) error { return s.manager.PauseTask(id) }

// SFTPResume resumes a paused transfer task.
func (s *Service) SFTPResume(id string) error { return s.manager.ResumeTask(id) }

// SFTPCancel cancels a transfer task, keeping its record.
func (s *Service) SFTPCancel(id string) error { return s.manager.CancelTask(id) }

// SFTPRemoveTask removes a finished or cancelled task record.
func (s *Service) SFTPRemoveTask(id string) error { return s.manager.RemoveTask(id) }

// ListSFTPTasks returns all transfer task snapshots.
func (s *Service) ListSFTPTasks() []sftppkg.TaskInfo { return s.manager.ListTasks() }

// Close stops the transfer engine and its SSH clients.
func (s *Service) Close() {
	if s != nil && s.manager != nil {
		s.manager.Close()
	}
}
