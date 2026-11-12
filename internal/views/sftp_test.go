package views

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	sftppkg "nextopsx/internal/sftp"
	hoststore "nextopsx/internal/store/hosts"
)

func sftpTestApp() *app.App {
	a := &app.App{RightTab: 2}
	a.Selected = &hoststore.TreeNode{
		ID: "h1", Name: "web-01", NodeType: "host", Protocol: "ssh",
		Addr: "10.0.0.5", Port: 22, User: "root",
	}
	a.SFTP = app.SFTPState{
		HostID: "h1", HostName: "web-01", Path: "/var/log",
		SelectedRow: -1,
		Entries: []sftppkg.Entry{
			{Name: "nginx", IsDir: true, Mode: "drwxr-xr-x", ModTime: time.Now().Unix()},
			{Name: "app.log", Size: 2048, Mode: "-rw-r--r--", ModTime: time.Now().Unix()},
		},
		Tasks: []sftppkg.TaskInfo{{
			ID: "t1", HostID: "h1", Direction: "upload",
			RemotePath: "/var/log/report.pdf", Total: 100, Done: 40,
			Status: string(sftppkg.TaskRunning),
		}},
		TaskTab: app.SFTPTasksRunning,
	}
	return a
}

func TestSFTPPanelRenders(t *testing.T) {
	a := sftpTestApp()
	tt := ui.NewTester(Shell(a), 1300, 800)
	for _, want := range []string{
		"SFTP · web-01", "/var/log", "nginx", "app.log", "目录", "文件",
		"2.0 KB", "上一级", "刷新", "新建目录", "上传", "下载", "重命名", "删除",
		"任务 (1)",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSFTPTaskPanel(t *testing.T) {
	a := sftpTestApp()
	a.SFTP.TaskPanelOpen = true
	tt := ui.NewTester(Shell(a), 1300, 800)
	for _, want := range []string{
		"传输任务", "进行中 (1)", "report.pdf", "上传", "传输中", "暂停", "取消",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSFTPDialogs(t *testing.T) {
	a := sftpTestApp()
	a.SFTP.NewFolderOpen = true
	tt := ui.NewTester(Shell(a), 1300, 800)
	if !tt.HasText("新建目录") || !tt.HasText("位置：/var/log") {
		t.Fatalf("new folder dialog missing: %q", tt.Texts())
	}

	a.SFTP.NewFolderOpen = false
	a.SFTP.RenameOpen = true
	a.SFTP.RenameName = "old.log"
	tt.Frame()
	if !tt.HasText("重命名") || !tt.HasText("新名称") {
		t.Fatalf("rename dialog missing: %q", tt.Texts())
	}

	a.SFTP.RenameOpen = false
	a.SFTP.ConfirmDelete = true
	a.SFTP.DeletePath = "/var/log/app.log"
	tt.Frame()
	if !tt.HasText("删除文件") || !tt.HasText("确定删除「/var/log/app.log」吗？") {
		t.Fatalf("delete dialog missing: %q", tt.Texts())
	}
}

func TestSFTPEmptyAndError(t *testing.T) {
	a := sftpTestApp()
	a.SFTP.Entries = nil
	tt := ui.NewTester(Shell(a), 1300, 800)
	if !tt.HasText("目录为空") {
		t.Fatalf("empty hint missing: %q", tt.Texts())
	}

	a.SFTP.LoadErr = "列目录失败：连接断开"
	tt.Frame()
	if !tt.HasText("列目录失败：连接断开") || !tt.HasText("重试") {
		t.Fatalf("error state missing: %q", tt.Texts())
	}
}

func TestSFTPOperationErrorKeepsListing(t *testing.T) {
	a := sftpTestApp()
	a.SFTP.Err = "删除失败：permission denied"
	tt := ui.NewTester(Shell(a), 1300, 800)
	if !tt.HasText("app.log") || !tt.HasText("nginx") {
		t.Fatalf("listing disappeared after an operation error: %q", tt.Texts())
	}
	if !tt.HasText("删除失败：permission denied") {
		t.Fatalf("operation error missing from status bar: %q", tt.Texts())
	}
}
