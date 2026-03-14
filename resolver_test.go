package localechain

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Resolve — per-key fallback across chain
// ---------------------------------------------------------------------------

func TestResolve_PerKeyFallbackAcrossChain(t *testing.T) {
	defer Reset()
	Configure()

	messages := map[string]map[string]string{
		"pt": {
			"hello":   "Ola (pt)",
			"goodbye": "Adeus (pt)",
			"thanks":  "Obrigado (pt)",
		},
		"pt-PT": {
			"hello":   "Ola (pt-PT)",
			"goodbye": "Adeus (pt-PT)",
		},
		"pt-BR": {
			"hello": "Ola (pt-BR)",
		},
	}

	result := Resolve("pt-BR", messages)

	// pt-BR has "hello" -> use pt-BR value
	assertKV(t, result, "hello", "Ola (pt-BR)")
	// pt-BR missing "goodbye" -> fallback to pt-PT
	assertKV(t, result, "goodbye", "Adeus (pt-PT)")
	// pt-BR missing "thanks", pt-PT missing "thanks" -> fallback to pt
	assertKV(t, result, "thanks", "Obrigado (pt)")
}

func TestResolve_GoI18nBugScenario(t *testing.T) {
	// Specific go-i18n bug scenario: locale pt-BR has key "hello" but
	// not "goodbye" -> "goodbye" should come from pt-PT or pt in the chain.
	defer Reset()
	Configure()

	messages := map[string]map[string]string{
		"pt": {
			"goodbye": "Adeus (pt)",
		},
		"pt-PT": {
			"goodbye": "Adeus (pt-PT)",
		},
		"pt-BR": {
			"hello": "Ola (pt-BR)",
		},
	}

	result := Resolve("pt-BR", messages)

	assertKV(t, result, "hello", "Ola (pt-BR)")
	// "goodbye" should come from pt-PT (first in chain for pt-BR)
	assertKV(t, result, "goodbye", "Adeus (pt-PT)")
}

func TestResolve_MissingKeysWalkFullChain(t *testing.T) {
	defer Reset()
	Configure()

	// en-NZ chain: ["en-AU", "en-GB", "en"]
	messages := map[string]map[string]string{
		"en": {
			"color":  "color",
			"centre": "center",
			"lift":   "elevator",
		},
		"en-GB": {
			"color":  "colour",
			"centre": "centre",
		},
		"en-AU": {
			"color": "colour (AU)",
		},
		"en-NZ": {
			// No keys — everything falls back
		},
	}

	result := Resolve("en-NZ", messages)

	// en-NZ empty -> fall back to en-AU -> "colour (AU)"
	assertKV(t, result, "color", "colour (AU)")
	// en-NZ empty, en-AU missing "centre" -> fall back to en-GB -> "centre"
	assertKV(t, result, "centre", "centre")
	// en-NZ, en-AU, en-GB all missing "lift" -> fall back to en -> "elevator"
	assertKV(t, result, "lift", "elevator")
}

func TestResolve_UnknownLocaleReturnsItsOwnMessages(t *testing.T) {
	defer Reset()
	Configure()

	messages := map[string]map[string]string{
		"en": {
			"hello": "Hello",
		},
		"xx-YY": {
			"hello": "Xxello",
		},
	}

	result := Resolve("xx-YY", messages)

	// Unknown locale has no chain, so only its own messages are returned
	assertKV(t, result, "hello", "Xxello")
	if len(result) != 1 {
		t.Errorf("expected 1 key for unknown locale, got %d: %v", len(result), result)
	}
}

func TestResolve_UnknownLocaleWithNoMessages(t *testing.T) {
	defer Reset()
	Configure()

	messages := map[string]map[string]string{
		"en": {
			"hello": "Hello",
		},
	}

	result := Resolve("xx-YY", messages)

	if len(result) != 0 {
		t.Errorf("expected empty map for unknown locale with no messages, got %v", result)
	}
}

