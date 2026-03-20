# 🗺️ Public Roadmap — aiplang v2.9 → v3.0

What's being built, in what order, and how you can help.

---

## ✅ Shipped in v2.9.1

- `.env` auto-load
- PostgreSQL — `~db postgres $DATABASE_URL`
- WebSockets — `~realtime`
- In-memory cache — `~cache key 60`
- Module imports — `~import ./auth.aip`
- 5 ready-to-use templates (blog, ecommerce, todo, analytics, chat)
- GitHub Actions — automated tests ✅
- SECURITY.md — threat model
- PROMPT_GUIDE.md — how to write great prompts
- VS Code syntax highlighting + snippets

---

## 🔨 Next — v3.0

| Feature | Status | ETA |
|---|---|---|
| Online playground — paste a prompt, see the app | 🔨 Planned | 1 month |
| VS Code extension on Marketplace | 🔨 Planned | 2 months |
| `npx aiplang export` — generate Next.js/Express code | 🔨 Planned | 2 months |
| Customizable admin panel with .aip blocks | 🔨 Planned | 3 months |
| Real AST parser (replaces string matching) | 🔨 Planned | 3 months |

---

## 💡 Open backlog (your vote matters)

- [ ] `~cron "0 9 * * *" jobName` — scheduled jobs
- [ ] `~webhook /api/github` — generic webhook handlers
- [ ] `~graphql` — auto GraphQL endpoint
- [ ] Embeddable playground (iframe)
- [ ] `aiplang tune` — fine-tune with your own examples
- [ ] Multi-LLM support (Gemini, Llama, Grok)
- [ ] `~multi-tenant` — multi-organization apps

---

## How to vote

👍 the items above that matter most to you, or open a new thread in [💡 Ideas](../categories/ideas) with details.

---

_Last updated: March 2026 · [@isacamartin](https://github.com/isacamartin)_
