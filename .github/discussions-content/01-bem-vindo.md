# 👋 Bem-vindo ao aiplang — comece por aqui

Olá! Seja bem-vindo à comunidade **aiplang** 🐴

## O que é aiplang?

aiplang é uma linguagem web criada para ser gerada por IA (Claude, GPT) — não escrita por humanos.

Um único arquivo `.aip` = app completo com frontend + backend + banco de dados + auth.

```aip
~db sqlite ./app.db
~auth jwt $JWT_SECRET expire=7d

model User {
  id       : uuid : pk auto
  email    : text : required unique
  password : text : required hashed
}

api POST /api/auth/login {
  $user = User.findBy(email=$body.email)
  ~check password $body.password $user.password | 401
  return jwt($user) 200
}

%home dark /
~theme accent=#6366f1 radius=1rem font=Inter
nav{MeuApp>/login:Entrar}
hero{Bem-vindo|Gerado por IA em 30 segundos.} animate:blur-in
foot{© 2025}
```

## Por onde começar?

```bash
# Instalar
npm install -g aiplang

# Criar projeto
npx aiplang init meu-app

# Rodar
cd meu-app && npx aiplang serve
```

## Links importantes

| | |
|---|---|
| 📦 npm | `npm install -g aiplang` |
| 📖 GitHub | https://github.com/isacamartin/aiplang |
| 🧠 Prompt Guide | Como gerar apps bons com Claude |
| 📝 Templates | /templates no repositório |
| 🔒 Segurança | SECURITY.md |

## Categorias aqui no Discussions

- **💬 General** — dúvidas gerais, conversas
- **💡 Ideas** — sugestão de features
- **🙏 Q&A** — perguntas técnicas
- **🎉 Show & Tell** — mostre seu app gerado

---

Seja bem-vindo. Se tiver dúvidas, abra um tópico em Q&A. Se quiser mostrar o que criou, vá em Show & Tell! 🚀
