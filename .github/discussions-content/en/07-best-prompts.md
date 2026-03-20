# 🧠 Best prompts for generating apps with aiplang

A collection of prompts that work well. Share yours in the comments!

---

## Base system prompt (Claude Project)

Paste this as System Prompt in your Claude Project and upload `aiplang-knowledge.md`:

```
You are an aiplang code generator. When asked to build any web app, page, or component:
1. Respond ONLY with valid .aip syntax — no markdown, no explanation, no ```blocks```
2. Never use React, HTML, Next.js, or any other framework
3. Always start backend directives (~db, ~auth) before model{} and api{} blocks
4. Pages always start with %id theme /route
5. Separate pages with ---
6. Use dark theme unless specified otherwise
7. For dynamic data: declare @var = [] or @var = {} and use ~mount
8. Tables with edit/delete unless explicitly readonly
9. Forms must have => @list.push($result) or => redirect /path
10. Never add explanation — output only .aip code
```

---

## Prompts that work

### Simple landing page
```
Generate an aiplang landing page for [product]:
- Hero with impactful title and subtitle
- 3 features in row3 with icons
- Pricing with 3 plans (Free, Pro, Enterprise)
- FAQ with 4 common questions
- Customer testimonial
- Theme: accent=#[color], dark, font=Syne
```

### Complete SaaS
```
Generate a SaaS in aiplang:

Config: ~db sqlite, ~auth jwt, ~admin /admin, ~use helmet

Models:
- User: nome, email (unique), password (hashed), plan (enum: free,pro), role (enum: user,admin)
- [Model2]: [fields]

APIs:
- POST /api/auth/register — validate + unique + hash + jwt 201
- POST /api/auth/login — findBy + check + jwt 200
- GET /api/me — guard auth
- GET /api/[resource] — guard auth, return paginate
- POST /api/[resource] — guard auth, validate, insert, return 201

4 pages: home (hero + pricing), login, signup, dashboard (stats + table + form)
Theme: accent=#6366f1 radius=1rem font=Inter dark
```

### API only (no frontend)
```
Generate only the APIs in aiplang for a [domain] system:

~db postgres $DATABASE_URL
~auth jwt $JWT_SECRET expire=7d

Models: [list here]

Required APIs:
- [List each route with method, path, guard, validation and what it returns]
```

### Real-time dashboard
```
Generate an aiplang dashboard:
- Stats: [metric1], [metric2], [metric3] — refresh every 30s
- Table of [entity] with edit and delete
- Form to create new [entity]
- Polling: ~interval 30000 on all metrics
- Theme dark, accent=#0ea5e9
```

---

## Prompts that FAIL

❌ `"create an app"` — too vague, always generates something generic

❌ `"make a Twitter clone"` — without detailing models, generates poorly

❌ `"app with payments"` — without specifying provider, plans, and flow

---

## Your prompt here

Post in the comments prompts that worked well for you! 👇

Suggested format:
```
**Type:** [SaaS / Blog / Dashboard / Landing / etc]
**Prompt:**
[paste here]
**Result:** [what it generated]
```
