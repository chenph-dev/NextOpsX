package app

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	sftppkg "nextopsx/internal/sftp"
	hoststore "nextopsx/internal/store/hosts"
)

// SFTP task tabs, in display order.
const (
	SFTPTasksRunning = iota
	SFTPTasksQueued
	SFTPTasksPaused
	SFTPTasksCancelled
	SFTPTasksDone
)

// SFTPState is the remote file browser and transfer panel state.
type SFTPState struct {
	HostID   string
	HostName string
	Path     string
	Entries  []sftppkg.Entry
	Loading  bool
	LoadErr  string
	Err      string
	Notice   string

	SelectedRow int
	Grid        ui.ListState

	Tasks         []sftppkg.TaskInfo
	TaskPanelOpen bool
	TaskTab       int

	NewFolderOpen bool
	NewFolderName string
	RenameOpen    bool
	RenameName    string
	RenamePath    string
	ConfirmDelete bool
	DeletePath    string

	listSeq uint64
}

// EnsureSFTP attaches SFTP for the selected SSH asset when needed.
func (a *App) EnsureSFTP() {
	n := a.Selected
	if n == nil || n.NodeType != "host" || !a.IsSSHProtocol(n.Protocol) {
		return
	}
	if a.SFTP.HostID != n.ID {
		a.AttachSFTPFor(n)
		return
	}
	if !a.SFTP.Loading {
		a.SFTPRefresh()
	}
}

// AttachSFTPFor loads the root of an SSH asset's filesystem.
func (a *App) AttachSFTPFor(n *hoststore.TreeNode) {
	if n == nil || n.NodeType != "host" || !a.IsSSHProtocol(n.Protocol) {
		return
	}
	if a.SFTP.HostID == n.ID {
		a.SFTP.HostName = n.Name
		return
	}
	tasks := a.SFTP.Tasks
	a.SFTP = SFTPState{
		HostID: n.ID, HostName: n.Name, Path: "/", Loading: true,
		SelectedRow: -1, Tasks: tasks, TaskTab: SFTPTasksRunning,
	}
	a.loadSFTPDir(n.ID, "/")
	a.SFTPRefreshTasks()
}

func (a *App) loadSFTPDir(hostID, dir string) {
	a.SFTP.listSeq++
	seq := a.SFTP.listSeq
	a.SFTP.Loading = true
	a.SFTP.LoadErr = ""
	if a.Backend == nil {
		a.SFTP.Loading = false
		a.SFTP.LoadErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		entries, err := b.ListSFTP(hostID, dir)
		a.Post(func() {
			if hostID != a.SFTP.HostID || seq != a.SFTP.listSeq {
				return
			}
			a.SFTP.Loading = false
			if err != nil {
				a.SFTP.LoadErr = "列目录失败：" + err.Error()
				return
			}
			a.SFTP.Entries = entries
			a.SFTP.Path = dir
			a.SFTP.SelectedRow = -1
		})
	}()
}

// SFTPRefresh reloads the current directory.
func (a *App) SFTPRefresh() {
	if a.SFTP.HostID == "" {
		return
	}
	a.SFTP.Notice = ""
	a.SFTP.Err = ""
	a.loadSFTPDir(a.SFTP.HostID, a.SFTP.Path)
}

// SFTPParentPath returns the parent of the current remote path.
func (a *App) SFTPParentPath() string {
	p := strings.TrimRight(a.SFTP.Path, "/")
	if p == "" {
		return "/"
	}
	parent := path.Dir(p)
	if parent == "." {
		return "/"
	}
	return parent
}

// SFTPGoParent navigates one directory up.
func (a *App) SFTPGoParent() {
	if a.SFTP.Path == "/" {
		return
	}
	a.loadSFTPDir(a.SFTP.HostID, a.SFTPParentPath())
}

