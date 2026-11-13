package app

import (
	"strings"
	"testing"

	sftppkg "nextopsx/internal/sftp"
	hoststore "nextopsx/internal/store/hosts"
)

func TestSFTPPathHelpers(t *testing.T) {
	a := &App{}
	a.SFTP.Path = "/var/log"
	if got := a.SFTPParentPath(); got != "/var" {
		t.Fatalf("parent = %q", got)
	}
	if got := a.SFTPPathFor("nginx"); got != "/var/log/nginx" {
		t.Fatalf("path for = %q", got)
	}
	a.SFTP.Path = "/"
	if got := a.SFTPParentPath(); got != "/" {
		t.Fatalf("root parent = %q", got)
	}
	if got := a.SFTPPathFor("etc"); got != "/etc" {
		t.Fatalf("root join = %q", got)
	}
}

func TestSFTPCompletionNotifies(t *testing.T) {
	a := &App{}
	var got [][2]string
	a.SetNotifier(func(title, body string) {
		got = append(got, [2]string{title, body})
	})

	a.HandleSFTPEvent("sftp:start", sftppkg.TaskInfo{
		ID: "t1", RemotePath: "/tmp/report.csv", Status: string(sftppkg.TaskRunning),
	})
	a.HandleSFTPEvent("sftp:progress", sftppkg.TaskInfo{
		ID: "t1", RemotePath: "/tmp/report.csv", Status: string(sftppkg.TaskDone),
	})
	if len(got) != 1 || got[0][0] != "传输完成" || got[0][1] != "report.csv" {
		t.Fatalf("done notification = %v", got)
	}

	// 同一终态的重复事件不再通知。
	a.HandleSFTPEvent("sftp:progress", sftppkg.TaskInfo{
		ID: "t1", RemotePath: "/tmp/report.csv", Status: string(sftppkg.TaskDone),
	})
	if len(got) != 1 {
		t.Fatalf("duplicate notification = %v", got)
	}

	a.HandleSFTPEvent("sftp:start", sftppkg.TaskInfo{
		ID: "t2", RemotePath: "/tmp/b.txt", Status: string(sftppkg.TaskRunning),
	})
	a.HandleSFTPEvent("sftp:progress", sftppkg.TaskInfo{
		ID: "t2", RemotePath: "/tmp/b.txt", Status: string(sftppkg.TaskError), Error: "disk full",
	})
	if len(got) != 2 || got[1][0] != "传输失败" || !strings.Contains(got[1][1], "disk full") {
		t.Fatalf("error notification = %v", got)
	}
}

func TestSFTPTaskEventsAndFilters(t *testing.T) {
	a := &App{}
	a.HandleSFTPEvent("sftp:start", sftppkg.TaskInfo{
		ID: "t1", HostID: "h1", Direction: "upload",
		RemotePath: "/tmp/a.txt", Total: 100, Done: 10,
		Status: string(sftppkg.TaskRunning),
	})
	if len(a.SFTP.Tasks) != 1 || !a.SFTP.TaskPanelOpen {
		t.Fatalf("start event = %+v open=%v", a.SFTP.Tasks, a.SFTP.TaskPanelOpen)
	}
	a.HandleSFTPEvent("sftp:progress", sftppkg.TaskInfo{
		ID: "t1", HostID: "h1", Direction: "upload",
		RemotePath: "/tmp/a.txt", Total: 100, Done: 100,
		Status: string(sftppkg.TaskDone),
	})
	if a.SFTP.Tasks[0].Done != 100 || a.SFTP.Tasks[0].Status != string(sftppkg.TaskDone) {
		t.Fatalf("progress event = %+v", a.SFTP.Tasks[0])
	}

	a.SFTP.Tasks = append(a.SFTP.Tasks,
		sftppkg.TaskInfo{ID: "t2", Status: string(sftppkg.TaskQueued)},
		sftppkg.TaskInfo{ID: "t3", Status: string(sftppkg.TaskPaused)},
		sftppkg.TaskInfo{ID: "t4", Status: string(sftppkg.TaskCancelled)},
		sftppkg.TaskInfo{ID: "t5", Status: string(sftppkg.TaskError)},
	)
	if n := a.SFTPTaskCount(SFTPTasksRunning); n != 0 {
		t.Fatalf("running = %d", n)
	}
	if n := a.SFTPTaskCount(SFTPTasksQueued); n != 1 {
		t.Fatalf("queued = %d", n)
	}
	if n := a.SFTPTaskCount(SFTPTasksDone); n != 2 {
		t.Fatalf("done+error = %d", n)
	}
	if got := a.SFTPTaskActiveCount(); got != 1 {
		t.Fatalf("active = %d", got)
	}
	if got := SFTPStatusText(string(sftppkg.TaskPaused)); got != "已暂停" {
		t.Fatalf("status = %q", got)
	}
	if got := SFTPFormatSize(2048); got != "2.0 KB" {
		t.Fatalf("size = %q", got)
	}
}

