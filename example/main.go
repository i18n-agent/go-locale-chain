package main

import (
	"encoding/json"
	"fmt"
	"os"

	localechain "github.com/i18n-agent/go-locale-chain"
)

func main() {
	// Configure with an override so pt-BR falls back through pt all the way to en.
	// The default chain for pt-BR is [pt-PT, pt] which doesn't include en.
	// Adding en as the last fallback ensures every key is resolved.
	localechain.ConfigureWithOverrides(map[string][]string{
		"pt-BR": {"pt", "en"},
	})

	// Resolve translations by loading JSON files on demand.
	result, err := localechain.ResolveWithLoader("pt-BR", func(locale string) (map[string]string, error) {
		data, err := os.ReadFile(fmt.Sprintf("locales/%s.json", locale))
		if err != nil {
			return nil, err // locale file doesn't exist — skip
		}
		var msgs map[string]string
		if err := json.Unmarshal(data, &msgs); err != nil {
			return nil, err
		}
		return msgs, nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Print resolved translations.
	// greeting  -> "Oi"                      (from pt-BR)
	// farewell  -> "Adeus"                   (fallback to pt)
	// welcome   -> "Welcome to LocaleChain"  (fallback to en)
	fmt.Printf("greeting = %q\n", result["greeting"])
	fmt.Printf("farewell = %q\n", result["farewell"])
	fmt.Printf("welcome  = %q\n", result["welcome"])
}
