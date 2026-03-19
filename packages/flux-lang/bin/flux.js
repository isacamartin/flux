#!/usr/bin/env node
'use strict'

const fs   = require('fs')
const path = require('path')

const VERSION     = '1.2.0'
const RUNTIME_DIR = path.join(__dirname, '..', 'runtime')
const cmd         = process.argv[2]
const args        = process.argv.slice(3)

const ICONS = {
  bolt:'⚡',leaf:'🌱',map:'🗺',chart:'📊',lock:'🔒',star:'⭐',
  heart:'❤',check:'✓',alert:'⚠',user:'👤',car:'🚗',money:'💰',
  phone:'📱',shield:'🛡',fire:'🔥',rocket:'🚀',clock:'🕐',
  globe:'🌐',gear:'⚙',pin:'📍',flash:'⚡',eye:'◉',tag:'◈',
  plus:'+',minus:'−',edit:'✎',trash:'🗑',search:'⌕',bell:'🔔',
  home:'⌂',mail:'✉',quote:'"',thumb:'👍',
}

const esc   = s => s==null?'':String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')
const ic    = n => ICONS[n] || n
const isDyn = s => s&&(s.includes('@')||s.includes('$'))
const hSize = n => n<1024?`${n}B`:`${(n/1024).toFixed(1)}KB`

// ── HELP ──────────────────────────────────────────────────────────
if (!cmd || cmd==='--help' || cmd==='-h') {
  console.log(`
  flux-web v${VERSION}
  AI-first web language — full apps in ~20 lines.

  Usage:
    npx flux-web init [name]                 create project
    npx flux-web init --template saas        create from template
    npx flux-web init --template landing
    npx flux-web init --template crud
    npx flux-web serve [dir]                 dev server + hot reload
    npx flux-web build [dir/file]            compile → static HTML
    npx flux-web new <page>                  new page template
    npx flux-web --version

  GitHub: https://github.com/isacamartin/flux-web
  npm:    https://npmjs.com/package/flux-web
  `)
  process.exit(0)
}
if (cmd==='--version'||cmd==='-v') { console.log(`flux-web v${VERSION}`); process.exit(0) }

// ── TEMPLATES ─────────────────────────────────────────────────────
const TEMPLATES = {
  saas: (name, year) => `# ${name} — SaaS app
%home dark /

@stats = {}
~mount GET /api/stats => @stats

nav{${name}>/features:Features>/pricing:Pricing>/login:Sign in}
hero{Ship faster with AI|Zero config, infinite scale.>/signup:Start free>/demo:View demo}
stats{@stats.users:Users|@stats.mrr:MRR|@stats.uptime:Uptime}
row3{rocket>Deploy instantly>Push to git, live in 3 seconds.|shield>Enterprise ready>SOC2, GDPR, SSO built-in.|chart>Full observability>Real-time errors and performance.}
testimonial{Sarah Chen, CEO @ Acme|"${name} cut our deployment time by 90%. Absolutely game-changing."|img:https://i.pravatar.cc/64?img=47}
foot{© ${year} ${name}>/privacy:Privacy>/terms:Terms}

---

%pricing light /pricing

nav{${name}>/login:Sign in}
hero{Simple, transparent pricing|No hidden fees. Cancel anytime.}
pricing{
  Starter > Free > 3 projects, 1GB, community support > /signup:Get started|
  Pro > $29/mo > Unlimited projects, priority support, analytics > /signup:Start trial|
  Enterprise > Custom > SSO, SLA, dedicated CSM, on-prem > /contact:Talk to sales
}
faq{
  How do I get started? > Sign up free, no credit card required. Deploy your first app in 5 minutes.|
  Can I cancel anytime? > Yes. Cancel with one click, no questions asked.|
  Do you offer refunds? > Full refund within 14 days, no questions asked.
}
foot{© ${year} ${name}}

---

%login dark /login

nav{${name}}
hero{Welcome back|Sign in to your account.}
form POST /api/auth/login => redirect /dashboard {
  Email : email : you@company.com
  Password : password :
}
foot{© ${year} ${name}}`,

  landing: (name, year) => `# ${name} — Landing page
%home dark /

nav{${name}>/about:About>/login:Sign in}
hero{The future is now|${name} — built for the next generation.>/signup:Get started for free|img:https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=1200&q=80}
row3{rocket>Fast>Zero config, instant results.|bolt>Simple>One command to deploy.|globe>Global>CDN in 180+ countries.}
testimonial{Alex Rivera, CTO @ TechCorp|"We went from idea to production in a single afternoon. Incredible."|img:https://i.pravatar.cc/64?img=12}
foot{© ${year} ${name}}`,

  crud: (name, year) => `# ${name} — CRUD app
%users dark /users

@users = []
~mount GET /api/users => @users

nav{${name}>/users:Users>/settings:Settings}
sect{User Management}
table @users { Name:name | Email:email | Plan:plan | Status:status | edit PUT /api/users/{id} | delete /api/users/{id} | empty: No users yet. }
sect{Add User}
form POST /api/users => @users.push($result) {
  Full name : text : Alice Johnson
  Email : email : alice@company.com
  Plan : select : starter,pro,enterprise
}
foot{© ${year} ${name}}`,

  default: (name, year) => `# ${name}
%home dark /

nav{${name}>/login:Sign in}
hero{Welcome to ${name}|Edit pages/home.flux to get started.>/signup:Get started}
row3{rocket>Fast>Renders in under 1ms.|bolt>AI-native>Written by Claude in seconds.|globe>Deploy anywhere>Static files. Any host.}
foot{© ${year} ${name}}`,
}

