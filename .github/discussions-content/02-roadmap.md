# 🗺️ Roadmap público — aiplang v2.9 → v3.0

O que está sendo construído, em que ordem, e como você pode ajudar.

---

## ✅ Entregue em v2.9.1

- `.env` auto-load
- PostgreSQL — `~db postgres $DATABASE_URL`
- WebSockets — `~realtime`
- Cache in-memory — `~cache key 60`
- Módulos — `~import ./auth.aip`
- 5 templates prontos (blog, ecommerce, todo, analytics, chat)
- GitHub Actions — testes automáticos ✅
- SECURITY.md — threat model
- PROMPT_GUIDE.md — guia de prompts
- VS Code syntax highlighting + snippets

---

## 🔨 Próximo — v3.0

| Feature | Status | Prazo |
|---|---|---|
| Playground online — cole um prompt, veja o app rodando | 🔨 Planejado | 1 mês |
| VS Code extension na Marketplace | 🔨 Planejado | 2 meses |
| `npx aiplang export` — gera Next.js/Express equivalente | 🔨 Planejado | 2 meses |
| Admin panel customizável com blocos .aip | 🔨 Planejado | 3 meses |
| AST parser real (substitui string matching) | 🔨 Planejado | 3 meses |

---

## 💡 Backlog aberto (sua opinião importa)

- [ ] `~cron "0 9 * * *" jobName` — cron jobs
- [ ] `~webhook /api/github` — webhook handlers genéricos
- [ ] `~graphql` — endpoint GraphQL automático
- [ ] Playground embeddável como iframe
- [ ] `aiplang tune` — fine-tuning com seus exemplos
- [ ] Suporte multi-LLM (Gemini, Llama, Grok)
- [ ] `~multi-tenant` — apps multi-organização

---

## Como votar?

👍 nos itens acima que mais importam para você, ou abra um novo tópico em [💡 Ideas](../categories/ideas) com detalhes.

---

_Atualizado em: Março 2026 · [@isacamartin](https://github.com/isacamartin)_
