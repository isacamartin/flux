# ⚡ aiplang vs Next.js — quando usar cada um

Pergunta frequente. Resposta direta e honesta.

---

## TL;DR

| Situação | Use |
|---|---|
| Prototipagem rápida, MVP, teste de ideia | **aiplang** |
| LLM gerando o app (Claude, GPT) | **aiplang** |
| Landing page, dashboard interno, SaaS simples | **aiplang** |
| App em produção com equipe, SSR, TypeScript | **Next.js** |
| E-commerce de alto volume | **Next.js + Shopify** |
| App com milhões de usuários | **Next.js** |

---

## Onde aiplang vence

### 20× menos tokens para o LLM
O LLM gasta ~490 tokens para gerar um SaaS completo em aiplang vs ~10.200 para Next.js.

### Zero config
Sem tsconfig, tailwind.config, prisma schema, next.config, package.json manual.

### Um arquivo = app completo
```aip
~db sqlite ./app.db
~auth jwt $SECRET expire=7d
~stripe $STRIPE_KEY webhook=$WH_SECRET success=/ok cancel=/pricing

model User { ... }
api POST /api/auth/login { ... }

%home dark /
hero{Meu SaaS|Pronto em 30 segundos.}
```

### Auth + ORM + rate-limit + cache embutidos
Next.js precisa de NextAuth + Prisma + rate-limit + Redis separados.

---

## Onde Next.js vence

- **SSR/ISR** — renderização server-side para SEO pesado
- **Edge Functions** — latência ultra-baixa global
- **TypeScript nativo** — tipagem forte em projetos grandes
- **Ecosystem** — 10M+ de pacotes npm, comunidade enorme
- **Vercel** — deploy integrado, analytics, preview deployments
- **Confiança** — 8 anos de histórico, usado por empresas Fortune 500

---

## A combinação inteligente

Use **aiplang para prototipar** → valide a ideia → **migre para Next.js quando escalar**.

```bash
# Protótipo em 30 minutos:
npx aiplang start app.aip

# Quando precisar escalar:
npx aiplang export --target nextjs  # em breve v3.0
```

---

O que você usa? Comente abaixo 👇
