package i18n

import (
	"regexp"
	"strings"
	"testing"
	"unicode"
)

func TestDefaultLanguage(t *testing.T) {
	SetLanguage(ZhCN)
	if Current() != ZhCN {
		t.Fatalf("default language = %q, want %q", Current(), ZhCN)
	}
	if got := T("shell.assets"); got != "资产" {
		t.Fatalf("T(shell.assets) = %q", got)
	}
}

func TestSetLanguageAndFallback(t *testing.T) {
	t.Cleanup(func() { SetLanguage(ZhCN) })
	SetLanguage(EnUS)
	if Current() != EnUS {
		t.Fatalf("language = %q", Current())
	}
	if got := T("shell.assets"); got != "Assets" {
		t.Fatalf("en T(shell.assets) = %q", got)
	}

	// A key registered only in zh-CN falls back to Chinese.
	zhOnly := "test.zh.only"
	add(ZhCN, map[string]string{zhOnly: "仅中文"})
	if got := T(zhOnly); got != "仅中文" {
		t.Fatalf("fallback = %q", got)
	}

	// Unknown keys surface as the key itself.
	if got := T("test.missing.key"); got != "test.missing.key" {
		t.Fatalf("missing key = %q", got)
	}
}

func TestTf(t *testing.T) {
	t.Cleanup(func() { SetLanguage(ZhCN) })
	key := "test.format"
	add(ZhCN, map[string]string{key: "共 %d 项：%s"})
	add(EnUS, map[string]string{key: "%d items: %s"})
	if got := Tf(key, 3, "x"); got != "共 3 项：x" {
		t.Fatalf("zh Tf = %q", got)
	}
	SetLanguage(EnUS)
	if got := Tf(key, 3, "x"); got != "3 items: x" {
		t.Fatalf("en Tf = %q", got)
	}
}

func TestLangIndex(t *testing.T) {
	cases := []struct {
		lang  Lang
		index int
	}{
		{ZhCN, 0},
		{EnUS, 1},
		{Lang("fr-FR"), 0},
	}
	for _, c := range cases {
		if got := LangIndex(c.lang); got != c.index {
			t.Errorf("LangIndex(%q) = %d, want %d", c.lang, got, c.index)
		}
	}
	if got := LangAt(1); got != EnUS {
		t.Errorf("LangAt(1) = %q", got)
	}
	if got := LangAt(99); got != Default {
		t.Errorf("LangAt(99) = %q", got)
	}
}

// TestCatalogParity catches a key added to one language but not the other.
// Keys with a "test." prefix are created by tests themselves.
func TestCatalogParity(t *testing.T) {
	for key := range catalogs[EnUS] {
		if len(key) >= 5 && key[:5] == "test." {
			continue
		}
		if _, ok := catalogs[ZhCN][key]; !ok {
			t.Errorf("en-US key %q has no zh-CN translation", key)
		}
	}
	for key := range catalogs[ZhCN] {
		if len(key) >= 5 && key[:5] == "test." {
			continue
		}
		if _, ok := catalogs[EnUS][key]; !ok {
			t.Errorf("zh-CN key %q has no en-US translation", key)
		}
	}
}

// TestEnglishCatalogHasNoChinese catches entries copied from zh-CN and left
// untranslated in en-US.
func TestEnglishCatalogHasNoChinese(t *testing.T) {
	for key, value := range catalogs[EnUS] {
		if strings.ContainsFunc(value, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Errorf("en-US %q still contains Chinese: %q", key, value)
		}
	}
}

var formatVerb = regexp.MustCompile(`%[+#0\- .0-9]*[a-zA-Z%]`)

// TestCatalogFormatVerbs makes sure a translation kept every format verb of
// its zh-CN original, so Tf cannot produce "%!d(MISSING)" or a wrong verb.
func TestCatalogFormatVerbs(t *testing.T) {
	for key, zh := range catalogs[ZhCN] {
		en, ok := catalogs[EnUS][key]
		if !ok {
			continue // parity test reports this
		}
		zhVerbs := formatVerb.FindAllString(zh, -1)
		enVerbs := formatVerb.FindAllString(en, -1)
		if len(zhVerbs) != len(enVerbs) {
			t.Errorf("%s: zh verbs %v, en verbs %v", key, zhVerbs, enVerbs)
			continue
		}
		for i := range zhVerbs {
			if zhVerbs[i] != enVerbs[i] {
				t.Errorf("%s: verb %d is %q in zh, %q in en", key, i, zhVerbs[i], enVerbs[i])
			}
		}
	}
}
