package workspace

// openAssetTab adds an asset to the tab bar if it is not there yet.
func (a *State) openAssetTab(id string) {
	if id == "" || a.tabIndex(id) >= 0 {
		return
	}
	a.Tabs = append(a.Tabs, id)
	a.tabScrolled = ""
}

// TakeTabScroll reports whether the tab bar should scroll id into view,
// once after the active tab changes. The user can still scroll the bar to
// browse other tabs without it snapping back.
func (a *State) TakeTabScroll(id string) bool {
	if id == "" || a.tabScrolled == id {
		return false
	}
	a.tabScrolled = id
	return true
}

const tabScrollStep = 200

// ScrollTabsLeft moves the tab bar toward its start.
func (a *State) ScrollTabsLeft() { a.scrollTabs(-tabScrollStep) }

// ScrollTabsRight moves the tab bar toward its end.
func (a *State) ScrollTabsRight() { a.scrollTabs(tabScrollStep) }

func (a *State) scrollTabs(delta float32) {
	x := a.TabScroll.X + delta
	if x < 0 {
		x = 0
	}
	if x > a.TabScroll.MaxX {
		x = a.TabScroll.MaxX
	}
	a.TabScroll.X = x
}

func (a *State) tabIndex(id string) int {
	for i, tab := range a.Tabs {
		if tab == id {
			return i
		}
	}
	return -1
}

// IsAssetTabOpen reports whether the asset has an open tab.
func (a *State) IsAssetTabOpen(id string) bool { return a.tabIndex(id) >= 0 }

// HostProtocol returns the protocol of an asset in the tree.
func (a *State) HostProtocol(id string) string {
	if n := findTreeNode(a.Roots, id); n != nil {
		return n.Protocol
	}
	return ""
}

// ActivateAssetTab selects the asset and shows its preserved workbench.
func (a *State) ActivateAssetTab(id string) {
	n := findTreeNode(a.Roots, id)
	if n == nil {
		a.CloseAssetTab(id)
		return
	}
	a.tabScrolled = ""
	a.Select(n)
}

// CloseAssetTab closes an asset tab together with its session and workbench
// state, activating a neighboring tab when the closed one was selected.
func (a *State) CloseAssetTab(id string) {
	if s := a.SessionFor(id); s != nil {
		a.CloseSession(id)
		return
	}
	a.closeWorkbenchState(id)
	idx := a.removeAssetTab(id)
	if idx < 0 || a.Selected == nil || a.Selected.ID != id {
		return
	}
	a.Layout.SFTPOpen = false
	a.activateNeighborTab(idx)
}

func (a *State) removeAssetTab(id string) int {
	idx := a.tabIndex(id)
	if idx < 0 {
		return -1
	}
	a.Tabs = append(a.Tabs[:idx], a.Tabs[idx+1:]...)
	return idx
}

func (a *State) activateNeighborTab(idx int) {
	if len(a.Tabs) == 0 {
		return
	}
	if idx >= len(a.Tabs) {
		idx = len(a.Tabs) - 1
	}
	a.ActivateAssetTab(a.Tabs[idx])
}

// NextTab activates the next open asset tab.
func (a *State) NextTab() { a.cycleAssetTab(1) }

// PrevTab activates the previous open asset tab.
func (a *State) PrevTab() { a.cycleAssetTab(-1) }

func (a *State) cycleAssetTab(delta int) {
	if len(a.Tabs) == 0 {
		a.cycleSession(delta)
		return
	}
	idx := -1
	if a.Selected != nil {
		idx = a.tabIndex(a.Selected.ID)
	}
	if idx < 0 {
		a.ActivateAssetTab(a.Tabs[0])
		return
	}
	n := len(a.Tabs)
	a.ActivateAssetTab(a.Tabs[((idx+delta)%n+n)%n])
}

// tabAfter returns the tab after id, wrapping around, or "".
func (a *State) tabAfter(id string) string {
	idx := a.tabIndex(id)
	if idx < 0 || len(a.Tabs) < 2 {
		return ""
	}
	return a.Tabs[(idx+1)%len(a.Tabs)]
}

func (a *State) saveActiveWorkbenches() {
	if a.DB.HostID != "" {
		a.saveDBActive()
	}
	if a.Redis.HostID != "" {
		a.saveRedisActive()
	}
	if a.ES.HostID != "" {
		a.saveESActive()
	}
}

func (a *State) saveDBActive() {
	if a.DB.HostID == "" {
		return
	}
	if a.dbStates == nil {
		a.dbStates = map[string]DBState{}
	}
	a.dbStates[a.DB.HostID] = a.DB
}

func (a *State) saveRedisActive() {
	if a.Redis.HostID == "" {
		return
	}
	if a.redisStates == nil {
		a.redisStates = map[string]RedisState{}
	}
	a.redisStates[a.Redis.HostID] = a.Redis
}

func (a *State) saveESActive() {
	if a.ES.HostID == "" {
		return
	}
	if a.esStates == nil {
		a.esStates = map[string]ESState{}
	}
	a.esStates[a.ES.HostID] = a.ES
}

func (a *State) closeWorkbenchState(id string) {
	delete(a.dbStates, id)
	delete(a.redisStates, id)
	delete(a.esStates, id)
	if a.DB.HostID == id {
		a.DB = DBState{}
	}
	if a.Redis.HostID == id {
		a.Redis = RedisState{}
	}
	if a.ES.HostID == id {
		a.ES = ESState{}
	}
}

// resumeDBState clears flags left by work that was dropped while the asset
// was inactive, and reloads a schema that never arrived.
func (a *State) resumeDBState() {
	a.DB.SchemaLoading = false
	for _, tab := range a.DB.Tabs {
		if !tab.Running {
			continue
		}
		tab.Running = false
		// A table page is safe to reload; a query may not be, so leave it
		// for the user to rerun.
		if tab.Kind == DBTabTable {
			a.loadDBTablePage(tab, tab.Page, false)
		}
	}
	if a.DB.HostID != "" && a.DB.Schema == nil && a.DB.SchemaErr == "" {
		a.loadDBSchema(a.DB.HostID)
	}
}

// resumeRedisState clears flags left by work that was dropped while the
// asset was inactive, and reloads keys that never arrived.
func (a *State) resumeRedisState() {
	a.Redis.KeysLoading = false
	a.Redis.Detail.Loading = false
	a.Redis.InfoLoading = false
	a.Redis.CommandRunning = false
	if a.Redis.HostID == "" || a.Redis.KeysLoaded || a.Redis.KeysErr != "" {
		return
	}
	if len(a.Redis.pageStarts) == 0 {
		a.RedisRefreshKeys(true)
		return
	}
	a.RedisReloadKeys()
}

// resumeESState clears flags left by work that was dropped while the asset
// was inactive, and reloads indices that never arrived.
func (a *State) resumeESState() {
	a.ES.IndicesLoading = false
	a.ES.QueryRunning = false
	a.ES.InfoLoading = false
	a.ES.CommandRunning = false
	if a.ES.HostID != "" && !a.ES.IndicesLoaded && a.ES.IndicesErr == "" {
		a.ESRefreshIndices()
	}
}
