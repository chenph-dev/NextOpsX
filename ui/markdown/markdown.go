// Package markdown parses and renders the subset of Markdown used by
// the intelligent ops panel and settings pages.
package markdown

// The Markdown block/inline parser below is adapted from MyGo's
// internal markdown parser (github.com/egoist/mygo, MIT), which renders
// release notes. Keeping it here lets the native AI panel show headings,
// lists, quotes, fenced code, tables, emphasis, code spans and links without
// a web view or an extra dependency.

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"nextopsx/ui/style"
)

// BlockKind is what a Block is.
type BlockKind uint8

const (
	Paragraph BlockKind = iota
	Heading
	List
	Quote
	Code
	Rule
	Table
)

// TableAlign is the alignment of a Markdown table column.
type TableAlign uint8

const (
	AlignLeft TableAlign = iota
	AlignCenter
	AlignRight
)

// Block is a block of release notes.
type Block struct {
	Kind BlockKind
	// Level is the level of a heading, from 1 to 6.
	Level int
	// Inlines are the text of a paragraph or heading.
	Inlines []Inline
	// Text is the text of fenced code.
	Text string
	// Ordered is set for a numbered list, whose Items hold the blocks of
	// each item; Blocks are those of a quote.
	Ordered bool
	Items   [][]Block
	Blocks  []Block
	// Header, Rows and Aligns are those of a Table.
	Header [][]Inline
	Rows   [][][]Inline
	Aligns []TableAlign
}

// InlineKind is what an Inline is.
type InlineKind uint8

const (
	Text InlineKind = iota
	CodeSpan
	Emphasis
	Strong
	Link
)

// Inline is a run of text in a paragraph or heading.
type Inline struct {
	Kind InlineKind
	// Text is the text of a Text or CodeSpan.
	Text string
	// Children are the text of an Emphasis, Strong or Link.
	Children []Inline
	// URL is where a Link goes: an http, https or mailto URL.
	URL string
}

// mdParse parses Markdown text.
func mdParse(src string) []Block {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\t", "    ")
	return parseBlocks(strings.Split(src, "\n"))
}

// parseBlocks parses lines as blocks.
func parseBlocks(lines []string) []Block {
	var blocks []Block
	var para []string
	flush := func() {
		if len(para) > 0 {
			blocks = append(blocks, Block{Kind: Paragraph, Inlines: parseInline(strings.Join(para, "\n"))})
			para = nil
		}
	}
	for i := 0; i < len(lines); {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			flush()
			i++
		case isFence(trimmed):
			flush()
			fence := trimmed[:3]
			var code []string
			for i++; i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), fence); i++ {
				code = append(code, lines[i])
			}
			i++
			blocks = append(blocks, Block{Kind: Code, Text: strings.Join(code, "\n")})
		case heading(trimmed) > 0:
			flush()
			n := heading(trimmed)
			text := strings.TrimRight(strings.TrimSpace(trimmed[n:]), "#")
			blocks = append(blocks, Block{Kind: Heading, Level: n, Inlines: parseInline(strings.TrimSpace(text))})
			i++
		case isRule(trimmed):
			flush()
			blocks = append(blocks, Block{Kind: Rule})
			i++
		case isTableStart(lines, i):
			flush()
			var table Block
			table, i = parseTable(lines, i)
			blocks = append(blocks, table)
		case strings.HasPrefix(trimmed, ">"):
			flush()
			var quote []string
			for ; i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">"); i++ {
				q := strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")
				quote = append(quote, strings.TrimPrefix(q, " "))
			}
			blocks = append(blocks, Block{Kind: Quote, Blocks: parseBlocks(quote)})
		default:
			if _, _, ok := listMarker(line); ok && (len(para) == 0 || startsList(line)) {
				flush()
				var list Block
				list, i = parseList(lines, i)
				blocks = append(blocks, list)
				continue
			}
			para = append(para, trimmed)
			i++
		}
	}
	flush()
	return blocks
}

