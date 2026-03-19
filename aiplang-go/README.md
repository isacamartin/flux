# aiplang v2

> Full-stack AI-first language. Frontend + Backend + Database + Auth in one file.

## What it is

aiplang v2 is a complete web language. One `.aip` file replaces:
- Next.js (frontend)
- Express/Fastify (backend API)
- Prisma/GORM (database ORM)
- JWT/Passport (authentication)
- Zod/Joi (validation)

Written by AI. Not designed for humans. Optimized for token density.

---

## Token comparison — same SaaS app

| Stack | Files | Tokens | Time for Claude |
|---|---|---|---|
| **aiplang v2** | **1** | **~600** | **~0.6s** |
| Next.js + Express + Prisma | 15+ | ~15,000 | ~15s |
| Rails | 10+ | ~12,000 | ~12s |
| Django | 10+ | ~11,000 | ~11s |

**20× fewer tokens than any other stack.**

---

## Example — complete SaaS in 80 lines

```aiplang
~env DATABASE_URL required
~env JWT_SECRET required

~db postgres $DATABASE_URL
~auth jwt $JWT_SECRET expire=7d
~middleware cors | rate-limit 100/min | log

model User {
  id         : uuid      : pk auto
  name       : text      : required
  email      : text      : required unique
  password   : text      : required hashed
  plan       : enum      : starter,pro,enterprise : default=starter
  role       : enum      : user,admin : default=user
  created_at : timestamp : auto
}

api POST /api/auth/register {
  ~validate name required | email required email | password min=8
  ~unique User email $body.email | 409
  insert User($body)
  return jwt($inserted) 201
}

api POST /api/auth/login {
  ~validate email required | password required
  $user = User.findBy(email=$body.email)
  ~check password $body.password $user.password | 401
  return jwt($user) 200
}

api GET /api/users {
  ~guard admin
  ~query page=1 limit=20
  return User.all(limit=$limit, order=created_at desc)
}

api PUT /api/users/:id {
  ~guard auth | owner
  update User($id, $body)
  return $updated
}

api DELETE /api/users/:id {
  ~guard auth | admin
  delete User($id)
  return 204
}

%home dark /

nav{MyApp>/pricing:Pricing>/login:Sign in}
hero{Ship faster|AI-native platform.>/signup:Start free}
row3{rocket>Deploy instantly>3 seconds.|shield>Enterprise>SOC2, GDPR, SSO.|chart>Observability>Real-time.}
foot{© 2025 MyApp}

---

%dashboard dark /dashboard

@users = []
@stats = {}
~mount GET /api/users => @users
~mount GET /api/stats => @stats
~interval 30000 GET /api/stats => @stats

nav{MyApp>/logout:Sign out}
stats{@stats.total:Users|@stats.mrr:MRR|@stats.active:Active}
table @users { Name:name | Email:email | Plan:plan | edit PUT /api/users/{id} | delete /api/users/{id} }
foot{MyApp Dashboard © 2025}

---

%login dark /login

nav{MyApp}
form POST /api/auth/login => redirect /dashboard { Email:email | Password:password }
foot{© 2025 MyApp}
```

That's a complete SaaS app. Auth, database, API, frontend, dashboard.

---

## Quick start

```bash
# Install
go install github.com/isacamartin/aiplang/cmd/aiplangd@latest

# Run
DATABASE_URL=sqlite://./app.db JWT_SECRET=secret aiplangd dev myapp.flux

# Or with .env file
aiplangd dev myapp.flux
```

---

## Language reference

### Config directives
```aiplang
~env NAME required | NAME=default
~db postgres $DATABASE_URL | mysql $URL | sqlite ./file.db
~auth jwt $SECRET expire=7d | session $SECRET
~middleware cors | rate-limit 100/min | log
```

### Models
```aiplang
model Name {
  field : type : modifiers
}
```

**Types:** `uuid` `text` `int` `float` `bool` `timestamp` `json` `enum`

**Modifiers:** `pk` `auto` `required` `unique` `hashed` `default=value` `ref Model`

### API routes
```aiplang
api METHOD /path {
  ~guard auth | admin | owner
  ~validate field required | field email | field min=N
  ~query param=default
  ~hash field
  ~check password $body.pw $user.pw | 401
  ~unique Model field $body.field | 409
  $var = Model.findBy(field=value)
  insert Model($body)
  update Model($id, $body)
  delete Model($id)
  return $expr statusCode
}
```

### Frontend pages
```aiplang
%id theme /route
@var = defaultValue
~mount GET /api/path => @var
~interval 30000 GET /api/path => @var
nav{} hero{} stats{} rowN{} table @var{} form METHOD /path => action {} pricing{} faq{} testimonial{} gallery{} foot{}
---
%nextpage theme /nextroute
...
```

---

## Architecture

```
.flux file
    ↓ aiplangd (Go binary)
    ├── Parse models → auto-migrate SQL tables
    ├── Parse API routes → Go HTTP handlers
    ├── Parse pages → SSR HTML with hydration
    └── Start HTTP server (net/http, ~50K req/s)
```

**Performance targets:**
- API: >50,000 req/sec (Go net/http, no framework overhead)
- Cold start: <10ms
- Memory: ~20MB
- Binary: ~10MB (single static binary)

---

## License

MIT
