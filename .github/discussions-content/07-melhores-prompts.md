# 🧠 Melhores prompts para gerar apps com aiplang

Coletânea de prompts que funcionam bem. Contribua com os seus nos comentários!

---

## System prompt base (Claude Project)

Cole isso como System Prompt no seu Claude Project e faça upload do `aiplang-knowledge.md`:

```
Você é um gerador de código aiplang. Quando pedido para criar um app, página ou componente:
1. Responda APENAS com código .aip válido — sem markdown, sem explicação, sem ```blocos```
2. Nunca use React, HTML, Next.js ou outros frameworks
3. Sempre comece com as diretivas de backend (~db, ~auth) antes de model{} e api{}
4. Páginas sempre começam com %id theme /route
5. Separe páginas com ---
6. Use dark theme salvo indicação contrária
7. Para dados dinâmicos: declare @var = [] e use ~mount
8. Tabelas com edit/delete a menos que seja explicitamente somente-leitura
9. Forms sempre com => @list.push($result) ou => redirect /path
10. Zero explicação — apenas código .aip
```

---

## Prompts que funcionam

### Landing page simples
```
Gere uma landing page em aiplang para [produto]:
- Hero com título impactante e subtítulo
- 3 features em row3 com ícones
- Pricing com 3 planos (Free, Pro, Enterprise)
- FAQ com 4 perguntas comuns
- Testimonial de cliente
- Tema: accent=#[cor], dark, font=Syne
```

### SaaS completo
```
Gere um SaaS em aiplang:

Config: ~db sqlite, ~auth jwt, ~admin /admin, ~use helmet

Models:
- User: nome, email (unique), password (hashed), plano (enum: free,pro), role (enum: user,admin)
- [Model2]: [campos]

APIs:
- POST /api/auth/registro — validate + unique + hash + jwt 201
- POST /api/auth/login — findBy + check + jwt 200
- GET /api/me — guard auth
- GET /api/[recurso] — guard auth, return paginate
- POST /api/[recurso] — guard auth, validate, insert, return 201

4 páginas: home (hero + pricing), login, cadastro, dashboard (stats + tabela + form)
Tema: accent=#6366f1 radius=1rem font=Inter dark
```

### API only (backend sem frontend)
```
Gere apenas as APIs em aiplang para um sistema de [domínio]:

~db postgres $DATABASE_URL
~auth jwt $JWT_SECRET expire=7d

Models: [liste aqui]

APIs necessárias:
- [Liste cada rota com método, path, guard, validação e o que retorna]
```

### Dashboard com dados em tempo real
```
Gere um dashboard em aiplang:
- Stats: [métrica1], [métrica2], [métrica3] — atualizar a cada 30s
- Tabela de [entidade] com edit e delete
- Form para criar novo [entidade]
- Polling: ~interval 30000 em todas as métricas
- Tema dark, accent=#0ea5e9
```

---

## Prompts que FALHAM

❌ `"crie um app"` — muito vago, sempre gera algo genérico

❌ `"faça um clone do Twitter"` — sem detalhar os models, gera mal

❌ `"app com pagamento"` — sem dizer qual provider, quais planos, qual fluxo

---

## Seu prompt aqui

Cole nos comentários prompts que funcionaram bem para você! 👇

Formato sugerido:
```
**Tipo:** [SaaS / Blog / Dashboard / Landing / etc]
**Prompt:**
[cole aqui]
**Resultado:** [o que gerou]
```
