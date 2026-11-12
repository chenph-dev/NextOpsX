package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo"

	"nextopsx/internal/connector"
	hoststore "nextopsx/internal/store/hosts"
)

// HostParamField is one driver-specific parameter rendered in the host form.
// Values live in a slice so the input can bind to the element's address.
type HostParamField struct {
	Key         string
	Label       string
	Placeholder string
	Required    bool
	File        bool
	Value       string
}

// ProtocolOption is a choice in the host form's protocol picker.
type ProtocolOption struct {
	Value     string
	Label     string
	NeedsHost bool
}

// HostForm is the state of the new/edit asset dialog.
type HostForm struct {
	Open             bool
	EditingID        string
	IsFolder         bool
	ParentID         string
	Name             string
	Protocol         string
	Addr             string
	Port             string
	User             string
	AuthType         string
	AuthLabel        string
	Secret           string
	AutoApprove      string
	AutoApproveLabel string
	RdpPort          string
	Params           []HostParamField

	OriginalProtocol string
	OriginalAuthType string

	Err     string
	TestMsg string
	TestOK  bool
	Busy    bool

	seq uint64
}

// HostsState holds asset dialogs and the connector catalog.
type HostsState struct {
	Connectors []*connector.Connector

	Form HostForm
	Err  string

	RenameOpen bool
	RenameID   string
	RenameName string

	formSeq uint64

	MoveOpen bool
	MoveID   string

	DeleteOpen     bool
	DeleteID       string
	DeleteName     string
	DeleteIsFolder bool
}

// ensureHostConnectors loads the registry once. The registry is static, so
// this also works in tests with no backend.
func (a *App) ensureHostConnectors() {
	if len(a.Hosts.Connectors) == 0 {
		a.Hosts.Connectors = connector.List()
	}
}

// HostProtocolOptions returns ssh/winrm first, then database connectors.
func (a *App) HostProtocolOptions() []ProtocolOption {
	a.ensureHostConnectors()
	var out []ProtocolOption
	add := func(protocol string) {
		for _, c := range a.Hosts.Connectors {
			if strings.EqualFold(c.Protocol, protocol) {
				out = append(out, ProtocolOption{Value: c.Protocol, Label: c.Name, NeedsHost: c.NeedsHost})
				return
			}
		}
	}
	add("ssh")
	add("winrm")
	for _, c := range a.Hosts.Connectors {
		if strings.EqualFold(c.Protocol, "ssh") || strings.EqualFold(c.Protocol, "winrm") {
			continue
		}
		out = append(out, ProtocolOption{Value: c.Protocol, Label: c.Name, NeedsHost: c.NeedsHost})
	}
	return out
}

// HostProtocolName returns a protocol's display name.
func (a *App) HostProtocolName(protocol string) string {
	if protocol == "" {
		return "SSH"
	}
	a.ensureHostConnectors()
	for _, c := range a.Hosts.Connectors {
		if strings.EqualFold(c.Protocol, protocol) {
			return c.Name
		}
	}
	return protocol
}

func (a *App) hostConnector(protocol string) *connector.Connector {
	if protocol == "" {
		protocol = "ssh"
	}
	a.ensureHostConnectors()
	for _, c := range a.Hosts.Connectors {
		if strings.EqualFold(c.Protocol, protocol) {
			return c
		}
	}
	return nil
}

// HostFormNeedsHost reports whether the selected protocol uses host/port/user.
func (a *App) HostFormNeedsHost() bool {
	c := a.hostConnector(a.Hosts.Form.Protocol)
	if c == nil {
		return true
	}
	return c.NeedsHost
}

// HostFormIsSSH reports whether the form edits an SSH asset.
func (a *App) HostFormIsSSH() bool {
	p := a.Hosts.Form.Protocol
	return p == "" || strings.EqualFold(p, "ssh")
}

// HostFormIsWinRM reports whether the form edits a WinRM asset.
func (a *App) HostFormIsWinRM() bool {
	return strings.EqualFold(a.Hosts.Form.Protocol, "winrm")
}

