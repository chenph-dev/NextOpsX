package backend

// OpenRdp launches the Windows Remote Desktop client for a WinRM asset.
// It only succeeds on Windows (see handler/rdp).
func (b *Backend) OpenRdp(hostID string) error {
	return b.rdp.OpenRdp(hostID)
}