// parseList parses the list starting at lines[i] and returns it with the
// index of the line after it.
func parseList(lines []string, i int) (Block, int) {
	indent, ordered, _ := listMarker(lines[i])
	list := Block{Kind: List, Ordered: ordered}
	for i < len(lines) {
		ind, ord, ok := listMarker(lines[i])
		if !ok || ind != indent || ord != ordered {
			break
		}
		// The item holds the rest of its first line, the lines indented to
		// its content, lists nested less deep than that, and the lines
		// that continue its paragraph.
		width := markerWidth(lines[i])
		item := []string{lines[i][width:]}
		blank := false
	lines:
		for i++; i < len(lines); i++ {
			line := lines[i]
			if strings.TrimSpace(line) == "" {
				blank = true
				item = append(item, "")
				continue
			}
			ind := leadingSpaces(line)
			_, _, marker := listMarker(line)
			switch {
			case ind >= width:
				item = append(item, line[width:])
			case ind > indent && marker:
				item = append(item, line[ind:])
			case !blank && !marker && !isBlockStart(line):
				item = append(item, strings.TrimSpace(line))
			default:
				break lines
			}
			blank = false
		}
		for len(item) > 0 && item[len(item)-1] == "" {
			item = item[:len(item)-1]
		}
		list.Items = append(list.Items, parseBlocks(item))
	}
	return list, i
}

// isTableStart reports whether lines[i] is a table header followed by a
// delimiter row.
func isTableStart(lines []string, i int) bool {
	if i+1 >= len(lines) || !strings.Contains(lines[i], "|") {
		return false
	}
	header := splitTableRow(lines[i])
	if len(header) == 0 {
		return false
	}
	aligns, ok := tableAlignments(lines[i+1])
	return ok && len(aligns) == len(header)
}

// parseTable parses a GFM table at lines[i].
func parseTable(lines []string, i int) (Block, int) {
	header := splitTableRow(lines[i])
	aligns, _ := tableAlignments(lines[i+1])
	b := Block{Kind: Table, Aligns: aligns}
	for _, cell := range header {
		b.Header = append(b.Header, parseInline(cell))
	}
	i += 2
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.Contains(line, "|") || isBlockStart(trimmed) {
			break
		}
		cells := splitTableRow(line)
		row := make([][]Inline, len(header))
		for j := range row {
			if j < len(cells) {
				row[j] = parseInline(cells[j])
			}
		}
		b.Rows = append(b.Rows, row)
		i++
	}
	return b, i
}

// splitTableRow splits a table row at unescaped pipes outside code spans.
func splitTableRow(line string) []string {
	s := strings.TrimSpace(line)
	if strings.HasPrefix(s, "|") {
		s = s[1:]
	}
	if strings.HasSuffix(s, "|") && !strings.HasSuffix(s, "\\|") {
		s = s[:len(s)-1]
	}
	var cells []string
	var b strings.Builder
	codeDelim := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && s[i+1] == '|':
			b.WriteByte('|')
			i++
		case c == '`':
			n := run(s[i:], '`')
			switch {
			case codeDelim == 0 && hasClosingBackticks(s[i+n:], n):
				codeDelim = n
			case codeDelim == n:
				codeDelim = 0
			}
			b.WriteString(s[i : i+n])
			i += n - 1
		case c == '|' && codeDelim == 0:
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	cells = append(cells, strings.TrimSpace(b.String()))
	return cells
}

// hasClosingBackticks reports whether s holds a run of exactly n backticks,
// closing an inline code span. A lone unmatched backtick stays literal, so
// pipes after it still separate table cells.
func hasClosingBackticks(s string, n int) bool {
	for i := 0; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		m := run(s[i:], '`')
		if m == n {
			return true
		}
		i += m
	}
	return false
}

func tableAlignments(line string) ([]TableAlign, bool) {
	cells := splitTableRow(line)
	if len(cells) == 0 {
		return nil, false
	}
	aligns := make([]TableAlign, len(cells))
	for i, cell := range cells {
		align, ok := tableAlign(cell)
		if !ok {
			return nil, false
		}
		aligns[i] = align
	}
	return aligns, true
}

func tableAlign(cell string) (TableAlign, bool) {
	cell = strings.TrimSpace(cell)
	left := strings.HasPrefix(cell, ":")
	right := strings.HasSuffix(cell, ":")
	core := strings.TrimSpace(strings.Trim(cell, ":"))
	if len(core) < 1 || strings.Trim(core, "-") != "" {
		return AlignLeft, false
	}
	switch {
	case left && right:
		return AlignCenter, true
	case right:
		return AlignRight, true
	default:
		return AlignLeft, true
	}
}

