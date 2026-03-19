// package fluxcompiler — FLUX SSG compiler
// Parses .flux source → pre-rendered HTML
// Static blocks → pure HTML (Google indexes, zero JS)
// Dynamic blocks → HTML skeleton + data-fx-* attributes (runtime hydrates)

package fluxcompiler

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────
// AST
// ──────────────────────────────────────────────

type Page struct {
	ID      string
	Theme   string
	Route   string
	State   map[string]string
	Queries []Query
	Blocks  []Block
}

type Query struct {
	Trigger  string
	Interval int
	Method   string
	Path     string
	Target   string
	Action   string
}

type Block struct {
	Kind    string
	Cols    int
	Items   []Item
	Binding string
	ColDefs []ColDef
	Empty   string
	Method  string
	BPath   string
	Action  string
	Fields  []FormField
	Cond    string
	Inner   string
}

type Item   []Field
type Field  struct{ IsLink bool; Text, Path, Label string }
type ColDef struct{ Label, Key string }
type FormField struct {
	Label, Type, Placeholder, Name string
}

// ──────────────────────────────────────────────
// PARSER
// ──────────────────────────────────────────────

func ParseFlux(src string) []Page {
	var pages []Page
	for _, section := range strings.Split(src, "\n---\n") {
		if p := parsePage(strings.TrimSpace(section)); p != nil {
			pages = append(pages, *p)
		}
	}
	return pages
}

func parsePage(src string) *Page {
	var lines []string
	for _, l := range strings.Split(src, "\n") {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	p := &Page{ID: "page", Theme: "dark", Route: "/", State: map[string]string{}}
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "%"):
			parts := strings.Fields(line[1:])
			if len(parts) > 0 { p.ID = parts[0] }
			if len(parts) > 1 { p.Theme = parts[1] }
			if len(parts) > 2 { p.Route = parts[2] }

		case strings.HasPrefix(line, "@") && strings.Contains(line, "="):
			eq := strings.Index(line, "=")
			p.State[strings.TrimSpace(line[1:eq])] = strings.TrimSpace(line[eq+1:])

		case strings.HasPrefix(line, "$") && strings.Contains(line, "="):
			// computed — skip for SSG, handled by hydration runtime

		case strings.HasPrefix(line, "~"):
			if q := parseQuery(line[1:]); q != nil {
				p.Queries = append(p.Queries, *q)
			}

		default:
			if b := parseBlock(line); b != nil {
				p.Blocks = append(p.Blocks, *b)
			}
		}
	}
	return p
}

func parseQuery(s string) *Query {
	parts := strings.Fields(s)
	if len(parts) < 3 {
		return nil
	}
	ai := -1
	for i, p := range parts {
		if p == "=>" { ai = i; break }
	}
	q := &Query{}
	switch parts[0] {
	case "mount":
		q.Trigger = "mount"
		q.Method = parts[1]; q.Path = parts[2]
		if ai != -1 { q.Action = strings.Join(parts[ai+1:], " ") } else if len(parts) > 3 { q.Target = parts[3] }
	case "interval":
		q.Trigger = "interval"
		fmt.Sscanf(parts[1], "%d", &q.Interval)
		if len(parts) > 2 { q.Method = parts[2] }
		if len(parts) > 3 { q.Path = parts[3] }
		if ai != -1 { q.Action = strings.Join(parts[ai+1:], " ") } else if len(parts) > 4 { q.Target = parts[4] }
	}
	return q
}

