package localechain

import "sync"

// resolver holds the active fallback configuration.
// It is guarded by mu for concurrent access safety.
var (
	mu       sync.RWMutex
	resolver *chainResolver
)

// chainResolver stores the resolved fallback map used by Resolve and ChainFor.
type chainResolver struct {
	fallbacks map[string][]string
}

// Configure initialises the global resolver with the default fallback chains.
// It is safe to call from multiple goroutines.
func Configure() {
	mu.Lock()
	defer mu.Unlock()
	resolver = &chainResolver{
		fallbacks: copyMap(DefaultFallbacks),
	}
}

// ConfigureWithOverrides initialises the global resolver by merging the
// provided overrides on top of the default fallback chains. Overrides
// replace entire chains for matching locale keys.
func ConfigureWithOverrides(overrides map[string][]string) {
	mu.Lock()
	defer mu.Unlock()
	resolver = &chainResolver{
		fallbacks: MergeFallbacks(DefaultFallbacks, overrides),
	}
}

// ConfigureCustom initialises the global resolver with a fully custom
// fallback map. When mergeDefaults is true the custom map is merged on
// top of DefaultFallbacks; when false only the supplied map is used.
func ConfigureCustom(fallbacks map[string][]string, mergeDefaults bool) {
	mu.Lock()
	defer mu.Unlock()
	if mergeDefaults {
		resolver = &chainResolver{
			fallbacks: MergeFallbacks(DefaultFallbacks, fallbacks),
		}
	} else {
		resolver = &chainResolver{
			fallbacks: copyMap(fallbacks),
		}
	}
}

// Reset clears the global resolver, returning the package to its
// unconfigured state. Subsequent calls to Resolve or ChainFor will
// behave as if no fallback chains are registered (until Configure is
// called again).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	resolver = nil
}

// copyMap returns a deep copy of a fallback map so that mutations to
// the caller's map do not affect the resolver.
func copyMap(m map[string][]string) map[string][]string {
	cp := make(map[string][]string, len(m))
	for k, v := range m {
		s := make([]string, len(v))
		copy(s, v)
		cp[k] = s
	}
	return cp
}
