package app

import "nextopsx/i18n"

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
				a.ErrMsg = i18n.Tf("rdp.openFailed", err.Error())
			})
		}
	}()
}
