// Package i18n holds the app's UI string catalogs. MyGo v0.4.8 has no i18n
// package, so this is intentionally small: a language value, one catalog per
// language, and lookup helpers with a zh-CN fallback.
//
// Catalogs are registered by files named catalog_*.go so each feature area
// can own its strings. Translation keys are dotted, lower-case, and grouped
// by area, for example "settings.general.language" or "ai.command.approve".
package i18n

import (
	"fmt"
	"maps"
	"sync/atomic"
)

// Lang is a supported UI language.
type Lang string

const (
	// ZhCN is Simplified Chinese, the default and fallback language.
	ZhCN Lang = "zh-CN"
	// EnUS is English (United States).
	EnUS Lang = "en-US"
)

// Default is the language used before preferences are loaded and whenever a
// stored value is not recognized.
const Default = ZhCN

// Languages lists the selectable languages in display order.
var Languages = []Lang{ZhCN, EnUS}

// LangAt maps a segmented-control index to a language. Out-of-range indexes
// fall back to the default.
func LangAt(index int) Lang {
	if index >= 0 && index < len(Languages) {
		return Languages[index]
	}
	return Default
}

// LangIndex returns the index of lang in Languages, or the default's index.
func LangIndex(lang Lang) int {
	lang = Normalize(lang)
	for i, l := range Languages {
		if l == lang {
			return i
		}
	}
	return 0
}

// Normalize maps unknown values to the default language.
func Normalize(lang Lang) Lang {
	switch lang {
	case ZhCN, EnUS:
		return lang
	default:
		return Default
	}
}

var current atomic.Value // Lang

func init() { current.Store(Default) }

// Current returns the active language.
func Current() Lang {
	if v := current.Load(); v != nil {
		return v.(Lang)
	}
	return Default
}

// SetLanguage activates lang for subsequently rendered strings. Views rebuild
// every frame, so the next redraw picks up the change.
func SetLanguage(lang Lang) { current.Store(Normalize(lang)) }

// catalogs are filled by catalog_*.go files' init functions. They are only
// written during init, so reads afterwards need no lock.
var catalogs = map[Lang]map[string]string{
	ZhCN: {},
	EnUS: {},
}

func add(lang Lang, entries map[string]string) {
	maps.Copy(catalogs[lang], entries)
}

// T returns the string for key in the current language. A missing translation
// falls back to zh-CN, then to the key itself so gaps are visible in tests and
// in the UI rather than rendering as an empty label.
func T(key string) string {
	lang := Current()
	if lang != ZhCN {
		if value, ok := catalogs[lang][key]; ok {
			return value
		}
	}
	if value, ok := catalogs[ZhCN][key]; ok {
		return value
	}
	return key
}

// Tf is T followed by fmt.Sprintf with args.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

// Has reports whether key has a translation in lang. It is used by tests to
// catch keys that were referenced but never registered.
func Has(lang Lang, key string) bool {
	_, ok := catalogs[lang][key]
	return ok
}
