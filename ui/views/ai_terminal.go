package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"nextopsx/i18n"
	convstore "nextopsx/store/conversations"
	"nextopsx/ui/app"
)

// Terminal palette: intentionally fixed rather than following the window
// theme, so command output reads as a terminal in both light and dark UIs.
var (
	termBg     = ui.RGB(24, 24, 27)
	termBorder = ui.RGB(63, 63, 70)
	termFg     = ui.RGB(228, 228, 231)
	termMuted  = ui.RGB(161, 161, 170)
)

// aiTerminalBubble renders one executed command and its output as a terminal.
// The header carries the command, exit status and line count; the body shows
// the output with ANSI SGR colors parsed into spans.
func aiTerminalBubble(c *ui.Context, a *app.App, msg convstore.Message, meta aiToolMeta, status string) {
	command := meta.Command
	if command == "" {
		command = msg.ToolName
	}
	if command == "" {
		command = i18n.T("ai.tool.output")
	}

	open := a.AI.ToolOpen[msg.ID]
	head := "▸ "
	if open {
		head = "▾ "
	}
	head += "$ " + truncateCommand(command, 72)
	switch {
	case status == app.AIStatusRejected:
		head += i18n.T("ai.tool.rejected")
	case meta.Cancelled:
		head += i18n.T("ai.tool.cancelled")
	case meta.Command != "" && msg.ToolResult != "":
		head += fmt.Sprintf(" · exit %d", meta.ExitCode)
	}
	if lines := countOutputLines(msg.Content); lines > 0 {
		head += i18n.Tf("ai.tool.lines", lines)
	}

	copyText := "$ " + command
	if strings.TrimSpace(msg.Content) != "" {
		copyText += "\n" + msg.Content
	}

	ui.Column(c).AlignItems(ui.Start).Gap(2).Children(func() {
		ui.Column(c).MaxWidthPercent(100).AlignItems(ui.Stretch).
			Background(termBg).Border(1, termBorder).Radius(8).
			Padding(8, 10).Gap(6).Children(func() {
			ui.Button(c, head).Font("monospace").FontSize(fontSize(12)).
				TextColor(termFg).Background(termBg).SingleLine().MinWidth(0).
				Justify(ui.Start).Tooltip(command).
				OnClick(func() { a.ToggleToolOutput(msg.ID) })
			if open {
				aiTerminalOutput(c, msg.Content, 280)
			}
		})
		ui.Row(c).Justify(ui.Start).Children(func() {
			aiBubbleToolbar(c, termMuted, aiCopyAction(copyText))
		})
	})
}

// aiTerminalOutput renders command output in a scrollable terminal area.
func aiTerminalOutput(c *ui.Context, output string, maxHeight float32) {
	if strings.TrimSpace(output) == "" {
		ui.Text(c, i18n.T("ai.tool.noOutput")).Font("monospace").
			FontSize(fontSize(12)).TextColor(termMuted)
		return
	}
	ui.Scroll(c).MaxHeight(maxHeight).Children(func() {
		ui.RichText(c, ansiSpans(output)...).Font("monospace").
			FontSize(fontSize(12)).TextColor(termFg).Wrap()
	})
}

// aiLiveTerminalOutput renders a bounded tail of streaming output. Reparsing
// the whole accumulated buffer on every frame would make a chatty command
// O(n²); the finalized tool message still renders the complete output.
func aiLiveTerminalOutput(c *ui.Context, output string, maxHeight float32) {
	aiTerminalOutput(c, liveOutputTail(output, 8<<10), maxHeight)
}

func liveOutputTail(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	s = s[len(s)-limit:]
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	return "…\n" + s
}

