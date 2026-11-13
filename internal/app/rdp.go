package app

// OpenRDP launches the Windows Remote Desktop client for the selected WinRM
// asset. Errors are surfaced in the asset header's error strip.
func (a *App) OpenRDP() {
	n := a.Selected
	if a.Backend == nil || n == nil || n.NodeType != "host" || n.Protocol != "winrm" {
		return
	}
	hostID := n.ID
	a.ErrMsg = ""
	b := a.Backend
	go func() {
		if err := b.OpenRdp(hostID); err != nil {
			a.Post(func() {
				// 用户可能已切换到别的资产，别把错误挂到无关资产上。
				if a.Selected == nil || a.Selected.ID != hostID {
					return
				}
				a.ErrMsg = "启动远程桌面失败：" + err.Error()
			})
		}
	}()
}
