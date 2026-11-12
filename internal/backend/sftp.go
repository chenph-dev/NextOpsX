package backend

import (
	"context"

	sftppkg "nextopsx/internal/sftp"
)

// ListSFTP lists a remote directory.
func (b *Backend) ListSFTP(hostID, dir string) ([]sftppkg.Entry, error) {
	return b.sftp.List(context.Background(), hostID, dir)
}

// SFTPMkdir creates a remote directory.
func (b *Backend) SFTPMkdir(hostID, dir string) error {
	return b.sftp.Mkdir(context.Background(), hostID, dir)
}

// SFTPDelete removes a remote file or empty directory.
func (b *Backend) SFTPDelete(hostID, p string) error {
	return b.sftp.Remove(context.Background(), hostID, p)
}

// SFTPRename renames or moves a remote entry.
func (b *Backend) SFTPRename(hostID, oldPath, newPath string) error {
	return b.sftp.Rename(context.Background(), hostID, oldPath, newPath)
}

// StartSFTPUpload queues a local file upload into remoteDir.
func (b *Backend) StartSFTPUpload(hostID, localPath, remoteDir string) (string, error) {
	return b.sftp.StartUpload(context.Background(), hostID, localPath, remoteDir)
}

// StartSFTPDownload queues a remote file download into localPath.
func (b *Backend) StartSFTPDownload(hostID, remotePath, localPath string) (string, error) {
	return b.sftp.StartDownload(context.Background(), hostID, remotePath, localPath)
}

// SFTPPause pauses a transfer task.
func (b *Backend) SFTPPause(id string) error { return b.sftp.PauseTask(id) }

// SFTPResume resumes a paused transfer task.
func (b *Backend) SFTPResume(id string) error { return b.sftp.ResumeTask(id) }

// SFTPCancel cancels a transfer task, keeping its record.
func (b *Backend) SFTPCancel(id string) error { return b.sftp.CancelTask(id) }

// SFTPRemoveTask removes a finished or cancelled task record.
func (b *Backend) SFTPRemoveTask(id string) error { return b.sftp.RemoveTask(id) }

// ListSFTPTasks returns all transfer task snapshots.
func (b *Backend) ListSFTPTasks() []sftppkg.TaskInfo { return b.sftp.ListTasks() }
