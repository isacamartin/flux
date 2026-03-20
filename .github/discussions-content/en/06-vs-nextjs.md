# ⚡ aiplang vs Next.js — when to use which

A common question. Direct and honest answer.

---

## TL;DR

| Situation | Use |
|---|---|
| Rapid prototyping, MVP, idea validation | **aiplang** |
| LLM generating the app (Claude, GPT) | **aiplang** |
| Landing page, internal dashboard, simple SaaS | **aiplang** |
| Production app with a team, SSR, TypeScript | **Next.js** |
| High-volume e-commerce | **Next.js + Shopify** |
| App with millions of users | **Next.js** |

---

## Where aiplang wins

### 20× fewer tokens for the LLM
The LLM uses ~490 tokens to generate a complete SaaS in aiplang vs ~10,200 for Next.js.

### Zero config
No tsconfig, tailwind.config, prisma schema, next.config, manual package.json.

### One file = complete app
```aip
~db sqlite ./app.db
~auth jwt $SECRET expire=7d
~stripe $STRIPE_KEY webhook=$WH success=/ok cancel=/pricing

model User { ... }
api POST /api/auth/login { ... }

%home dark /
hero{My SaaS|Ready in 30 seconds.}
```

### Auth + ORM + rate-limit + cache built-in
Next.js needs NextAuth + Prisma + rate-limit + Redis separately.

---

## Where Next.js wins

- **SSR/ISR** — server-side rendering for heavy SEO needs
- **Edge Functions** — ultra-low latency globally
- **TypeScript** — strong typing for large teams
- **Ecosystem** — 10M+ npm packages, huge community
- **Vercel** — integrated deploy, analytics, preview deployments
- **Trust** — 8 years of history, used by Fortune 500 companies

---

## The smart combination

Use **aiplang to prototype** → validate the idea → **migrate to Next.js when you scale**.

```bash
# Prototype in 30 minutes:
npx aiplang start app.aip

# When you need to scale:
npx aiplang export --target nextjs  # coming in v3.0
```

---

What do you use? Comment below 👇