// ── INIT ──────────────────────────────────────────────────────────
if (cmd==='init') {
  const tplIdx = args.indexOf('--template')
  const tplName = tplIdx !== -1 ? args[tplIdx+1] : 'default'
  const nameArg = args.find(a => !a.startsWith('--') && a !== tplName) || 'flux-app'
  const name = nameArg, dir = path.resolve(name), year = new Date().getFullYear()

  if (fs.existsSync(dir)) { console.error(`\n  ✗  Directory "${name}" already exists.\n`); process.exit(1) }
  fs.mkdirSync(path.join(dir,'pages'), {recursive:true})
  fs.mkdirSync(path.join(dir,'public'), {recursive:true})
  for (const f of ['flux-runtime.js','flux-hydrate.js']) {
    const src = path.join(RUNTIME_DIR,f)
    if (fs.existsSync(src)) fs.copyFileSync(src, path.join(dir,'public',f))
  }
  fs.writeFileSync(path.join(dir,'public','index.html'), `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${name}</title></head>
<body><div id="app"></div><script src="flux-runtime.js"></script><script>
fetch('../pages/home.flux').then(r=>r.text()).then(src=>FLUX.boot(src,document.getElementById('app')))
</script></body></html>`)

  const tplFn = TEMPLATES[tplName] || TEMPLATES.default
  fs.writeFileSync(path.join(dir,'pages','home.flux'), tplFn(name, year))
  fs.writeFileSync(path.join(dir,'package.json'), JSON.stringify({
    name, version:'0.1.0',
    scripts:{dev:'npx flux-web serve', build:'npx flux-web build pages/ --out dist/'},
    devDependencies:{'flux-web':`^${VERSION}`}
  },null,2))
  fs.writeFileSync(path.join(dir,'.gitignore'), 'dist/\nnode_modules/\n')

  const tplLabel = tplName !== 'default' ? ` (template: ${tplName})` : ''
  console.log(`\n  ✓  Created ${name}/${tplLabel}\n\n     pages/home.flux  ← edit this\n\n  Next:\n     cd ${name} && npx flux-web serve\n`)
  process.exit(0)
}

// ── NEW ────────────────────────────────────────────────────────────
if (cmd==='new') {
  const name = args[0]
  if (!name) { console.error('\n  ✗  Usage: flux-web new <page-name>\n'); process.exit(1) }
  const dir  = fs.existsSync('pages') ? 'pages' : '.'
  const file = path.join(dir, `${name}.flux`)
  if (fs.existsSync(file)) { console.error(`\n  ✗  ${file} already exists.\n`); process.exit(1) }
  const cap = name.charAt(0).toUpperCase()+name.slice(1)
  fs.writeFileSync(file, `# ${name}\n%${name} dark /${name}\n\nnav{AppName>/home:Home}\nhero{${cap}|Page description.>/action:Get started}\nfoot{© ${new Date().getFullYear()} AppName}\n`)
  console.log(`\n  ✓  Created ${file}\n`)
  process.exit(0)
}

