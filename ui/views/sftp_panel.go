package views

import (
	"fmt"
	"path"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/domain/sftp"
	"nextopsx/i18n"
	"nextopsx/ui/app"
	"nextopsx/ui/assets"
)

// SFTPPanel is the remote file browser and transfer panel for SSH assets.
func SFTPPanel(c *ui.Context, a *app.App) {
	t := c.Theme()
	ui.Column(c).Fill().MinWidth(0).Children(func() {
		sftpToolbar(c, a, t)
		ui.Divider(c)
		sftpFileList(c, a, t)
		sftpTransferStrip(c, a, t)
		sftpStatusBar(c, a, t)
	})
	sftpDialogs(c, a, t)
}

func sftpToolbar(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Shrink(0).Children(func() {
		ui.Row(c).Padding(6, 8, 4, 8).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "SFTP").FontSize(fontSize(12)).Bold()
			ui.Text(c, a.SFTP.Path).Font("monospace").FontSize(fontSize(11)).
				TextColor(t.TextMuted).Grow(1).MinWidth(0).SingleLine()
			sftpIconButton(c, t, "close", i18n.T("sftp.drawer.close"), false,
				a.CloseSFTPDrawer)
		})
		ui.Row(c).Padding(0, 6, 6, 6).Gap(2).AlignItems(ui.Center).Children(func() {
			selected := a.SFTPSelectedEntry()
			sftpIconButton(c, t, "arrow-up", i18n.T("sftp.toolbar.up"),
				a.SFTP.Path == "/", a.SFTPGoParent)
			sftpIconButton(c, t, "refresh", i18n.T("sftp.action.refresh"),
				false, a.SFTPRefresh)
			sftpIconButton(c, t, "folder-plus", i18n.T("sftp.action.newFolder"),
				false, a.SFTPOpenNewFolder)
			sftpIconButton(c, t, "upload", i18n.T("sftp.action.upload"),
				false, a.SFTPUpload)
			sftpIconButton(c, t, "download", i18n.T("sftp.action.download"),
				selected == nil || selected.IsDir, a.SFTPDownload)
			sftpIconButton(c, t, "pencil", i18n.T("sftp.action.rename"),
				selected == nil, a.SFTPOpenRename)
			sftpIconButton(c, t, "trash", i18n.T("sftp.action.delete"),
				selected == nil, a.SFTPOpenDelete)

			taskBtn := sftpIconButton(c, t, "activity",
				i18n.Tf("sftp.toolbar.tasks", a.SFTPTaskActiveCount()), false, func() {
					a.SFTP.TaskPanelOpen = !a.SFTP.TaskPanelOpen
				})
			ui.Popover(c, taskBtn, &a.SFTP.TaskPanelOpen, func() {
				sftpTasksPanel(c, a, t)
			})
		})
	})
}

func sftpIconButton(c *ui.Context, t *ui.Theme, icon, label string, disabled bool,
	onClick func()) ui.Element {
	btn := ui.ButtonBase(c).Padding(4).Radius(5).Label(label).Tooltip(label).
		Cursor(ui.CursorPointer).OnClick(onClick)
	if disabled {
		btn = btn.Disabled(true)
	}
	if btn.Hovered() {
		btn = btn.Background(t.SurfaceHover)
	}
	btn.Children(func() {
		if s := assets.SVG(icon); s != nil {
			ui.Icon(c, s).Size(14, 14).TextColor(t.TextMuted)
		}
	})
	return btn
}