// aiToolTextBubble renders a non-command tool result as Markdown, so tool
// output that is text (file contents, query results, notes) keeps headings,
// lists and code formatting instead of looking like a raw dump.
func aiToolTextBubble(c *ui.Context, msg convstore.Message, t *ui.Theme) {
	name := msg.ToolName
	if name == "" {
		name = i18n.T("ai.tool.output")
	}
	empty := strings.TrimSpace(msg.Content) == ""
	ui.Column(c).AlignItems(ui.Start).Gap(2).Children(func() {
		ui.Column(c).MaxWidth(720).AlignItems(ui.Stretch).Padding(8, 10).Radius(8).
			Background(t.Surface).Border(1, t.Border).Children(func() {
			ui.Text(c, name).FontSize(fontSize(11)).TextColor(t.TextMuted)
			if empty {
				ui.Text(c, i18n.T("ai.tool.noOutput")).FontSize(fontSize(12)).
					TextColor(t.TextMuted)
				return
			}
			MarkdownView(c, msg.Content)
		})
		if !empty {
			ui.Row(c).Justify(ui.Start).Children(func() {
				aiBubbleToolbar(c, t.TextMuted, aiCopyAction(msg.Content))
			})
		}
	})
}

// truncateCommand keeps a collapsed command header within the bubble; the
// full command is still what the copy action puts on the clipboard.
func truncateCommand(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// ansiStyle is the SGR state while scanning terminal output.
type ansiStyle struct {
	fg, bg            ui.Color
	hasFg, hasBg      bool
	bold, dim         bool
	italic, underline bool
	inverse           bool
}

var ansiBasic = []ui.Color{
	ui.RGB(63, 63, 70),    // black
	ui.RGB(248, 113, 113), // red
	ui.RGB(74, 222, 128),  // green
	ui.RGB(250, 204, 21),  // yellow
	ui.RGB(96, 165, 250),  // blue
	ui.RGB(232, 121, 249), // magenta
	ui.RGB(34, 211, 238),  // cyan
	ui.RGB(228, 228, 231), // white
}

var ansiBright = []ui.Color{
	ui.RGB(113, 113, 122),
	ui.RGB(252, 165, 165),
	ui.RGB(134, 239, 172),
	ui.RGB(253, 224, 71),
	ui.RGB(147, 197, 253),
	ui.RGB(240, 171, 252),
	ui.RGB(103, 232, 249),
	ui.RGB(250, 250, 250),
}

// ansiSpans parses ANSI SGR colors and styles from terminal output. Other
// escape sequences (cursor movement, erase, OSC hyperlinks) are removed.
func ansiSpans(s string) []ui.Span {
	// Progress output uses carriage returns; show each update on its own line
	// rather than letting \r corrupt the paragraph layout.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	var spans []ui.Span
	var buf strings.Builder
	st := ansiStyle{}

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		sp := st.span(buf.String())
		buf.Reset()
		if n := len(spans); n > 0 && sameSpanStyle(spans[n-1], sp) {
			spans[n-1].Text += sp.Text
			return
		}
		spans = append(spans, sp)
	}

	for i := 0; i < len(s); {
		if s[i] != 0x1b {
			buf.WriteByte(s[i])
			i++
			continue
		}
		if i+1 >= len(s) {
			break
		}
		switch s[i+1] {
		case '[': // CSI
			j := i + 2
			for j < len(s) && (s[j] >= 0x30 && s[j] <= 0x3f || s[j] >= 0x20 && s[j] <= 0x2f) {
				j++
			}
			if j >= len(s) {
				i = len(s)
				break
			}
			if s[j] == 'm' {
				// Flush with the old style before the SGR changes it.
				flush()
				st.applySGR(s[i+2 : j])
			}
			i = j + 1
		case ']': // OSC: skip to BEL or ST
			j := i + 2
			for j < len(s) {
				if s[j] == 0x07 {
					j++
					break
				}
				if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
					j += 2
					break
				}
				j++
			}
			i = j
		default:
			// Charset selection escapes carry one parameter byte.
			if s[i+1] == '(' || s[i+1] == ')' || s[i+1] == '#' || s[i+1] == '%' {
				i += 3
			} else {
				i += 2
			}
		}
	}
	flush()
	if len(spans) == 0 {
		return []ui.Span{{Text: ""}}
	}
	return spans
}