func parseBlock(line string) *Block {
	// table @var { ... }
	if strings.HasPrefix(line, "table ") {
		idx := strings.Index(line[6:], "{")
		if idx == -1 { return nil }
		binding := strings.TrimSpace(line[6 : 6+idx])
		content := strings.TrimSuffix(strings.TrimSpace(line[6+idx+1:]), "}")
		return &Block{Kind: "table", Binding: binding, ColDefs: parseColDefs(content), Empty: parseEmpty(content)}
	}
	// list @var { ... }
	if strings.HasPrefix(line, "list ") {
		idx := strings.Index(line[5:], "{")
		if idx == -1 { return nil }
		binding := strings.TrimSpace(line[5 : 5+idx])
		content := strings.TrimSuffix(strings.TrimSpace(line[5+idx+1:]), "}")
		return &Block{Kind: "list", Binding: binding, ColDefs: parseColDefs(content)}
	}
	// form METHOD /path => action { ... }
	if strings.HasPrefix(line, "form ") {
		bi := strings.Index(line, "{")
		if bi == -1 { return nil }
		head := strings.TrimSpace(line[5:bi])
		content := strings.TrimSuffix(strings.TrimSpace(line[bi+1:]), "}")
		method, path, action := "", "", ""
		ai := strings.Index(head, "=>")
		if ai != -1 { action = strings.TrimSpace(head[ai+2:]); head = strings.TrimSpace(head[:ai]) }
		parts := strings.Fields(head)
		if len(parts) > 0 { method = parts[0] }
		if len(parts) > 1 { path = parts[1] }
		return &Block{Kind: "form", Method: method, BPath: path, Action: action, Fields: parseFormFields(content)}
	}
	// if @cond { inner }
	if strings.HasPrefix(line, "if ") {
		bi := strings.Index(line, "{")
		if bi == -1 { return nil }
		return &Block{Kind: "if", Cond: strings.TrimSpace(line[3:bi]), Inner: strings.TrimSuffix(strings.TrimSpace(line[bi+1:]), "}")}
	}
	// regular block: nav{...}
	bi := strings.Index(line, "{")
	if bi == -1 { return nil }
	head := strings.TrimSpace(line[:bi])
	body := strings.TrimSpace(line[bi+1 : strings.LastIndex(line, "}")])
	kind := head
	cols := 3
	for i, c := range head {
		if c >= '0' && c <= '9' { kind = head[:i]; fmt.Sscanf(head[i:], "%d", &cols); break }
	}
	return &Block{Kind: kind, Cols: cols, Items: parseItems(body)}
}

func parseItems(body string) []Item {
	var items []Item
	for _, raw := range strings.Split(body, "|") {
		raw = strings.TrimSpace(raw)
		if raw == "" { continue }
		var item Item
		for _, f := range strings.Split(raw, ">") {
			f = strings.TrimSpace(f)
			if strings.HasPrefix(f, "/") {
				parts := strings.SplitN(f, ":", 2)
				label := ""
				if len(parts) > 1 { label = strings.TrimSpace(parts[1]) }
				item = append(item, Field{IsLink: true, Path: strings.TrimSpace(parts[0]), Label: label})
			} else {
				item = append(item, Field{Text: f})
			}
		}
		if len(item) > 0 { items = append(items, item) }
	}
	return items
}

func parseColDefs(s string) []ColDef {
	var cols []ColDef
	for _, c := range strings.Split(s, "|") {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(c, "empty:") { continue }
		parts := strings.SplitN(c, ":", 2)
		if len(parts) == 2 { cols = append(cols, ColDef{Label: strings.TrimSpace(parts[0]), Key: strings.TrimSpace(parts[1])}) }
	}
	return cols
}

func parseEmpty(s string) string {
	for _, p := range strings.Split(s, "|") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "empty:") { return strings.TrimSpace(p[6:]) }
	}
	return "No data."
}

func parseFormFields(s string) []FormField {
	var fields []FormField
	for _, f := range strings.Split(s, "|") {
		parts := strings.SplitN(strings.TrimSpace(f), ":", 3)
		if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" { continue }
		label := strings.TrimSpace(parts[0])
		ftype, ph := "text", ""
		if len(parts) > 1 { ftype = strings.TrimSpace(parts[1]) }
		if len(parts) > 2 { ph = strings.TrimSpace(parts[2]) }
		fields = append(fields, FormField{
			Label: label, Type: ftype, Placeholder: ph,
			Name: strings.ToLower(strings.ReplaceAll(label, " ", "_")),
		})
	}
	return fields
}

// ──────────────────────────────────────────────
// HTML RENDERER
// ──────────────────────────────────────────────

var Icons = map[string]string{
	"bolt": "⚡", "leaf": "🌱", "map": "🗺", "chart": "📊", "lock": "🔒",
	"star": "⭐", "heart": "❤", "check": "✓", "alert": "⚠", "user": "👤",
	"car": "🚗", "money": "💰", "phone": "📱", "shield": "🛡", "fire": "🔥",
	"rocket": "🚀", "clock": "🕐", "globe": "🌐", "gear": "⚙", "pin": "📍",
	"flash": "⚡", "eye": "◉", "tag": "◈",
}

