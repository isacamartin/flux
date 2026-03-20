# ❓ How to generate an app with Claude in 60 seconds

Quick guide for getting started.

---

## Step 1 — Install

```bash
npm install -g aiplang
```

---

## Step 2 — Create a project

```bash
npx aiplang init my-app
cd my-app
npx aiplang serve
```

Opens at `http://localhost:3000` with hot reload.

---

## Step 3 — Generate with Claude

Paste this as your system prompt in Claude:

> You are an aiplang code generator. Respond ONLY with valid .aip syntax — no markdown, no explanation, no ```blocks```. Never use React, HTML, Next.js or other frameworks.

Then upload the `aiplang-knowledge.md` file from the npm package as context.

Now describe what you want:

> Generate a blog app with: JWT auth, Post model (title, body, published), CRUD for posts (admin creates/deletes, everyone reads), 3 pages (home with list, login, dashboard), theme accent=#10b981 dark

---

## Step 4 — Run the generated app

```bash
# Copy the generated .aip to pages/home.aip
# Or save as app.aip for full-stack

# Frontend only (static):
npx aiplang serve

# Full-stack (with DB, auth, APIs):
npx aiplang start app.aip
```

---

## Common question: what's the difference between `serve` and `start`?

| Command | When to use |
|---|---|
| `npx aiplang serve` | Static pages, no backend |
| `npx aiplang build pages/` | Compile for CDN deploy |
| `npx aiplang start app.aip` | Full app with DB + APIs |

---

Got stuck? Post below with your `.aip` file and the error 👇
