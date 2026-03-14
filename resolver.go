package localechain

// Resolve builds a flat key/value map for the given locale by walking the
// fallback chain and performing per-key fallback across all provided
// message bundles.
//
// The messages parameter maps locale tags to their key/value translations.
// The returned map contains every key found across the chain, with the
// value taken from the most specific (closest) locale that defines it.
//
// Load order (lowest to highest priority):
//
//	[...chain reversed, requestedLocale]
//
// If the resolver has not been configured the function still works: it
// returns the messages for the requested locale (if present) with no
// fallback.
func Resolve(locale string, messages map[string]map[string]string) map[string]string {
	loadOrder := buildLoadOrder(locale)

	result := make(map[string]string)
	for _, loc := range loadOrder {
		msgs, ok := messages[loc]
		if !ok {
			continue
		}
		for k, v := range msgs {
			result[k] = v
		}
	}

	// If the chain produced nothing and the exact locale has messages,
	// that is already covered above. When even the requested locale has
	// no messages the caller gets an empty map rather than nil.
	return result
}

// ResolveWithLoader is like Resolve but retrieves messages on demand
// through a loader function. The loader is called once per locale in
// the chain (including the requested locale). Errors from the loader
// are treated as "locale not available" and the locale is skipped.
func ResolveWithLoader(locale string, loader func(string) (map[string]string, error)) (map[string]string, error) {
	loadOrder := buildLoadOrder(locale)

	result := make(map[string]string)
	for _, loc := range loadOrder {
		msgs, err := loader(loc)
		if err != nil {
			// Skip locales that fail to load — they may not exist.
			continue
		}
		for k, v := range msgs {
			result[k] = v
		}
	}
	return result, nil
}

// ChainFor returns the full fallback chain for a locale as configured
// in the global resolver. The returned slice does NOT include the locale
// itself — only its fallbacks in priority order (first = most preferred
// fallback). Returns nil if the locale has no configured chain or the
// resolver is not configured.
func ChainFor(locale string) []string {
	mu.RLock()
	defer mu.RUnlock()

	if resolver == nil {
		return nil
	}

	chain, ok := resolver.fallbacks[locale]
	if !ok {
		return nil
	}

	// Return a copy so callers cannot mutate internal state.
	cp := make([]string, len(chain))
	copy(cp, chain)
	return cp
}

// buildLoadOrder constructs the deduped ordered list of locales to load
// for the given locale. Order is lowest-to-highest priority so that
// later entries override earlier ones when merging key/value maps:
//
//	[...chain reversed, requestedLocale]
//
// Duplicates are removed, keeping the first occurrence (which is the
// lowest-priority position — higher-priority duplicates later in the
// chain would override anyway).
func buildLoadOrder(locale string) []string {
	mu.RLock()
	var chain []string
	if resolver != nil {
		chain = resolver.fallbacks[locale]
	}
	mu.RUnlock()

	// Reverse chain so the most generic fallback is loaded first
	// (lowest priority) and the most specific fallback is loaded last.
	reversed := make([]string, len(chain))
	for i, loc := range chain {
		reversed[len(chain)-1-i] = loc
	}

	// Build load order: [...reversed chain, requestedLocale]
	seen := make(map[string]struct{}, len(reversed)+1)
	order := make([]string, 0, len(reversed)+1)

	for _, loc := range reversed {
		if _, dup := seen[loc]; !dup {
			seen[loc] = struct{}{}
			order = append(order, loc)
		}
	}
	if _, dup := seen[locale]; !dup {
		seen[locale] = struct{}{}
		order = append(order, locale)
	}

	return order
}
