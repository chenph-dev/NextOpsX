package ansi

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTruncateCommand(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := Truncate(long, 72)
	if n := len([]rune(got)); n != 72 {
		t.Fatalf("length = %d, want 72: %q", n, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("missing ellipsis: %q", got)
	}
	if short := Truncate("ls -la", 72); short != "ls -la" {
		t.Fatalf("short command changed: %q", short)
	}
}

func TestANSISpansInvalidExtendedColor(t *testing.T) {
	spans := Spans("\x1b[38;5;999mX\x1b[0mY")
	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.Text)
		if s.Color == (ui.Color{}) {
			t.Fatalf("invalid color made text transparent: %+v", s)
		}
	}
	if text.String() != "XY" {
		t.Fatalf("text = %q, want %q", text.String(), "XY")
	}
}

func TestLiveOutputTail(t *testing.T) {
	s := "old\n" + strings.Repeat("x", 200)
	got := LiveTail(s, 100)
	if !strings.HasPrefix(got, "…\n") {
		t.Fatalf("missing tail marker: %q", got)
	}
	if !strings.HasSuffix(got, "x") {
		t.Fatalf("did not keep the newest output: %q", got)
	}
	if short := LiveTail("short", 100); short != "short" {
		t.Fatalf("short output changed: %q", short)
	}
}

func TestANSISpansParsesColors(t *testing.T) {
	spans := Spans("\x1b[31mred\x1b[0m plain\x1b[2J")
	var got strings.Builder
	for _, s := range spans {
		got.WriteString(s.Text)
	}
	if got.String() != "red plain" {
		t.Fatalf("text = %q, want %q", got.String(), "red plain")
	}
	red := false
	for _, s := range spans {
		if s.Text == "red" && s.Color == Basic[1] {
			red = true
		}
	}
	if !red {
		t.Fatalf("red span not colored: %+v", spans)
	}
}

func TestANSISpansExtendedColors(t *testing.T) {
	spans := Spans("\x1b[38;5;196mX\x1b[0m\x1b[38;2;1;2;3mY\x1b[0m")
	var got strings.Builder
	for _, s := range spans {
		got.WriteString(s.Text)
	}
	if got.String() != "XY" {
		t.Fatalf("text = %q", got.String())
	}
	var x, y bool
	for _, s := range spans {
		switch s.Text {
		case "X":
			x = s.Color == ui.RGB(255, 0, 0)
		case "Y":
			y = s.Color == ui.RGB(1, 2, 3)
		}
	}
	if !x || !y {
		t.Fatalf("extended colors not parsed: %+v", spans)
	}
}
