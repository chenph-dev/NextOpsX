package i18n

// WinRM console strings: a one-shot PowerShell command console for Windows
// assets. It is not a PTY terminal.
var (
	winrmZh = map[string]string{
		"winrm.empty":             "输入 PowerShell 命令开始。",
		"winrm.placeholder":       "输入 PowerShell 命令，Enter 执行",
		"winrm.running":           "执行中…",
		"winrm.status.connecting": "连接中…",
		"winrm.status.failed":     "连接失败",
		"winrm.retry":             "重新连接",
		"winrm.error.backend":     "后端未初始化",
		"winrm.error.executor":    "无法创建 WinRM 执行器",
	}

	winrmEn = map[string]string{
		"winrm.empty":             "Run a PowerShell command to get started.",
		"winrm.placeholder":       "Enter a PowerShell command and press Enter",
		"winrm.running":           "Running…",
		"winrm.status.connecting": "Connecting…",
		"winrm.status.failed":     "Connection failed",
		"winrm.retry":             "Reconnect",
		"winrm.error.backend":     "Backend not initialized",
		"winrm.error.executor":    "Cannot create the WinRM executor",
	}
)

func init() {
	add(ZhCN, winrmZh)
	add(EnUS, winrmEn)
}