func sftpFileList(c *ui.Context, a *app.App, t *ui.Theme) {
	if a.SFTP.Loading {
		ui.Row(c).Grow(1).Padding(24).Center().Children(func() {
			ui.Spinner(c)
			ui.Text(c, i18n.T("sftp.list.loading")).FontSize(fontSize(12)).TextColor(t.TextMuted)
		})
		return
	}
	if a.SFTP.LoadErr != "" {
		ui.Column(c).Grow(1).Padding(16).Gap(8).Children(func() {
			ui.Text(c, a.SFTP.LoadErr).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			ui.Button(c, i18n.T("sftp.list.retry")).OnClick(a.SFTPRefresh)
		})
		return
	}
	if len(a.SFTP.Entries) == 0 {
		ui.Row(c).Grow(1).Padding(24).Center().Children(func() {
			ui.Text(c, i18n.T("sftp.list.empty")).FontSize(fontSize(13)).TextColor(t.TextMuted)
		})
		return
	}

	a.SFTP.Grid.Selected = &a.SFTP.SelectedRow
	cols := []ui.TableColumn{
		{Title: i18n.T("sftp.column.name"), MinWidth: 80},
		{Title: i18n.T("sftp.column.size"), Width: 72, Align: ui.End},
		{Title: i18n.T("sftp.column.modified"), Width: 96},
	}
	table := ui.Table(c, &a.SFTP.Grid, cols, len(a.SFTP.Entries), func(row, col int) {
		if row < 0 || row >= len(a.SFTP.Entries) {
			return
		}
		e := a.SFTP.Entries[row]
		color := t.TextMuted
		if row == a.SFTP.SelectedRow {
			color = t.Text
		}
		switch col {
		case 0:
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				sftpEntryIcon(c, e, t)
				ui.Text(c, e.Name).FontSize(fontSize(12)).SingleLine()
			})
		case 1:
			size := ""
			if !e.IsDir {
				size = app.SFTPFormatSize(e.Size)
			}
			ui.Text(c, size).FontSize(fontSize(11)).TextColor(color).SingleLine()
		case 2:
			ui.Text(c, time.Unix(e.ModTime, 0).Format("01-02 15:04")).
				FontSize(fontSize(11)).TextColor(color).SingleLine()
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

// sftpTransferStrip shows the active transfer at a glance, keeping the full
// task list one click away.
func sftpTransferStrip(c *ui.Context, a *app.App, t *ui.Theme) {
	var task *sftp.TaskInfo
	for i := range a.SFTP.Tasks {
		if a.SFTP.Tasks[i].Status == string(sftp.TaskRunning) {
			task = &a.SFTP.Tasks[i]
			break
		}
	}
	if task == nil {
		return
	}
	ratio := 0.0
	if task.Total > 0 {
		ratio = float64(task.Done) / float64(task.Total)
	}
	id := task.ID
	name := path.Base(task.RemotePath)
	if name == "." || name == "/" || name == "" {
		name = task.RemotePath
	}
	ui.Column(c).Shrink(0).Padding(6, 10, 7, 10).Gap(5).
		Background(t.Surface).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).
		Children(func() {
			ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
				icon := "upload"
				if task.Direction == "download" {
					icon = "download"
				}
				if s := assets.SVG(icon); s != nil {
					ui.Icon(c, s).Size(12, 12).TextColor(t.Accent)
				}
				ui.Text(c, name).FontSize(fontSize(11)).
					TextColor(t.Text).Grow(1).MinWidth(0).SingleLine()
				ui.Text(c, fmt.Sprintf("%d%%", int(ratio*100))).
					FontSize(fontSize(10)).TextColor(t.TextMuted)
				sftpIconButton(c, t, "pause", i18n.T("sftp.action.pause"),
					false, func() { a.SFTPPause(id) })
				sftpIconButton(c, t, "close", i18n.T("sftp.action.cancel"),
					false, func() { a.SFTPCancel(id) })
			})
			ui.Progress(c, ratio).FillWidth()
		})
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
			text = i18n.T("sftp.status.ready")
		}
		ui.Text(c, text).FontSize(fontSize(12)).TextColor(t.TextMuted).SingleLine()
	})
}

