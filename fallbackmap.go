package localechain

// DefaultFallbacks contains the canonical locale fallback chains.
// Each key is a locale tag that maps to an ordered list of fallback
// locales tried from first to last when a translation key is missing.
var DefaultFallbacks = map[string][]string{
	// Chinese (5)
	"zh-Hant-HK": {"zh-Hant-TW", "zh-Hant"},
	"zh-Hant-MO": {"zh-Hant-HK", "zh-Hant-TW", "zh-Hant"},
	"zh-Hant-TW": {"zh-Hant"},
	"zh-Hans-SG":  {"zh-Hans"},
	"zh-Hans-MY":  {"zh-Hans"},

	// Portuguese (4)
	"pt-BR": {"pt-PT", "pt"},
	"pt-PT": {"pt"},
	"pt-AO": {"pt-PT", "pt"},
	"pt-MZ": {"pt-PT", "pt"},

	// Spanish (20)
	"es-419": {"es"},
	"es-MX":  {"es-419", "es"},
	"es-AR":  {"es-419", "es"},
	"es-CO":  {"es-419", "es"},
	"es-CL":  {"es-419", "es"},
	"es-PE":  {"es-419", "es"},
	"es-VE":  {"es-419", "es"},
	"es-EC":  {"es-419", "es"},
	"es-GT":  {"es-419", "es"},
	"es-CU":  {"es-419", "es"},
	"es-BO":  {"es-419", "es"},
	"es-DO":  {"es-419", "es"},
	"es-HN":  {"es-419", "es"},
	"es-PY":  {"es-419", "es"},
	"es-SV":  {"es-419", "es"},
	"es-NI":  {"es-419", "es"},
	"es-CR":  {"es-419", "es"},
	"es-PA":  {"es-419", "es"},
	"es-UY":  {"es-419", "es"},
	"es-PR":  {"es-419", "es"},

	// French (11)
	"fr-CA": {"fr"},
	"fr-BE": {"fr"},
	"fr-CH": {"fr"},
	"fr-LU": {"fr"},
	"fr-MC": {"fr"},
	"fr-SN": {"fr"},
	"fr-CI": {"fr"},
	"fr-ML": {"fr"},
	"fr-CM": {"fr"},
	"fr-MG": {"fr"},
	"fr-CD": {"fr"},

	// German (4)
	"de-AT": {"de"},
	"de-CH": {"de"},
	"de-LU": {"de"},
	"de-LI": {"de"},

	// Italian (1)
	"it-CH": {"it"},

	// Dutch (1)
	"nl-BE": {"nl"},

	// English (8)
	"en-GB": {"en"},
	"en-AU": {"en-GB", "en"},
	"en-NZ": {"en-AU", "en-GB", "en"},
	"en-IN": {"en-GB", "en"},
	"en-CA": {"en"},
	"en-ZA": {"en-GB", "en"},
	"en-IE": {"en-GB", "en"},
	"en-SG": {"en-GB", "en"},

	// Arabic (16)
	"ar-SA": {"ar"},
	"ar-EG": {"ar"},
	"ar-AE": {"ar"},
	"ar-MA": {"ar"},
	"ar-DZ": {"ar"},
	"ar-IQ": {"ar"},
	"ar-KW": {"ar"},
	"ar-QA": {"ar"},
	"ar-BH": {"ar"},
	"ar-OM": {"ar"},
	"ar-JO": {"ar"},
	"ar-LB": {"ar"},
	"ar-TN": {"ar"},
	"ar-LY": {"ar"},
	"ar-SD": {"ar"},
	"ar-YE": {"ar"},

	// Norwegian (2)
	"nb": {"no"},
	"nn": {"nb", "no"},

	// Malay (3)
	"ms-MY": {"ms"},
	"ms-SG": {"ms"},
	"ms-BN": {"ms"},
}

// MergeFallbacks merges override chains on top of a base set.
// Overrides replace entire chains for matching keys; base keys
// not present in overrides are preserved unchanged.
func MergeFallbacks(base, overrides map[string][]string) map[string][]string {
	merged := make(map[string][]string, len(base)+len(overrides))
	for k, v := range base {
		cp := make([]string, len(v))
		copy(cp, v)
		merged[k] = cp
	}
	for k, v := range overrides {
		cp := make([]string, len(v))
		copy(cp, v)
		merged[k] = cp
	}
	return merged
}