func TestAttachSFTPWithoutBackend(t *testing.T) {
	a := &App{}
	n := &hoststore.TreeNode{ID: "h1", Name: "web", NodeType: "host", Protocol: "ssh"}
	a.Selected = n
	a.EnsureSFTP()
	if a.SFTP.HostID != "h1" || a.SFTP.Loading {
		t.Fatalf("state = %+v", a.SFTP)
	}
	if a.SFTP.LoadErr != "后端未初始化" {
		t.Fatalf("load err = %q", a.SFTP.LoadErr)
	}

	nonSSH := &hoststore.TreeNode{ID: "h2", Name: "win", NodeType: "host", Protocol: "winrm"}
	a.Selected = nonSSH
	a.EnsureSFTP()
	if a.SFTP.HostID != "h1" {
		t.Fatalf("sftp attached for non-SSH: %q", a.SFTP.HostID)
	}
}

func TestAttachSFTPUpdatesNameAndRefreshesOnReentry(t *testing.T) {
	a := &App{}
	first := &hoststore.TreeNode{ID: "h1", Name: "web-01", NodeType: "host", Protocol: "ssh"}
	a.AttachSFTPFor(first)
	if a.SFTP.HostName != "web-01" || a.SFTP.LoadErr != "后端未初始化" {
		t.Fatalf("state = %+v", a.SFTP)
	}

	renamed := &hoststore.TreeNode{ID: "h1", Name: "web-prod", NodeType: "host", Protocol: "ssh"}
	a.AttachSFTPFor(renamed)
	if a.SFTP.HostName != "web-prod" {
		t.Fatalf("name after rename = %q", a.SFTP.HostName)
	}

	a.Selected = renamed
	a.SFTP.LoadErr = ""
	a.EnsureSFTP()
	if a.SFTP.LoadErr == "" {
		t.Fatal("re-entering the SFTP tab did not refresh the listing")
	}
}

func TestSFTPRenameUsesCapturedPath(t *testing.T) {
	a := &App{}
	a.SFTP.RenamePath = "/var/log/old.txt"
	a.SFTP.RenameName = "new.txt"
	a.SFTP.SelectedRow = -1
	a.SFTPRename()
	if a.SFTP.Err != "后端未初始化" {
		t.Fatalf("rename should use the captured path, err = %q", a.SFTP.Err)
	}

	a = &App{}
	a.SFTP.Path = "/var/log"
	a.SFTP.SelectedRow = 0
	a.SFTP.Entries = []sftppkg.Entry{{Name: "old.txt"}}
	a.SFTPOpenRename()
	if a.SFTP.RenamePath != "/var/log/old.txt" || a.SFTP.RenameName != "old.txt" {
		t.Fatalf("rename dialog = %q / %q", a.SFTP.RenamePath, a.SFTP.RenameName)
	}
}