// IsSSHProtocol reports whether an asset protocol supports the interactive
// SSH terminal (an empty protocol means SSH, for legacy assets).
func (a *App) IsSSHProtocol(protocol string) bool {
	return protocol == "" || strings.EqualFold(protocol, "ssh")
}

// HostFormParentLabel names the folder a new asset goes into.
func (a *App) HostFormParentLabel() string {
	if a.Hosts.Form.ParentID == "" {
		return "根目录"
	}
	if a.Selected != nil && a.Selected.ID == a.Hosts.Form.ParentID {
		return a.Selected.Name
	}
	if n := findTreeNode(a.Roots, a.Hosts.Form.ParentID); n != nil {
		return n.Name
	}
	return "根目录"
}

// nextHostFormSeq returns a monotonically increasing id for one opening of
// the host form. It lives on HostsState so replacing HostForm cannot reset
// it and let a stale callback match a newly opened form.
func (a *App) nextHostFormSeq() uint64 {
	a.Hosts.formSeq++
	return a.Hosts.formSeq
}

// OpenNewHostForm opens the form for a new asset in the selected folder.
func (a *App) OpenNewHostForm() {
	a.Hosts.RenameOpen = false
	a.Hosts.MoveOpen = false
	a.Hosts.DeleteOpen = false
	a.Hosts.Err = ""
	a.Hosts.Form = HostForm{
		Open: true, Protocol: "ssh", Port: "22", User: "root",
		AuthType: "password", AuthLabel: "密码",
		AutoApprove: "inherit", AutoApproveLabel: "继承全局", RdpPort: "3389",
		ParentID: a.selectedParentID(),
		Params:   hostParamFields(a.hostConnector("ssh"), nil),
	}
	a.Hosts.Form.seq = a.nextHostFormSeq()
}

// OpenNewFolderForm opens the form for a folder in the selected folder.
func (a *App) OpenNewFolderForm() {
	a.Hosts.RenameOpen = false
	a.Hosts.MoveOpen = false
	a.Hosts.DeleteOpen = false
	a.Hosts.Err = ""
	a.Hosts.Form = HostForm{
		Open: true, IsFolder: true,
		AutoApprove: "inherit", AutoApproveLabel: "继承全局",
		ParentID: a.selectedParentID(),
	}
	a.Hosts.Form.seq = a.nextHostFormSeq()
}

// OpenEditHostForm opens the form for the selected host.
func (a *App) OpenEditHostForm() {
	n := a.Selected
	if n == nil || n.NodeType != "host" {
		return
	}
	proto := n.Protocol
	if proto == "" {
		proto = "ssh"
	}
	auth := n.AuthType
	if auth == "" {
		auth = "password"
	}
	port := n.Port
	if port == 0 {
		port = defaultPort(proto)
	}
	a.Hosts.RenameOpen = false
	a.Hosts.MoveOpen = false
	a.Hosts.DeleteOpen = false
	a.Hosts.Err = ""
	a.Hosts.Form = HostForm{
		Open: true, EditingID: n.ID, ParentID: n.ParentID,
		Name: n.Name, Protocol: proto, Addr: n.Addr,
		Port: strconv.Itoa(port), User: n.User, AuthType: auth,
		AuthLabel:        authDisplay(auth),
		AutoApprove:      canonicalAutoApprove(n.AutoApprove),
		AutoApproveLabel: autoApproveDisplay(n.AutoApprove),
		RdpPort:          strconv.Itoa(rdpPortOrDefault(n.RdpPort)),
		Params:           hostParamFields(a.hostConnector(proto), n.Params),
		OriginalProtocol: proto, OriginalAuthType: auth,
	}
	a.Hosts.Form.seq = a.nextHostFormSeq()
}

// CloseHostForm discards the form.
func (a *App) CloseHostForm() {
	a.Hosts.Form.Open = false
	a.Hosts.Form.Err = ""
	a.Hosts.Form.TestMsg = ""
	a.Hosts.Form.Busy = false
	a.Hosts.Form.seq = a.nextHostFormSeq()
}

