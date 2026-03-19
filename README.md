# FLUX

> AI-first web language. One file = full app. Written by AI, not humans.

## Quick start

```bash
npx flux-web init my-app
cd my-app
npx flux-web serve
```

Ask Claude to generate a page → paste into `pages/home.flux` → see it live.

## npm package

```bash
npm install -g flux-web
```

→ [npmjs.com/package/flux-web](https://npmjs.com/package/flux-web)

## For Claude

Upload `FLUX-PROJECT-KNOWLEDGE.md` to a Claude Project to make Claude generate FLUX automatically.

## Docs

See [packages/flux-lang/README.md](packages/flux-lang/README.md) for full language reference.

## Performance

| | FLUX | React/Next | Laravel | Go+GORM |
|---|---|---|---|---|
| Tokens (same app) | **620** | 15,200 | 11,800 | 8,400 |
| First paint | **45ms** | 320ms | 210ms | 55ms |
| JS downloaded | **10KB** | 280KB | — | — |
| Lighthouse | **98** | 62 | 74 | 95 |

## License

MIT
