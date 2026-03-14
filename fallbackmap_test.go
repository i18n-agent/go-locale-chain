package localechain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Chain data correctness
// ---------------------------------------------------------------------------

func TestDefaultFallbacks_AllLanguageGroupsPresent(t *testing.T) {
	// Each group maps to at least one locale key in DefaultFallbacks.
	groups := map[string][]string{
		"Chinese":    {"zh-Hant-HK", "zh-Hant-MO", "zh-Hant-TW", "zh-Hans-SG", "zh-Hans-MY"},
		"Portuguese": {"pt-BR", "pt-PT", "pt-AO", "pt-MZ"},
		"Spanish":    {"es-419", "es-MX", "es-AR", "es-CO", "es-CL", "es-PE", "es-VE", "es-EC", "es-GT", "es-CU", "es-BO", "es-DO", "es-HN", "es-PY", "es-SV", "es-NI", "es-CR", "es-PA", "es-UY", "es-PR"},
		"French":     {"fr-CA", "fr-BE", "fr-CH", "fr-LU", "fr-MC", "fr-SN", "fr-CI", "fr-ML", "fr-CM", "fr-MG", "fr-CD"},
		"German":     {"de-AT", "de-CH", "de-LU", "de-LI"},
		"Italian":    {"it-CH"},
		"Dutch":      {"nl-BE"},
		"English":    {"en-GB", "en-AU", "en-NZ", "en-IN", "en-CA", "en-ZA", "en-IE", "en-SG"},
		"Arabic":     {"ar-SA", "ar-EG", "ar-AE", "ar-MA", "ar-DZ", "ar-IQ", "ar-KW", "ar-QA", "ar-BH", "ar-OM", "ar-JO", "ar-LB", "ar-TN", "ar-LY", "ar-SD", "ar-YE"},
		"Norwegian":  {"nb", "nn"},
		"Malay":      {"ms-MY", "ms-SG", "ms-BN"},
	}

	for group, locales := range groups {
		t.Run(group, func(t *testing.T) {
			for _, loc := range locales {
				if _, ok := DefaultFallbacks[loc]; !ok {
					t.Errorf("expected locale %q in group %s to be present in DefaultFallbacks", loc, group)
				}
			}
		})
	}
}

func TestDefaultFallbacks_NoChainsEmpty(t *testing.T) {
	for locale, chain := range DefaultFallbacks {
		if len(chain) == 0 {
			t.Errorf("chain for %q should not be empty", locale)
		}
	}
}