// SyncHostProtocol rebuilds driver parameters and default ports after the
// protocol picker changes. The bound value is already updated when a view
// calls this from OnChange.
func (a *App) SyncHostProtocol() {
	f := &a.Hosts.Form
	if f.IsFolder {
		return
	}
	old := map[string]string{}
	for _, p := range f.Params {
		old[p.Key] = p.Value
	}
	f.Params = hostParamFields(a.hostConnector(f.Protocol), nil)
	for i := range f.Params {
		if v, ok := old[f.Params[i].Key]; ok {
			f.Params[i].Value = v
		}
	}
	if !a.HostFormIsSSH() {
		f.AuthType = "password"
		f.AuthLabel = "密码"
	}
	if a.HostFormNeedsHost() {
		f.Port = strconv.Itoa(defaultPort(f.Protocol))
	} else {
		f.Port = ""
	}
	if a.HostFormIsWinRM() && f.RdpPort == "" {
		f.RdpPort = "3389"
	}
	f.Err = ""
	f.TestMsg = ""
}

func (a *App) selectedParentID() string {
	n := a.Selected
	if n == nil {
		return ""
	}
	if n.NodeType == "folder" {
		return n.ID
	}
	return n.ParentID
}

func defaultPort(protocol string) int {
	switch strings.ToLower(protocol) {
	case "winrm":
		return 5985
	case "mysql":
		return 3306
	case "postgres":
		return 5432
	case "redis":
		return 6379
	case "elasticsearch":
		return 9200
	case "clickhouse":
		return 9000
	case "sqlserver":
		return 1433
	default:
		return 22
	}
}

func rdpPortOrDefault(v int) int {
	if v == 0 {
		return 3389
	}
	return v
}

// SyncHostAuth maps the Chinese auth picker label to its canonical value.
func (a *App) SyncHostAuth() {
	a.Hosts.Form.AuthType = canonicalAuth(a.Hosts.Form.AuthLabel)
}

// SyncHostAutoApprove maps the auto-approve picker label to its canonical
// inherit/on/off value.
func (a *App) SyncHostAutoApprove() {
	a.Hosts.Form.AutoApprove = canonicalAutoApprove(a.Hosts.Form.AutoApproveLabel)
}

func authDisplay(auth string) string {
	if strings.EqualFold(auth, "privatekey") {
		return "私钥"
	}
	return "密码"
}

func canonicalAuth(v string) string {
	switch v {
	case "privatekey", "私钥":
		return "privatekey"
	default:
		return "password"
	}
}

func canonicalAutoApprove(v string) string {
	switch v {
	case "on", "自动放行":
		return "on"
	case "off", "人工审批":
		return "off"
	default:
		return "inherit"
	}
}

// AuthDisplay returns the Chinese label for an auth type.
func (a *App) AuthDisplay(auth string) string { return authDisplay(auth) }

// AutoApproveDisplay returns the Chinese label for an auto-approve value.
func (a *App) AutoApproveDisplay(v string) string { return autoApproveDisplay(v) }

func autoApproveDisplay(v string) string {
	switch canonicalAutoApprove(v) {
	case "on":
		return "自动放行"
	case "off":
		return "人工审批"
	default:
		return "继承全局"
	}
}

func hostParamFields(c *connector.Connector, existing map[string]any) []HostParamField {
	if c == nil {
		return nil
	}
	out := make([]HostParamField, 0, len(c.Params))
	for _, p := range c.Params {
		f := HostParamField{
			Key: p.Key, Label: p.Label, Placeholder: p.Placeholder,
			Required: p.Required, File: p.Type == connector.ParamFile,
		}
		if v, ok := existing[p.Key]; ok && v != nil {
			f.Value = fmt.Sprint(v)
		} else if p.Default != nil {
			f.Value = fmt.Sprint(p.Default)
		}
		if f.Label == "" {
			f.Label = p.Key
		}
		out = append(out, f)
	}
	return out
}