func ic(name string) string {
	if e, ok := Icons[name]; ok { return e }
	return name
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func isDynamic(s string) bool {
	return strings.Contains(s, "@") || strings.Contains(s, "$")
}

// CompileSSG renders a .flux page to a full standalone HTML file.
// Static blocks → pure HTML
// Dynamic blocks → HTML placeholders with data-fx-* for hydration
func CompileSSG(pages []Page, assetBase string) map[string]string {
	out := map[string]string{}
	for _, page := range pages {
		html := renderPage(page, pages, assetBase)
		route := page.Route
		if route == "/" {
			out["index.html"] = html
		} else {
			clean := strings.Trim(route, "/")
			out[clean+"/index.html"] = html
		}
	}
	return out
}

func renderPage(page Page, allPages []Page, assetBase string) string {
	// Determine if any dynamic hydration is needed
	needsJS := len(page.Queries) > 0
	for _, b := range page.Blocks {
		if b.Kind == "table" || b.Kind == "list" || b.Kind == "form" || b.Kind == "if" {
			needsJS = true
		}
		if !needsJS {
			for _, item := range b.Items {
				for _, f := range item {
					if isDynamic(f.Text) { needsJS = true }
				}
			}
		}
	}

	var body strings.Builder
	for _, block := range page.Blocks {
		body.WriteString(renderBlock(block, page.Theme))
	}

	// Build page config JSON for hydration runtime
	pageConfigJSON := ""
	if needsJS {
		pageConfigJSON = buildPageConfig(page, allPages)
	}

	// Hydration script (only if needed)
	hydrateScript := ""
	if needsJS {
		hydrateScript = fmt.Sprintf(`
<script>
window.__FLUX_PAGE__ = %s;
</script>
<script src="%sflux-hydrate.js" defer></script>`, pageConfigJSON, assetBase)
	}

	// CSS
	css := buildCSS(page.Theme)

	// Build canonical and alternate links for SEO
	canonicalURL := page.Route

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<link rel="canonical" href="%s">
<meta name="robots" content="index,follow">
<style>%s</style>
</head>
<body>
%s%s</body>
</html>`, esc(titleCase(page.ID)), esc(canonicalURL), css, body.String(), hydrateScript)
}

func titleCase(s string) string {
	if len(s) == 0 { return s }
	return strings.ToUpper(s[:1]) + s[1:]
}

func renderBlock(b Block, theme string) string {
	switch b.Kind {
	case "nav":    return renderNav(b, theme)
	case "hero":   return renderHero(b, theme)
	case "stats":  return renderStats(b, theme)
	case "row":    return renderRow(b, theme)
	case "sect":   return renderSect(b, theme)
	case "foot":   return renderFoot(b, theme)
	case "table":  return renderTableSSG(b, theme)  // skeleton — hydrated
	case "list":   return renderListSSG(b, theme)   // skeleton — hydrated
	case "form":   return renderFormSSG(b, theme)   // rendered — hydrated for submit
	case "if":     return renderIfSSG(b, theme)     // hidden — hydrated
	}
	return ""
}

// ── Static blocks (pure HTML, Google indexes) ──────────────────────

func renderNav(b Block, theme string) string {
	if len(b.Items) == 0 { return "" }
	item := b.Items[0]
	brand := ""
	linkStart := 0
	if len(item) > 0 && !item[0].IsLink {
		brand = esc(item[0].Text)
		linkStart = 1
	}
	var links strings.Builder
	for _, f := range item[linkStart:] {
		if f.IsLink {
			links.WriteString(fmt.Sprintf(`<a href="%s" class="fx-nav-link">%s</a>`, esc(f.Path), esc(f.Label)))
		}
	}
	return fmt.Sprintf(`<nav class="fx-nav"><span class="fx-brand">%s</span><div class="fx-nav-links">%s</div></nav>%s`, brand, links.String(), "\n")
}

func renderHero(b Block, theme string) string {
	var inner strings.Builder
	h1done := false
	for _, item := range b.Items {
		for _, f := range item {
			if f.IsLink {
				inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-cta">%s</a>`, esc(f.Path), esc(f.Label)))
			} else if !h1done {
				// h1 is critical for SEO — always static
				text := f.Text
				if isDynamic(text) {
					// Render placeholder + data attr for hydration
					inner.WriteString(fmt.Sprintf(`<h1 class="fx-title" data-fx-bind="%s">%s</h1>`, esc(text), esc(stripBindings(text))))
				} else {
					inner.WriteString(fmt.Sprintf(`<h1 class="fx-title">%s</h1>`, esc(text)))
				}
				h1done = true
			} else {
				if isDynamic(f.Text) {
					inner.WriteString(fmt.Sprintf(`<p class="fx-sub" data-fx-bind="%s">%s</p>`, esc(f.Text), esc(stripBindings(f.Text))))
				} else {
					inner.WriteString(fmt.Sprintf(`<p class="fx-sub">%s</p>`, esc(f.Text)))
				}
			}
		}
	}
	return fmt.Sprintf(`<section class="fx-hero"><div class="fx-hero-inner">%s</div></section>`+"\n", inner.String())
}