func sftpTasksPanel(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Column(c).Width(640).MaxHeight(470).Padding(10).Gap(8).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("sftp.tasks.title")).FontSize(fontSize(13)).Bold()
			ui.Spacer(c)
			ui.Text(c, i18n.Tf("sftp.tasks.runningCount", a.SFTPTaskCount(app.SFTPTasksRunning))).
				FontSize(fontSize(12)).TextColor(t.TextMuted)
		})
		labels := []string{
			i18n.Tf("sftp.tabs.running", a.SFTPTaskCount(app.SFTPTasksRunning)),
			i18n.Tf("sftp.tabs.queued", a.SFTPTaskCount(app.SFTPTasksQueued)),
			i18n.Tf("sftp.tabs.paused", a.SFTPTaskCount(app.SFTPTasksPaused)),
			i18n.Tf("sftp.tabs.cancelled", a.SFTPTaskCount(app.SFTPTasksCancelled)),
			i18n.Tf("sftp.tabs.done", a.SFTPTaskCount(app.SFTPTasksDone)),
		}
		ui.Tabs(c, &a.SFTP.TaskTab, labels...)
		tasks := a.SFTPVisibleTasks()
		if len(tasks) == 0 {
			ui.Text(c, i18n.T("sftp.tasks.empty")).FontSize(fontSize(12)).TextColor(t.TextMuted)
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
	direction := i18n.T("sftp.action.upload")
	if task.Direction == "download" {
		direction = i18n.T("sftp.action.download")
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
		ui.Button(c, i18n.T("sftp.task.pause")).FontSize(fontSize(11)).OnClick(func() { a.SFTPPause(task.ID) })
		ui.Button(c, i18n.T("sftp.action.cancel")).FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskPaused):
		ui.Button(c, i18n.T("sftp.task.resume")).FontSize(fontSize(11)).OnClick(func() { a.SFTPResume(task.ID) })
		ui.Button(c, i18n.T("sftp.action.cancel")).FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskQueued):
		ui.Button(c, i18n.T("sftp.action.cancel")).FontSize(fontSize(11)).OnClick(func() { a.SFTPCancel(task.ID) })
	case string(sftp.TaskCancelled), string(sftp.TaskDone), string(sftp.TaskError):
		ui.Button(c, i18n.T("sftp.task.remove")).FontSize(fontSize(11)).OnClick(func() { a.SFTPRemoveTask(task.ID) })
	}
}

func sftpDialogs(c *ui.Context, a *app.App, t *ui.Theme) {
	ui.Modal(c, &a.SFTP.NewFolderOpen, func() {
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, i18n.T("sftp.action.newFolder")).FontSize(fontSize(16)).Bold()
			ui.Text(c, i18n.Tf("sftp.dialog.location", a.SFTP.Path)).FontSize(fontSize(12)).TextColor(t.TextMuted)
			c.Key("sftp-new-folder")
			ui.TextInput(c, &a.SFTP.NewFolderName).Label(i18n.T("sftp.dialog.folderName")).Grow(1)
			if a.SFTP.Err != "" {
				ui.Text(c, a.SFTP.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, i18n.T("sftp.action.cancel")).OnClick(func() { a.SFTP.NewFolderOpen = false })
				ui.PrimaryButton(c, i18n.T("sftp.dialog.create")).OnClick(a.SFTPCreateFolder)
			})
		})
	})
	ui.Modal(c, &a.SFTP.RenameOpen, func() {
		ui.Column(c).Width(420).Padding(18).Gap(10).Children(func() {
			ui.Text(c, i18n.T("sftp.action.rename")).FontSize(fontSize(16)).Bold()
			c.Key("sftp-rename")
			ui.TextInput(c, &a.SFTP.RenameName).Label(i18n.T("sftp.dialog.newName")).Grow(1)
			if a.SFTP.Err != "" {
				ui.Text(c, a.SFTP.Err).FontSize(fontSize(13)).TextColor(t.Danger).Wrap()
			}
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				ui.Button(c, i18n.T("sftp.action.cancel")).OnClick(func() { a.SFTP.RenameOpen = false })
				ui.PrimaryButton(c, i18n.T("sftp.dialog.save")).OnClick(a.SFTPRename)
			})
		})
	})
	if ui.AlertDialog(c, &a.SFTP.ConfirmDelete, i18n.T("sftp.delete.title"),
		i18n.Tf("sftp.delete.confirm", a.SFTP.DeletePath),
		i18n.T("sftp.action.cancel"), i18n.T("sftp.delete.confirmAction")) == 1 {
		a.SFTPConfirmDelete()
	}
}
