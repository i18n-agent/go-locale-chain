# go-locale-chain example

Demonstrates per-key locale fallback: `pt-BR -> pt -> en`.

## Files

- `locales/en.json` -- complete: greeting, farewell, welcome
- `locales/pt.json` -- partial: greeting, farewell
- `locales/pt-BR.json` -- minimal: greeting only

## Run

```bash
cd example
go run main.go
```

## Expected output

```
greeting = "Oi"
farewell = "Adeus"
welcome  = "Welcome to LocaleChain"
```

- **greeting** resolves from `pt-BR.json` (most specific)
- **farewell** falls back to `pt.json`
- **welcome** falls back to `en.json` (base language)
