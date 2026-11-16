package views

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/app"
	"nextopsx/internal/i18n"
	logsstore "nextopsx/internal/store/logs"
)

func auditTestApp(n int) *app.App {
	a := &app.App{View: app.ViewSettings}
	a.Settings.Tab = app.SettingsAudit
	a.Audit.Selected = -1
	a.Audit.Logs = make([]logsstore.CallLog, n)
	for i := range a.Audit.Logs {
		a.Audit.Logs[i] = logsstore.CallLog{
			ID:          fmt.Sprintf("log-%03d", i),
			Ts:          int64(1700000000 + i),
			Component:   "model",
			Name:        fmt.Sprintf("call-%03d", i),
			Provider:    "OpenAI",
			TokensIn:    i,
			TokensOut:   i * 2,
			TokensTotal: i * 3,
			DurationMS:  10,
			OK:          true,
		}
	}
	return a
}

func TestAuditColumnsWidened(t *testing.T) {
	cols := auditColumns()
	want := []float32{130, 80, 160, 130, 80, 80, 90, 88, 74}
	if len(cols) != len(want) {
		t.Fatalf("columns = %d, want %d", len(cols), len(want))
	}
	for i, width := range want {
		if cols[i].Width != width {
			t.Errorf("column %d width = %v, want %v", i, cols[i].Width, width)
		}
	}
}

func TestAuditPagination(t *testing.T) {
	i18n.SetLanguage(i18n.ZhCN)
	t.Cleanup(func() { i18n.SetLanguage(i18n.ZhCN) })

	a := auditTestApp(120)
	tt := ui.NewTester(func(c *ui.Context) { SettingsView(c, a) }, 1200, 800)
	if !tt.HasText("第 1 / 3 页") {
		t.Fatalf("first page label missing: %q", tt.Texts())
	}
	if !tt.HasText("共 120 条") {
		t.Fatalf("total count missing: %q", tt.Texts())
	}
	if !tt.HasText("call-000") {
		t.Fatalf("first page row missing: %q", tt.Texts())
	}
	if tt.HasText("call-050") {
		t.Fatalf("second page row shown on first page: %q", tt.Texts())
	}

	if err := tt.Click("下一页"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Audit.Page != 1 {
		t.Fatalf("page = %d, want 1", a.Audit.Page)
	}
	if a.Audit.Selected != -1 {
		t.Fatalf("selection = %d, want none after page change", a.Audit.Selected)
	}
	if !tt.HasText("第 2 / 3 页") {
		t.Fatalf("second page label missing: %q", tt.Texts())
	}
	if !tt.HasText("call-050") {
		t.Fatalf("second page row missing: %q", tt.Texts())
	}
	if tt.HasText("call-000") {
		t.Fatalf("first page row shown on second page: %q", tt.Texts())
	}

	if err := tt.Click("上一页"); err != nil {
		t.Fatal(err)
	}
	if a.Audit.Page != 0 {
		t.Fatalf("page = %d, want 0", a.Audit.Page)
	}
}