func renderStats(b Block, theme string) string {
	var cells strings.Builder
	for _, item := range b.Items {
		raw := ""
		if len(item) > 0 { raw = item[0].Text }
		parts := strings.SplitN(raw, ":", 2)
		val := strings.TrimSpace(parts[0])
		lbl := ""
		if len(parts) == 2 { lbl = strings.TrimSpace(parts[1]) }

		if isDynamic(raw) {
			// Render with placeholder text for SEO + data attr for hydration
			cells.WriteString(fmt.Sprintf(
				`<div class="fx-stat"><div class="fx-stat-val" data-fx-bind="%s">%s</div><div class="fx-stat-lbl">%s</div></div>`,
				esc(val), esc(stripBindings(val)), esc(lbl),
			))
		} else {
			cells.WriteString(fmt.Sprintf(
				`<div class="fx-stat"><div class="fx-stat-val">%s</div><div class="fx-stat-lbl">%s</div></div>`,
				esc(val), esc(lbl),
			))
		}
	}
	return fmt.Sprintf(`<div class="fx-stats">%s</div>`+"\n", cells.String())
}

func renderRow(b Block, theme string) string {
	cols := b.Cols
	if cols == 0 { cols = 3 }
	var cards strings.Builder
	for _, item := range b.Items {
		var inner strings.Builder
		for fi, f := range item {
			if f.IsLink {
				inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-card-link">%s →</a>`, esc(f.Path), esc(f.Label)))
			} else if fi == 0 {
				inner.WriteString(fmt.Sprintf(`<div class="fx-icon">%s</div>`, ic(f.Text)))
			} else if fi == 1 {
				inner.WriteString(fmt.Sprintf(`<h3 class="fx-card-title">%s</h3>`, esc(f.Text)))
			} else {
				inner.WriteString(fmt.Sprintf(`<p class="fx-card-body">%s</p>`, esc(f.Text)))
			}
		}
		cards.WriteString(fmt.Sprintf(`<div class="fx-card">%s</div>`, inner.String()))
	}
	return fmt.Sprintf(`<div class="fx-grid fx-grid-%d">%s</div>`+"\n", cols, cards.String())
}

func renderSect(b Block, theme string) string {
	var inner strings.Builder
	for i, item := range b.Items {
		for _, f := range item {
			if f.IsLink {
				inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-sect-link">%s</a>`, esc(f.Path), esc(f.Label)))
			} else if i == 0 {
				inner.WriteString(fmt.Sprintf(`<h2 class="fx-sect-title">%s</h2>`, esc(f.Text)))
			} else {
				inner.WriteString(fmt.Sprintf(`<p class="fx-sect-body">%s</p>`, esc(f.Text)))
			}
		}
	}
	return fmt.Sprintf(`<section class="fx-sect">%s</section>`+"\n", inner.String())
}

func renderFoot(b Block, theme string) string {
	var inner strings.Builder
	for _, item := range b.Items {
		for _, f := range item {
			if f.IsLink {
				inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-footer-link">%s</a>`, esc(f.Path), esc(f.Label)))
			} else {
				inner.WriteString(fmt.Sprintf(`<p class="fx-footer-text">%s</p>`, esc(f.Text)))
			}
		}
	}
	return fmt.Sprintf(`<footer class="fx-footer">%s</footer>`+"\n", inner.String())
}

// ── Dynamic blocks (skeleton HTML + data-fx-* for hydration) ───────

