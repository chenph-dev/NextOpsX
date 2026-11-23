package views_test

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	hoststore "nextopsx/store/hosts"
	"nextopsx/ui/views"
	"nextopsx/ui/workspace"
)

func TestAssetTabBarScrolls(t *testing.T) {
	a := &workspace.State{}
	for i := 0; i < 12; i++ {
		a.Roots = append(a.Roots, hoststore.TreeNode{
			ID: fmt.Sprintf("h%d", i), Name: fmt.Sprintf("asset-%02d", i),
			NodeType: "host", Protocol: "mysql",
		})
		a.Tabs = append(a.Tabs, fmt.Sprintf("h%d", i))
	}
	a.Selected = &a.Roots[0]
	tt := ui.NewTester(func(c *ui.Context) { views.Tabs(c, a) }, 420, 60)
	if _, ok := tt.Find("asset-00"); !ok {
		t.Fatalf("first tab missing: %q", tt.Texts())
	}

	// A horizontal scroll reaches the tabs past the right edge.
	tt.Scroll(200, 17, 4000, 0)
	tt.Frame()
	last, ok := tt.Find("asset-11")
	if !ok {
		t.Fatalf("last tab unreachable: %q", tt.Texts())
	}
	if last.X < 0 || last.X+last.W > 420 {
		t.Fatalf("last tab not in view: x=%.1f w=%.1f", last.X, last.W)
	}

	// Back at the start, activating the last tab scrolls it into view.
	tt.Scroll(200, 17, -4000, 0)
	tt.Frame()
	a.ActivateAssetTab("h11")
	tt.Frame()
	tt.Frame()
	last, ok = tt.Find("asset-11")
	if !ok {
		t.Fatalf("active tab not scrolled into view: %q", tt.Texts())
	}
	if last.X < 0 || last.X+last.W > 420 {
		t.Fatalf("active tab still clipped: x=%.1f w=%.1f", last.X, last.W)
	}
}

func TestAssetTabBarArrowsScroll(t *testing.T) {
	a := &workspace.State{}
	for i := 0; i < 12; i++ {
		a.Roots = append(a.Roots, hoststore.TreeNode{
			ID: fmt.Sprintf("h%d", i), Name: fmt.Sprintf("asset-%02d", i),
			NodeType: "host", Protocol: "mysql",
		})
		a.Tabs = append(a.Tabs, fmt.Sprintf("h%d", i))
	}
	a.Selected = &a.Roots[0]
	tt := ui.NewTester(func(c *ui.Context) { views.Tabs(c, a) }, 420, 60)
	tt.Frame()
	if a.TabScroll.MaxX <= 0 {
		t.Fatalf("tab bar should overflow, maxX=%v", a.TabScroll.MaxX)
	}

	for i := 0; i < 20; i++ {
		if err := tt.Click("向右滚动页签"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if a.TabScroll.X <= 0 {
		t.Fatalf("right arrow did not scroll: x=%v", a.TabScroll.X)
	}
	if _, ok := tt.Find("asset-11"); !ok {
		t.Fatalf("last tab unreachable: %q", tt.Texts())
	}

	before := a.TabScroll.X
	if err := tt.Click("向左滚动页签"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.TabScroll.X >= before {
		t.Fatalf("left arrow did not scroll: %v -> %v", before, a.TabScroll.X)
	}
}

func TestAssetTabBarRendersNonSSHAssets(t *testing.T) {
	a := &workspace.State{Roots: []hoststore.TreeNode{
		{ID: "a", Name: "mysql-a", NodeType: "host", Protocol: "mysql"},
		{ID: "b", Name: "mysql-b", NodeType: "host", Protocol: "mysql"},
		{ID: "r", Name: "redis-c", NodeType: "host", Protocol: "redis"},
	}}
	a.OpenAsset(&a.Roots[0])
	a.OpenAsset(&a.Roots[1])
	a.OpenAsset(&a.Roots[2])
	tt := ui.NewTester(views.Shell(a), 1200, 800)

	// Each open asset is named once in the tree and once in the tab bar.
	for _, name := range []string{"mysql-a", "mysql-b", "redis-c"} {
		count := 0
		for _, text := range tt.Texts() {
			if text == name {
				count++
			}
		}
		if count < 2 {
			t.Fatalf("%q rendered %d times, want tree + tab: %q", name, count, tt.Texts())
		}
	}
}