// hostInputFromForm validates the form and builds the store input.
func (a *App) hostInputFromForm() (hoststore.HostInput, error) {
	f := &a.Hosts.Form
	name := strings.TrimSpace(f.Name)
	if name == "" {
		return hoststore.HostInput{}, fmt.Errorf("请输入名称")
	}
	autoApprove := canonicalAutoApprove(f.AutoApprove)
	if f.AutoApproveLabel != "" {
		autoApprove = canonicalAutoApprove(f.AutoApproveLabel)
	}
	in := hoststore.HostInput{
		Name: name, ParentID: f.ParentID, Protocol: f.Protocol,
		AutoApprove: autoApprove,
	}
	if f.IsFolder {
		return in, nil
	}
	if in.Protocol == "" {
		in.Protocol = "ssh"
	}
	needsHost := a.HostFormNeedsHost()
	if needsHost {
		in.Addr = strings.TrimSpace(f.Addr)
		if in.Addr == "" {
			return in, fmt.Errorf("请输入地址")
		}
		port, err := strconv.Atoi(strings.TrimSpace(f.Port))
		if err != nil || port < 1 || port > 65535 {
			return in, fmt.Errorf("端口需为 1-65535")
		}
		in.Port = port
		in.User = strings.TrimSpace(f.User)
		if in.User == "" {
			return in, fmt.Errorf("请输入用户名")
		}
		auth := canonicalAuth(f.AuthType)
		if f.AuthLabel != "" {
			auth = canonicalAuth(f.AuthLabel)
		}
		if !a.HostFormIsSSH() {
			auth = "password"
		}
		if auth != "password" && auth != "privatekey" {
			return in, fmt.Errorf("请选择认证方式")
		}
		if f.EditingID == "" && strings.TrimSpace(f.Secret) == "" {
			return in, fmt.Errorf("请输入密码或私钥")
		}
		if f.EditingID != "" && strings.TrimSpace(f.Secret) == "" {
			if f.OriginalProtocol != "" && in.Protocol != f.OriginalProtocol {
				return in, fmt.Errorf("切换协议后请重新输入凭据")
			}
			if f.OriginalAuthType != "" && auth != f.OriginalAuthType {
				return in, fmt.Errorf("切换认证方式后请重新输入凭据")
			}
		}
		in.AuthType = auth
		in.Secret = f.Secret
		rdp, _ := strconv.Atoi(strings.TrimSpace(f.RdpPort))
		in.RdpPort = rdpPortOrDefault(rdp)
	}
	params := map[string]any{}
	for _, p := range f.Params {
		v := strings.TrimSpace(p.Value)
		if p.Required && v == "" {
			return in, fmt.Errorf("请填写%s", p.Label)
		}
		if v != "" {
			params[p.Key] = v
		}
	}
	in.Params = params
	return in, nil
}

// SaveHostForm validates and persists the form (asset or folder).
func (a *App) SaveHostForm() {
	f := &a.Hosts.Form
	if f.Busy {
		return
	}
	in, err := a.hostInputFromForm()
	if err != nil {
		f.Err = err.Error()
		return
	}
	f.Err = ""
	f.Busy = true
	f.TestMsg = ""
	seq := f.seq
	if a.Backend == nil {
		f.Busy = false
		f.Open = false
		return
	}
	b := a.Backend
	editingID, isFolder := f.EditingID, f.IsFolder
	go func() {
		var createdID string
		var err error
		switch {
		case isFolder:
			createdID, err = b.CreateFolder(in.Name, in.ParentID)
		case editingID == "":
			createdID, err = b.CreateHost(in)
		default:
			err = b.UpdateHost(editingID, in)
		}
		a.Post(func() {
			f := &a.Hosts.Form
			current := seq == f.seq
			if current {
				f.Busy = false
				f.Err = ""
				if err != nil {
					f.Err = "保存失败：" + err.Error()
				}
			}
			if err != nil {
				return
			}
			if current {
				f.Open = false
			}
			if editingID != "" && a.DB.HostID == editingID {
				if a.IsDBProtocol(in.Protocol) {
					a.DB.HostName = in.Name
					a.DB.Protocol = in.Protocol
					a.RefreshDBSchema()
				}
			}
			target := createdID
			if target == "" {
				target = editingID
			}
			a.refreshHosts(target)
		})
	}()
}