func renderTableSSG(b Block, theme string) string {
	// Render full table structure — tbody is empty, hydration fills it
	var headers strings.Builder
	for _, col := range b.ColDefs {
		headers.WriteString(fmt.Sprintf(`<th class="fx-th">%s</th>`, esc(col.Label)))
	}
	colKeysJSON, _ := json.Marshal(func() []string {
		keys := make([]string, len(b.ColDefs))
		for i, c := range b.ColDefs { keys[i] = c.Key }
		return keys
	}())
	return fmt.Sprintf(
		`<div class="fx-table-wrap"><table class="fx-table" data-fx-table="%s" data-fx-cols='%s'><thead><tr>%s</tr></thead><tbody class="fx-tbody"><tr><td colspan="%d" class="fx-td-empty">%s</td></tr></tbody></table></div>`+"\n",
		esc(b.Binding), string(colKeysJSON), headers.String(), len(b.ColDefs), esc(b.Empty),
	)
}

func renderListSSG(b Block, theme string) string {
	colKeysJSON, _ := json.Marshal(func() []string {
		keys := make([]string, len(b.ColDefs))
		for i, c := range b.ColDefs { keys[i] = c.Key }
		return keys
	}())
	return fmt.Sprintf(
		`<div class="fx-list-wrap" data-fx-list="%s" data-fx-cols='%s'></div>`+"\n",
		esc(b.Binding), string(colKeysJSON),
	)
}

func renderFormSSG(b Block, theme string) string {
	// Form is fully rendered — SEO sees the form labels
	// data-fx-form tells hydration runtime to handle submit
	var fields strings.Builder
	for _, f := range b.Fields {
		if f.Type == "select" {
			fields.WriteString(fmt.Sprintf(
				`<div class="fx-field"><label class="fx-label">%s</label><select class="fx-input" name="%s"><option value="">Select...</option></select></div>`,
				esc(f.Label), esc(f.Name),
			))
		} else {
			fields.WriteString(fmt.Sprintf(
				`<div class="fx-field"><label class="fx-label">%s</label><input class="fx-input" type="%s" name="%s" placeholder="%s"></div>`,
				esc(f.Label), esc(f.Type), esc(f.Name), esc(f.Placeholder),
			))
		}
	}
	return fmt.Sprintf(
		`<div class="fx-form-wrap"><form class="fx-form" data-fx-form="%s" data-fx-method="%s" data-fx-action="%s">%s<div class="fx-form-msg"></div><button type="submit" class="fx-btn">Submit</button></form></div>`+"\n",
		esc(b.BPath), esc(b.Method), esc(b.Action), fields.String(),
	)
}

func renderIfSSG(b Block, theme string) string {
	// Render hidden — hydration shows/hides based on condition
	return fmt.Sprintf(
		`<div class="fx-if-wrap" data-fx-if="%s" style="display:none"></div>`+"\n",
		esc(b.Cond),
	)
}

// stripBindings removes @var and $var from a string, leaving the rest
// Used for SSG placeholder text that Google can index
func stripBindings(s string) string {
	result := strings.ReplaceAll(s, "@", "")
	result = strings.ReplaceAll(result, "$", "")
	// Remove common state var names from placeholder
	for _, word := range []string{"stats.", "user.", "users.", "data."} {
		result = strings.ReplaceAll(result, word, "")
	}
	return strings.TrimSpace(result)
}

// ──────────────────────────────────────────────
// PAGE CONFIG JSON (for hydration runtime)
// ──────────────────────────────────────────────

func buildPageConfig(page Page, allPages []Page) string {
	type queryJSON struct {
		Trigger  string `json:"trigger"`
		Interval int    `json:"interval,omitempty"`
		Method   string `json:"method"`
		Path     string `json:"path"`
		Target   string `json:"target,omitempty"`
		Action   string `json:"action,omitempty"`
	}
	type pageJSON struct {
		ID      string            `json:"id"`
		Theme   string            `json:"theme"`
		Routes  []string          `json:"routes"`
		State   map[string]string `json:"state"`
		Queries []queryJSON       `json:"queries"`
	}

	routes := make([]string, len(allPages))
	for i, p := range allPages { routes[i] = p.Route }

	queries := make([]queryJSON, len(page.Queries))
	for i, q := range page.Queries {
		queries[i] = queryJSON{
			Trigger: q.Trigger, Interval: q.Interval,
			Method: q.Method, Path: q.Path,
			Target: q.Target, Action: q.Action,
		}
	}

	cfg := pageJSON{
		ID: page.ID, Theme: page.Theme, Routes: routes,
		State: page.State, Queries: queries,
	}
	b, _ := json.Marshal(cfg)
	return string(b)
}

// ──────────────────────────────────────────────
// CSS (inlined, theme-specific)
// ──────────────────────────────────────────────