func TestDefaultFallbacks_NoSelfReferences(t *testing.T) {
	for locale, chain := range DefaultFallbacks {
		for _, fallback := range chain {
			if fallback == locale {
				t.Errorf("chain for %q contains itself as a fallback", locale)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Specific chain data verification
// ---------------------------------------------------------------------------

func TestDefaultFallbacks_SpecificChains(t *testing.T) {
	tests := []struct {
		locale string
		want   []string
	}{
		// Portuguese
		{"pt-BR", []string{"pt-PT", "pt"}},
		{"pt-PT", []string{"pt"}},
		{"pt-AO", []string{"pt-PT", "pt"}},
		{"pt-MZ", []string{"pt-PT", "pt"}},

		// Spanish
		{"es-419", []string{"es"}},
		{"es-MX", []string{"es-419", "es"}},
		{"es-AR", []string{"es-419", "es"}},

		// French
		{"fr-CA", []string{"fr"}},
		{"fr-BE", []string{"fr"}},
		{"fr-CH", []string{"fr"}},

		// German
		{"de-AT", []string{"de"}},
		{"de-CH", []string{"de"}},

		// Chinese
		{"zh-Hant-HK", []string{"zh-Hant-TW", "zh-Hant"}},
		{"zh-Hant-MO", []string{"zh-Hant-HK", "zh-Hant-TW", "zh-Hant"}},
		{"zh-Hant-TW", []string{"zh-Hant"}},
		{"zh-Hans-SG", []string{"zh-Hans"}},

		// English
		{"en-GB", []string{"en"}},
		{"en-AU", []string{"en-GB", "en"}},
		{"en-NZ", []string{"en-AU", "en-GB", "en"}},
		{"en-IN", []string{"en-GB", "en"}},

		// Norwegian
		{"nb", []string{"no"}},
		{"nn", []string{"nb", "no"}},

		// Other
		{"it-CH", []string{"it"}},
		{"nl-BE", []string{"nl"}},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			got := DefaultFallbacks[tt.locale]
			if len(got) != len(tt.want) {
				t.Fatalf("DefaultFallbacks[%q] = %v (len %d), want %v (len %d)",
					tt.locale, got, len(got), tt.want, len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("DefaultFallbacks[%q][%d] = %q, want %q",
						tt.locale, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MergeFallbacks
// ---------------------------------------------------------------------------

func TestMergeFallbacks_OverridesReplaceMatchingKeys(t *testing.T) {
	base := map[string][]string{
		"pt-BR": {"pt-PT", "pt"},
		"fr-CA": {"fr"},
	}
	overrides := map[string][]string{
		"pt-BR": {"pt"},
	}

	result := MergeFallbacks(base, overrides)

	assertChainEqual(t, result, "pt-BR", []string{"pt"})
	assertChainEqual(t, result, "fr-CA", []string{"fr"})
}

func TestMergeFallbacks_NewLocalesAdded(t *testing.T) {
	base := map[string][]string{
		"pt-BR": {"pt-PT", "pt"},
	}
	overrides := map[string][]string{
		"zh-Hant": {"zh-Hans", "zh"},
	}

	result := MergeFallbacks(base, overrides)

	assertChainEqual(t, result, "pt-BR", []string{"pt-PT", "pt"})
	assertChainEqual(t, result, "zh-Hant", []string{"zh-Hans", "zh"})
}

func TestMergeFallbacks_DefaultsPreserved(t *testing.T) {
	base := map[string][]string{
		"de-AT": {"de"},
		"de-CH": {"de"},
		"de-LU": {"de"},
	}
	overrides := map[string][]string{
		"de-AT": {"de-DE", "de"},
	}

	result := MergeFallbacks(base, overrides)

	// de-AT is overridden
	assertChainEqual(t, result, "de-AT", []string{"de-DE", "de"})
	// de-CH and de-LU preserved from base
	assertChainEqual(t, result, "de-CH", []string{"de"})
	assertChainEqual(t, result, "de-LU", []string{"de"})
}

func TestMergeFallbacks_DoesNotMutateInputs(t *testing.T) {
	base := map[string][]string{
		"pt-BR": {"pt-PT", "pt"},
	}
	overrides := map[string][]string{
		"pt-BR": {"pt"},
	}

	// Capture originals
	baseCopy := map[string][]string{
		"pt-BR": {"pt-PT", "pt"},
	}
	overridesCopy := map[string][]string{
		"pt-BR": {"pt"},
	}

	result := MergeFallbacks(base, overrides)

	// Mutate result to verify isolation
	result["pt-BR"][0] = "MUTATED"

	assertChainEqual(t, base, "pt-BR", baseCopy["pt-BR"])
	assertChainEqual(t, overrides, "pt-BR", overridesCopy["pt-BR"])
}

func TestMergeFallbacks_EmptyOverrides(t *testing.T) {
	base := map[string][]string{
		"fr-CA": {"fr"},
		"de-AT": {"de"},
	}
	overrides := map[string][]string{}

	result := MergeFallbacks(base, overrides)

	assertChainEqual(t, result, "fr-CA", []string{"fr"})
	assertChainEqual(t, result, "de-AT", []string{"de"})
}

func TestMergeFallbacks_EmptyBase(t *testing.T) {
	base := map[string][]string{}
	overrides := map[string][]string{
		"fr-CA": {"fr"},
	}

	result := MergeFallbacks(base, overrides)

	assertChainEqual(t, result, "fr-CA", []string{"fr"})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertChainEqual(t *testing.T, m map[string][]string, locale string, want []string) {
	t.Helper()
	got, ok := m[locale]
	if !ok {
		t.Fatalf("locale %q not found in map", locale)
	}
	if len(got) != len(want) {
		t.Fatalf("chain for %q = %v (len %d), want %v (len %d)",
			locale, got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("chain for %q[%d] = %q, want %q", locale, i, got[i], want[i])
		}
	}
}