// SFTPSelectedEntry returns the selected entry, if any.
func (a *App) SFTPSelectedEntry() *sftppkg.Entry {
	i := a.SFTP.SelectedRow
	if i < 0 || i >= len(a.SFTP.Entries) {
		return nil
	}
	return &a.SFTP.Entries[i]
}

// SFTPSelectedPath returns the full remote path of the selection.
func (a *App) SFTPSelectedPath() string {
	e := a.SFTPSelectedEntry()
	if e == nil {
		return ""
	}
	return a.SFTPPathFor(e.Name)
}

// SFTPOpenSelected enters the selected directory.
func (a *App) SFTPOpenSelected() {
	e := a.SFTPSelectedEntry()
	if e == nil || !e.IsDir {
		return
	}
	a.loadSFTPDir(a.SFTP.HostID, a.SFTPPathFor(e.Name))
}

// SFTPOpenNewFolder opens the new-folder dialog.
func (a *App) SFTPOpenNewFolder() {
	a.SFTP.NewFolderName = ""
	a.SFTP.NewFolderOpen = true
}

// SFTPCreateFolder creates the folder named in the dialog.
func (a *App) SFTPCreateFolder() {
	name := strings.TrimSpace(a.SFTP.NewFolderName)
	if name == "" {
		a.SFTP.Err = "请输入目录名称"
		return
	}
	hostID := a.SFTP.HostID
	dir := a.SFTPPathFor(name)
	a.SFTP.Err = ""
	if a.Backend == nil {
		a.SFTP.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		err := b.SFTPMkdir(hostID, dir)
		a.Post(func() {
			if hostID != a.SFTP.HostID {
				return
			}
			if err != nil {
				a.SFTP.Err = "新建目录失败：" + err.Error()
				return
			}
			a.SFTP.NewFolderOpen = false
			a.loadSFTPDir(hostID, a.SFTP.Path)
		})
	}()
}

// SFTPOpenRename opens the rename dialog for the selection.
func (a *App) SFTPOpenRename() {
	e := a.SFTPSelectedEntry()
	if e == nil {
		return
	}
	a.SFTP.RenameName = e.Name
	a.SFTP.RenamePath = a.SFTPSelectedPath()
	a.SFTP.RenameOpen = true
}

// SFTPRename applies the rename dialog.
func (a *App) SFTPRename() {
	name := strings.TrimSpace(a.SFTP.RenameName)
	oldPath := a.SFTP.RenamePath
	if oldPath == "" {
		a.SFTP.Err = "没有可重命名的条目"
		return
	}
	if name == "" {
		a.SFTP.Err = "请输入新名称"
		return
	}
	hostID := a.SFTP.HostID
	newPath := path.Join(path.Dir(oldPath), name)
	if a.Backend == nil {
		a.SFTP.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		err := b.SFTPRename(hostID, oldPath, newPath)
		a.Post(func() {
			if hostID != a.SFTP.HostID {
				return
			}
			if err != nil {
				a.SFTP.Err = "重命名失败：" + err.Error()
				return
			}
			a.SFTP.RenameOpen = false
			a.SFTP.RenamePath = ""
			a.loadSFTPDir(hostID, a.SFTP.Path)
		})
	}()
}

// SFTPOpenDelete opens the delete confirmation.
func (a *App) SFTPOpenDelete() {
	p := a.SFTPSelectedPath()
	if p == "" {
		return
	}
	a.SFTP.DeletePath = p
	a.SFTP.ConfirmDelete = true
}

// SFTPConfirmDelete deletes the confirmed path.
func (a *App) SFTPConfirmDelete() {
	p := a.SFTP.DeletePath
	if p == "" || a.Backend == nil {
		return
	}
	hostID := a.SFTP.HostID
	a.SFTP.Err = ""
	b := a.Backend
	go func() {
		err := b.SFTPDelete(hostID, p)
		a.Post(func() {
			if hostID != a.SFTP.HostID {
				return
			}
			if err != nil {
				a.SFTP.Err = "删除失败：" + err.Error()
				return
			}
			a.SFTP.DeletePath = ""
			a.loadSFTPDir(hostID, a.SFTP.Path)
		})
	}()
}