// listMarker reports whether line starts a list item, with the indentation
// of its marker and whether the list is ordered.
func listMarker(line string) (indent int, ordered, ok bool) {
	indent = leadingSpaces(line)
	rest := line[indent:]
	if len(rest) >= 2 && strings.ContainsRune("-*+", rune(rest[0])) && rest[1] == ' ' {
		return indent, false, !isRule(strings.TrimSpace(rest))
	}
	n := 0
	for n < len(rest) && n < 9 && rest[n] >= '0' && rest[n] <= '9' {
		n++
	}
	if n > 0 && len(rest) > n+1 && (rest[n] == '.' || rest[n] == ')') && rest[n+1] == ' ' {
		return indent, true, true
	}
	return 0, false, false
}

// markerWidth returns the column where the content of the list item that
// line starts begins: after its marker and the space after it.
func markerWidth(line string) int {
	w := leadingSpaces(line)
	for w < len(line) && line[w] != ' ' {
		w++
	}
	start := w
	for w < len(line) && line[w] == ' ' && w-start < 4 {
		w++
	}
	return w
}

// startsList reports whether line starts a list item that may interrupt a
// paragraph: a bullet, or an ordered item numbered 1.
func startsList(line string) bool {
	_, ordered, ok := listMarker(line)
	if !ok {
		return false
	}
	return !ordered || strings.HasPrefix(strings.TrimSpace(line), "1")
}

