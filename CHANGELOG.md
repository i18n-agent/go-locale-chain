# Changelog

## [1.0.0] - 2026-03-14

### Added
- `Configure()` initializes the global resolver with 75 default fallback chains covering Chinese, Portuguese, Spanish, French, German, Italian, Dutch, English, Arabic, Norwegian, and Malay regional variants
- `ConfigureWithOverrides()` merges custom overrides on top of default fallback chains
- `ConfigureCustom()` initializes with a fully custom fallback map, with optional default merging
- `Resolve()` performs per-key fallback across the full chain using pre-loaded message bundles
- `ResolveWithLoader()` performs per-key fallback with on-demand locale loading via callback
- `ChainFor()` returns the configured fallback chain for a given locale
- `Reset()` clears the global resolver for reconfiguration
- `MergeFallbacks()` utility for merging two fallback maps
- `DefaultFallbacks` map with 75 built-in chains covering 11 language families
- Goroutine-safe design using `sync.RWMutex`
- Zero external dependencies
