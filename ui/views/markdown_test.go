package views

import (
	"testing"

	"github.com/egoist/mygo/ui"

	convstore "nextopsx/store/conversations"
)

func TestMarkdownParseBlocks(t *testing.T) {
	src := "# 标题\n\n正文 **粗体** 与 `代码`\n\n- 一\n- 二\n\n```go\nfmt.Println(\"x\")\n```\n\n> 引用\n"
	blocks := mdParse(src)
	if len(blocks) != 5 {
		t.Fatalf("blocks = %d, want 5: %#v", len(blocks), blocks)
	}
	if blocks[0].Kind != Heading || blocks[0].Level != 1 {
		t.Fatalf("heading = %#v", blocks[0])
	}
	if blocks[1].Kind != Paragraph {
		t.Fatalf("paragraph = %#v", blocks[1])
	}
	if blocks[2].Kind != List || len(blocks[2].Items) != 2 {
		t.Fatalf("list = %#v", blocks[2])
	}
	if blocks[3].Kind != Code || blocks[3].Text != "fmt.Println(\"x\")" {
		t.Fatalf("code = %#v", blocks[3])
	}
	if blocks[4].Kind != Quote || len(blocks[4].Blocks) != 1 {
		t.Fatalf("quote = %#v", blocks[4])
	}
}

func TestMarkdownInlineStyles(t *testing.T) {
	in := parseInline("**粗体** 与 `代码` 和 [链接](https://example.com)")
	var kinds []InlineKind
	for _, v := range in {
		kinds = append(kinds, v.Kind)
	}
	want := []InlineKind{Strong, Text, CodeSpan, Text, Link}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds = %v, want %v", kinds, want)
		}
	}
}

func TestMarkdownTableParse(t *testing.T) {
	src := "| 指标 | 数值 | 状态 |\n|:---|---:|:---:|\n| CPU | 12% | **正常** |\n| 内存 | 16 GB | 正常 |"
	blocks := mdParse(src)
	if len(blocks) != 1 || blocks[0].Kind != Table {
		t.Fatalf("blocks = %#v", blocks)
	}
	b := blocks[0]
	if len(b.Header) != 3 || len(b.Rows) != 2 {
		t.Fatalf("header=%d rows=%d", len(b.Header), len(b.Rows))
	}
	wantAlign := []TableAlign{AlignLeft, AlignRight, AlignCenter}
	for i, want := range wantAlign {
		if b.Aligns[i] != want {
			t.Fatalf("align[%d] = %v, want %v", i, b.Aligns[i], want)
		}
	}
	if len(b.Header[0]) == 0 || b.Header[0][0].Text != "指标" {
		t.Fatalf("header cell = %#v", b.Header[0])
	}
	if len(b.Rows[0][2]) == 0 || b.Rows[0][2][0].Kind != Strong {
		t.Fatalf("bold cell = %#v", b.Rows[0][2])
	}
}

func TestMarkdownViewTableRenders(t *testing.T) {
	a := aiTestApp()
	a.AI.Configured = true
	a.AI.Messages = []convstore.Message{{
		ID: "m1", Role: "assistant",
		Content: "| 指标 | 数值 |\n|---|---|\n| CPU | 12% |",
	}}
	tt := ui.NewTester(Shell(a), 1000, 900)
	for _, want := range []string{"指标", "数值", "CPU", "12%"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestSplitTableRowCodeSpanPipes(t *testing.T) {
	cells := splitTableRow("| ``a|b`` | c |")
	if len(cells) != 2 || cells[0] != "``a|b``" || cells[1] != "c" {
		t.Fatalf("double-backtick cells = %#v", cells)
	}
	cells = splitTableRow("| a`b | c |")
	if len(cells) != 2 || cells[0] != "a`b" || cells[1] != "c" {
		t.Fatalf("stray-backtick cells = %#v", cells)
	}
}

func TestMarkdownTableOneDashDelimiter(t *testing.T) {
	blocks := mdParse("| a | b |\n|-|-|\n| 1 | 2 |")
	if len(blocks) != 1 || blocks[0].Kind != Table {
		t.Fatalf("blocks = %#v", blocks)
	}
	if len(blocks[0].Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(blocks[0].Rows))
	}
}
