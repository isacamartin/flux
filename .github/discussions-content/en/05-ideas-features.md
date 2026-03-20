# 💡 Most wanted features — vote here

List of the most discussed ideas. 👍 the ones you want and add yours in the comments.

---

## 🔥 Being considered for v3.0

**1. Online playground**
Paste a prompt → see the app running without installing anything. Like StackBlitz but for aiplang.

**2. Export to Next.js**
```bash
npx aiplang export --target nextjs
```
Generates an equivalent Next.js project for migration.

**3. VS Code extension on Marketplace**
Syntax highlighting + autocomplete + hover docs for `.aip` files.

---

## 💭 Backlog

**4. `~cron` for scheduled jobs**
```aip
~cron "0 9 * * *" sendReport
```

**5. `~graphql` auto endpoint**
```aip
~graphql /api/graphql
```

**6. Multi-tenant**
```aip
~multi-tenant field=organization_id
```
Auto-filters all queries by tenant.

**7. `~webhook` for incoming events**
```aip
~webhook /api/github secret=$GH_SECRET
```

**8. Customizable admin panel**
Override parts of the auto-generated admin with `.aip` blocks.

---

## How to suggest a new feature

Post in the comments with:
1. The problem it solves
2. What the `.aip` syntax would look like
3. A real use case

---

👍 vote on the ones that matter most!
