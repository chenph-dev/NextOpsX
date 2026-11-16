package i18n

// SFTP file browser and transfer panel strings. The two maps below are kept
// split so each key appears in both languages side by side, which makes
// missing translations easy to spot.
var (
	sftpZh = map[string]string{
		// Shared actions and toolbar.
		"sftp.action.cancel":    "取消",
		"sftp.action.refresh":   "刷新",
		"sftp.action.upload":    "上传",
		"sftp.action.download":  "下载",
		"sftp.action.newFolder": "新建目录",
		"sftp.action.rename":    "重命名",
		"sftp.action.delete":    "删除",
		"sftp.toolbar.up":       "上一级",
		"sftp.toolbar.tasks":    "任务 (%d)",

		// File listing.
		"sftp.list.loading":    "正在读取目录…",
		"sftp.list.retry":      "重试",
		"sftp.list.empty":      "目录为空",
		"sftp.column.name":     "名称",
		"sftp.column.type":     "类型",
		"sftp.column.size":     "大小",
		"sftp.column.modified": "修改时间",
		"sftp.entry.file":      "文件",
		"sftp.entry.dir":       "目录",
		"sftp.status.ready":    "就绪",

		// Transfer task panel.
		"sftp.tasks.title":        "传输任务",
		"sftp.tasks.runningCount": "进行中 %d",
		"sftp.tabs.running":       "进行中 (%d)",
		"sftp.tabs.queued":        "排队 (%d)",
		"sftp.tabs.paused":        "已暂停 (%d)",
		"sftp.tabs.cancelled":     "已取消 (%d)",
		"sftp.tabs.done":          "已完成 (%d)",
		"sftp.tasks.empty":        "没有任务",
		"sftp.task.pause":         "暂停",
		"sftp.task.resume":        "继续",
		"sftp.task.remove":        "移除",

		// Dialogs.
		"sftp.dialog.location":      "位置：%s",
		"sftp.dialog.folderName":    "目录名称",
		"sftp.dialog.create":        "创建",
		"sftp.dialog.newName":       "新名称",
		"sftp.dialog.save":          "保存",
		"sftp.delete.title":         "删除文件",
		"sftp.delete.confirm":       "确定删除「%s」吗？目录必须为空，此操作无法撤销。",
		"sftp.delete.confirmAction": "确认删除",

		// Backend and operation errors.
		"sftp.backendNotReady":          "后端未初始化",
		"sftp.error.listDir":            "列目录失败：%s",
		"sftp.error.folderNameRequired": "请输入目录名称",
		"sftp.error.createFolder":       "新建目录失败：%s",
		"sftp.error.noRenameEntry":      "没有可重命名的条目",
		"sftp.error.newNameRequired":    "请输入新名称",
		"sftp.error.rename":             "重命名失败：%s",
		"sftp.error.delete":             "删除失败：%s",
		"sftp.dialog.upload.title":      "选择要上传的文件",
		"sftp.error.selectFile":         "选择文件失败：%s",
		"sftp.error.dirChanged":         "目录已切换，请重新上传",
		"sftp.error.upload":             "上传失败：%s",
		"sftp.notice.uploadQueued":      "已加入上传队列（%s）",
		"sftp.dialog.download.title":    "选择保存位置",
		"sftp.error.selectSaveLocation": "选择保存位置失败：%s",
		"sftp.error.download":           "下载失败：%s",
		"sftp.notice.downloadQueued":    "已加入下载队列（%s）",

		// Notifications and task status.
		"sftp.notify.done":           "传输完成",
		"sftp.notify.failed":         "传输失败",
		"sftp.error.taskAction":      "任务操作失败：%s",
		"sftp.taskName.fallback":     "文件传输",
		"sftp.notify.errorSeparator": "：",
		"sftp.status.queued":         "排队中",
		"sftp.status.running":        "传输中",
		"sftp.status.paused":         "已暂停",
		"sftp.status.cancelled":      "已取消",
		"sftp.status.done":           "已完成",
		"sftp.status.error":          "失败",
	}

	sftpEn = map[string]string{
		// Shared actions and toolbar.
		"sftp.action.cancel":    "Cancel",
		"sftp.action.refresh":   "Refresh",
		"sftp.action.upload":    "Upload",
		"sftp.action.download":  "Download",
		"sftp.action.newFolder": "New folder",
		"sftp.action.rename":    "Rename",
		"sftp.action.delete":    "Delete",
		"sftp.toolbar.up":       "Up",
		"sftp.toolbar.tasks":    "Tasks (%d)",

		// File listing.
		"sftp.list.loading":    "Reading directory…",
		"sftp.list.retry":      "Retry",
		"sftp.list.empty":      "Directory is empty",
		"sftp.column.name":     "Name",
		"sftp.column.type":     "Type",
		"sftp.column.size":     "Size",
		"sftp.column.modified": "Modified",
		"sftp.entry.file":      "File",
		"sftp.entry.dir":       "Directory",
		"sftp.status.ready":    "Ready",

		// Transfer task panel.
		"sftp.tasks.title":        "Transfers",
		"sftp.tasks.runningCount": "%d in progress",
		"sftp.tabs.running":       "In progress (%d)",
		"sftp.tabs.queued":        "Queued (%d)",
		"sftp.tabs.paused":        "Paused (%d)",
		"sftp.tabs.cancelled":     "Cancelled (%d)",
		"sftp.tabs.done":          "Completed (%d)",
		"sftp.tasks.empty":        "No tasks",
		"sftp.task.pause":         "Pause",
		"sftp.task.resume":        "Resume",
		"sftp.task.remove":        "Remove",

		// Dialogs.
		"sftp.dialog.location":      "Location: %s",
		"sftp.dialog.folderName":    "Folder name",
		"sftp.dialog.create":        "Create",
		"sftp.dialog.newName":       "New name",
		"sftp.dialog.save":          "Save",
		"sftp.delete.title":         "Delete file",
		"sftp.delete.confirm":       "Delete \"%s\"? Directories must be empty; this cannot be undone.",
		"sftp.delete.confirmAction": "Delete",

		// Backend and operation errors.
		"sftp.backendNotReady":          "Backend not initialized",
		"sftp.error.listDir":            "Failed to list directory: %s",
		"sftp.error.folderNameRequired": "Enter a folder name",
		"sftp.error.createFolder":       "Failed to create folder: %s",
		"sftp.error.noRenameEntry":      "No entry to rename",
		"sftp.error.newNameRequired":    "Enter a new name",
		"sftp.error.rename":             "Rename failed: %s",
		"sftp.error.delete":             "Delete failed: %s",
		"sftp.dialog.upload.title":      "Choose files to upload",
		"sftp.error.selectFile":         "Failed to select file: %s",
		"sftp.error.dirChanged":         "The directory changed; please upload again",
		"sftp.error.upload":             "Upload failed: %s",
		"sftp.notice.uploadQueued":      "Added to upload queue (%s)",
		"sftp.dialog.download.title":    "Choose save location",
		"sftp.error.selectSaveLocation": "Failed to choose save location: %s",
		"sftp.error.download":           "Download failed: %s",
		"sftp.notice.downloadQueued":    "Added to download queue (%s)",

		// Notifications and task status.
		"sftp.notify.done":           "Transfer complete",
		"sftp.notify.failed":         "Transfer failed",
		"sftp.error.taskAction":      "Task action failed: %s",
		"sftp.taskName.fallback":     "File transfer",
		"sftp.notify.errorSeparator": ": ",
		"sftp.status.queued":         "Queued",
		"sftp.status.running":        "Transferring",
		"sftp.status.paused":         "Paused",
		"sftp.status.cancelled":      "Cancelled",
		"sftp.status.done":           "Completed",
		"sftp.status.error":          "Failed",
	}
)

func init() {
	add(ZhCN, sftpZh)
	add(EnUS, sftpEn)
}
