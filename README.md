# go-locale-chain

Smart locale fallback chains for Go -- because pt-BR users deserve pt-PT, not English.

## The Problem

Go's popular [`go-i18n`](https://github.com/nicksnyder/go-i18n) library accepts multiple languages in `i18n.NewLocalizer(bundle, "pt-BR", "pt-PT", "pt")`, but per-key cascade is broken: once a locale matches (has any translations loaded), missing keys don't fall through to the next locale ([Issue #239](https://github.com/nicksnyder/go-i18n/issues/239), [Issue #30](https://github.com/nicksnyder/go-i18n/issues/30)).

**Example:** Your app has `pt-PT` and `pt` translation files. A user requests `pt-BR`. Even if you pass all three locales to `NewLocalizer`, once `pt-BR` matches a single key, any keys missing from `pt-BR` return empty strings instead of falling back to `pt-PT` or `pt`.

The same thing happens with `es-MX` -> `es`, `fr-CA` -> `fr`, `de-AT` -> `de`, and every other regional variant.

Your users see empty strings or English when a perfectly good translation exists in a sibling locale.

## The Solution

One function call. Works standalone or alongside `go-i18n`.

`go-locale-chain` resolves translations with proper per-key fallback across the full chain. Every key is resolved from the most specific locale that defines it, walking the chain until a value is found.

## Installation

```bash
go get github.com/i18n-agent/go-locale-chain
```

Requires Go 1.21+.

## Quick Start

```go
package main

import (
    "fmt"
    localechain "github.com/i18n-agent/go-locale-chain"
)

func main() {
    // 1. Configure fallback chains (call once at startup)
    localechain.Configure()

    // 2. Load your translation bundles
    bundle := map[string]map[string]string{
        "pt":    {"hello": "Ola", "goodbye": "Adeus", "thanks": "Obrigado"},
        "pt-PT": {"hello": "Ola (PT)", "goodbye": "Adeus (PT)"},
        "pt-BR": {"hello": "Ola (BR)"},
    }

    // 3. Resolve with per-key fallback
    result := localechain.Resolve("pt-BR", bundle)

    fmt.Println(result["hello"])   // "Ola (BR)"   -- from pt-BR
    fmt.Println(result["goodbye"]) // "Adeus (PT)" -- from pt-PT (fallback)
    fmt.Println(result["thanks"])  // "Obrigado"   -- from pt (fallback)
}
```

## API Reference

### Configure

```go
// Use all 75 built-in fallback chains
localechain.Configure()
```

### ConfigureWithOverrides

```go
// Override specific chains while keeping all defaults
localechain.ConfigureWithOverrides(map[string][]string{
    "pt-BR": {"pt"},            // Simplify pt-BR chain
    "sv-FI": {"sv"},            // Add new chain
})
```

Your overrides replace matching keys in the default map. All other defaults remain.

### ConfigureCustom

```go
// Full control -- only use your chains
localechain.ConfigureCustom(map[string][]string{
    "pt-BR": {"pt-PT", "pt"},
    "es-MX": {"es-419", "es"},
}, false) // false = don't merge with defaults

// Or merge custom chains on top of defaults
localechain.ConfigureCustom(map[string][]string{
    "sv-FI": {"sv"},
}, true) // true = merge with defaults
```

### Resolve

```go
// Resolve translations with per-key fallback
result := localechain.Resolve("pt-BR", messages)
// messages: map[locale]map[key]value
// result:   map[key]value (merged from chain)
```

Returns a flat `map[string]string` with every key found across the chain, using the value from the most specific locale that defines it.

### ResolveWithLoader

```go
// Resolve with on-demand loading (e.g., from files or database)
result, err := localechain.ResolveWithLoader("pt-BR", func(locale string) (map[string]string, error) {
    return loadMessagesFromFile(locale)
})
```

The loader is called once per locale in the chain. Errors are treated as "locale not available" and the locale is skipped.

### ChainFor

```go
// Inspect the fallback chain for a locale
chain := localechain.ChainFor("pt-BR")
// Returns: ["pt-PT", "pt"]
```

Returns the configured fallback chain (not including the locale itself). Returns `nil` if the locale has no chain or the resolver is not configured.

### Reset

```go
// Clear the global resolver
localechain.Reset()
```

Returns the package to its unconfigured state. `Resolve` still works but without fallback chains.

### MergeFallbacks

```go
// Merge two fallback maps (useful for building custom configurations)
merged := localechain.MergeFallbacks(base, overrides)
```

Overrides replace entire chains for matching keys. Base keys not present in overrides are preserved.

### DefaultFallbacks

```go
// Access the built-in fallback map directly
chains := localechain.DefaultFallbacks
```

A `map[string][]string` with 75 built-in chains covering 11 language families.

## Usage with go-i18n

`go-locale-chain` works alongside `go-i18n` to fix the per-key fallback gap. Use `ResolveWithLoader` to load messages from go-i18n's `Bundle`:

```go
package main

import (
    "encoding/json"
    "os"

    "github.com/nicksnyder/go-i18n/v2/i18n"
    localechain "github.com/i18n-agent/go-locale-chain"
    "golang.org/x/text/language"
)

func main() {
    // Set up go-i18n bundle
    bundle := i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
    bundle.LoadMessageFile("locales/en.json")
    bundle.LoadMessageFile("locales/pt.json")
    bundle.LoadMessageFile("locales/pt-PT.json")
    bundle.LoadMessageFile("locales/pt-BR.json")

    // Configure fallback chains
    localechain.Configure()

    // Resolve with per-key fallback using go-i18n's localizer per chain step
    result, _ := localechain.ResolveWithLoader("pt-BR", func(locale string) (map[string]string, error) {
        localizer := i18n.NewLocalizer(bundle, locale)
        messages := make(map[string]string)
        // Load your known message IDs through the localizer
        for _, id := range []string{"hello", "goodbye", "thanks"} {
            msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: id})
            if err == nil {
                messages[id] = msg
            }
        }
        return messages, nil
    })

    // result now has proper per-key fallback across pt-BR -> pt-PT -> pt
    _ = result
}
```

## Standalone Usage

No external dependencies required. Works with any translation format:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    localechain "github.com/i18n-agent/go-locale-chain"
)

func main() {
    localechain.Configure()

    // Load translations from JSON files
    result, _ := localechain.ResolveWithLoader("zh-Hant-MO", func(locale string) (map[string]string, error) {
        data, err := os.ReadFile(fmt.Sprintf("locales/%s.json", locale))
        if err != nil {
            return nil, err
        }
        var msgs map[string]string
        if err := json.Unmarshal(data, &msgs); err != nil {
            return nil, err
        }
        return msgs, nil
    })

    // zh-Hant-MO -> zh-Hant-HK -> zh-Hant-TW -> zh-Hant
    // Each key resolved from the most specific locale that defines it
    _ = result
}
```

## Concurrency Safety

All exported functions are safe for concurrent use. The global resolver is protected by a `sync.RWMutex`. Call `Configure` once at startup and use `Resolve`/`ChainFor` freely from any goroutine.

## Default Fallback Map

75 built-in chains covering 11 language families.

### Chinese

| Locale | Fallback Chain |
|--------|---------------|
| zh-Hant-HK | zh-Hant-TW -> zh-Hant -> (default) |
| zh-Hant-MO | zh-Hant-HK -> zh-Hant-TW -> zh-Hant -> (default) |
| zh-Hant-TW | zh-Hant -> (default) |
| zh-Hans-SG | zh-Hans -> (default) |
| zh-Hans-MY | zh-Hans -> (default) |

### Portuguese

| Locale | Fallback Chain |
|--------|---------------|
| pt-BR | pt-PT -> pt -> (default) |
| pt-PT | pt -> (default) |
| pt-AO | pt-PT -> pt -> (default) |
| pt-MZ | pt-PT -> pt -> (default) |

### Spanish

| Locale | Fallback Chain |
|--------|---------------|
| es-419 | es -> (default) |
| es-MX | es-419 -> es -> (default) |
| es-AR | es-419 -> es -> (default) |
| es-CO | es-419 -> es -> (default) |
| es-CL | es-419 -> es -> (default) |
| es-PE | es-419 -> es -> (default) |
| es-VE | es-419 -> es -> (default) |
| es-EC | es-419 -> es -> (default) |
| es-GT | es-419 -> es -> (default) |
| es-CU | es-419 -> es -> (default) |
| es-BO | es-419 -> es -> (default) |
| es-DO | es-419 -> es -> (default) |
| es-HN | es-419 -> es -> (default) |
| es-PY | es-419 -> es -> (default) |
| es-SV | es-419 -> es -> (default) |
| es-NI | es-419 -> es -> (default) |
| es-CR | es-419 -> es -> (default) |
| es-PA | es-419 -> es -> (default) |
| es-UY | es-419 -> es -> (default) |
| es-PR | es-419 -> es -> (default) |

### French

| Locale | Fallback Chain |
|--------|---------------|
| fr-CA | fr -> (default) |
| fr-BE | fr -> (default) |
| fr-CH | fr -> (default) |
| fr-LU | fr -> (default) |
| fr-MC | fr -> (default) |
| fr-SN | fr -> (default) |
| fr-CI | fr -> (default) |
| fr-ML | fr -> (default) |
| fr-CM | fr -> (default) |
| fr-MG | fr -> (default) |
| fr-CD | fr -> (default) |

### German

| Locale | Fallback Chain |
|--------|---------------|
| de-AT | de -> (default) |
| de-CH | de -> (default) |
| de-LU | de -> (default) |
| de-LI | de -> (default) |

### Italian

| Locale | Fallback Chain |
|--------|---------------|
| it-CH | it -> (default) |

### Dutch

| Locale | Fallback Chain |
|--------|---------------|
| nl-BE | nl -> (default) |

### English

| Locale | Fallback Chain |
|--------|---------------|
| en-GB | en -> (default) |
| en-AU | en-GB -> en -> (default) |
| en-NZ | en-AU -> en-GB -> en -> (default) |
| en-IN | en-GB -> en -> (default) |
| en-CA | en -> (default) |
| en-ZA | en-GB -> en -> (default) |
| en-IE | en-GB -> en -> (default) |
| en-SG | en-GB -> en -> (default) |

### Arabic

| Locale | Fallback Chain |
|--------|---------------|
| ar-SA | ar -> (default) |
| ar-EG | ar -> (default) |
| ar-AE | ar -> (default) |
| ar-MA | ar -> (default) |
| ar-DZ | ar -> (default) |
| ar-IQ | ar -> (default) |
| ar-KW | ar -> (default) |
| ar-QA | ar -> (default) |
| ar-BH | ar -> (default) |
| ar-OM | ar -> (default) |
| ar-JO | ar -> (default) |
| ar-LB | ar -> (default) |
| ar-TN | ar -> (default) |
| ar-LY | ar -> (default) |
| ar-SD | ar -> (default) |
| ar-YE | ar -> (default) |

### Norwegian

| Locale | Fallback Chain |
|--------|---------------|
| nb | no -> (default) |
| nn | nb -> no -> (default) |

### Malay

| Locale | Fallback Chain |
|--------|---------------|
| ms-MY | ms -> (default) |
| ms-SG | ms -> (default) |
| ms-BN | ms -> (default) |

## How It Works

1. `Configure()` initializes the global resolver with the default fallback map (or your custom overrides).
2. When `Resolve()` is called, it builds a load order by reversing the fallback chain for the requested locale.
3. Translations are merged in lowest-to-highest priority order, so the most specific locale wins for each key.
4. The result is a flat `map[string]string` with every key resolved from the closest locale that defines it.
5. All shared state is protected by `sync.RWMutex` for goroutine safety.

## Publishing

Go modules are published by tagging the repository. No package registry upload needed.

```bash
git tag v1.0.0
git push --tags
```

The [Go Module Proxy](https://proxy.golang.org/) (GOPROXY) auto-indexes tagged versions from GitHub. Users can then install with:

```bash
go get github.com/i18n-agent/go-locale-chain@v1.0.0
```

## Requirements

- Go 1.21+
- Zero external dependencies

## FAQ

**Does this replace go-i18n?**
No. `go-locale-chain` complements `go-i18n` by fixing the per-key fallback gap. Use go-i18n for message loading, pluralization, and template rendering. Use `go-locale-chain` for proper fallback chain resolution.

**Performance impact?**
Negligible. The fallback chain is only walked when resolving translations. The load order is computed once per `Resolve` call, and map merging is O(n) where n is the total number of keys across the chain.

**Can I use this without go-i18n?**
Yes. `go-locale-chain` has zero external dependencies. It works with any translation format -- JSON files, TOML, YAML, database records, or in-memory maps. See the Standalone Usage section.

**Is it goroutine-safe?**
Yes. All exported functions are safe for concurrent use. `Configure` should be called once at startup. `Resolve`, `ResolveWithLoader`, and `ChainFor` can be called from any goroutine.

**What happens if I call Resolve without Configure?**
It still works -- it returns the messages for the requested locale with no fallback. You get the same behavior as not having this library.

**Can I reset the configuration?**
Yes. Call `localechain.Reset()` to clear the global resolver. You can reconfigure at any time with `Configure()` or its variants.

## Contributing

- Open issues for bugs or feature requests.
- PRs welcome, especially for adding new locale fallback chains.
- Run `go test ./...` before submitting.

## License

MIT License - see [LICENSE](LICENSE) file.

Built by [i18nagent.ai](https://i18nagent.ai)