func TestResolve_WithoutConfigure(t *testing.T) {
	// When resolver is not configured, Resolve should still return
	// the messages for the requested locale without fallback.
	Reset()

	messages := map[string]map[string]string{
		"pt": {
			"hello": "Ola (pt)",
		},
		"pt-BR": {
			"thanks": "Obrigado (pt-BR)",
		},
	}

	result := Resolve("pt-BR", messages)

	// Only pt-BR messages, no fallback to pt
	assertKV(t, result, "thanks", "Obrigado (pt-BR)")
	if _, ok := result["hello"]; ok {
		t.Error("without Configure, Resolve should not fall back to pt")
	}
}

func TestResolve_HigherPriorityOverridesLower(t *testing.T) {
	defer Reset()
	Configure()

	// pt-BR chain: ["pt-PT", "pt"]
	// Load order (low to high): pt, pt-PT, pt-BR
	messages := map[string]map[string]string{
		"pt": {
			"greeting": "Ola (pt)",
		},
		"pt-PT": {
			"greeting": "Ola (pt-PT)",
		},
		"pt-BR": {
			"greeting": "Ola (pt-BR)",
		},
	}

	result := Resolve("pt-BR", messages)

	// pt-BR (highest priority) should win
	assertKV(t, result, "greeting", "Ola (pt-BR)")
}

func TestResolve_EmptyMessages(t *testing.T) {
	defer Reset()
	Configure()

	messages := map[string]map[string]string{}
	result := Resolve("pt-BR", messages)

	if len(result) != 0 {
		t.Errorf("expected empty map for empty messages, got %v", result)
	}
}

func TestResolve_ChineseChainMultiLevel(t *testing.T) {
	defer Reset()
	Configure()

	// zh-Hant-MO chain: ["zh-Hant-HK", "zh-Hant-TW", "zh-Hant"]
	messages := map[string]map[string]string{
		"zh-Hant": {
			"save":   "儲存 (zh-Hant)",
			"cancel": "取消 (zh-Hant)",
			"delete": "刪除 (zh-Hant)",
		},
		"zh-Hant-TW": {
			"save":   "儲存 (zh-Hant-TW)",
			"cancel": "取消 (zh-Hant-TW)",
		},
		"zh-Hant-HK": {
			"save": "儲存 (zh-Hant-HK)",
		},
		"zh-Hant-MO": {
			// Empty — everything falls back
		},
	}

	result := Resolve("zh-Hant-MO", messages)

	assertKV(t, result, "save", "儲存 (zh-Hant-HK)")
	assertKV(t, result, "cancel", "取消 (zh-Hant-TW)")
	assertKV(t, result, "delete", "刪除 (zh-Hant)")
}

// ---------------------------------------------------------------------------
// ResolveWithLoader
// ---------------------------------------------------------------------------

