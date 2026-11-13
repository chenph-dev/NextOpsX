package views

import (
	"fmt"
	"path"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/assets"
	"nextopsx/internal/sftp"
)

// SFTPPanel is the remote file browser and transfer panel for SSH assets.
func SFTPPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		sftpToolbar(c, a, t)
		ui.Divider(c)
		sftpFileList(c, a, t)
		sftpStatusBar(c, a, t)
	})
	sftpDialogs(c, a, t)
}

func sftpToolbar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Row(c).Padding(6, 10).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "SFTP · "+a.SFTP.HostName).FontSize(fontSize(13)).Bold().SingleLine()
		ui.Text(c, a.SFTP.Path).Font("monospace").FontSize(fontSize(13)).
			TextColor(t.TextMuted).Grow(1).SingleLine()

		up := ui.Button(c, "上一级").FontSize(fontSize(12))
		if a.SFTP.Path == "/" {
			up = up.Disabled(true)
		}
		up.OnClick(a.SFTPGoParent)
		ui.Button(c, "刷新").FontSize(fontSize(12)).OnClick(a.SFTPRefresh)
		ui.Button(c, "新建目录").FontSize(fontSize(12)).OnClick(a.SFTPOpenNewFolder)
		ui.Button(c, "上传").FontSize(fontSize(12)).OnClick(a.SFTPUpload)

		selected := a.SFTPSelectedEntry()
		download := ui.Button(c, "下载").FontSize(fontSize(12))
		if selected == nil || selected.IsDir {
			download = download.Disabled(true)
		}
		download.OnClick(a.SFTPDownload)

		rename := ui.Button(c, "重命名").FontSize(fontSize(12))
		if selected == nil {
			rename = rename.Disabled(true)
		}
		rename.OnClick(a.SFTPOpenRename)

		del := ui.Button(c, "删除").FontSize(fontSize(12))
		if selected == nil {
			del = del.Disabled(true)
		}
		del.OnClick(a.SFTPOpenDelete)

		taskBtn := ui.Button(c, fmt.Sprintf("任务 (%d)", a.SFTPTaskActiveCount())).
			FontSize(fontSize(12))
		taskBtn = taskBtn.OnClick(func() {
			a.SFTP.TaskPanelOpen = !a.SFTP.TaskPanelOpen
		})
		ui.Popover(c, taskBtn, &a.SFTP.TaskPanelOpen, func() {
			sftpTasksPanel(c, a, t)
		})
	})
}

func sftpFileList(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.SFTP.Loading {
		ui.Row(c).Grow(1).Padding(24).Center().Children(func() {
			ui.Spinner(c)
			ui.Text(c, "正在读取目录…").FontSize(fontSize(12)).TextColor(t.TextMuted)
		})
		return
	}
	if a.SFTP.LoadErr != "" {
		ui.Column(c).Grow(1).Padding(16).Gap(8).Children(func() {
			ui.Text(c, a.SFTP.LoadErr).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			ui.Button(c, "重试").OnClick(a.SFTPRefresh)
		})
		return
	}
	if len(a.SFTP.Entries) == 0 {
		ui.Row(c).Grow(1).Padding(24).Center().Children(func() {
			ui.Text(c, "目录为空").FontSize(fontSize(13)).TextColor(t.TextMuted)
		})
		return
	}

	a.SFTP.Grid.Selected = &a.SFTP.SelectedRow
	cols := []ui.TableColumn{
		{Title: "名称", MinWidth: 260},
		{Title: "类型", Width: 72, Align: ui.Center},
		{Title: "大小", Width: 100, Align: ui.End},
		{Title: "修改时间", Width: 156},
	}
	table := ui.Table(c, &a.SFTP.Grid, cols, len(a.SFTP.Entries), func(row, col int) {
		if row < 0 || row >= len(a.SFTP.Entries) {
			return
		}
		e := a.SFTP.Entries[row]
		switch col {
		case 0:
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				sftpEntryIcon(c, e, t)
				ui.Text(c, e.Name).FontSize(fontSize(13)).SingleLine()
			})
		case 1:
			kind := "文件"
			if e.IsDir {
				kind = "目录"
			}
			ui.Text(c, kind).FontSize(fontSize(12)).TextColor(t.TextMuted)
		case 2:
			size := ""
			if !e.IsDir {
				size = app.SFTPFormatSize(e.Size)
			}
			ui.Text(c, size).FontSize(fontSize(12)).TextColor(t.TextMuted)
		case 3:
			ui.Text(c, time.Unix(e.ModTime, 0).Format("2006-01-02 15:04")).
				FontSize(fontSize(12)).TextColor(t.TextMuted)
		}
	}).Grow(1).MinWidth(0)
	if table.Submitted() {
		a.SFTPOpenSelected()
	}
}

func sftpEntryIcon(c *ui.Context, e sftp.Entry, t *ui.Theme) {
	if e.IsDir {
		if s := assets.FolderSVG(); s != nil {
			ui.Icon(c, s).Size(14, 14).TextColor(t.Accent)
		}
		return
	}
	if s := assets.ObjectSVG("file"); s != nil {
		ui.Icon(c, s).Size(14, 14).TextColor(t.TextMuted)
	}
}

func sftpStatusBar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Divider(c)
	ui.Row(c).Padding(3, 10).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.SFTP.Path).Font("monospace").FontSize(fontSize(12)).
			TextColor(t.TextMuted).SingleLine()
		ui.Spacer(c)
		if a.SFTP.Err != "" {
			ui.Text(c, a.SFTP.Err).FontSize(fontSize(12)).TextColor(t.Danger).SingleLine()
			return
		}
		text := a.SFTP.Notice
		if text == "" {
			text = "就绪"
		}
		ui.Text(c, text).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
	})
}

