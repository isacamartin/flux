# aiplang v2.0 — Full-Stack AI-First Language Specification

## Philosophy

aiplang v2 is a complete full-stack language.
One .aip file = frontend + backend + database + auth + API.
Written by AI. Not for humans. Optimized for token density.

---

## File structure

```
~env                          ← environment declarations
~db                           ← database connection
~auth                         ← authentication config
~cache                        ← cache config

model Name { fields }         ← data models (generates DB tables)

api METHOD /path { handler }  ← backend API routes

%page theme /route            ← frontend pages
blocks...
---
%page2 theme /route2
...
```

---

## Environment

```aiplang
~env DATABASE_URL required
~env JWT_SECRET required
~env PORT=3000
~env NODE_ENV=production
```

---

## Database

```aiplang
~db postgres $DATABASE_URL
~db mysql    $DATABASE_URL
~db sqlite   ./data.db
~db mongo    $MONGO_URL
```

---

## Models (generates SQL tables + CRUD ops)

```aiplang
model User {
  id         : uuid      : pk auto
  name       : text      : required
  email      : text      : required unique
  password   : text      : required hashed
  plan       : enum      : starter,pro,enterprise : default=starter
  status     : enum      : active,inactive,banned : default=active
  created_at : timestamp : auto
  updated_at : timestamp : auto
}

model Post {
  id         : uuid      : pk auto
  title      : text      : required
  body       : text      : required
  author_id  : ref User  : required
  published  : bool      : default=false
  created_at : timestamp : auto
}
```

### Field types
`uuid` `text` `int` `float` `bool` `timestamp` `json` `enum`

### Field modifiers
`pk` `auto` `required` `unique` `hashed` `default=value`
`ref ModelName` (foreign key)

---

## API Routes

```aiplang
api GET /api/users {
  ~guard auth
  ~query page=1 limit=20
  return User.all(limit=$limit, offset=($page-1)*$limit, order=created_at desc)
}

api GET /api/users/:id {
  ~guard auth
  return User.find($id) | 404
}

api POST /api/users {
  ~validate name required | email required email | password min=8
  ~hash password
  insert User($body)
  return $inserted 201
}

api PUT /api/users/:id {
  ~guard auth | owner
  ~validate name? | email? email
  update User($id, $body)
  return $updated
}

api DELETE /api/users/:id {
  ~guard auth | admin
  delete User($id)
  return 204
}

api POST /api/auth/login {
  ~validate email required | password required
  $user = User.findBy(email=$body.email)
  ~check password $body.password $user.password | 401
  return jwt($user) 200
}

api POST /api/auth/register {
  ~validate name required | email required email | password min=8
  ~unique User email $body.email | 409
  ~hash $body.password
  insert User($body)
  return jwt($inserted) 201
}

api GET /api/me {
  ~guard auth
  return $auth.user
}
```

### API directives
```
~guard auth          require valid JWT/session
~guard admin         require admin role
~guard owner         require record belongs to user
~validate field rule | field rule
~query param=default extract query string params
~hash field          bcrypt hash before insert
~unique Model field value | statusCode  check uniqueness
~check password plain hash | statusCode verify bcrypt
```

### API expressions
```
Model.all(limit=N, offset=N, order=field dir, where=expr)
Model.find($id)
Model.findBy(field=value)
insert Model($body)
update Model($id, $body)
delete Model($id)
jwt($user)           generate JWT token
$body                parsed request body
$params.field        URL params (:id → $id or $params.id)
$query.field         query string
$auth.user           authenticated user
$inserted            result of insert
$updated             result of update
```

---

## Frontend Pages

Same as aiplang v1 + new reactive bindings:

```aiplang
%dashboard dark /dashboard

@users = []
@stats = {}
~mount GET /api/users => @users
~mount GET /api/stats => @stats
~interval 30000 GET /api/stats => @stats

nav{AppName>/logout:Sign out}
stats{@stats.total:Users|@stats.mrr:MRR|@stats.uptime:Uptime}
table @users { Name:name | Email:email | Plan:plan | Status:status | edit PUT /api/users/{id} | delete /api/users/{id} }
form POST /api/users => @users.push($result) { Name:text:Alice | Email:email:alice@co.com | Plan:select:starter,pro,enterprise }
foot{© 2025 AppName}
```

---

## Auth config