// ── BUILD ──────────────────────────────────────────────────────────
if (cmd==='build') {
  const outIdx = args.indexOf('--out')
  const outDir = outIdx!==-1 ? args[outIdx+1] : 'dist'
  const input  = args.filter((a,i)=>!a.startsWith('--')&&i!==outIdx+1)[0]||'pages/'
  const files  = []
  if (fs.existsSync(input)&&fs.statSync(input).isDirectory()) {
    fs.readdirSync(input).filter(f=>f.endsWith('.flux')).forEach(f=>files.push(path.join(input,f)))
  } else if (input.endsWith('.flux')&&fs.existsSync(input)) { files.push(input) }
  if (!files.length) { console.error(`\n  ✗  No .flux files in: ${input}\n`); process.exit(1) }

  const src   = files.map(f=>fs.readFileSync(f,'utf8')).join('\n---\n')
  const pages = parseFlux(src)
  if (!pages.length) { console.error('\n  ✗  No pages found.\n'); process.exit(1) }

  fs.mkdirSync(outDir,{recursive:true})
  console.log(`\n  flux-web build — ${files.length} file(s)\n`)
  let total=0
  for (const page of pages) {
    const html  = renderPage(page, pages)
    const fname = page.route==='/'?'index.html':page.route.replace(/^\//,'')+'/index.html'
    const out   = path.join(outDir,fname)
    fs.mkdirSync(path.dirname(out),{recursive:true})
    fs.writeFileSync(out,html)
    const note = html.includes('flux-hydrate')?'+hydrate':'zero JS ✓'
    console.log(`  ✓  ${out.padEnd(40)} ${hSize(html.length)} (${note})`)
    total += html.length
  }
  const hf = path.join(RUNTIME_DIR,'flux-hydrate.js')
  if (fs.existsSync(hf)) {
    const dst=path.join(outDir,'flux-hydrate.js')
    fs.copyFileSync(hf,dst)
    total += fs.statSync(dst).size
    console.log(`  ✓  ${dst.padEnd(40)} ${hSize(fs.statSync(dst).size)}`)
  }
  if (fs.existsSync('public')) {
    fs.readdirSync('public').filter(f=>!f.endsWith('.flux'))
      .forEach(f=>fs.copyFileSync(path.join('public',f),path.join(outDir,f)))
  }
  console.log(`\n  ${pages.length} page(s) — ${hSize(total)} total\n\n  Preview: npx serve ${outDir}\n  Deploy:  Vercel, Netlify, S3, any static host\n`)
  process.exit(0)
}

// ── SERVE (hot reload) ────────────────────────────────────────────
if (cmd==='serve'||cmd==='dev') {
  const root = path.resolve(args[0]||'.')
  const port = parseInt(process.env.PORT||'3000')
  const MIME = {
    '.html':'text/html;charset=utf-8','.js':'application/javascript',
    '.css':'text/css','.flux':'text/plain','.json':'application/json',
    '.wasm':'application/wasm','.svg':'image/svg+xml',
    '.png':'image/png','.jpg':'image/jpeg','.ico':'image/x-icon',
  }

  // Track file mtimes for hot reload
  const mtimes = {}
  let clients  = []

  const checkChanges = () => {
    const pagesDir = path.join(root,'pages')
    if (!fs.existsSync(pagesDir)) return
    fs.readdirSync(pagesDir).filter(f=>f.endsWith('.flux')).forEach(f => {
      const fp  = path.join(pagesDir,f)
      const mt  = fs.statSync(fp).mtimeMs
      if (mtimes[fp] && mtimes[fp] !== mt) {
        clients.forEach(c => { try { c.write('data: reload\n\n') } catch {} })
      }
      mtimes[fp] = mt
    })
  }
  setInterval(checkChanges, 500)

  require('http').createServer((req, res) => {
    const p = req.url.split('?')[0]

    // SSE endpoint for hot reload
    if (p==='/__flux_reload') {
      res.writeHead(200,{'Content-Type':'text/event-stream','Cache-Control':'no-cache','Access-Control-Allow-Origin':'*'})
      res.write('data: connected\n\n')
      clients.push(res)
      req.on('close',()=>{ clients=clients.filter(c=>c!==res) })
      return
    }

    let urlPath = p==='/'?'/index.html':p
    let fp = null
    for (const c of [path.join(root,'public',urlPath),path.join(root,urlPath)]) {
      if (fs.existsSync(c)&&fs.statSync(c).isFile()) { fp=c; break }
    }
    if (!fp&&urlPath.endsWith('.flux')) {
      const c=path.join(root,'pages',path.basename(urlPath))
      if (fs.existsSync(c)) fp=c
    }
    if (!fp) { res.writeHead(404); res.end('Not found'); return }

    let content = fs.readFileSync(fp)
    // Inject hot reload script into HTML responses
    if (path.extname(fp)==='.html') {
      const inject = `\n<script>
const __es=new EventSource('/__flux_reload')
__es.onmessage=e=>{if(e.data==='reload')window.location.reload()}
</script>`
      content = content.toString().replace('</body>',inject+'</body>')
    }
    res.writeHead(200,{'Content-Type':MIME[path.extname(fp)]||'application/octet-stream','Access-Control-Allow-Origin':'*'})
    res.end(content)
  }).listen(port,()=>{
    console.log(`\n  ✓  flux-web dev server\n\n  →  http://localhost:${port}\n\n  Hot reload: ON — edit .flux files and browser auto-refreshes.\n  Ctrl+C to stop.\n`)
  })
  return
}

console.error(`\n  ✗  Unknown command: ${cmd}\n  Run flux-web --help\n`)
process.exit(1)

// ═══════════════════════════════════════════════════════════════
// PARSER
// ═══════════════════════════════════════════════════════════════

function parseFlux(src) {
  return src.split(/\n---\n/).map(s=>parsePage(s.trim())).filter(Boolean)
}

function parsePage(src) {
  const lines = src.split('\n').map(l=>l.trim()).filter(l=>l&&!l.startsWith('#'))
  if (!lines.length) return null
  const p = {id:'page',theme:'dark',route:'/',customTheme:null,state:{},queries:[],blocks:[]}
  for (const line of lines) {
    if (line.startsWith('%')) {
      const pts = line.slice(1).trim().split(/\s+/)
      p.id    = pts[0]||'page'
      p.route = pts[2]||'/'
      // theme can be "dark" | "light" | "acid" | "#hex1,#hex2" | "theme=#hex,#hex"
      const rawTheme = pts[1]||'dark'
      if (rawTheme.includes('#') || rawTheme.startsWith('theme=')) {
        const colors = rawTheme.replace('theme=','').split(',')
        p.theme = 'custom'
        p.customTheme = { bg: colors[0], text: colors[1]||'#f1f5f9', accent: colors[2]||'#2563eb', surface: colors[3]||null }
      } else {
        p.theme = rawTheme
      }
    } else if (line.startsWith('@')&&line.includes('=')) {
      const eq=line.indexOf('='); p.state[line.slice(1,eq).trim()]=line.slice(eq+1).trim()
    } else if (line.startsWith('~')) {
      const q=parseQuery(line.slice(1).trim()); if(q) p.queries.push(q)
    } else {
      const b=parseBlock(line); if(b) p.blocks.push(b)
    }
  }
  return p
}

function parseQuery(s) {
  const pts=s.split(/\s+/), ai=pts.indexOf('=>')
  if(pts[0]==='mount')    return{trigger:'mount',method:pts[1],path:pts[2],target:ai===-1?pts[3]:null,action:ai!==-1?pts.slice(ai+1).join(' '):null}
  if(pts[0]==='interval') return{trigger:'interval',interval:parseInt(pts[1]),method:pts[2],path:pts[3],target:ai===-1?pts[4]:null,action:ai!==-1?pts.slice(ai+1).join(' '):null}
  return null
}

function parseBlock(line) {
  // table
  if (line.startsWith('table ')) {
    const idx=line.indexOf('{'); if(idx===-1) return null
    const binding=line.slice(6,idx).trim()
    const content=line.slice(idx+1,line.lastIndexOf('}')).trim()
    const editMatch=content.match(/edit\s+(PUT|PATCH)\s+(\S+)/)
    const deleteMatch=content.match(/delete\s+(?:DELETE\s+)?(\S+)/)
    const clean=content.replace(/edit\s+(PUT|PATCH)\s+\S+/g,'').replace(/delete\s+(?:DELETE\s+)?\S+/g,'')
    return{kind:'table',binding,cols:parseCols(clean),empty:parseEmpty(clean),
      editPath:editMatch?.[2]||null,editMethod:editMatch?.[1]||'PUT',
      deletePath:deleteMatch?.[1]||null,deleteKey:'id'}
  }
  // form
  if (line.startsWith('form ')) {
    const bi=line.indexOf('{'); if(bi===-1) return null
    let head=line.slice(5,bi).trim()
    const content=line.slice(bi+1,line.lastIndexOf('}')).trim()
    let action=''
    const ai=head.indexOf('=>')
    if(ai!==-1){action=head.slice(ai+2).trim();head=head.slice(0,ai).trim()}
    const [method,bpath]=head.split(/\s+/)
    return{kind:'form',method:method||'POST',bpath:bpath||'',action,fields:parseFields(content)}
  }
  // pricing
  if (line.startsWith('pricing{')) {
    const body=line.slice(8,line.lastIndexOf('}')).trim()
    const plans=body.split('|').map(p=>{
      const pts=p.trim().split('>').map(x=>x.trim())
      return{name:pts[0],price:pts[1],desc:pts[2],linkRaw:pts[3]}
    }).filter(p=>p.name)
    return{kind:'pricing',plans}
  }
  // faq
  if (line.startsWith('faq{')) {
    const body=line.slice(4,line.lastIndexOf('}')).trim()
    const items=body.split('|').map(i=>{
      const idx=i.indexOf('>')
      return{q:i.slice(0,idx).trim(),a:i.slice(idx+1).trim()}
    }).filter(i=>i.q&&i.a)
    return{kind:'faq',items}
  }
  // testimonial{Name, Title|"Quote"|img:url}
  if (line.startsWith('testimonial{')) {
    const body=line.slice(12,line.lastIndexOf('}')).trim()
    const parts=body.split('|').map(x=>x.trim())
    const imgPart=parts.find(p=>p.startsWith('img:'))
    return{kind:'testimonial',author:parts[0],quote:parts[1]?.replace(/^"|"$/g,''),img:imgPart?.slice(4)||null}
  }
  // gallery
  if (line.startsWith('gallery{')) {
    const body=line.slice(8,line.lastIndexOf('}')).trim()
    const imgs=body.split('|').map(x=>x.trim()).filter(Boolean)
    return{kind:'gallery',imgs}
  }
  // btn
  if (line.startsWith('btn{')) {
    const parts=line.slice(4,line.lastIndexOf('}')).split('>').map(p=>p.trim())
    const label=parts[0]||'Click'
    const method=parts[1]?.split(' ')[0]||'POST'
    const bpath=parts[1]?.split(' ').slice(1).join(' ')||'#'
    const confirm=parts.find(p=>p.startsWith('confirm:'))?.slice(8)||null
    const action=parts.find(p=>!p.startsWith('confirm:')&&p!==parts[0]&&p!==parts[1])||''
    return{kind:'btn',label,method,bpath,action,confirm}
  }
  // select
  if (line.startsWith('select ')) {
    const bi=line.indexOf('{')
    const varName=bi!==-1?line.slice(7,bi).trim():line.slice(7).trim()
    const body=bi!==-1?line.slice(bi+1,line.lastIndexOf('}')).trim():''
    return{kind:'select',binding:varName,options:body.split('|').map(o=>o.trim()).filter(Boolean)}
  }
  // if
  if (line.startsWith('if ')) {
    const bi=line.indexOf('{'); if(bi===-1) return null
    return{kind:'if',cond:line.slice(3,bi).trim(),inner:line.slice(bi+1,line.lastIndexOf('}')).trim()}
  }
  // regular
  const bi=line.indexOf('{'); if(bi===-1) return null
  const head=line.slice(0,bi).trim()
  const body=line.slice(bi+1,line.lastIndexOf('}')).trim()
  const m=head.match(/^([a-z]+)(\d+)$/)
  return{kind:m?m[1]:head,cols:m?parseInt(m[2]):3,items:parseItems(body)}
}

function parseItems(body) {
  return body.split('|').map(raw=>{
    raw=raw.trim(); if(!raw) return null
    return raw.split('>').map(f=>{
      f=f.trim()
      if(f.startsWith('img:')) return{isImg:true,src:f.slice(4)}
      if(f.startsWith('/')) { const[p,l]=f.split(':'); return{isLink:true,path:(p||'').trim(),label:(l||'').trim()} }
      return{isLink:false,text:f}
    })
  }).filter(Boolean)
}
function parseCols(s){return s.split('|').map(c=>{c=c.trim();if(c.startsWith('empty:')||!c)return null;const[l,k]=c.split(':').map(x=>x.trim());return k?{label:l,key:k}:null}).filter(Boolean)}
function parseEmpty(s){const m=s.match(/empty:\s*([^|]+)/);return m?m[1].trim():'No data.'}
function parseFields(s){return s.split('|').map(f=>{const[label,type,ph]=f.split(':').map(x=>x.trim());return label?{label,type:type||'text',placeholder:ph||'',name:label.toLowerCase().replace(/\s+/g,'_')}:null}).filter(Boolean)}

// ═══════════════════════════════════════════════════════════════
// RENDERER
// ═══════════════════════════════════════════════════════════════

function renderPage(page, allPages) {
  const needsJS = page.queries.length>0 ||
    page.blocks.some(b=>['table','list','form','if','btn','select','faq'].includes(b.kind))
  const body    = page.blocks.map(b=>renderBlock(b,page)).join('')
  const config  = needsJS?JSON.stringify({id:page.id,theme:page.theme,routes:allPages.map(p=>p.route),state:page.state,queries:page.queries}):''
  const hydrate = needsJS?`\n<script>window.__FLUX_PAGE__=${config};</script>\n<script src="./flux-hydrate.js" defer></script>`:''
  const customVars = page.customTheme ? genCustomThemeVars(page.customTheme) : ''

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>${esc(page.id.charAt(0).toUpperCase()+page.id.slice(1))}</title>
<link rel="canonical" href="${esc(page.route)}">
<meta name="robots" content="index,follow">
<style>${css(page.theme)}${customVars}</style>
</head>
<body>
${body}${hydrate}
</body>
</html>`
}

function genCustomThemeVars(ct) {
  return `
body{background:${ct.bg};color:${ct.text}}
.fx-nav{background:${ct.bg}cc;border-bottom:1px solid ${ct.text}18}
.fx-cta{background:${ct.accent};color:#fff}
.fx-btn{background:${ct.accent};color:#fff}
.fx-card{background:${ct.surface||ct.bg};border:1px solid ${ct.text}15}
.fx-form{background:${ct.surface||ct.bg};border:1px solid ${ct.text}15}
.fx-input{background:${ct.bg};border:1px solid ${ct.text}30;color:${ct.text}}
.fx-stat-lbl,.fx-card-body,.fx-sub,.fx-sect-body,.fx-footer-text{color:${ct.text}88}
.fx-th,.fx-nav-link{color:${ct.text}77}
.fx-footer{border-top:1px solid ${ct.text}15}
.fx-th{border-bottom:1px solid ${ct.text}15}`
}

function renderBlock(b, page) {
  switch(b.kind) {
    case 'nav':         return rNav(b)
    case 'hero':        return rHero(b)
    case 'stats':       return rStats(b)
    case 'row':         return rRow(b)
    case 'sect':        return rSect(b)
    case 'foot':        return rFoot(b)
    case 'table':       return rTable(b)
    case 'form':        return rForm(b)
    case 'btn':         return rBtn(b)
    case 'select':      return rSelectBlock(b)
    case 'pricing':     return rPricing(b)
    case 'faq':         return rFaq(b)
    case 'testimonial': return rTestimonial(b)
    case 'gallery':     return rGallery(b)
    case 'if':          return `<div class="fx-if-wrap" data-fx-if="${esc(b.cond)}" style="display:none"></div>\n`
    default: return ''
  }
}

function rNav(b) {
  if(!b.items?.[0]) return ''
  const it=b.items[0]
  const brand=!it[0]?.isLink?`<span class="fx-brand">${esc(it[0].text)}</span>`:''
  const start=!it[0]?.isLink?1:0
  const links=it.slice(start).filter(f=>f.isLink)
    .map(f=>`<a href="${esc(f.path)}" class="fx-nav-link">${esc(f.label)}</a>`).join('')
  // hamburger for mobile
  return `<nav class="fx-nav">
  ${brand}
  <button class="fx-hamburger" onclick="this.classList.toggle('open');document.querySelector('.fx-nav-links').classList.toggle('open')" aria-label="Menu">
    <span></span><span></span><span></span>
  </button>
  <div class="fx-nav-links">${links}</div>
</nav>\n`
}

function rHero(b) {
  let h1='',sub='',img='',ctas=''
  for(const item of b.items) for(const f of item){
    if(f.isImg) img=`<img src="${esc(f.src)}" class="fx-hero-img" alt="hero" loading="eager">`
    else if(f.isLink) ctas+=`<a href="${esc(f.path)}" class="fx-cta">${esc(f.label)}</a>`
    else if(!h1){ h1=`<h1 class="fx-title">${esc(f.text)}</h1>` }
    else sub+=`<p class="fx-sub">${esc(f.text)}</p>`
  }
  const hasImg = img !== ''
  return `<section class="fx-hero${hasImg?' fx-hero-split':''}">
  <div class="fx-hero-inner">${h1}${sub}${ctas}</div>
  ${img}
</section>\n`
}

function rStats(b) {
  const cells=b.items.map(item=>{
    const[val,lbl]=(item[0]?.text||'').split(':')
    const bind=isDyn(val?.trim())?` data-fx-bind="${esc(val.trim())}"`:'  '
    return`<div class="fx-stat"><div class="fx-stat-val"${bind}>${esc(val?.trim())}</div><div class="fx-stat-lbl">${esc(lbl?.trim())}</div></div>`
  }).join('')
  return `<div class="fx-stats">${cells}</div>\n`
}

function rRow(b) {
  const cards=b.items.map(item=>{
    const inner=item.map((f,fi)=>{
      if(f.isImg) return `<img src="${esc(f.src)}" class="fx-card-img" alt="" loading="lazy">`
      if(f.isLink) return `<a href="${esc(f.path)}" class="fx-card-link">${esc(f.label)} →</a>`
      if(fi===0) return `<div class="fx-icon">${ic(f.text)}</div>`
      if(fi===1) return `<h3 class="fx-card-title">${esc(f.text)}</h3>`
      return `<p class="fx-card-body">${esc(f.text)}</p>`
    }).join('')
    return `<div class="fx-card">${inner}</div>`
  }).join('')
  return `<div class="fx-grid fx-grid-${b.cols||3}">${cards}</div>\n`
}

function rSect(b) {
  let inner=''
  b.items.forEach((item,ii)=>item.forEach(f=>{
    if(f.isLink) inner+=`<a href="${esc(f.path)}" class="fx-sect-link">${esc(f.label)}</a>`
    else if(ii===0) inner+=`<h2 class="fx-sect-title">${esc(f.text)}</h2>`
    else inner+=`<p class="fx-sect-body">${esc(f.text)}</p>`
  }))
  return `<section class="fx-sect">${inner}</section>\n`
}

function rFoot(b) {
  let inner=''
  for(const item of b.items) for(const f of item){
    if(f.isLink) inner+=`<a href="${esc(f.path)}" class="fx-footer-link">${esc(f.label)}</a>`
    else inner+=`<p class="fx-footer-text">${esc(f.text)}</p>`
  }
  return `<footer class="fx-footer">${inner}</footer>\n`
}

function rTable(b) {
  const ths=b.cols.map(c=>`<th class="fx-th">${esc(c.label)}</th>`).join('')
  const keys=JSON.stringify(b.cols.map(c=>c.key))
  const colMap=JSON.stringify(b.cols.map(c=>({label:c.label,key:c.key})))
  const editAttr=b.editPath?` data-fx-edit="${esc(b.editPath)}" data-fx-edit-method="${esc(b.editMethod)}"`:''
  const delAttr=b.deletePath?` data-fx-delete="${esc(b.deletePath)}"`:''
  const actTh=(b.editPath||b.deletePath)?'<th class="fx-th fx-th-actions">Actions</th>':''
  const span=b.cols.length+((b.editPath||b.deletePath)?1:0)
  return `<div class="fx-table-wrap"><table class="fx-table" data-fx-table="${esc(b.binding)}" data-fx-cols='${keys}' data-fx-col-map='${colMap}'${editAttr}${delAttr}><thead><tr>${ths}${actTh}</tr></thead><tbody class="fx-tbody"><tr><td colspan="${span}" class="fx-td-empty">${esc(b.empty)}</td></tr></tbody></table></div>\n`
}

function rForm(b) {
  const fields=b.fields.map(f=>{
    const inp=f.type==='select'
      ?`<select class="fx-input" name="${esc(f.name)}"><option value="">Select...</option></select>`
      :`<input class="fx-input" type="${esc(f.type)}" name="${esc(f.name)}" placeholder="${esc(f.placeholder)}">`
    return`<div class="fx-field"><label class="fx-label">${esc(f.label)}</label>${inp}</div>`
  }).join('')
  return `<div class="fx-form-wrap"><form class="fx-form" data-fx-form="${esc(b.bpath)}" data-fx-method="${esc(b.method)}" data-fx-action="${esc(b.action)}">${fields}<div class="fx-form-msg"></div><button type="submit" class="fx-btn">Submit</button></form></div>\n`
}

function rBtn(b) {
  const ca=b.confirm?` data-fx-confirm="${esc(b.confirm)}"`:''
  const aa=b.action?` data-fx-action="${esc(b.action)}"`:''
  return `<div class="fx-btn-wrap"><button class="fx-btn fx-standalone-btn" data-fx-btn="${esc(b.bpath)}" data-fx-method="${esc(b.method)}"${aa}${ca}>${esc(b.label)}</button></div>\n`
}

function rSelectBlock(b) {
  const opts=b.options.map(o=>`<option value="${esc(o)}">${esc(o)}</option>`).join('')
  return `<div class="fx-select-wrap"><select class="fx-input fx-select-block" data-fx-model="${esc(b.binding)}">${opts}</select></div>\n`
}

function rPricing(b) {
  const cards=b.plans.map((plan,i)=>{
    let linkHref='#',linkLabel='Get started'
    if(plan.linkRaw){
      const m=plan.linkRaw.match(/\/([^:]+):(.+)/)
      if(m){linkHref='/'+m[1];linkLabel=m[2]}
    }
    const featured=i===1?' fx-pricing-featured':''
    return `<div class="fx-pricing-card${featured}">
      ${i===1?'<div class="fx-pricing-badge">Most popular</div>':''}
      <div class="fx-pricing-name">${esc(plan.name)}</div>
      <div class="fx-pricing-price">${esc(plan.price)}</div>
      <p class="fx-pricing-desc">${esc(plan.desc)}</p>
      <a href="${esc(linkHref)}" class="fx-cta fx-pricing-cta">${esc(linkLabel)}</a>
    </div>`
  }).join('')
  return `<div class="fx-pricing">${cards}</div>\n`
}

function rFaq(b) {
  const items=b.items.map(item=>`
    <div class="fx-faq-item" onclick="this.classList.toggle('open')">
      <div class="fx-faq-q">${esc(item.q)}<span class="fx-faq-arrow">▸</span></div>
      <div class="fx-faq-a">${esc(item.a)}</div>
    </div>`).join('')
  return `<section class="fx-sect"><div class="fx-faq">${items}</div></section>\n`
}

function rTestimonial(b) {
  const img=b.img?`<img src="${esc(b.img)}" class="fx-testi-img" alt="${esc(b.author)}" loading="lazy">`:'<div class="fx-testi-avatar">${b.author?.charAt(0)||\'?\'}</div>'
  return `<section class="fx-testi-wrap">
  <div class="fx-testi">
    ${img}
    <blockquote class="fx-testi-quote">"${esc(b.quote)}"</blockquote>
    <div class="fx-testi-author">${esc(b.author)}</div>
  </div>
</section>\n`
}

function rGallery(b) {
  const imgs=b.imgs.map(src=>`<div class="fx-gallery-item"><img src="${esc(src)}" alt="" loading="lazy"></div>`).join('')
  return `<div class="fx-gallery">${imgs}</div>\n`
}

// ═══════════════════════════════════════════════════════════════
// CSS
// ═══════════════════════════════════════════════════════════════

function css(theme) {
  const base=`*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}html{scroll-behavior:smooth}body{font-family:-apple-system,'Segoe UI',system-ui,sans-serif;-webkit-font-smoothing:antialiased;min-height:100vh}a{text-decoration:none;color:inherit}input,button,select{font-family:inherit}img{max-width:100%;height:auto}
.fx-nav{display:flex;align-items:center;justify-content:space-between;padding:1rem 2.5rem;position:sticky;top:0;z-index:50;backdrop-filter:blur(12px);flex-wrap:wrap;gap:.5rem}
.fx-brand{font-size:1.25rem;font-weight:800;letter-spacing:-.03em}
.fx-nav-links{display:flex;align-items:center;gap:1.75rem}
.fx-nav-link{font-size:.875rem;font-weight:500;opacity:.65;transition:opacity .15s}.fx-nav-link:hover{opacity:1}
.fx-hamburger{display:none;flex-direction:column;gap:5px;background:none;border:none;cursor:pointer;padding:.25rem}
.fx-hamburger span{display:block;width:22px;height:2px;background:currentColor;transition:all .2s;border-radius:1px}
.fx-hamburger.open span:nth-child(1){transform:rotate(45deg) translate(5px,5px)}
.fx-hamburger.open span:nth-child(2){opacity:0}
.fx-hamburger.open span:nth-child(3){transform:rotate(-45deg) translate(5px,-5px)}
@media(max-width:640px){
  .fx-hamburger{display:flex}
  .fx-nav-links{display:none;width:100%;flex-direction:column;align-items:flex-start;gap:.75rem;padding:.75rem 0}
  .fx-nav-links.open{display:flex}
}
.fx-hero{display:flex;align-items:center;justify-content:center;min-height:92vh;padding:4rem 1.5rem}
.fx-hero-split{display:grid;grid-template-columns:1fr 1fr;gap:3rem;align-items:center;padding:4rem 2.5rem;min-height:70vh}
@media(max-width:768px){.fx-hero-split{grid-template-columns:1fr;min-height:auto}}
.fx-hero-img{width:100%;border-radius:1.25rem;object-fit:cover;max-height:500px}
.fx-hero-inner{max-width:56rem;text-align:center;display:flex;flex-direction:column;align-items:center;gap:1.5rem}
.fx-hero-split .fx-hero-inner{text-align:left;align-items:flex-start;max-width:none}
.fx-title{font-size:clamp(2.5rem,8vw,5.5rem);font-weight:900;letter-spacing:-.04em;line-height:1}
.fx-sub{font-size:clamp(1rem,2vw,1.25rem);line-height:1.75;max-width:40rem}
.fx-cta{display:inline-flex;align-items:center;padding:.875rem 2.5rem;border-radius:.75rem;font-weight:700;font-size:1rem;letter-spacing:-.01em;transition:transform .15s;margin:.25rem}.fx-cta:hover{transform:translateY(-1px)}
.fx-stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:3rem;padding:5rem 2.5rem;text-align:center}
.fx-stat-val{font-size:clamp(2.5rem,5vw,4rem);font-weight:900;letter-spacing:-.04em;line-height:1}
.fx-stat-lbl{font-size:.75rem;font-weight:600;text-transform:uppercase;letter-spacing:.1em;margin-top:.5rem}
.fx-grid{display:grid;gap:1.25rem;padding:1rem 2.5rem 5rem}
.fx-grid-2{grid-template-columns:repeat(auto-fit,minmax(280px,1fr))}
.fx-grid-3{grid-template-columns:repeat(auto-fit,minmax(240px,1fr))}
.fx-grid-4{grid-template-columns:repeat(auto-fit,minmax(200px,1fr))}
.fx-card{border-radius:1rem;padding:1.75rem;transition:transform .2s,box-shadow .2s}.fx-card:hover{transform:translateY(-2px)}
.fx-card-img{width:100%;border-radius:.75rem;object-fit:cover;height:180px;margin-bottom:1rem}
.fx-icon{font-size:2rem;margin-bottom:1rem}
.fx-card-title{font-size:1.0625rem;font-weight:700;letter-spacing:-.02em;margin-bottom:.5rem}
.fx-card-body{font-size:.875rem;line-height:1.65}
.fx-card-link{font-size:.8125rem;font-weight:600;display:inline-block;margin-top:1rem;opacity:.6;transition:opacity .15s}.fx-card-link:hover{opacity:1}
.fx-sect{padding:5rem 2.5rem}
.fx-sect-title{font-size:clamp(1.75rem,4vw,3rem);font-weight:800;letter-spacing:-.04em;margin-bottom:1.5rem;text-align:center}
.fx-sect-body{font-size:1rem;line-height:1.75;text-align:center;max-width:48rem;margin:0 auto}
.fx-form-wrap{padding:3rem 2.5rem;display:flex;justify-content:center}
.fx-form{width:100%;max-width:28rem;border-radius:1.25rem;padding:2.5rem}
.fx-field{margin-bottom:1.25rem}
.fx-label{display:block;font-size:.8125rem;font-weight:600;margin-bottom:.5rem}
.fx-input{width:100%;padding:.75rem 1rem;border-radius:.625rem;font-size:.9375rem;outline:none;transition:box-shadow .15s}.fx-input:focus{box-shadow:0 0 0 3px rgba(37,99,235,.35)}
.fx-btn{width:100%;padding:.875rem 1.5rem;border:none;border-radius:.625rem;font-size:.9375rem;font-weight:700;cursor:pointer;margin-top:.5rem;transition:transform .15s,opacity .15s;letter-spacing:-.01em}.fx-btn:hover{transform:translateY(-1px)}.fx-btn:disabled{opacity:.5;cursor:not-allowed;transform:none}
.fx-btn-wrap{padding:0 2.5rem 1.5rem}.fx-standalone-btn{width:auto;padding:.75rem 2rem;margin-top:0}
.fx-form-msg{font-size:.8125rem;padding:.5rem 0;min-height:1.5rem;text-align:center}.fx-form-err{color:#f87171}.fx-form-ok{color:#4ade80}
.fx-table-wrap{overflow-x:auto;padding:0 2.5rem 4rem}
.fx-table{width:100%;border-collapse:collapse;font-size:.875rem}
.fx-th{text-align:left;padding:.875rem 1.25rem;font-size:.75rem;font-weight:700;text-transform:uppercase;letter-spacing:.06em}.fx-th-actions{opacity:.6}
.fx-tr{transition:background .1s}.fx-td{padding:.875rem 1.25rem}.fx-td-empty{padding:2rem 1.25rem;text-align:center;opacity:.4}.fx-td-actions{white-space:nowrap;padding:.5rem 1rem !important}
.fx-action-btn{border:none;cursor:pointer;font-size:.75rem;font-weight:600;padding:.3rem .75rem;border-radius:.375rem;margin-right:.375rem;font-family:inherit;transition:opacity .15s,transform .1s}.fx-action-btn:hover{opacity:.85;transform:translateY(-1px)}.fx-edit-btn{background:#1e40af;color:#93c5fd}.fx-delete-btn{background:#7f1d1d;color:#fca5a5}
.fx-select-wrap{padding:.5rem 2.5rem}.fx-select-block{width:auto;min-width:200px;margin-top:0}
.fx-pricing{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:1.5rem;padding:2rem 2.5rem 5rem;align-items:start}
.fx-pricing-card{border-radius:1.25rem;padding:2rem;position:relative;transition:transform .2s}
.fx-pricing-featured{transform:scale(1.03)}
.fx-pricing-badge{position:absolute;top:-12px;left:50%;transform:translateX(-50%);background:#2563eb;color:#fff;font-size:.7rem;font-weight:700;padding:.25rem .875rem;border-radius:999px;white-space:nowrap;letter-spacing:.05em}
.fx-pricing-name{font-size:.875rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;margin-bottom:.5rem;opacity:.7}
.fx-pricing-price{font-size:3rem;font-weight:900;letter-spacing:-.05em;line-height:1;margin-bottom:.75rem}
.fx-pricing-desc{font-size:.875rem;line-height:1.65;margin-bottom:1.5rem;opacity:.7}
.fx-pricing-cta{display:block;text-align:center;padding:.75rem;border-radius:.625rem;font-weight:700;font-size:.9rem;transition:opacity .15s}.fx-pricing-cta:hover{opacity:.85}
.fx-faq{max-width:48rem;margin:0 auto}
.fx-faq-item{border-radius:.75rem;margin-bottom:.625rem;cursor:pointer;overflow:hidden;transition:background .15s}
.fx-faq-q{display:flex;justify-content:space-between;align-items:center;padding:1rem 1.25rem;font-size:.9375rem;font-weight:600}
.fx-faq-arrow{transition:transform .2s;font-size:.75rem;opacity:.5}
.fx-faq-item.open .fx-faq-arrow{transform:rotate(90deg)}
.fx-faq-a{max-height:0;overflow:hidden;padding:0 1.25rem;font-size:.875rem;line-height:1.7;transition:max-height .3s ease,padding .3s}
.fx-faq-item.open .fx-faq-a{max-height:300px;padding:.75rem 1.25rem 1.25rem}
.fx-testi-wrap{padding:5rem 2.5rem;display:flex;justify-content:center}
.fx-testi{max-width:42rem;text-align:center;display:flex;flex-direction:column;align-items:center;gap:1.25rem}
.fx-testi-img{width:64px;height:64px;border-radius:50%;object-fit:cover}
.fx-testi-avatar{width:64px;height:64px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:1.5rem;font-weight:700;background:#1e293b}
.fx-testi-quote{font-size:1.25rem;line-height:1.7;font-style:italic;opacity:.9}
.fx-testi-author{font-size:.875rem;font-weight:600;opacity:.5}
.fx-gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:.75rem;padding:1rem 2.5rem 4rem}
.fx-gallery-item{border-radius:.75rem;overflow:hidden;aspect-ratio:4/3}
.fx-gallery-item img{width:100%;height:100%;object-fit:cover;transition:transform .3s}.fx-gallery-item:hover img{transform:scale(1.04)}
.fx-if-wrap{display:contents}
.fx-footer{padding:3rem 2.5rem;text-align:center}
.fx-footer-text{font-size:.8125rem}.fx-footer-link{font-size:.8125rem;margin:0 .75rem;opacity:.5;transition:opacity .15s}.fx-footer-link:hover{opacity:1}`

  const T={
    dark: `body{background:#030712;color:#f1f5f9}.fx-nav{border-bottom:1px solid #1e293b;background:rgba(3,7,18,.85)}.fx-nav-link{color:#cbd5e1}.fx-sub{color:#94a3b8}.fx-cta{background:#2563eb;color:#fff;box-shadow:0 8px 24px rgba(37,99,235,.35)}.fx-stat-lbl{color:#64748b}.fx-card{background:#0f172a;border:1px solid #1e293b}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.5)}.fx-card-body{color:#64748b}.fx-sect-body{color:#64748b}.fx-form{background:#0f172a;border:1px solid #1e293b}.fx-label{color:#94a3b8}.fx-input{background:#020617;border:1px solid #1e293b;color:#f1f5f9}.fx-input::placeholder{color:#334155}.fx-btn{background:#2563eb;color:#fff;box-shadow:0 4px 14px rgba(37,99,235,.4)}.fx-th{color:#475569;border-bottom:1px solid #1e293b}.fx-tr:hover{background:#0f172a}.fx-td{border-bottom:1px solid rgba(255,255,255,.03)}.fx-footer{border-top:1px solid #1e293b}.fx-footer-text{color:#334155}.fx-pricing-card{background:#0f172a;border:1px solid #1e293b}.fx-pricing-featured{border-color:#2563eb}.fx-faq-item{background:#0f172a}.fx-faq-item:hover{background:#111827}.fx-faq-q{color:#f1f5f9}`,
    light:`body{background:#fff;color:#0f172a}.fx-nav{border-bottom:1px solid #e2e8f0;background:rgba(255,255,255,.85)}.fx-nav-link{color:#475569}.fx-sub{color:#475569}.fx-cta{background:#2563eb;color:#fff}.fx-stat-lbl{color:#94a3b8}.fx-card{background:#f8fafc;border:1px solid #e2e8f0}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.08)}.fx-card-body{color:#475569}.fx-sect-body{color:#475569}.fx-form{background:#f8fafc;border:1px solid #e2e8f0}.fx-label{color:#475569}.fx-input{background:#fff;border:1px solid #cbd5e1;color:#0f172a}.fx-btn{background:#2563eb;color:#fff}.fx-th{color:#94a3b8;border-bottom:1px solid #e2e8f0}.fx-tr:hover{background:#f8fafc}.fx-footer{border-top:1px solid #e2e8f0}.fx-footer-text{color:#94a3b8}.fx-pricing-card{background:#f8fafc;border:1px solid #e2e8f0}.fx-pricing-featured{border-color:#2563eb}.fx-faq-item{background:#f8fafc}.fx-faq-item:hover{background:#f1f5f9}`,
    acid: `body{background:#000;color:#a3e635}.fx-nav{border-bottom:1px solid #1a2e05;background:rgba(0,0,0,.9)}.fx-nav-link{color:#86efac}.fx-sub{color:#4d7c0f}.fx-cta{background:#a3e635;color:#000;font-weight:800}.fx-stat-lbl{color:#365314}.fx-card{background:#0a0f00;border:1px solid #1a2e05}.fx-card-body{color:#365314}.fx-sect-body{color:#365314}.fx-form{background:#0a0f00;border:1px solid #1a2e05}.fx-label{color:#4d7c0f}.fx-input{background:#000;border:1px solid #1a2e05;color:#a3e635}.fx-btn{background:#a3e635;color:#000;font-weight:800}.fx-th{color:#365314;border-bottom:1px solid #1a2e05}.fx-footer{border-top:1px solid #1a2e05}.fx-footer-text{color:#1a2e05}.fx-pricing-card{background:#0a0f00;border:1px solid #1a2e05}.fx-faq-item{background:#0a0f00}`,
  }
  return base+(T[theme]||T.dark)
}
