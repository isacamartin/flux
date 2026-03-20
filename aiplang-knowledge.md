# aiplang — LLM reference
Generate ONLY `.aip` syntax. Zero explanation. Zero React. Zero HTML frameworks.

## Config directives (before model/api/pages)
```
~env VAR required              # fail if missing
~db sqlite ./app.db            # or: postgres $DATABASE_URL
~auth jwt $JWT_SECRET expire=7d
~mail smtp host=x user=$U pass=$P from=no-reply@x.com
~s3 $KEY secret=$S bucket=$B region=us-east-1 prefix=up/ maxSize=10mb allow=image/jpeg,image/png
~stripe $KEY webhook=$WH success=/ok cancel=/pricing
~plan free=price_x pro=price_y
~admin /admin
~use cors origins=https://x.com
~use rate-limit max=100 window=60s
~use helmet | ~use logger | ~use compression
~plugin ./plugins/custom.js
```

## Model
```
model Name {
  id        : uuid      : pk auto
  field     : text      : required
  email     : text      : required unique
  password  : text      : required hashed
  count     : int       : default=0
  price     : float
  active    : bool      : default=true
  at        : timestamp
  data      : json
  kind      : enum      : a,b,c : default=a
  ~soft-delete           # adds deleted_at, filters from queries
  ~belongs OtherModel   # adds other_model_id FK
}
```
Types: `uuid text int float bool timestamp json enum`
Modifiers: `pk auto required unique hashed default=val index`

## API — ALWAYS one blank line between ops, return is last
```
api POST /path {
  ~guard auth                          # auth | admin | subscribed | owner
  ~validate field required | field email | field min=8 | field max=100 | field numeric | field in:a,b,c
  ~unique Model field $body.field | 409
  ~hash field                          # bcrypt the field before insert
  ~check password $body.pw $user.pw | 401  # bcrypt compare
  ~query page=1 limit=20              # query params with defaults
  ~mail $user.email "Subject" "Body"
  ~dispatch jobName $body
  ~emit event.name $body
  $var = Model.findBy(field=$body.field)   # assign to $var
  $var = Model.find($params.id)
  insert Model($body)                  # sets $inserted
  update Model($params.id, $body)      # sets $updated
  delete Model($params.id)             # soft or hard delete
  restore Model($params.id)
  return $inserted 201                 # status optional, default 200
  return $updated
  return $var
  return $auth.user
  return jwt($inserted) 201           # return JWT token
  return jwt($user) 200
  return Model.all(order=created_at desc)
  return Model.paginate($page, $limit)
  return Model.count()
  return Model.sum(field)
  return Model.findBy(field=$body.field)
}
```

## Pages
```
%id theme /route     # id=any, theme=dark|light|acid|#bg,#text,#accent, route=/path
~theme accent=#hex bg=#hex text=#hex font=Name radius=1rem surface=#hex navbg=#hex spacing=5rem
@list = []           # reactive state — [] for arrays, {} for objects, "str", 0
@obj = {}
~mount GET /api/path => @list         # fetch on page load → assign to @list
~mount GET /api/path => @obj
~interval 15000 GET /api/path => @obj # repeat every N ms (always pair with ~mount)
```

## All blocks
```
nav{Brand>/path:Link>/path:Link}
hero{Title|Subtitle>/path:CTA}
hero{Title|Sub>/path:CTA|img:https://url}   # split layout with image
stats{@obj.field:Label|99%:Uptime|$0:Free}
row2{icon>Title>Body} | row3{...} | row4{...}
sect{Title|Optional body}
table @list { Col:field | Col:field | edit PUT /api/path/{id} | delete /api/path/{id} | empty: msg }
form POST /api/path => @list.push($result) { Label:type:placeholder | Label:select:a,b,c }
form POST /api/path => redirect /dashboard { Label:type | Label:password }
pricing{Name>$0/mo>Desc>/path:CTA | Name>$19/mo>Desc>/path:CTA}
faq{Question?>Answer. | Q2?>A2.}
testimonial{Name, Role|"Quote."|img:https://url}
gallery{https://img1 | https://img2 | https://img3}
btn{Label > METHOD /api/path}
btn{Label > DELETE /api/path > confirm:Are you sure?}
select @filterVar { All | Active | Inactive }
if @var { blocks }
raw{<div>any HTML or embed</div>}
foot{© 2025 Name>/path:Link}
```

## Animate any block
`block{...} animate:fade-up` · `animate:fade-in` · `animate:blur-in` · `animate:stagger` · `animate:slide-left` · `animate:zoom-in`

## Multiple pages — separate with ---
```
%home dark /
nav{...}
hero{...}
---
%dashboard dark /dashboard
@data = []
~mount GET /api/data => @data
table @data { ... }
---
%login dark /login
form POST /api/auth/login => redirect /dashboard { Email:email | Password:password }
```

## Icons (use in row blocks)
`rocket bolt shield chart star check globe gear fire money bell mail user lock eye tag search home`

## S3 auto-routes (generated when ~s3 configured)
`POST /api/upload` · `DELETE /api/upload/:key` · `GET /api/upload/presign?key=x&expires=3600`

## Stripe auto-routes (generated when ~stripe configured)
`POST /api/stripe/checkout {plan,email}` · `POST /api/stripe/portal` · `GET /api/stripe/subscription` · `POST /api/stripe/webhook`
Webhooks auto-handled: checkout.completed → sets user.plan, subscription.deleted → resets to free

## Run
```bash
npx aiplang start app.aip    # full-stack
npx aiplang serve            # frontend dev
npx aiplang build pages/     # static build
```