```aiplang
~auth jwt $JWT_SECRET expire=7d
~auth session $SESSION_SECRET expire=30d
~auth google $GOOGLE_CLIENT_ID $GOOGLE_SECRET
~auth github $GITHUB_CLIENT_ID $GITHUB_SECRET
```

---

## Cache

```aiplang
~cache redis $REDIS_URL ttl=300
~cache memory ttl=60
```

---

## Middleware (runs before all routes)

```aiplang
~middleware cors origins=* | rate-limit 100/min | log
```

---

## Complete example — SaaS app

```aiplang
~env DATABASE_URL required
~env JWT_SECRET required
~env PORT=3000

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

model Subscription {
  id         : uuid      : pk auto
  user_id    : ref User  : required
  plan       : enum      : starter,pro,enterprise
  status     : enum      : active,cancelled,past_due
  started_at : timestamp : auto
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

api GET /api/me {
  ~guard auth
  return $auth.user
}

api GET /api/users {
  ~guard admin
  ~query page=1 limit=20
  return User.all(limit=$limit, offset=($page-1)*$limit, order=created_at desc)
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

nav{AppName>/pricing:Pricing>/login:Sign in>/signup:Get started}
hero{Ship faster with AI|Zero config, infinite scale.>/signup:Start free}
row3{rocket>Deploy instantly>3 seconds from push to live.|shield>Enterprise>SOC2, GDPR, SSO.|chart>Observability>Real-time errors and performance.}
pricing{Starter>Free>3 projects, 1GB>/signup:Get started|Pro>$29/mo>Unlimited, analytics>/signup:Start trial|Enterprise>Custom>SSO, SLA>/contact:Talk to sales}
foot{© 2025 AppName>/privacy:Privacy>/terms:Terms}

---

%dashboard dark /dashboard

@user = {}
@users = []
@stats = {}
~mount GET /api/me => @user
~mount GET /api/users => @users
~mount GET /api/stats => @stats
~interval 30000 GET /api/stats => @stats

nav{AppName>/logout:Sign out}
stats{@stats.total:Users|@stats.mrr:MRR|@stats.active:Active}
table @users { Name:name | Email:email | Plan:plan | Status:status | edit PUT /api/users/{id} | delete /api/users/{id} }
foot{© 2025 AppName}

---

%login dark /login

nav{AppName}
hero{Welcome back|Sign in to continue.}
form POST /api/auth/login => redirect /dashboard { Email:email:you@co.com | Password:password: }
foot{© 2025 AppName}

---

%signup dark /signup

nav{AppName}
hero{Create your account|Start for free, no credit card.}
form POST /api/auth/register => redirect /dashboard { Name:text:Alice Johnson | Email:email:alice@co.com | Password:password: }
foot{© 2025 AppName}
```

---

## CLI (aiplangd — the aiplang daemon)

```bash
# Development
aiplangd dev app.flux              # start dev server with hot reload

# Production
aiplangd build app.flux            # compile to optimized Go binary
aiplangd start app.flux            # start production server
aiplangd migrate app.flux          # run DB migrations only
aiplangd migrate app.flux --reset  # drop and recreate tables

# Deploy
aiplangd deploy app.flux           # deploy to Fly.io / Railway / Render
```

---

## Output — what aiplangd generates

```
.aip source
  ↓ compiler
Go HTTP server (net/http)
  ├── /api/* routes (generated handlers)
  ├── /* frontend routes (SSR HTML)
  └── /static/* (CSS, JS, hydration runtime)

Database layer (via GORM or raw SQL)
  ├── auto-migration on startup
  └── generated CRUD operations

Single binary deployment
  └── app (static binary, ~10MB)
```

---

## Performance targets

| Metric | aiplang v2 | Next.js | Rails |
|---|---|---|---|
| Req/sec (API) | >50,000 | ~8,000 | ~5,000 |
| Cold start | <10ms | ~300ms | ~500ms |
| Binary size | ~10MB | N/A | N/A |
| Memory | ~20MB | ~200MB | ~150MB |
| DB query | direct SQL | ORM overhead | ORM overhead |

---

## Token density (AI generation)

Same app as above:
- aiplang v2:   ~600 tokens (frontend + backend + DB)
- Next.js + Express + Prisma: ~12,000 tokens
- Ratio: **20× fewer tokens**