func isBlockStart(line string) bool {
	t := strings.TrimSpace(line)
	return heading(t) > 0 || isFence(t) || isRule(t) || strings.HasPrefix(t, ">")
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

func heading(t string) int {
	n := 0
	for n < len(t) && n < 7 && t[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || (len(t) > n && t[n] != ' ') {
		return 0
	}
	return n
}

func isFence(t string) bool { return strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") }

func isRule(t string) bool {
	if len(t) < 3 || !strings.ContainsRune("-*_", rune(t[0])) {
		return false
	}
	n := 0
	for _, c := range t {
		switch {
		case c == rune(t[0]):
			n++
		case c != ' ':
			return false
		}
	}
	return n >= 3
}

// inlines collects the inlines of a text, merging runs of text.
type inlines []Inline

func (in *inlines) text(s string) {
	if n := len(*in); n > 0 && (*in)[n-1].Kind == Text {
		(*in)[n-1].Text += s
		return
	}
	*in = append(*in, Inline{Kind: Text, Text: s})
}

// link adds children linked to href, or the children alone when the URL is
// not one the window opens.
func (in *inlines) link(href string, children []Inline) {
	if !SafeURL(href) {
		for _, c := range children {
			if c.Kind == Text {
				in.text(c.Text)
			} else {
				*in = append(*in, c)
			}
		}
		return
	}
	*in = append(*in, Inline{Kind: Link, URL: href, Children: children})
}

// parseInline parses the text of a paragraph or heading.
func parseInline(s string) []Inline {
	var in inlines
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && isPunct(s[i+1]):
			in.text(s[i+1 : i+2])
			i += 2
			continue
		case c == '\n':
			in.text(" ")
			i++
			continue
		case c == '`':
			n := run(s[i:], '`')
			if end := strings.Index(s[i+n:], s[i:i+n]); end >= 0 {
				code := strings.TrimSpace(strings.ReplaceAll(s[i+n:i+n+end], "\n", " "))
				in = append(in, Inline{Kind: CodeSpan, Text: code})
				i += n + end + n
				continue
			}
			in.text(s[i : i+n])
			i += n
			continue
		case c == '[' || c == '!' && strings.HasPrefix(s[i:], "!["):
			image := c == '!'
			start := i
			if image {
				start++
			}
			if text, href, n, ok := parseLink(s[start:]); ok {
				var children []Inline
				if image {
					children = []Inline{{Kind: Text, Text: text}}
				} else {
					children = parseInline(text)
				}
				in.link(href, children)
				i = start + n
				continue
			}
		case c == '<':
			if end := strings.IndexByte(s[i:], '>'); end > 0 {
				if u := s[i+1 : i+end]; !strings.ContainsAny(u, " <") && SafeURL(u) {
					in.link(u, []Inline{{Kind: Text, Text: u}})
					i += end + 1
					continue
				}
			}
		case c == '*' || c == '_':
			n := min(run(s[i:], c), 2)
			if c == '_' && i > 0 && isWord(s[i-1]) {
				break
			}
			delim := s[i : i+n]
			rest := s[i+n:]
			if rest != "" && rest[0] != ' ' && rest[0] != '\n' {
				if end := closing(rest, delim); end > 0 {
					kind := Emphasis
					if n == 2 {
						kind = Strong
					}
					in = append(in, Inline{Kind: kind, Children: parseInline(rest[:end])})
					i += n + end + n
					continue
				}
			}
			in.text(delim)
			i += n
			continue
		case c == 'h' && (i == 0 || !isWord(s[i-1])) && (strings.HasPrefix(s[i:], "https://") || strings.HasPrefix(s[i:], "http://")):
			end := i
			for end < len(s) && s[end] > ' ' && s[end] != '<' {
				end++
			}
			u := strings.TrimRight(s[i:end], ".,:;!?\"')")
			if strings.Count(u, "(") > strings.Count(u, ")") && end > i+len(u) && s[i+len(u)] == ')' {
				u += ")"
			}
			in.link(u, []Inline{{Kind: Text, Text: u}})
			i += len(u)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		in.text(s[i : i+size])
		i += size
	}
	return in
}

// parseLink parses "[text](url)" or `[text](url "title")` at the start of
// s, returning the text, the URL and the length of the link.
func parseLink(s string) (text, href string, n int, ok bool) {
	depth := 0
	end := -1
	for i := 0; i < len(s) && end < 0; i++ {
		switch s[i] {
		case '\\':
			i++
		case '[':
			depth++
		case ']':
			if depth--; depth == 0 {
				end = i
			}
		}
	}
	if end < 0 || !strings.HasPrefix(s[end+1:], "(") {
		return "", "", 0, false
	}
	close := strings.IndexByte(s[end+2:], ')')
	if close < 0 {
		return "", "", 0, false
	}
	dest := strings.TrimSpace(s[end+2 : end+2+close])
	if sp := strings.IndexAny(dest, " \n"); sp >= 0 {
		dest = dest[:sp] // drop the title
	}
	dest = strings.TrimSuffix(strings.TrimPrefix(dest, "<"), ">")
	return s[1:end], dest, end + 2 + close + 1, true
}

// SafeURL reports whether u is a URL the window opens: http, https or
// mailto.
func SafeURL(u string) bool {
	l := strings.ToLower(u)
	return strings.HasPrefix(l, "https://") || strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "mailto:")
}

// closing returns the index of the delimiter closing an emphasis in s.
func closing(s, delim string) int {
	for i := 1; i+len(delim) <= len(s); i++ {
		switch {
		case s[i] == '`':
			// Code spans take precedence.
			n := run(s[i:], '`')
			if end := strings.Index(s[i+n:], s[i:i+n]); end >= 0 {
				i += n + end + n - 1
			}
		case s[i] == '\\':
			i++
		case strings.HasPrefix(s[i:], delim) && s[i-1] != ' ' && s[i-1] != '\n':
			if len(delim) == 1 && i+1 < len(s) && s[i+1] == delim[0] {
				i++ // part of a strong delimiter
				continue
			}
			if delim[0] == '_' && i+len(delim) < len(s) && isWord(s[i+len(delim)]) {
				continue
			}
			return i
		}
	}
	return -1
}

func run(s string, c byte) int {
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	return n
}

func isPunct(c byte) bool { return strings.IndexByte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", c) >= 0 }

func isWord(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80
}

// mdStyle is the inherited inline style while descending into emphasis.
type mdStyle struct {
	weight    int
	italic    bool
	color     ui.Color
	underline bool
}

// View renders Markdown text with native widgets.
func View(c *ui.Context, src string) {
	t := c.Theme()
	blocks := mdParse(src)
	if len(blocks) == 0 {
		ui.Text(c, src).FontSize(style.FontSize(13)).Wrap()
		return
	}
	ui.Column(c).Gap(6).Children(func() {
		mdRenderBlocks(c, blocks, t)
	})
}

func mdRenderBlocks(c *ui.Context, blocks []Block, t *ui.Theme) {
	for _, b := range blocks {
		switch b.Kind {
		case Paragraph:
			mdParagraph(c, b.Inlines, 12, false, t)
		case Heading:
			size := float32(13)
			switch b.Level {
			case 1:
				size = 18
			case 2:
				size = 16
			case 3:
				size = 14
			}
			mdParagraph(c, b.Inlines, size, true, t)
		case Code:
			ui.Scroll(c).MaxHeight(320).Children(func() {
				ui.Column(c).Background(t.Background).Border(1, t.Border).Radius(4).
					Padding(6, 8).Children(func() {
					ui.Text(c, b.Text).Font("monospace").FontSize(style.FontSize(12)).Wrap()
				})
			})
		case Rule:
			ui.Divider(c)
		case Quote:
			ui.Row(c).Gap(8).AlignItems(ui.Stretch).Children(func() {
				ui.Box(c).Width(2).Background(t.Border)
				ui.Column(c).Grow(1).Children(func() {
					mdRenderBlocks(c, b.Blocks, t)
				})
			})
		case List:
			for i, item := range b.Items {
				marker := "•"
				if b.Ordered {
					marker = fmt.Sprintf("%d.", i+1)
				}
				ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
					ui.Text(c, marker).Width(20).FontSize(style.FontSize(13)).TextColor(t.TextMuted)
					ui.Column(c).Grow(1).Children(func() {
						mdRenderBlocks(c, item, t)
					})
				})
			}
		case Table:
			mdTable(c, b, t)
		}
	}
}