// TestHostConnection runs the backend's connection check for the form.
func (a *App) TestHostConnection() {
	f := &a.Hosts.Form
	if f.Busy || f.IsFolder {
		return
	}
	in, err := a.hostInputFromForm()
	if err != nil {
		f.Err = err.Error()
		return
	}
	if a.HostFormNeedsHost() && strings.TrimSpace(in.Secret) == "" && f.EditingID != "" {
		f.TestMsg = "编辑已有资产时，请重新输入密码/私钥后再测试"
		f.TestOK = false
		return
	}
	if a.Backend == nil {
		f.TestMsg = "无法测试：后端未初始化"
		f.TestOK = false
		return
	}
	seq := f.seq
	f.Err = ""
	f.Busy = true
	f.TestMsg = "正在测试连接…"
	b := a.Backend
	go func() {
		ok, err := b.TestConnection(in)
		a.Post(func() {
			f := &a.Hosts.Form
			if seq != f.seq {
				return
			}
			f.Busy = false
			if err != nil {
				f.TestMsg = "连接失败：" + err.Error()
				f.TestOK = false
				return
			}
			if ok {
				f.TestMsg = "连接成功"
				f.TestOK = true
			} else {
				f.TestMsg = "连接失败"
				f.TestOK = false
			}
		})
	}()
}

// PickHostParamFile opens the native file dialog for a file parameter.
func (a *App) PickHostParamFile(index int) {
	if index < 0 || index >= len(a.Hosts.Form.Params) {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "选择文件",
			ButtonLabel: "选择",
			Filters: []mygo.FileFilter{
				{Name: "SQLite 数据库", Extensions: []string{"db", "sqlite", "sqlite3"}},
			},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		a.Post(func() {
			if index < len(a.Hosts.Form.Params) {
				a.Hosts.Form.Params[index].Value = paths[0]
			}
		})
	}()
}

// TermHostLabel names the asset owning the background terminal.
func (a *App) TermHostLabel() string {
	if a.TermHostID == "" {
		return ""
	}
	if n := findTreeNode(a.Roots, a.TermHostID); n != nil {
		return n.Name
	}
	return "未知资产"
}

// RefreshHosts reloads the tree, keeping the current selection when it
// still exists.
func (a *App) RefreshHosts() {
	a.refreshHosts("")
}

// refreshHosts reloads the tree and re-points the selection at selectID.
func (a *App) refreshHosts(selectID string) {
	if a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		roots, err := b.ListHosts()
		a.Post(func() {
			if err != nil {
				a.Hosts.Err = "刷新资产失败：" + err.Error()
				return
			}
			a.Roots = roots
			a.Hosts.Err = ""
			a.mergeOpen(roots)
			if selectID != "" {
				if n := findTreeNode(roots, selectID); n != nil {
					a.Selected = n
					a.attachAsset(n)
				}
				return
			}
			if a.Selected == nil {
				return
			}
			if n := findTreeNode(roots, a.Selected.ID); n != nil {
				a.Selected = n
				a.attachAsset(n)
			} else {
				a.Selected = nil
			}
		})
	}()
}

// mergeOpen keeps existing open/closed folder state and opens new folders.
func (a *App) mergeOpen(nodes []hoststore.TreeNode) {
	if a.Open == nil {
		a.Open = map[string]*bool{}
	}
	for i := range nodes {
		n := &nodes[i]
		if n.NodeType == "folder" {
			if _, ok := a.Open[n.ID]; !ok {
				open := true
				a.Open[n.ID] = &open
			}
		}
		a.mergeOpen(n.Children)
	}
}

// treeContains reports whether target is id or below it in the tree.
func treeContains(nodes []hoststore.TreeNode, id, target string) bool {
	n := findTreeNode(nodes, id)
	return n != nil && subtreeContains(*n, target)
}

