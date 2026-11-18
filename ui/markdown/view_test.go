package markdown_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/ui/markdown"
)

func TestViewTableRenders(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		markdown.View(c, "| 指标 | 数值 |\n|---|---|\n| CPU | 12% |")
	}, 1000, 900)
	for _, want := range []string{"指标", "数值", "CPU", "12%"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}