func sftpTasksPanel(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(640).MaxHeight(470).Padding(10).Gap(8).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "传输任务").FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Text(c, fmt.Sprintf("进行中 %d", a.SFTPTaskCount(app.SFTPTasksRunning))).
				FontSize(fontSize(12)).TextColor(t.TextMuted)
		})
		labels := []string{
			fmt.Sprintf("进行中 (%d)", a.SFTPTaskCount(app.SFTPTasksRunning)),
			fmt.Sprintf("排队 (%d)", a.SFTPTaskCount(app.SFTPTasksQueued)),
			fmt.Sprintf("已暂停 (%d)", a.SFTPTaskCount(app.SFTPTasksPaused)),
			fmt.Sprintf("已取消 (%d)", a.SFTPTaskCount(app.SFTPTasksCancelled)),
			fmt.Sprintf("已完成 (%d)", a.SFTPTaskCount(app.SFTPTasksDone)),
		}
		ui.Tabs(c, &a.SFTP.TaskTab, labels...)
		tasks := a.SFTPVisibleTasks()
		if len(tasks) == 0 {
			ui.Text(c, "没有任务").FontSize(fontSize(12)).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(340).Children(func() {
			ui.Column(c).Gap(6).Children(func() {
				for _, task := range tasks {
					sftpTaskRow(c, a, task, t)
				}
			})
		})
	})
}

func sftpTaskRow(c *ui.Context, a *app.App, task sftp.TaskInfo, t *ui.Theme) {
	name := path.Base(task.RemotePath)
	if name == "." || name == "/" || name == "" {
		name = task.RemotePath
	}
	direction := "上传"
	if task.Direction == "download" {
		direction = "下载"
	}
	ratio := 0.0
	if task.Total > 0 {
		ratio = float64(task.Done) / float64(task.Total)
	}
	statusColor := t.TextMuted
	switch task.Status {
	case string(sftp.TaskRunning):
		statusColor = t.Accent
	case string(sftp.TaskDone):
		statusColor = t.Success
	case string(sftp.TaskError):
		statusColor = t.Danger
	case string(sftp.TaskPaused):
		statusColor = t.Warning
	}
	ui.Column(c).Border(1, t.Border).Radius(6).Padding(6, 8).Gap(4).
		Background(t.Surface).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, direction).FontSize(fontSize(11)).TextColor(t.TextMuted)
			ui.Text(c, name).FontSize(fontSize(13)).SingleLine().Grow(1)
			ui.Text(c, app.SFTPStatusText(task.Status)).FontSize(fontSize(12)).
				TextColor(statusColor)
			sftpTaskActions(c, a, task, t)
		})
		ui.Progress(c, ratio)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, app.SFTPFormatSize(task.Done)+" / "+app.SFTPFormatSize(task.Total)).
				FontSize(fontSize(11)).TextColor(t.TextMuted)
			ui.Text(c, task.RemotePath).FontSize(fontSize(11)).TextColor(t.TextMuted).
				SingleLine().Grow(1)
		})
		if task.Error != "" {
			ui.Text(c, task.Error).FontSize(fontSize(11)).TextColor(t.Danger).Wrap()
		}
	})
}

func sftpTaskActions(c *ui.Context, a *app.App, task sftp.TaskInfo, t *ui.Theme) {
	switch task.Status {
	case string(sftp.TaskRunning):
		ui.Button(c, "暂停").FontSize(fontSize(11)).OnClick(func() { a.SFTPPause(task.ID) })
		ui.Button(c, "取消").FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskPaused):
		ui.Button(c, "继续").FontSize(fontSize(11)).OnClick(func() { a.SFTPResume(task.ID) })
		ui.Button(c, "取消").FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskQueued):
		ui.Button(c, "取消").FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskCancelled), string(sftp.TaskDone), string(sftp.TaskError):
		ui.Button(c, "移除").FontSize(fontSize(11)).OnClick(func() { a.SFTPRemoveTask(task.ID) })
	}
}

func sftpDialogs(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Modal(c, &a.SFTP.NewFolderOpen, func() {
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, "新建目录").FontSize(fontSize(16)).Bold()
			ui.Text(c, "位置："+a.SFTP.Path).FontSize(fontSize(12)).TextColor(t.TextMuted)
			c.Key("sftp-new-folder")
			ui.TextInput(c, &a.SFTP.NewFolderName).Label("目录名称").Grow(1)
			if a.SFTP.Err != "" {
				ui.Text(c, a.SFTP.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.SFTP.NewFolderOpen = false })
				ui.PrimaryButton(c, "创建").OnClick(a.SFTPCreateFolder)
			})
		})
	})
	ui.Modal(c, &a.SFTP.RenameOpen, func() {
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, "重命名").FontSize(fontSize(16)).Bold()
			c.Key("sftp-rename")
			ui.TextInput(c, &a.SFTP.RenameName).Label("新名称").Grow(1)
			if a.SFTP.Err != "" {
				ui.Text(c, a.SFTP.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, "取消").OnClick(func() { a.SFTP.RenameOpen = false })
				ui.PrimaryButton(c, "保存").OnClick(a.SFTPRename)
			})
		})
	})
	if ui.AlertDialog(c, &a.SFTP.ConfirmDelete, "删除文件",
		fmt.Sprintf("确定删除「%s」吗？目录必须为空，此操作无法撤销。", a.SFTP.DeletePath),
		"取消", "确认删除") == 1 {
		a.SFTPConfirmDelete()
	}
}