// SFTPUpload picks a local file and queues an upload into the current dir.
func (a *App) SFTPUpload() {
	hostID, dir := a.SFTP.HostID, a.SFTP.Path
	seq := a.SFTP.listSeq
	if hostID == "" || a.Backend == nil {
		return
	}
	a.SFTP.Err = ""
	b := a.Backend
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "选择要上传的文件",
			ButtonLabel: "上传",
		})
		if err != nil {
			a.Post(func() { a.SFTP.Err = "选择文件失败：" + err.Error() })
			return
		}
		if len(paths) == 0 {
			return
		}
		a.Post(func() {
			if hostID != a.SFTP.HostID || seq != a.SFTP.listSeq || dir != a.SFTP.Path {
				a.SFTP.Err = "目录已切换，请重新上传"
				return
			}
			go func() {
				id, err := b.StartSFTPUpload(hostID, paths[0], dir)
				a.Post(func() {
					if hostID != a.SFTP.HostID {
						return
					}
					if err != nil {
						a.SFTP.Err = "上传失败：" + err.Error()
						return
					}
					a.SFTP.Notice = "已加入上传队列（" + id + "）"
					a.SFTPRefreshTasks()
				})
			}()
		})
	}()
}

// SFTPDownload picks a local destination and queues a download.
func (a *App) SFTPDownload() {
	e := a.SFTPSelectedEntry()
	if e == nil || e.IsDir || a.Backend == nil {
		return
	}
	hostID := a.SFTP.HostID
	remote := a.SFTPSelectedPath()
	a.SFTP.Err = ""
	b := a.Backend
	go func() {
		local, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Title:       "选择保存位置",
			DefaultPath: filepath.Base(e.Name),
			ButtonLabel: "下载",
		})
		if err != nil {
			a.Post(func() { a.SFTP.Err = "选择保存位置失败：" + err.Error() })
			return
		}
		if strings.TrimSpace(local) == "" {
			return
		}
		id, err := b.StartSFTPDownload(hostID, remote, local)
		a.Post(func() {
			if hostID != a.SFTP.HostID {
				return
			}
			if err != nil {
				a.SFTP.Err = "下载失败：" + err.Error()
				return
			}
			a.SFTP.Notice = "已加入下载队列（" + id + "）"
			a.SFTPRefreshTasks()
		})
	}()
}

// SFTPRefreshTasks reloads the transfer task list.
func (a *App) SFTPRefreshTasks() {
	if a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		tasks := b.ListSFTPTasks()
		a.Post(func() { a.SFTP.Tasks = tasks })
	}()
}

// HandleSFTPEvent is the backend's transfer event sink.
func (a *App) HandleSFTPEvent(event string, task sftppkg.TaskInfo) {
	a.Post(func() {
		found := false
		prevStatus := ""
		for i := range a.SFTP.Tasks {
			if a.SFTP.Tasks[i].ID == task.ID {
				prevStatus = a.SFTP.Tasks[i].Status
				a.SFTP.Tasks[i] = task
				found = true
				break
			}
		}
		if !found {
			a.SFTP.Tasks = append(a.SFTP.Tasks, task)
		}
		if event == "sftp:start" {
			a.SFTP.TaskPanelOpen = true
		}
		if found && prevStatus != task.Status {
			switch task.Status {
			case string(sftppkg.TaskDone):
				a.notifyUser("传输完成", sftpTaskName(task))
			case string(sftppkg.TaskError):
				body := sftpTaskName(task)
				if task.Error != "" {
					body += "：" + task.Error
				}
				a.notifyUser("传输失败", body)
			}
		}
	})
}