func subtreeContains(n hoststore.TreeNode, target string) bool {
	if n.ID == target {
		return true
	}
	for _, c := range n.Children {
		if subtreeContains(c, target) {
			return true
		}
	}
	return false
}

func findTreeNode(nodes []hoststore.TreeNode, id string) *hoststore.TreeNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if n := findTreeNode(nodes[i].Children, id); n != nil {
			return n
		}
	}
	return nil
}

// OpenDeleteDialog asks for confirmation before deleting the selection.
func (a *App) OpenDeleteDialog() {
	n := a.Selected
	if n == nil {
		return
	}
	a.Hosts.DeleteOpen = true
	a.Hosts.DeleteID = n.ID
	a.Hosts.DeleteName = n.Name
	a.Hosts.DeleteIsFolder = n.NodeType == "folder"
}

// ConfirmDeleteNode deletes the node chosen in the confirmation dialog.
func (a *App) ConfirmDeleteNode() {
	id := a.Hosts.DeleteID
	if id == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	disconnectTerm := a.TermHostID != "" &&
		(a.TermHostID == id || treeContains(a.Roots, id, a.TermHostID))
	go func() {
		err := b.DeleteNode(id)
		a.Post(func() {
			if err != nil {
				a.Hosts.Err = "删除失败：" + err.Error()
				return
			}
			if disconnectTerm {
				a.Disconnect()
			}
			a.Hosts.Err = ""
			a.Hosts.DeleteID = ""
			a.Hosts.DeleteName = ""
			if a.Selected != nil && a.Selected.ID == id {
				a.Selected = nil
			}
			a.refreshHosts("")
		})
	}()
}

// OpenRenameDialog asks for a new name for the selection.
func (a *App) OpenRenameDialog() {
	n := a.Selected
	if n == nil {
		return
	}
	a.Hosts.Err = ""
	a.Hosts.RenameOpen = true
	a.Hosts.RenameID = n.ID
	a.Hosts.RenameName = n.Name
}

// SaveRename applies the rename dialog.
func (a *App) SaveRename() {
	id := a.Hosts.RenameID
	name := strings.TrimSpace(a.Hosts.RenameName)
	if id == "" {
		return
	}
	if name == "" {
		a.Hosts.Err = "请输入名称"
		return
	}
	if a.Backend == nil {
		return
	}
	b := a.Backend
	a.Hosts.RenameOpen = false
	go func() {
		err := b.RenameNode(id, name)
		a.Post(func() {
			if err != nil {
				a.Hosts.Err = "重命名失败：" + err.Error()
				return
			}
			a.Hosts.Err = ""
			a.refreshHosts(id)
		})
	}()
}

// OpenMoveDialog lists folders the selection can move into.
func (a *App) OpenMoveDialog() {
	n := a.Selected
	if n == nil {
		return
	}
	a.Hosts.Err = ""
	a.Hosts.MoveOpen = true
	a.Hosts.MoveID = n.ID
}

// MoveTargets returns candidate folders, excluding the moved node and its
// descendants (which would create a cycle).
func (a *App) MoveTargets() []hoststore.TreeNode {
	id := a.Hosts.MoveID
	var out []hoststore.TreeNode
	var walk func(nodes []hoststore.TreeNode, underMoved bool)
	walk = func(nodes []hoststore.TreeNode, underMoved bool) {
		for i := range nodes {
			n := nodes[i]
			if underMoved || n.ID == id {
				walk(n.Children, true)
				continue
			}
			if n.NodeType == "folder" {
				out = append(out, n)
			}
			walk(n.Children, false)
		}
	}
	walk(a.Roots, false)
	return out
}

// MoveNodeTo moves the selection under parentID ("" for the root).
func (a *App) MoveNodeTo(parentID string) {
	id := a.Hosts.MoveID
	if id == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	a.Hosts.MoveOpen = false
	go func() {
		err := b.MoveNode(id, parentID)
		a.Post(func() {
			if err != nil {
				a.Hosts.Err = "移动失败：" + err.Error()
				return
			}
			a.Hosts.Err = ""
			a.refreshHosts(id)
		})
	}()
}
