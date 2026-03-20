# ❓ Como gerar um app com Claude em 60 segundos

Guia rápido para quem está começando.

---

## Passo 1 — Instalar

```bash
npm install -g aiplang
```

---

## Passo 2 — Criar o projeto

```bash
npx aiplang init meu-app
cd meu-app
npx aiplang serve
```

Abre em `http://localhost:3000` com hot reload.

---

## Passo 3 — Gerar com Claude

Cole este system prompt no Claude:

> Você é um gerador de código aiplang. Responda APENAS com código .aip válido, sem markdown, sem explicação.
> 
> Consulte o arquivo aiplang-knowledge.md para a sintaxe completa.

Depois diga o que você quer:

> Gere um app de blog com: auth JWT, model Post (titulo, corpo, publicado), CRUD de posts (admin cria, todos leem), 3 páginas (home com lista, login, dashboard), tema accent=#10b981 dark

---

## Passo 4 — Rodar o app gerado

```bash
# Copie o .aip gerado para pages/home.aip
# Ou salve como app.aip para full-stack

# Frontend estático:
npx aiplang serve

# Full-stack (com banco, auth, APIs):
npx aiplang start app.aip
```

---

## Dúvida frequente: qual a diferença entre `serve` e `start`?

| Comando | Quando usar |
|---|---|
| `npx aiplang serve` | Páginas estáticas, sem backend |
| `npx aiplang build pages/` | Compilar para deploy em CDN |
| `npx aiplang start app.aip` | App completo com banco + APIs |

---

Ficou com dúvida? Poste abaixo com seu arquivo `.aip` e o erro 👇