func buildCSS(theme string) string {
	base := `*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}html{scroll-behavior:smooth}body{font-family:-apple-system,'Segoe UI',system-ui,sans-serif;-webkit-font-smoothing:antialiased;min-height:100vh}a{text-decoration:none;color:inherit}input,button,select{font-family:inherit}.fx-nav{display:flex;align-items:center;justify-content:space-between;padding:1rem 2.5rem;position:sticky;top:0;z-index:50;backdrop-filter:blur(12px)}.fx-brand{font-size:1.25rem;font-weight:800;letter-spacing:-.03em}.fx-nav-links{display:flex;align-items:center;gap:1.75rem}.fx-nav-link{font-size:.875rem;font-weight:500;opacity:.65;transition:opacity .15s}.fx-nav-link:hover{opacity:1}.fx-hero{display:flex;align-items:center;justify-content:center;min-height:92vh;padding:4rem 1.5rem}.fx-hero-inner{max-width:56rem;text-align:center;display:flex;flex-direction:column;align-items:center;gap:1.5rem}.fx-title{font-size:clamp(2.5rem,8vw,5.5rem);font-weight:900;letter-spacing:-.04em;line-height:1}.fx-sub{font-size:clamp(1rem,2vw,1.25rem);line-height:1.75;max-width:40rem}.fx-cta{display:inline-flex;align-items:center;padding:.875rem 2.5rem;border-radius:.75rem;font-weight:700;font-size:1rem;letter-spacing:-.01em;transition:transform .15s;margin:.25rem}.fx-cta:hover{transform:translateY(-1px)}.fx-stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:3rem;padding:5rem 2.5rem;text-align:center}.fx-stat-val{font-size:clamp(2.5rem,5vw,4rem);font-weight:900;letter-spacing:-.04em;line-height:1}.fx-stat-lbl{font-size:.75rem;font-weight:600;text-transform:uppercase;letter-spacing:.1em;margin-top:.5rem}.fx-grid{display:grid;gap:1.25rem;padding:1rem 2.5rem 5rem}.fx-grid-2{grid-template-columns:repeat(auto-fit,minmax(280px,1fr))}.fx-grid-3{grid-template-columns:repeat(auto-fit,minmax(240px,1fr))}.fx-grid-4{grid-template-columns:repeat(auto-fit,minmax(200px,1fr))}.fx-card{border-radius:1rem;padding:1.75rem;transition:transform .2s,box-shadow .2s}.fx-card:hover{transform:translateY(-2px)}.fx-icon{font-size:2rem;margin-bottom:1rem}.fx-card-title{font-size:1.0625rem;font-weight:700;letter-spacing:-.02em;margin-bottom:.5rem}.fx-card-body{font-size:.875rem;line-height:1.65}.fx-card-link{font-size:.8125rem;font-weight:600;display:inline-block;margin-top:1rem;opacity:.6;transition:opacity .15s}.fx-card-link:hover{opacity:1}.fx-sect{padding:5rem 2.5rem}.fx-sect-title{font-size:clamp(1.75rem,4vw,3rem);font-weight:800;letter-spacing:-.04em;margin-bottom:1.5rem;text-align:center}.fx-sect-body{font-size:1rem;line-height:1.75;text-align:center;max-width:48rem;margin:0 auto}.fx-form-wrap{padding:3rem 2.5rem;display:flex;justify-content:center}.fx-form{width:100%;max-width:28rem;border-radius:1.25rem;padding:2.5rem}.fx-field{margin-bottom:1.25rem}.fx-label{display:block;font-size:.8125rem;font-weight:600;margin-bottom:.5rem}.fx-input{width:100%;padding:.75rem 1rem;border-radius:.625rem;font-size:.9375rem;outline:none;transition:box-shadow .15s}.fx-input:focus{box-shadow:0 0 0 3px rgba(37,99,235,.35)}.fx-btn{width:100%;padding:.875rem 1.5rem;border:none;border-radius:.625rem;font-size:.9375rem;font-weight:700;cursor:pointer;margin-top:.5rem;transition:transform .15s,opacity .15s;letter-spacing:-.01em}.fx-btn:hover{transform:translateY(-1px)}.fx-btn:disabled{opacity:.5;cursor:not-allowed;transform:none}.fx-form-msg{font-size:.8125rem;padding:.5rem 0;min-height:1.5rem;text-align:center}.fx-form-err{color:#f87171}.fx-form-ok{color:#4ade80}.fx-table-wrap{overflow-x:auto;padding:0 2.5rem 4rem}.fx-table{width:100%;border-collapse:collapse;font-size:.875rem}.fx-th{text-align:left;padding:.875rem 1.25rem;font-size:.75rem;font-weight:700;text-transform:uppercase;letter-spacing:.06em}.fx-tr{transition:background .1s}.fx-td{padding:.875rem 1.25rem}.fx-td-empty{padding:2rem 1.25rem;text-align:center;opacity:.4}.fx-list-wrap{padding:1rem 2.5rem 4rem;display:flex;flex-direction:column;gap:.75rem}.fx-list-item{border-radius:.75rem;padding:1.25rem 1.5rem}.fx-list-field{font-size:.9375rem;line-height:1.5}.fx-if-wrap{display:contents}.fx-footer{padding:3rem 2.5rem;text-align:center}.fx-footer-text{font-size:.8125rem}.fx-footer-link{font-size:.8125rem;margin:0 .75rem;opacity:.5;transition:opacity .15s}.fx-footer-link:hover{opacity:1}`

	themes := map[string]string{
		"dark":  `body{background:#030712;color:#f1f5f9}.fx-nav{border-bottom:1px solid #1e293b;background:rgba(3,7,18,.85)}.fx-nav-link{color:#cbd5e1}.fx-sub{color:#94a3b8}.fx-cta{background:#2563eb;color:#fff;box-shadow:0 8px 24px rgba(37,99,235,.35)}.fx-stat-lbl{color:#64748b}.fx-card{background:#0f172a;border:1px solid #1e293b}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.5)}.fx-card-body{color:#64748b}.fx-sect-body{color:#64748b}.fx-form{background:#0f172a;border:1px solid #1e293b}.fx-label{color:#94a3b8}.fx-input{background:#020617;border:1px solid #1e293b;color:#f1f5f9}.fx-input::placeholder{color:#334155}.fx-btn{background:#2563eb;color:#fff;box-shadow:0 4px 14px rgba(37,99,235,.4)}.fx-th{color:#475569;border-bottom:1px solid #1e293b}.fx-tr:hover{background:#0f172a}.fx-td{border-bottom:1px solid rgba(255,255,255,.03)}.fx-footer{border-top:1px solid #1e293b}.fx-footer-text{color:#334155}`,
		"light": `body{background:#fff;color:#0f172a}.fx-nav{border-bottom:1px solid #e2e8f0;background:rgba(255,255,255,.85)}.fx-nav-link{color:#475569}.fx-sub{color:#475569}.fx-cta{background:#2563eb;color:#fff}.fx-stat-lbl{color:#94a3b8}.fx-card{background:#f8fafc;border:1px solid #e2e8f0}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.08)}.fx-card-body{color:#475569}.fx-sect-body{color:#475569}.fx-form{background:#f8fafc;border:1px solid #e2e8f0}.fx-label{color:#475569}.fx-input{background:#fff;border:1px solid #cbd5e1;color:#0f172a}.fx-btn{background:#2563eb;color:#fff}.fx-th{color:#94a3b8;border-bottom:1px solid #e2e8f0}.fx-tr:hover{background:#f8fafc}.fx-footer{border-top:1px solid #e2e8f0}.fx-footer-text{color:#94a3b8}`,
		"acid":  `body{background:#000;color:#a3e635}.fx-nav{border-bottom:1px solid #1a2e05;background:rgba(0,0,0,.9)}.fx-nav-link{color:#86efac}.fx-sub{color:#4d7c0f}.fx-cta{background:#a3e635;color:#000;font-weight:800}.fx-stat-lbl{color:#365314}.fx-card{background:#0a0f00;border:1px solid #1a2e05}.fx-card-body{color:#365314}.fx-form{background:#0a0f00;border:1px solid #1a2e05}.fx-label{color:#4d7c0f}.fx-input{background:#000;border:1px solid #1a2e05;color:#a3e635}.fx-btn{background:#a3e635;color:#000;font-weight:800}.fx-th{color:#365314;border-bottom:1px solid #1a2e05}.fx-footer{border-top:1px solid #1a2e05}.fx-footer-text{color:#1a2e05}`,
	}

	t, ok := themes[theme]
	if !ok { t = themes["dark"] }
	return base + t
}
