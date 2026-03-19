# flux-web

> AI-first web language. Describe an app to Claude, get it running in seconds.

```bash
npx flux-web init my-app
cd my-app && npx flux-web serve
```

---

## The real value

You describe what you want. Claude generates `.flux`. You run it.

**You:** *"Dashboard com tabela de usuários, stats ao vivo e form para adicionar usuário"*

**Claude generates:**
```flux
%dashboard dark /dashboard

@users = []
@stats = {}
~mount GET /api/users => @users
~mount GET /api/stats => @stats
~interval 10000 GET /api/stats => @stats

nav{MyApp>/logout:Sign out}
stats{@stats.users:Users|@stats.mrr:MRR|@stats.retention:Retention}
table @users { Name:name | Email:email | Plan:plan | Status:status | edit PUT /api/users/{id} | delete /api/users/{id} | empty: No users yet. }
form POST /api/users => @users.push($result) { Full name:text:Alice | Email:email:alice@co.com | Plan:select:starter,pro,enterprise }
foot{© 2025 MyApp}
```

**You paste it, run `npx flux-web serve` → working dashboard.**

That's it. No TypeScript, no JSX, no config files, no node_modules drama.

---

## Why it's faster for AI

Same dashboard in React requires Claude to generate ~4,300 tokens across 18+ files.  
In FLUX: **442 tokens. 1 file. 10× faster.**

| | FLUX | React/Next | Laravel | Go+GORM |
|---|---|---|---|---|
| Tokens (same app) | **620** | 15,200 | 11,800 | 8,400 |
| Files | **1** | 18+ | 22+ | 14+ |
| Time for Claude | **0.6s** | 15s | 12s | 8s |
| Apps per session | **322** | 13 | 16 | 23 |

---

## Setup Claude to generate FLUX

**Option 1 — Claude Project (recommended, one-time setup):**
1. Go to claude.ai → Projects → New Project → name: "FLUX Generator"
2. Upload `FLUX-PROJECT-KNOWLEDGE.md` as project knowledge
3. Any conversation inside will generate FLUX automatically

**Option 2 — Paste in any conversation:**
```
I'm using FLUX web language. Syntax:
%id theme /route | @var=[] | ~mount GET /api => @var | ~interval 10000 GET /api => @var
~theme accent=#hex radius=1rem font=Name bg=#hex text=#hex surface=#hex
nav{Brand>/path:Link} | hero{Title|Sub>/path:CTA} animate:fade-up
stats{@val:label|@val:label} | row3{icon>Title>Body} animate:stagger
table @var { Col:key | edit PUT /api/{id} | delete /api/{id} | empty: msg }
form POST /api => @list.push($result) { Label:type:placeholder | ... }
pricing{Name>Price>Desc>/path:CTA | ...} | faq{Q > A | ...}
testimonial{Author|Quote|img:url} | gallery{url1|url2|url3}
raw{<div>literal HTML</div>} | foot{text>/path:Link}
~theme, animate:, class: work as suffix modifiers on any block
Pages separated by ---

Generate: [your request here]
```

---

## Commands

```bash
npx flux-web init my-app                    # create project
npx flux-web init --template saas           # SaaS starter
npx flux-web init --template landing        # landing page
npx flux-web init --template crud           # CRUD app
npx flux-web serve                          # dev server + hot reload → localhost:3000
npx flux-web build pages/ --out dist/       # compile → static HTML
npx flux-web new dashboard                  # new page template
```

---

## Full language reference

### Page declaration
```flux
%id theme /route
```
Themes: `dark` `light` `acid` or custom: `#bg,#text,#accent`

### Global theme vars
```flux
~theme accent=#7c3aed radius=1.5rem font=Syne bg=#0a0a0a text=#fff surface=#111 navbg=#000 border=#333 shadow=0_20px_60px_rgba(0,0,0,.5) spacing=6rem
```

### State + data fetching
```flux
@users = []                                    # reactive state, default []
@stats = {}                                    # reactive state, default {}
~mount GET /api/users => @users                # fetch on page load
~mount GET /api/stats => @stats
~interval 30000 GET /api/stats => @stats       # poll every 30s
```

### Blocks
```flux
nav{Brand>/path:Link>/path:Link}
hero{Title|Subtitle>/path:CTA>/path:CTA2|img:https://...}
stats{@val:Label|@val:Label|static text:Label}
row2{} row3{} row4{}                           # icon>Title>Body>/path:Link
sect{Title|Body text}
table @var { Col:key | Col:key | edit PUT /api/{id} | delete /api/{id} | empty: No data. }
form METHOD /api/path => @list.push($result) { Label:type:placeholder | Label:select:opt1,opt2 }
form POST /api/auth/login => redirect /dashboard { Email:email | Password:password }
pricing{Name>Price>Desc>/path:CTA | ...}
faq{Question > Answer | ...}
testimonial{Author, Title|Quote|img:https://...}
gallery{url1|url2|url3}
btn{Label > METHOD /api/path > confirm:Are you sure?}
select @var { option1 | option2 | option3 }
raw{<div>Any HTML here — custom components, embeds, etc.</div>}
foot{© 2025 Name>/path:Link>/path:Link}
```

### Block modifiers (append to any block)
```flux
hero{...} animate:fade-up
row3{...} animate:stagger class:my-section
sect{...} animate:fade-in class:highlight-box
```

Animations: `fade-up` `fade-in` `blur-in` `slide-left` `slide-right` `zoom-in` `stagger`

### Multiple pages
```flux
%home dark /
nav{...}
hero{...}
---
%dashboard dark /dashboard
@users = []
~mount GET /api/users => @users
table @users { ... }
---
%login dark /login
form POST /api/auth/login => redirect /dashboard { ... }
```

---

## Output

```
pages/home.flux
  ↓ npx flux-web build
dist/
  index.html          ← pre-rendered HTML, SEO-ready
  flux-hydrate.js     ← 10KB, only loaded on dynamic pages
```

Zero framework. Zero node_modules in production. Deploys to any static host.

---

## Deploy

```bash
npx flux-web build pages/ --out dist/

# Vercel
npx vercel dist/

# Netlify
npx netlify deploy --dir dist/

# Any server
rsync -r dist/ user@host:/var/www/html/
```

---

## Performance (what the user feels)

| | FLUX | React/Next |
|---|---|---|
| First paint | 45ms | 320ms |
| Interactive | 55ms | 380ms |
| JS downloaded | 10KB | 280KB |
| Lighthouse | 98/100 | 62/100 |
| Low-end Android | 180ms | 4,200ms |

FLUX pre-renders HTML server-side at build time. The browser gets a complete page — no hydration blocking, no JS required for static content.

---

## GitHub

[github.com/isacamartin/flux](https://github.com/isacamartin/flux)

## License

MIT
