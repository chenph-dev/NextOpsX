package workspace

import (
	"testing"

	hoststore "nextopsx/store/hosts"
)

func assetTabTree() []hoststore.TreeNode {
	return []hoststore.TreeNode{
		{ID: "a", Name: "asset-a", NodeType: "host", Protocol: "mysql"},
		{ID: "b", Name: "asset-b", NodeType: "host", Protocol: "mysql"},
		{ID: "r1", Name: "redis-1", NodeType: "host", Protocol: "redis"},
		{ID: "r2", Name: "redis-2", NodeType: "host", Protocol: "redis"},
		{ID: "e1", Name: "es-1", NodeType: "host", Protocol: "elasticsearch"},
	}
}

func TestAssetTabsPreserveDBWorkbench(t *testing.T) {
	a := &State{Roots: assetTabTree()}
	na, nb := &a.Roots[0], &a.Roots[1]

	a.OpenAsset(na)
	if !a.IsAssetTabOpen("a") || a.DB.HostID != "a" || len(a.DB.Tabs) != 1 {
		t.Fatalf("open a: tabs=%v db=%+v", a.Tabs, a.DB)
	}
	a.NewDBQueryTab()
	if len(a.DB.Tabs) != 2 {
		t.Fatalf("a tabs = %d", len(a.DB.Tabs))
	}

	a.OpenAsset(nb)
	if a.DB.HostID != "b" || len(a.DB.Tabs) != 1 || len(a.Tabs) != 2 {
		t.Fatalf("open b: tabs=%v db=%+v", a.Tabs, a.DB)
	}

	a.ActivateAssetTab("a")
	if a.DB.HostID != "a" || len(a.DB.Tabs) != 2 {
		t.Fatalf("back to a: db=%+v", a.DB)
	}

	a.CloseAssetTab("b")
	if a.IsAssetTabOpen("b") || len(a.Tabs) != 1 {
		t.Fatalf("close b: tabs=%v", a.Tabs)
	}
}

func TestAssetTabsPreserveRedisAndES(t *testing.T) {
	a := &State{Roots: assetTabTree()}
	nr1, nr2, ne := &a.Roots[2], &a.Roots[3], &a.Roots[4]

	a.OpenAsset(nr1)
	a.Redis.Pattern = "user:*"
	a.OpenAsset(ne)
	a.OpenAsset(nr2)
	if len(a.Tabs) != 3 {
		t.Fatalf("tabs = %v", a.Tabs)
	}

	a.ActivateAssetTab("r1")
	if a.Redis.HostID != "r1" || a.Redis.Pattern != "user:*" {
		t.Fatalf("redis state = %+v", a.Redis)
	}

	a.ActivateAssetTab("e1")
	if a.ES.HostID != "e1" {
		t.Fatalf("es state = %+v", a.ES)
	}
}

func TestSelectOpenAssetSwitchesWorkbench(t *testing.T) {
	a := &State{Roots: assetTabTree()}
	na, nb := &a.Roots[0], &a.Roots[1]

	a.OpenAsset(na)
	a.NewDBQueryTab()
	a.OpenAsset(nb)
	a.Select(na)
	if a.DB.HostID != "a" || len(a.DB.Tabs) != 2 {
		t.Fatalf("select a: db=%+v", a.DB)
	}

	// An unopened asset only shows its selection panel; the active
	// workbench stays untouched.
	c := &hoststore.TreeNode{ID: "c", Name: "asset-c", NodeType: "host", Protocol: "mysql"}
	a.Select(c)
	if a.DB.HostID != "a" || a.IsAssetTabOpen("c") {
		t.Fatalf("select unopened c: db=%+v tabs=%v", a.DB, a.Tabs)
	}
}

func TestCloseActiveTabActivatesNeighbor(t *testing.T) {
	a := &State{Roots: assetTabTree()}
	na, nb := &a.Roots[0], &a.Roots[1]

	a.OpenAsset(na)
	a.OpenAsset(nb)
	a.CloseAssetTab("b")
	if a.Selected == nil || a.Selected.ID != "a" || a.DB.HostID != "a" {
		t.Fatalf("selected=%v db=%+v", a.Selected, a.DB)
	}
	if a.IsAssetTabOpen("b") || !a.IsAssetTabOpen("a") {
		t.Fatalf("tabs = %v", a.Tabs)
	}
}

func TestNextTabCyclesOpenAssets(t *testing.T) {
	a := &State{Roots: assetTabTree()}
	na, nb := &a.Roots[0], &a.Roots[1]

	a.OpenAsset(na)
	a.OpenAsset(nb)
	a.NextTab()
	if a.Selected == nil || a.Selected.ID != "a" {
		t.Fatalf("next tab: %v", a.Selected)
	}
	a.NextTab()
	if a.Selected == nil || a.Selected.ID != "b" {
		t.Fatalf("next tab again: %v", a.Selected)
	}
}

func TestCloseSessionsUnderDropsAssetTabs(t *testing.T) {
	a := &State{Roots: []hoststore.TreeNode{{
		ID: "f", Name: "folder", NodeType: "folder",
		Children: []hoststore.TreeNode{
			{ID: "a", Name: "asset-a", NodeType: "host", Protocol: "mysql"},
			{ID: "b", Name: "asset-b", NodeType: "host", Protocol: "mysql"},
		},
	}}}
	a.OpenAsset(&a.Roots[0].Children[0])
	a.OpenAsset(&a.Roots[0].Children[1])
	a.CloseSessionsUnder("f")
	if len(a.Tabs) != 0 || a.DB.HostID != "" {
		t.Fatalf("tabs=%v db=%+v", a.Tabs, a.DB)
	}
}