func (s *ansiStyle) applySGR(params string) {
	if params == "" {
		*s = ansiStyle{}
		return
	}
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "" {
			parts[i] = "0"
		}
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}
		switch {
		case n == 0:
			*s = ansiStyle{}
		case n == 1:
			s.bold = true
		case n == 2:
			s.dim = true
		case n == 3:
			s.italic = true
		case n == 4:
			s.underline = true
		case n == 7:
			s.inverse = true
		case n == 22:
			s.bold, s.dim = false, false
		case n == 23:
			s.italic = false
		case n == 24:
			s.underline = false
		case n == 27:
			s.inverse = false
		case n >= 30 && n <= 37:
			s.fg, s.hasFg = ansiBasic[n-30], true
		case n == 38:
			if color, used, ok := parseExtendedColor(parts[i+1:]); used > 0 {
				i += used
				if ok {
					s.fg, s.hasFg = color, true
				}
			}
		case n == 39:
			s.fg, s.hasFg = ui.Color{}, false
		case n >= 40 && n <= 47:
			s.bg, s.hasBg = ansiBasic[n-40], true
		case n == 48:
			if color, used, ok := parseExtendedColor(parts[i+1:]); used > 0 {
				i += used
				if ok {
					s.bg, s.hasBg = color, true
				}
			}
		case n == 49:
			s.bg, s.hasBg = ui.Color{}, false
		case n >= 90 && n <= 97:
			s.fg, s.hasFg = ansiBright[n-90], true
		case n >= 100 && n <= 107:
			s.bg, s.hasBg = ansiBright[n-100], true
		}
	}
}

func parseExtendedColor(rest []string) (ui.Color, int, bool) {
	if len(rest) >= 2 && rest[0] == "5" {
		if n, err := strconv.Atoi(rest[1]); err == nil {
			if color, ok := ansiColor(n); ok {
				return color, 2, true
			}
		}
		return ui.Color{}, 2, false
	}
	if len(rest) >= 4 && rest[0] == "2" {
		r, e1 := strconv.Atoi(rest[1])
		g, e2 := strconv.Atoi(rest[2])
		b, e3 := strconv.Atoi(rest[3])
		if e1 == nil && e2 == nil && e3 == nil {
			return ui.RGB(uint8(r), uint8(g), uint8(b)), 4, true
		}
		return ui.Color{}, 4, false
	}
	return ui.Color{}, 0, false
}

func ansiColor(n int) (ui.Color, bool) {
	switch {
	case n >= 0 && n < len(ansiBasic):
		return ansiBasic[n], true
	case n >= 8 && n < 16:
		return ansiBright[n-8], true
	case n >= 16 && n <= 231:
		n -= 16
		levels := [...]uint8{0, 95, 135, 175, 215, 255}
		r := n / 36
		g := (n % 36) / 6
		b := n % 6
		return ui.RGB(levels[r], levels[g], levels[b]), true
	case n >= 232 && n <= 255:
		v := uint8(8 + 10*(n-232))
		return ui.RGB(v, v, v), true
	}
	return ui.Color{}, false
}

func (s ansiStyle) span(text string) ui.Span {
	fg := termFg
	if s.hasFg {
		fg = s.fg
	}
	if s.dim {
		fg = termMuted
	}
	sp := ui.Span{
		Text:      text,
		Color:     fg,
		Italic:    s.italic,
		Underline: s.underline,
	}
	if s.bold {
		sp.Weight = 600
	}
	if s.inverse {
		sp.Color = termBg
		if s.hasBg {
			sp.Color = s.bg
		}
		sp.Background = fg
	} else if s.hasBg {
		sp.Background = s.bg
	}
	return sp
}

func sameSpanStyle(a, b ui.Span) bool {
	return a.Font == b.Font &&
		a.Size == b.Size &&
		a.Weight == b.Weight &&
		a.Italic == b.Italic &&
		a.Color == b.Color &&
		a.Underline == b.Underline &&
		a.Strikethrough == b.Strikethrough &&
		a.Background == b.Background
}