// sftpTaskName picks a readable file name for a notification.
func sftpTaskName(t sftppkg.TaskInfo) string {
	name := filepath.Base(t.RemotePath)
	if name == "" || name == "." || name == "/" || name == string(filepath.Separator) {
		name = filepath.Base(t.LocalPath)
	}
	if name == "" || name == "." || name == "/" {
		return "文件传输"
	}
	return name
}

// SFTPPause pauses a task.
func (a *App) SFTPPause(id string) { a.sftpTaskAction(id, "pause") }

// SFTPResume resumes a task.
func (a *App) SFTPResume(id string) { a.sftpTaskAction(id, "resume") }

// SFTPCancel cancels a task.
func (a *App) SFTPCancel(id string) { a.sftpTaskAction(id, "cancel") }

// SFTPRemoveTask removes a finished task record.
func (a *App) SFTPRemoveTask(id string) { a.sftpTaskAction(id, "remove") }

func (a *App) sftpTaskAction(id, action string) {
	if id == "" || a.Backend == nil {
		return
	}
	a.SFTP.Err = ""
	b := a.Backend
	go func() {
		var err error
		switch action {
		case "pause":
			err = b.SFTPPause(id)
		case "resume":
			err = b.SFTPResume(id)
		case "cancel":
			err = b.SFTPCancel(id)
		case "remove":
			err = b.SFTPRemoveTask(id)
		}
		tasks := b.ListSFTPTasks()
		a.Post(func() {
			a.SFTP.Tasks = tasks
			if err != nil {
				a.SFTP.Err = "任务操作失败：" + err.Error()
			}
		})
	}()
}

func sftpTaskInTab(tab int, status string) bool {
	switch tab {
	case SFTPTasksRunning:
		return status == string(sftppkg.TaskRunning)
	case SFTPTasksQueued:
		return status == string(sftppkg.TaskQueued)
	case SFTPTasksPaused:
		return status == string(sftppkg.TaskPaused)
	case SFTPTasksCancelled:
		return status == string(sftppkg.TaskCancelled)
	case SFTPTasksDone:
		return status == string(sftppkg.TaskDone) || status == string(sftppkg.TaskError)
	}
	return false
}

// SFTPVisibleTasks filters tasks by the selected task tab.
func (a *App) SFTPVisibleTasks() []sftppkg.TaskInfo {
	out := make([]sftppkg.TaskInfo, 0, len(a.SFTP.Tasks))
	for _, t := range a.SFTP.Tasks {
		if sftpTaskInTab(a.SFTP.TaskTab, t.Status) {
			out = append(out, t)
		}
	}
	return out
}

// SFTPTaskCount counts tasks in a tab without touching the active tab.
func (a *App) SFTPTaskCount(tab int) int {
	n := 0
	for _, t := range a.SFTP.Tasks {
		if sftpTaskInTab(tab, t.Status) {
			n++
		}
	}
	return n
}

// SFTPTaskActiveCount counts running and queued tasks for the toolbar badge.
func (a *App) SFTPTaskActiveCount() int {
	n := 0
	for _, t := range a.SFTP.Tasks {
		if t.Status == string(sftppkg.TaskRunning) || t.Status == string(sftppkg.TaskQueued) {
			n++
		}
	}
	return n
}

// SFTPPathFor returns the full path of an entry name in the current dir.
func (a *App) SFTPPathFor(name string) string {
	return path.Join(a.SFTP.Path, name)
}

// SFTPStatusText renders a task status in Chinese.
func SFTPStatusText(status string) string {
	switch status {
	case string(sftppkg.TaskQueued):
		return "排队中"
	case string(sftppkg.TaskRunning):
		return "传输中"
	case string(sftppkg.TaskPaused):
		return "已暂停"
	case string(sftppkg.TaskCancelled):
		return "已取消"
	case string(sftppkg.TaskDone):
		return "已完成"
	case string(sftppkg.TaskError):
		return "失败"
	default:
		return status
	}
}

// SFTPFormatSize formats bytes for display.
func SFTPFormatSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(n)/(1024*1024*1024))
	}
}
