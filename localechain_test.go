package localechain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func TestConfigure_SetsUpDefaultChains(t *testing.T) {
	defer Reset()
	Configure()

	tests := []struct {
		locale string
		want   []string
	}{
		{"pt-BR", []string{"pt-PT", "pt"}},
		{"es-MX", []string{"es-419", "es"}},
		{"fr-CA", []string{"fr"}},
		{"de-AT", []string{"de"}},
		{"en-AU", []string{"en-GB", "en"}},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			got := ChainFor(tt.locale)
			assertSliceEqual(t, tt.locale, got, tt.want)
		})
	}
}

func TestConfigure_ChainForReturnsNilForUnknownLocale(t *testing.T) {
	defer Reset()
	Configure()

	got := ChainFor("xx-YY")
	if got != nil {
		t.Errorf("ChainFor(\"xx-YY\") = %v, want nil", got)
	}
}

func TestConfigure_ChainForReturnsCopy(t *testing.T) {
	defer Reset()
	Configure()

	chain1 := ChainFor("pt-BR")
	chain1[0] = "MUTATED"

	chain2 := ChainFor("pt-BR")
	if chain2[0] == "MUTATED" {
		t.Error("ChainFor should return a copy; mutation leaked to internal state")
	}
}

// ---------------------------------------------------------------------------
// ConfigureWithOverrides
// ---------------------------------------------------------------------------

func TestConfigureWithOverrides_OverridesApplied(t *testing.T) {
	defer Reset()

	overrides := map[string][]string{
		"pt-BR":  {"pt"},            // Override existing
		"custom": {"base1", "base"}, // Add new
	}
	ConfigureWithOverrides(overrides)

	// Overridden chain
	got := ChainFor("pt-BR")
	assertSliceEqual(t, "pt-BR", got, []string{"pt"})

	// New chain
	got = ChainFor("custom")
	assertSliceEqual(t, "custom", got, []string{"base1", "base"})

	// Default preserved
	got = ChainFor("fr-CA")
	assertSliceEqual(t, "fr-CA", got, []string{"fr"})
}

// ---------------------------------------------------------------------------
// ConfigureCustom
// ---------------------------------------------------------------------------

func TestConfigureCustom_WithMergeDefaults(t *testing.T) {
	defer Reset()

	custom := map[string][]string{
		"xx-YY": {"xx"},
	}
	ConfigureCustom(custom, true)

	// Custom locale present
	got := ChainFor("xx-YY")
	assertSliceEqual(t, "xx-YY", got, []string{"xx"})

	// Default chains also present
	got = ChainFor("pt-BR")
	assertSliceEqual(t, "pt-BR", got, []string{"pt-PT", "pt"})
}

func TestConfigureCustom_WithoutMergeDefaults(t *testing.T) {
	defer Reset()

	custom := map[string][]string{
		"xx-YY": {"xx"},
	}
	ConfigureCustom(custom, false)

	// Custom locale present
	got := ChainFor("xx-YY")
	assertSliceEqual(t, "xx-YY", got, []string{"xx"})

	// Default chains NOT present
	got = ChainFor("pt-BR")
	if got != nil {
		t.Errorf("ChainFor(\"pt-BR\") = %v, want nil (defaults not merged)", got)
	}
}

func TestConfigureCustom_DoesNotMutateInput(t *testing.T) {
	defer Reset()

	custom := map[string][]string{
		"xx-YY": {"xx"},
	}
	ConfigureCustom(custom, false)

	// Mutate the internal state via ChainFor result
	chain := ChainFor("xx-YY")
	chain[0] = "MUTATED"

	// Original input map should be unchanged
	if custom["xx-YY"][0] != "xx" {
		t.Error("ConfigureCustom mutated the input map")
	}

	// Internal state should be unchanged
	chain2 := ChainFor("xx-YY")
	if chain2[0] == "MUTATED" {
		t.Error("mutation of ChainFor result leaked to internal state")
	}
}

// ---------------------------------------------------------------------------
// Reset
// ---------------------------------------------------------------------------

func TestReset_ClearsResolver(t *testing.T) {
	Configure()
	Reset()

	// After Reset, ChainFor should return nil
	got := ChainFor("pt-BR")
	if got != nil {
		t.Errorf("after Reset(), ChainFor(\"pt-BR\") = %v, want nil", got)
	}
}

func TestReset_ResolveStillWorksWithoutFallback(t *testing.T) {
	Configure()
	Reset()

	messages := map[string]map[string]string{
		"pt-BR": {"hello": "Ola"},
	}

	// Resolve should still return messages for the requested locale,
	// just without fallback chains.
	result := Resolve("pt-BR", messages)
	if result["hello"] != "Ola" {
		t.Errorf("after Reset(), Resolve should still return direct locale messages; got %v", result)
	}
}

func TestReset_ResolveReturnsEmptyForMissingLocale(t *testing.T) {
	Configure()
	Reset()

	messages := map[string]map[string]string{
		"en": {"hello": "Hello"},
	}

	result := Resolve("pt-BR", messages)
	if len(result) != 0 {
		t.Errorf("after Reset(), Resolve(\"pt-BR\") with no pt-BR messages should return empty map; got %v", result)
	}
}

func TestReset_CanReconfigure(t *testing.T) {
	defer Reset()

	Configure()
	Reset()
	Configure()

	got := ChainFor("pt-BR")
	assertSliceEqual(t, "pt-BR", got, []string{"pt-PT", "pt"})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertSliceEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v (len %d), want %v (len %d)",
			label, got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", label, i, got[i], want[i])
		}
	}
}