func mdTable(c *ui.Context, b Block, t *ui.Theme) {
	ui.Column(c).Border(1, t.Border).Radius(4).Children(func() {
		mdTableRow(c, b.Header, b.Aligns, true, t)
		for _, row := range b.Rows {
			ui.Divider(c)
			mdTableRow(c, row, b.Aligns, false, t)
		}
	})
}

func mdTableRow(c *ui.Context, cells [][]Inline, aligns []TableAlign, header bool, t *ui.Theme) {
	ui.Row(c).AlignItems(ui.Stretch).Children(func() {
		for i, cell := range cells {
			align := AlignLeft
			if i < len(aligns) {
				align = aligns[i]
			}
			col := ui.Column(c).Grow(1).MinWidth(60).Padding(5, 8)
			if header {
				col = col.Background(t.SurfaceHover)
			}
			col.Children(func() {
				mdTableCell(c, cell, align, header, t)
			})
		}
	})
}

func mdTableCell(c *ui.Context, inlines []Inline, align TableAlign, header bool, t *ui.Theme) {
	spans := mdSpans(inlines, mdStyle{}, t)
	if len(spans) == 0 {
		return
	}
	text := ui.RichText(c, spans...).FontSize(style.FontSize(12)).LineHeight(1.4).
		TextAlign(mdTableAlign(align))
	if header {
		text = text.Bold()
	}
}

func mdTableAlign(a TableAlign) ui.Align {
	switch a {
	case AlignCenter:
		return ui.Center
	case AlignRight:
		return ui.End
	default:
		return ui.Start
	}
}

func mdParagraph(c *ui.Context, inlines []Inline, size float32, bold bool, t *ui.Theme) {
	spans := mdSpans(inlines, mdStyle{}, t)
	if len(spans) == 0 {
		return
	}
	text := ui.RichText(c, spans...).FontSize(style.FontSize(size)).LineHeight(1.5)
	if bold {
		text = text.Bold()
	}
}

func mdSpans(inlines []Inline, st mdStyle, t *ui.Theme) []ui.Span {
	var out []ui.Span
	for _, in := range inlines {
		switch in.Kind {
		case Text:
			if in.Text == "" {
				continue
			}
			out = append(out, ui.Span{
				Text: in.Text, Weight: st.weight, Italic: st.italic,
				Color: st.color, Underline: st.underline,
			})
		case CodeSpan:
			out = append(out, ui.Span{
				Text: " " + in.Text + " ", Font: "monospace",
				Background: t.SurfaceHover, Weight: st.weight,
			})
		case Emphasis:
			next := st
			next.italic = true
			out = append(out, mdSpans(in.Children, next, t)...)
		case Strong:
			next := st
			next.weight = 700
			out = append(out, mdSpans(in.Children, next, t)...)
		case Link:
			next := st
			next.color = t.Accent
			next.underline = true
			out = append(out, mdSpans(in.Children, next, t)...)
		}
	}
	return out
}