func TestResolveWithLoader_PerKeyFallback(t *testing.T) {
	defer Reset()
	Configure()

	data := map[string]map[string]string{
		"pt": {
			"goodbye": "Adeus (pt)",
			"thanks":  "Obrigado (pt)",
		},
		"pt-PT": {
			"goodbye": "Adeus (pt-PT)",
		},
		"pt-BR": {
			"hello": "Ola (pt-BR)",
		},
	}

	loader := func(locale string) (map[string]string, error) {
		msgs, ok := data[locale]
		if !ok {
			return nil, errors.New("locale not found")
		}
		return msgs, nil
	}

	result, err := ResolveWithLoader("pt-BR", loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertKV(t, result, "hello", "Ola (pt-BR)")
	assertKV(t, result, "goodbye", "Adeus (pt-PT)")
	assertKV(t, result, "thanks", "Obrigado (pt)")
}

func TestResolveWithLoader_SkipsFailedLocales(t *testing.T) {
	defer Reset()
	Configure()

	data := map[string]map[string]string{
		"pt": {
			"hello":   "Ola (pt)",
			"goodbye": "Adeus (pt)",
		},
		"pt-BR": {
			"hello": "Ola (pt-BR)",
		},
	}

	loader := func(locale string) (map[string]string, error) {
		if locale == "pt-PT" {
			return nil, errors.New("file not found")
		}
		msgs, ok := data[locale]
		if !ok {
			return nil, errors.New("locale not found")
		}
		return msgs, nil
	}

	result, err := ResolveWithLoader("pt-BR", loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// pt-PT failed, so "goodbye" falls through to pt
	assertKV(t, result, "hello", "Ola (pt-BR)")
	assertKV(t, result, "goodbye", "Adeus (pt)")
}

func TestResolveWithLoader_AllLocalesFail(t *testing.T) {
	defer Reset()
	Configure()

	loader := func(locale string) (map[string]string, error) {
		return nil, errors.New("all fail")
	}

	result, err := ResolveWithLoader("pt-BR", loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty map when all locales fail, got %v", result)
	}
}

func TestResolveWithLoader_WithoutConfigure(t *testing.T) {
	Reset()

	data := map[string]map[string]string{
		"pt-BR": {
			"hello": "Ola (pt-BR)",
		},
	}

	loader := func(locale string) (map[string]string, error) {
		msgs, ok := data[locale]
		if !ok {
			return nil, errors.New("not found")
		}
		return msgs, nil
	}

	result, err := ResolveWithLoader("pt-BR", loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertKV(t, result, "hello", "Ola (pt-BR)")
	if len(result) != 1 {
		t.Errorf("without Configure, expected only direct locale messages; got %v", result)
	}
}

// ---------------------------------------------------------------------------
// ChainFor
// ---------------------------------------------------------------------------

func TestChainFor_ReturnsConfiguredChain(t *testing.T) {
	defer Reset()
	Configure()

	tests := []struct {
		locale string
		want   []string
	}{
		{"pt-BR", []string{"pt-PT", "pt"}},
		{"nn", []string{"nb", "no"}},
		{"zh-Hant-MO", []string{"zh-Hant-HK", "zh-Hant-TW", "zh-Hant"}},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			got := ChainFor(tt.locale)
			assertSliceEq(t, got, tt.want)
		})
	}
}

func TestChainFor_ReturnsNilWhenNotConfigured(t *testing.T) {
	Reset()

	got := ChainFor("pt-BR")
	if got != nil {
		t.Errorf("ChainFor without Configure should return nil, got %v", got)
	}
}

func TestChainFor_ReturnsNilForUnknownLocale(t *testing.T) {
	defer Reset()
	Configure()

	got := ChainFor("xx-YY")
	if got != nil {
		t.Errorf("ChainFor(\"xx-YY\") should return nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// buildLoadOrder (tested indirectly via Resolve behavior)
// ---------------------------------------------------------------------------

func TestResolve_LoadOrderDeduplicates(t *testing.T) {
	defer Reset()

	// Create a custom chain with duplicate entries.
	// Chain ["bb", "cc", "bb"] reversed is ["bb", "cc", "bb"].
	// Deduplication keeps first occurrence: ["bb", "cc"].
	// Then "aa" appended: ["bb", "cc", "aa"] (low to high priority).
	custom := map[string][]string{
		"aa": {"bb", "cc", "bb"}, // "bb" duplicated
	}
	ConfigureCustom(custom, false)

	messages := map[string]map[string]string{
		"bb": {"key": "from-bb", "only-bb": "bb-value"},
		"cc": {"key": "from-cc"},
		"aa": {"key": "from-aa"},
	}

	result := Resolve("aa", messages)

	// "aa" is last loaded (highest priority), should win
	assertKV(t, result, "key", "from-aa")
	// "only-bb" only exists in bb, should still be present
	assertKV(t, result, "only-bb", "bb-value")
}

func TestResolve_NorwegianChain(t *testing.T) {
	defer Reset()
	Configure()

	// nn chain: ["nb", "no"]
	messages := map[string]map[string]string{
		"no": {
			"morning": "God morgen (no)",
			"evening": "God kveld (no)",
		},
		"nb": {
			"morning": "God morgen (nb)",
		},
		"nn": {
			// Empty
		},
	}

	result := Resolve("nn", messages)

	// nb overrides no for "morning"
	assertKV(t, result, "morning", "God morgen (nb)")
	// nb missing "evening" -> falls through to no
	assertKV(t, result, "evening", "God kveld (no)")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertKV(t *testing.T, m map[string]string, key, want string) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Fatalf("key %q not found in result map (keys: %v)", key, mapKeys(m))
	}
	if got != want {
		t.Errorf("result[%q] = %q, want %q", key, got, want)
	}
}

func assertSliceEq(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
