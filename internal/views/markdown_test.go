package views

import "testing"

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
