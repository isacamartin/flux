// aiplang v2 HTML Renderer
// Renders Page AST → pre-rendered HTML with hydration

package aiplangserver

import (
	"fmt"
	"strings"

	fc "github.com/isacamartin/aiplang/compiler"
)

var icons2 = map[string]string{
	"bolt":"⚡","leaf":"🌱","map":"🗺","chart":"📊","lock":"🔒","star":"⭐",
	"heart":"❤","check":"✓","alert":"⚠","user":"👤","car":"🚗","money":"💰",
	"phone":"📱","shield":"🛡","fire":"🔥","rocket":"🚀","clock":"🕐",
	"globe":"🌐","gear":"⚙","pin":"📍","flash":"⚡","eye":"◉","tag":"◈",
	"plus":"+","minus":"−","edit":"✎","trash":"🗑","search":"⌕",
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func renderPageHTML(page fc.Page, allPages []fc.Page) string {
	needsJS := len(page.Queries) > 0 || hasInteractiveBlocks(page)

	var bodyParts []string
	for _, block := range page.Blocks {
		bodyParts = append(bodyParts, renderBlockHTML(block))
	}
	body := strings.Join(bodyParts, "")

	config := ""
	hydrateTag := ""
	if needsJS {
		routes := make([]string, len(allPages))
		for i, p := range allPages { routes[i] = `"` + p.Route + `"` }
		
		stateJSON := "{"
		first := true
		for k, v := range page.State {
			if !first { stateJSON += "," }
			stateJSON += fmt.Sprintf(`"%s":%s`, k, v)
			first = false
		}
		stateJSON += "}"

		queriesJSON := buildQueriesJSON(page.Queries)
		config = fmt.Sprintf(`{"id":"%s","theme":"%s","routes":[%s],"state":%s,"queries":%s}`,
			page.ID, page.Theme, strings.Join(routes, ","), stateJSON, queriesJSON)
		hydrateTag = fmt.Sprintf("\n<script>window.__aiplang_PAGE__=%s;</script>\n<script src=\"/static/aiplang-hydrate.js\" defer></script>", config)
	}

	customCSS := ""
	if page.CustomTheme != nil {
		customCSS = genCustomCSS(page.CustomTheme)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<link rel="canonical" href="%s">
<meta name="robots" content="index,follow">
<style>%s%s</style>
</head>
<body>
%s%s
</body>
</html>`,
		htmlEsc(capitalize(page.ID)),
		htmlEsc(page.Route),
		getBaseCSS(page.Theme),
		customCSS,
		body,
		hydrateTag,
	)
}

func hasInteractiveBlocks(page fc.Page) bool {
	for _, b := range page.Blocks {
		switch b.Kind {
		case "table", "form", "if", "btn", "select", "faq":
			return true
		}
	}
	return false
}

func renderBlockHTML(b fc.Block) string {
	line := b.RawLine
	switch b.Kind {
	case "nav":         return renderNavHTML(line)
	case "hero":        return renderHeroHTML(line)
	case "stats":       return renderStatsHTML(line)
	case "row":         return renderRowHTML(line)
	case "sect":        return renderSectHTML(line)
	case "foot":        return renderFootHTML(line)
	case "table":       return renderTableHTML(line)
	case "form":        return renderFormHTML(line)
	case "pricing":     return renderPricingHTML(line)
	case "faq":         return renderFaqHTML(line)
	case "testimonial": return renderTestimonialHTML(line)
	case "gallery":     return renderGalleryHTML(line)
	case "btn":         return renderBtnHTML(line)
	case "if":          return renderIfHTML(line)
	default: return ""
	}
}

// ── Block Renderers ───────────────────────────────────────────────

func renderNavHTML(line string) string {
	body := extractBody(line)
	items := parseItemsHTML(body)
	if len(items) == 0 { return "" }
	item := items[0]
	brand := ""
	start := 0
	if len(item) > 0 && !item[0].isLink {
		brand = fmt.Sprintf(`<span class="fx-brand">%s</span>`, htmlEsc(item[0].text))
		start = 1
	}
	var links strings.Builder
	for _, f := range item[start:] {
		if f.isLink {
			links.WriteString(fmt.Sprintf(`<a href="%s" class="fx-nav-link">%s</a>`, htmlEsc(f.path), htmlEsc(f.label)))
		}
	}
	return fmt.Sprintf(`<nav class="fx-nav">%s<button class="fx-hamburger" onclick="this.classList.toggle('open');document.querySelector('.fx-nav-links').classList.toggle('open')"><span></span><span></span><span></span></button><div class="fx-nav-links">%s</div></nav>`+"\n", brand, links.String())
}

func renderHeroHTML(line string) string {
	body := extractBody(line)
	items := parseItemsHTML(body)
	var h1, sub, img, ctas string
	for _, item := range items {
		for _, f := range item {
			if f.isImg { img = fmt.Sprintf(`<img src="%s" class="fx-hero-img" alt="hero" loading="eager">`, htmlEsc(f.src)) } else if f.isLink { ctas += fmt.Sprintf(`<a href="%s" class="fx-cta">%s</a>`, htmlEsc(f.path), htmlEsc(f.label)) } else if h1 == "" { h1 = fmt.Sprintf(`<h1 class="fx-title">%s</h1>`, htmlEsc(f.text)) } else { sub += fmt.Sprintf(`<p class="fx-sub">%s</p>`, htmlEsc(f.text)) }
		}
	}
	split := ""
	if img != "" { split = " fx-hero-split" }
	return fmt.Sprintf(`<section class="fx-hero%s"><div class="fx-hero-inner">%s%s%s</div>%s</section>`+"\n", split, h1, sub, ctas, img)
}

func renderStatsHTML(line string) string {
	body := extractBody(line)
	items := parseItemsHTML(body)
	var cells strings.Builder
	for _, item := range items {
		if len(item) > 0 {
			raw := item[0].text
			parts := strings.SplitN(raw, ":", 2)
			val := strings.TrimSpace(parts[0])
			lbl := ""
			if len(parts) > 1 { lbl = strings.TrimSpace(parts[1]) }
			bind := ""
			if strings.Contains(val, "@") || strings.Contains(val, "$") {
				bind = fmt.Sprintf(` data-fx-bind="%s"`, htmlEsc(val))
			}
			cells.WriteString(fmt.Sprintf(`<div class="fx-stat"><div class="fx-stat-val"%s>%s</div><div class="fx-stat-lbl">%s</div></div>`, bind, htmlEsc(val), htmlEsc(lbl)))
		}
	}
	return fmt.Sprintf(`<div class="fx-stats">%s</div>`+"\n", cells.String())
}

func renderRowHTML(line string) string {
	body := extractBody(line)
	cols := 3
	head := line[:strings.Index(line, "{")]
	if m := strings.TrimLeftFunc(head, func(r rune) bool { return r < '0' || r > '9' }); m != "" {
		fmt.Sscanf(m, "%d", &cols)
	}
	items := parseItemsHTML(body)
	var cards strings.Builder
	for _, item := range items {
		var inner strings.Builder
		for fi, f := range item {
			if f.isImg { inner.WriteString(fmt.Sprintf(`<img src="%s" class="fx-card-img" alt="" loading="lazy">`, htmlEsc(f.src))) } else if f.isLink { inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-card-link">%s →</a>`, htmlEsc(f.path), htmlEsc(f.label))) } else if fi == 0 { ico, ok := icons2[f.text]; if !ok { ico = f.text }; inner.WriteString(fmt.Sprintf(`<div class="fx-icon">%s</div>`, ico)) } else if fi == 1 { inner.WriteString(fmt.Sprintf(`<h3 class="fx-card-title">%s</h3>`, htmlEsc(f.text))) } else { inner.WriteString(fmt.Sprintf(`<p class="fx-card-body">%s</p>`, htmlEsc(f.text))) }
		}
		cards.WriteString(fmt.Sprintf(`<div class="fx-card">%s</div>`, inner.String()))
	}
	return fmt.Sprintf(`<div class="fx-grid fx-grid-%d">%s</div>`+"\n", cols, cards.String())
}

func renderSectHTML(line string) string {
	body := extractBody(line)
	items := parseItemsHTML(body)
	var inner strings.Builder
	for ii, item := range items {
		for _, f := range item {
			if f.isLink { inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-sect-link">%s</a>`, htmlEsc(f.path), htmlEsc(f.label))) } else if ii == 0 { inner.WriteString(fmt.Sprintf(`<h2 class="fx-sect-title">%s</h2>`, htmlEsc(f.text))) } else { inner.WriteString(fmt.Sprintf(`<p class="fx-sect-body">%s</p>`, htmlEsc(f.text))) }
		}
	}
	return fmt.Sprintf(`<section class="fx-sect">%s</section>`+"\n", inner.String())
}

func renderFootHTML(line string) string {
	body := extractBody(line)
	items := parseItemsHTML(body)
	var inner strings.Builder
	for _, item := range items {
		for _, f := range item {
			if f.isLink { inner.WriteString(fmt.Sprintf(`<a href="%s" class="fx-footer-link">%s</a>`, htmlEsc(f.path), htmlEsc(f.label))) } else { inner.WriteString(fmt.Sprintf(`<p class="fx-footer-text">%s</p>`, htmlEsc(f.text))) }
		}
	}
	return fmt.Sprintf(`<footer class="fx-footer">%s</footer>`+"\n", inner.String())
}

func renderTableHTML(line string) string {
	// table @var { Col:key | edit PUT /path/{id} | delete /path/{id} | empty: text }
	bi := strings.Index(line, "{")
	if bi == -1 { return "" }
	binding := strings.TrimSpace(line[6:bi])
	content := strings.TrimSuffix(strings.TrimSpace(line[bi+1:]), "}")

	editMatch := extractMatch(content, `edit\s+(PUT|PATCH)\s+(\S+)`)
	delMatch  := extractMatch(content, `delete\s+(?:DELETE\s+)?(\S+)`)
	clean := content
	if editMatch != "" { clean = strings.ReplaceAll(clean, editMatch, "") }

	var ths strings.Builder
	var keys []string
	var colMap []string
	for _, col := range strings.Split(clean, "|") {
		col = strings.TrimSpace(col)
		if col == "" || strings.HasPrefix(col, "empty:") || strings.HasPrefix(col, "delete") || strings.HasPrefix(col, "edit") { continue }
		parts := strings.SplitN(col, ":", 2)
		if len(parts) == 2 {
			label := strings.TrimSpace(parts[0])
			key   := strings.TrimSpace(parts[1])
			ths.WriteString(fmt.Sprintf(`<th class="fx-th">%s</th>`, htmlEsc(label)))
			keys = append(keys, `"`+key+`"`)
			colMap = append(colMap, fmt.Sprintf(`{"label":%q,"key":%q}`, label, key))
		}
	}

	empty := "No data."
	for _, p := range strings.Split(content, "|") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "empty:") { empty = strings.TrimSpace(p[6:]) }
	}

	editAttr, delAttr, actTh := "", "", ""
	if editMatch != "" {
		parts := strings.Fields(editMatch)
		if len(parts) >= 3 { editAttr = fmt.Sprintf(` data-fx-edit="%s" data-fx-edit-method="%s"`, htmlEsc(parts[2]), htmlEsc(parts[1])) }
		actTh = `<th class="fx-th fx-th-actions">Actions</th>`
	}
	if delMatch != "" {
		parts := strings.Fields(delMatch)
		if len(parts) >= 1 { delAttr = fmt.Sprintf(` data-fx-delete="%s"`, htmlEsc(parts[len(parts)-1])) }
		actTh = `<th class="fx-th fx-th-actions">Actions</th>`
	}

	colSpan := len(keys) + 1
	keysJSON := "[" + strings.Join(keys, ",") + "]"
	colMapJSON := "[" + strings.Join(colMap, ",") + "]"

	return fmt.Sprintf(`<div class="fx-table-wrap"><table class="fx-table" data-fx-table="%s" data-fx-cols='%s' data-fx-col-map='%s'%s%s><thead><tr>%s%s</tr></thead><tbody class="fx-tbody"><tr><td colspan="%d" class="fx-td-empty">%s</td></tr></tbody></table></div>`+"\n",
		htmlEsc(binding), keysJSON, colMapJSON, editAttr, delAttr, ths.String(), actTh, colSpan, htmlEsc(empty))
}

func renderFormHTML(line string) string {
	bi := strings.Index(line, "{")
	if bi == -1 { return "" }
	head := strings.TrimSpace(line[5:bi])
	content := strings.TrimSuffix(strings.TrimSpace(line[bi+1:]), "}")
	method, bpath, action := "POST", "#", ""
	ai := strings.Index(head, "=>")
	if ai != -1 { action = strings.TrimSpace(head[ai+2:]); head = strings.TrimSpace(head[:ai]) }
	parts := strings.Fields(head)
	if len(parts) > 0 { method = parts[0] }
	if len(parts) > 1 { bpath = parts[1] }
	var fields strings.Builder
	for _, f := range strings.Split(content, "|") {
		pts := strings.SplitN(strings.TrimSpace(f), ":", 3)
		if len(pts) < 2 || strings.TrimSpace(pts[0]) == "" { continue }
		label := strings.TrimSpace(pts[0])
		ftype := strings.TrimSpace(pts[1])
		ph := ""
		if len(pts) > 2 { ph = strings.TrimSpace(pts[2]) }
		name := strings.ToLower(strings.ReplaceAll(label, " ", "_"))
		var inp string
		if ftype == "select" {
			inp = fmt.Sprintf(`<select class="fx-input" name="%s"><option value="">Select...</option></select>`, htmlEsc(name))
		} else {
			inp = fmt.Sprintf(`<input class="fx-input" type="%s" name="%s" placeholder="%s">`, htmlEsc(ftype), htmlEsc(name), htmlEsc(ph))
		}
		fields.WriteString(fmt.Sprintf(`<div class="fx-field"><label class="fx-label">%s</label>%s</div>`, htmlEsc(label), inp))
	}
	return fmt.Sprintf(`<div class="fx-form-wrap"><form class="fx-form" data-fx-form="%s" data-fx-method="%s" data-fx-action="%s">%s<div class="fx-form-msg"></div><button type="submit" class="fx-btn">Submit</button></form></div>`+"\n",
		htmlEsc(bpath), htmlEsc(method), htmlEsc(action), fields.String())
}

func renderPricingHTML(line string) string {
	body := extractBody(line)
	var cards strings.Builder
	for i, plan := range strings.Split(body, "|") {
		pts := strings.Split(strings.TrimSpace(plan), ">")
		if len(pts) < 3 { continue }
		name, price, desc := strings.TrimSpace(pts[0]), strings.TrimSpace(pts[1]), strings.TrimSpace(pts[2])
		linkHref, linkLabel := "#", "Get started"
		if len(pts) > 3 {
			lp := strings.TrimSpace(pts[3])
			lpts := strings.SplitN(lp[1:], ":", 2)
			if len(lpts) == 2 { linkHref = "/"+lpts[0]; linkLabel = lpts[1] }
		}
		featured := ""
		badge := ""
		if i == 1 { featured = " fx-pricing-featured"; badge = `<div class="fx-pricing-badge">Most popular</div>` }
		cards.WriteString(fmt.Sprintf(`<div class="fx-pricing-card%s">%s<div class="fx-pricing-name">%s</div><div class="fx-pricing-price">%s</div><p class="fx-pricing-desc">%s</p><a href="%s" class="fx-cta fx-pricing-cta">%s</a></div>`,
			featured, badge, htmlEsc(name), htmlEsc(price), htmlEsc(desc), htmlEsc(linkHref), htmlEsc(linkLabel)))
	}
	return fmt.Sprintf(`<div class="fx-pricing">%s</div>`+"\n", cards.String())
}

func renderFaqHTML(line string) string {
	body := extractBody(line)
	var items strings.Builder
	for _, pair := range strings.Split(body, "|") {
		idx := strings.Index(pair, ">")
		if idx == -1 { continue }
		q := strings.TrimSpace(pair[:idx])
		a := strings.TrimSpace(pair[idx+1:])
		items.WriteString(fmt.Sprintf(`<div class="fx-faq-item" onclick="this.classList.toggle('open')"><div class="fx-faq-q">%s<span class="fx-faq-arrow">▸</span></div><div class="fx-faq-a">%s</div></div>`, htmlEsc(q), htmlEsc(a)))
	}
	return fmt.Sprintf(`<section class="fx-sect"><div class="fx-faq">%s</div></section>`+"\n", items.String())
}

func renderTestimonialHTML(line string) string {
	body := extractBody(line)
	parts := strings.Split(body, "|")
	if len(parts) < 2 { return "" }
	author := strings.TrimSpace(parts[0])
	quote  := strings.Trim(strings.TrimSpace(parts[1]), `"`)
	img := ""
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "img:") { img = fmt.Sprintf(`<img src="%s" class="fx-testi-img" alt="%s" loading="lazy">`, htmlEsc(p[4:]), htmlEsc(author)) }
	}
	if img == "" { img = fmt.Sprintf(`<div class="fx-testi-avatar">%s</div>`, htmlEsc(string([]rune(author)[0:1]))) }
	return fmt.Sprintf(`<section class="fx-testi-wrap"><div class="fx-testi">%s<blockquote class="fx-testi-quote">"%s"</blockquote><div class="fx-testi-author">%s</div></div></section>`+"\n", img, htmlEsc(quote), htmlEsc(author))
}

func renderGalleryHTML(line string) string {
	body := extractBody(line)
	var imgs strings.Builder
	for _, src := range strings.Split(body, "|") {
		src = strings.TrimSpace(src)
		if src != "" { imgs.WriteString(fmt.Sprintf(`<div class="fx-gallery-item"><img src="%s" alt="" loading="lazy"></div>`, htmlEsc(src))) }
	}
	return fmt.Sprintf(`<div class="fx-gallery">%s</div>`+"\n", imgs.String())
}

func renderBtnHTML(line string) string {
	body := extractBody(line)
	parts := strings.Split(body, ">")
	label, method, bpath, confirm, action := "Click", "POST", "#", "", ""
	if len(parts) > 0 { label = strings.TrimSpace(parts[0]) }
	if len(parts) > 1 { pts := strings.Fields(parts[1]); if len(pts)>0{method=pts[0]}; if len(pts)>1{bpath=strings.Join(pts[1:]," ")} }
	for _, p := range parts[2:] {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "confirm:") { confirm = fmt.Sprintf(` data-fx-confirm="%s"`, htmlEsc(p[8:])) } else { action = fmt.Sprintf(` data-fx-action="%s"`, htmlEsc(p)) }
	}
	return fmt.Sprintf(`<div class="fx-btn-wrap"><button class="fx-btn fx-standalone-btn" data-fx-btn="%s" data-fx-method="%s"%s%s>%s</button></div>`+"\n", htmlEsc(bpath), htmlEsc(method), confirm, action, htmlEsc(label))
}

func renderIfHTML(line string) string {
	bi := strings.Index(line, "{")
	if bi == -1 { return "" }
	cond := strings.TrimSpace(line[3:bi])
	return fmt.Sprintf(`<div class="fx-if-wrap" data-fx-if="%s" style="display:none"></div>`+"\n", htmlEsc(cond))
}

// ── Item Parser ───────────────────────────────────────────────────

type htmlField struct {
	isLink bool
	isImg  bool
	text   string
	path   string
	label  string
	src    string
}

func parseItemsHTML(body string) [][]htmlField {
	var items [][]htmlField
	for _, raw := range strings.Split(body, "|") {
		raw = strings.TrimSpace(raw)
		if raw == "" { continue }
		var item []htmlField
		for _, f := range strings.Split(raw, ">") {
			f = strings.TrimSpace(f)
			if strings.HasPrefix(f, "img:") {
				item = append(item, htmlField{isImg: true, src: f[4:]})
			} else if strings.HasPrefix(f, "/") {
				pts := strings.SplitN(f, ":", 2)
				label := ""
				if len(pts) > 1 { label = pts[1] }
				item = append(item, htmlField{isLink: true, path: pts[0], label: label})
			} else {
				item = append(item, htmlField{text: f})
			}
		}
		if len(item) > 0 { items = append(items, item) }
	}
	return items
}

func extractBody(line string) string {
	bi := strings.Index(line, "{")
	li := strings.LastIndex(line, "}")
	if bi == -1 || li == -1 { return "" }
	return strings.TrimSpace(line[bi+1 : li])
}

func extractMatch(s, pattern string) string {
	// Simple extraction without regex — find key patterns
	if strings.Contains(pattern, "edit") {
		for _, part := range strings.Split(s, "|") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "edit ") { return part }
		}
	}
	if strings.Contains(pattern, "delete") {
		for _, part := range strings.Split(s, "|") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "delete ") { return part }
		}
	}
	return ""
}

func buildQueriesJSON(queries []fc.LifecycleQuery) string {
	var parts []string
	for _, q := range queries {
		parts = append(parts, fmt.Sprintf(`{"trigger":%q,"interval":%d,"method":%q,"path":%q,"target":%q,"action":%q}`,
			q.Trigger, q.Interval, q.Method, q.Path, q.Target, q.Action))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func capitalize(s string) string {
	if s == "" { return s }
	return strings.ToUpper(s[:1]) + s[1:]
}

func genCustomCSS(ct *fc.CustomTheme) string {
	accent := ct.Accent
	if accent == "" { accent = "#2563eb" }
	return fmt.Sprintf(`body{background:%s;color:%s}.fx-cta,.fx-btn{background:%s;color:#fff}.fx-nav{background:%scc}.fx-card,.fx-form{border:1px solid %s30}`,
		ct.BG, ct.Text, accent, ct.BG, ct.Text)
}

func getBaseCSS(theme string) string {
	// Same CSS as aiplang npm package — omitted here for brevity
	// In production this would be the full CSS string
	base := `*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}html{scroll-behavior:smooth}body{font-family:-apple-system,'Segoe UI',system-ui,sans-serif;-webkit-font-smoothing:antialiased;min-height:100vh}a{text-decoration:none;color:inherit}input,button,select{font-family:inherit}img{max-width:100%;height:auto}.fx-nav{display:flex;align-items:center;justify-content:space-between;padding:1rem 2.5rem;position:sticky;top:0;z-index:50;backdrop-filter:blur(12px);flex-wrap:wrap;gap:.5rem}.fx-brand{font-size:1.25rem;font-weight:800;letter-spacing:-.03em}.fx-nav-links{display:flex;align-items:center;gap:1.75rem}.fx-nav-link{font-size:.875rem;font-weight:500;opacity:.65;transition:opacity .15s}.fx-nav-link:hover{opacity:1}.fx-hamburger{display:none;flex-direction:column;gap:5px;background:none;border:none;cursor:pointer;padding:.25rem}.fx-hamburger span{display:block;width:22px;height:2px;background:currentColor;transition:all .2s;border-radius:1px}@media(max-width:640px){.fx-hamburger{display:flex}.fx-nav-links{display:none;width:100%;flex-direction:column;align-items:flex-start;gap:.75rem;padding:.75rem 0}.fx-nav-links.open{display:flex}}.fx-hero{display:flex;align-items:center;justify-content:center;min-height:92vh;padding:4rem 1.5rem}.fx-hero-split{display:grid;grid-template-columns:1fr 1fr;gap:3rem;align-items:center;padding:4rem 2.5rem;min-height:70vh}@media(max-width:768px){.fx-hero-split{grid-template-columns:1fr}}.fx-hero-img{width:100%;border-radius:1.25rem;object-fit:cover;max-height:500px}.fx-hero-inner{max-width:56rem;text-align:center;display:flex;flex-direction:column;align-items:center;gap:1.5rem}.fx-hero-split .fx-hero-inner{text-align:left;align-items:flex-start;max-width:none}.fx-title{font-size:clamp(2.5rem,8vw,5.5rem);font-weight:900;letter-spacing:-.04em;line-height:1}.fx-sub{font-size:clamp(1rem,2vw,1.25rem);line-height:1.75;max-width:40rem}.fx-cta{display:inline-flex;align-items:center;padding:.875rem 2.5rem;border-radius:.75rem;font-weight:700;font-size:1rem;letter-spacing:-.01em;transition:transform .15s;margin:.25rem}.fx-cta:hover{transform:translateY(-1px)}.fx-stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:3rem;padding:5rem 2.5rem;text-align:center}.fx-stat-val{font-size:clamp(2.5rem,5vw,4rem);font-weight:900;letter-spacing:-.04em;line-height:1}.fx-stat-lbl{font-size:.75rem;font-weight:600;text-transform:uppercase;letter-spacing:.1em;margin-top:.5rem}.fx-grid{display:grid;gap:1.25rem;padding:1rem 2.5rem 5rem}.fx-grid-2{grid-template-columns:repeat(auto-fit,minmax(280px,1fr))}.fx-grid-3{grid-template-columns:repeat(auto-fit,minmax(240px,1fr))}.fx-grid-4{grid-template-columns:repeat(auto-fit,minmax(200px,1fr))}.fx-card{border-radius:1rem;padding:1.75rem;transition:transform .2s,box-shadow .2s}.fx-card:hover{transform:translateY(-2px)}.fx-card-img{width:100%;border-radius:.75rem;object-fit:cover;height:180px;margin-bottom:1rem}.fx-icon{font-size:2rem;margin-bottom:1rem}.fx-card-title{font-size:1.0625rem;font-weight:700;letter-spacing:-.02em;margin-bottom:.5rem}.fx-card-body{font-size:.875rem;line-height:1.65}.fx-sect{padding:5rem 2.5rem}.fx-sect-title{font-size:clamp(1.75rem,4vw,3rem);font-weight:800;letter-spacing:-.04em;margin-bottom:1.5rem;text-align:center}.fx-sect-body{font-size:1rem;line-height:1.75;text-align:center;max-width:48rem;margin:0 auto}.fx-form-wrap{padding:3rem 2.5rem;display:flex;justify-content:center}.fx-form{width:100%;max-width:28rem;border-radius:1.25rem;padding:2.5rem}.fx-field{margin-bottom:1.25rem}.fx-label{display:block;font-size:.8125rem;font-weight:600;margin-bottom:.5rem}.fx-input{width:100%;padding:.75rem 1rem;border-radius:.625rem;font-size:.9375rem;outline:none;transition:box-shadow .15s}.fx-input:focus{box-shadow:0 0 0 3px rgba(37,99,235,.35)}.fx-btn{width:100%;padding:.875rem 1.5rem;border:none;border-radius:.625rem;font-size:.9375rem;font-weight:700;cursor:pointer;margin-top:.5rem;transition:transform .15s,opacity .15s}.fx-btn:hover{transform:translateY(-1px)}.fx-btn:disabled{opacity:.5;cursor:not-allowed;transform:none}.fx-btn-wrap{padding:0 2.5rem 1.5rem}.fx-standalone-btn{width:auto;padding:.75rem 2rem;margin-top:0}.fx-form-msg{font-size:.8125rem;padding:.5rem 0;min-height:1.5rem;text-align:center}.fx-form-err{color:#f87171}.fx-form-ok{color:#4ade80}.fx-table-wrap{overflow-x:auto;padding:0 2.5rem 4rem}.fx-table{width:100%;border-collapse:collapse;font-size:.875rem}.fx-th{text-align:left;padding:.875rem 1.25rem;font-size:.75rem;font-weight:700;text-transform:uppercase;letter-spacing:.06em}.fx-th-actions{opacity:.6}.fx-tr{transition:background .1s}.fx-td{padding:.875rem 1.25rem}.fx-td-empty{padding:2rem 1.25rem;text-align:center;opacity:.4}.fx-td-actions{white-space:nowrap;padding:.5rem 1rem!important}.fx-action-btn{border:none;cursor:pointer;font-size:.75rem;font-weight:600;padding:.3rem .75rem;border-radius:.375rem;margin-right:.375rem;font-family:inherit;transition:opacity .15s}.fx-action-btn:hover{opacity:.85}.fx-edit-btn{background:#1e40af;color:#93c5fd}.fx-delete-btn{background:#7f1d1d;color:#fca5a5}.fx-pricing{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:1.5rem;padding:2rem 2.5rem 5rem;align-items:start}.fx-pricing-card{border-radius:1.25rem;padding:2rem;position:relative;transition:transform .2s}.fx-pricing-featured{transform:scale(1.03)}.fx-pricing-badge{position:absolute;top:-12px;left:50%;transform:translateX(-50%);background:#2563eb;color:#fff;font-size:.7rem;font-weight:700;padding:.25rem .875rem;border-radius:999px;white-space:nowrap}.fx-pricing-name{font-size:.875rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;margin-bottom:.5rem;opacity:.7}.fx-pricing-price{font-size:3rem;font-weight:900;letter-spacing:-.05em;line-height:1;margin-bottom:.75rem}.fx-pricing-desc{font-size:.875rem;line-height:1.65;margin-bottom:1.5rem;opacity:.7}.fx-pricing-cta{display:block;text-align:center;padding:.75rem;border-radius:.625rem;font-weight:700;font-size:.9rem}.fx-faq{max-width:48rem;margin:0 auto}.fx-faq-item{border-radius:.75rem;margin-bottom:.625rem;cursor:pointer;overflow:hidden}.fx-faq-q{display:flex;justify-content:space-between;align-items:center;padding:1rem 1.25rem;font-size:.9375rem;font-weight:600}.fx-faq-arrow{transition:transform .2s;font-size:.75rem;opacity:.5}.fx-faq-item.open .fx-faq-arrow{transform:rotate(90deg)}.fx-faq-a{max-height:0;overflow:hidden;padding:0 1.25rem;font-size:.875rem;line-height:1.7;transition:max-height .3s,padding .3s}.fx-faq-item.open .fx-faq-a{max-height:300px;padding:.75rem 1.25rem 1.25rem}.fx-testi-wrap{padding:5rem 2.5rem;display:flex;justify-content:center}.fx-testi{max-width:42rem;text-align:center;display:flex;flex-direction:column;align-items:center;gap:1.25rem}.fx-testi-img{width:64px;height:64px;border-radius:50%;object-fit:cover}.fx-testi-avatar{width:64px;height:64px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:1.5rem;font-weight:700;background:#1e293b}.fx-testi-quote{font-size:1.25rem;line-height:1.7;font-style:italic}.fx-testi-author{font-size:.875rem;font-weight:600;opacity:.5}.fx-gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:.75rem;padding:1rem 2.5rem 4rem}.fx-gallery-item{border-radius:.75rem;overflow:hidden;aspect-ratio:4/3}.fx-gallery-item img{width:100%;height:100%;object-fit:cover;transition:transform .3s}.fx-gallery-item:hover img{transform:scale(1.04)}.fx-if-wrap{display:contents}.fx-footer{padding:3rem 2.5rem;text-align:center}.fx-footer-text{font-size:.8125rem}.fx-footer-link{font-size:.8125rem;margin:0 .75rem;opacity:.5;transition:opacity .15s}.fx-footer-link:hover{opacity:1}`

	themes := map[string]string{
		"dark":  `body{background:#030712;color:#f1f5f9}.fx-nav{border-bottom:1px solid #1e293b;background:rgba(3,7,18,.85)}.fx-nav-link{color:#cbd5e1}.fx-sub{color:#94a3b8}.fx-cta{background:#2563eb;color:#fff;box-shadow:0 8px 24px rgba(37,99,235,.35)}.fx-stat-lbl{color:#64748b}.fx-card{background:#0f172a;border:1px solid #1e293b}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.5)}.fx-card-body{color:#64748b}.fx-sect-body{color:#64748b}.fx-form{background:#0f172a;border:1px solid #1e293b}.fx-label{color:#94a3b8}.fx-input{background:#020617;border:1px solid #1e293b;color:#f1f5f9}.fx-input::placeholder{color:#334155}.fx-btn{background:#2563eb;color:#fff;box-shadow:0 4px 14px rgba(37,99,235,.4)}.fx-th{color:#475569;border-bottom:1px solid #1e293b}.fx-tr:hover{background:#0f172a}.fx-td{border-bottom:1px solid rgba(255,255,255,.03)}.fx-footer{border-top:1px solid #1e293b}.fx-footer-text{color:#334155}.fx-pricing-card{background:#0f172a;border:1px solid #1e293b}.fx-faq-item{background:#0f172a}`,
		"light": `body{background:#fff;color:#0f172a}.fx-nav{border-bottom:1px solid #e2e8f0;background:rgba(255,255,255,.85)}.fx-nav-link{color:#475569}.fx-sub{color:#475569}.fx-cta{background:#2563eb;color:#fff}.fx-stat-lbl{color:#94a3b8}.fx-card{background:#f8fafc;border:1px solid #e2e8f0}.fx-card:hover{box-shadow:0 20px 40px rgba(0,0,0,.08)}.fx-card-body{color:#475569}.fx-sect-body{color:#475569}.fx-form{background:#f8fafc;border:1px solid #e2e8f0}.fx-label{color:#475569}.fx-input{background:#fff;border:1px solid #cbd5e1;color:#0f172a}.fx-btn{background:#2563eb;color:#fff}.fx-th{color:#94a3b8;border-bottom:1px solid #e2e8f0}.fx-tr:hover{background:#f8fafc}.fx-footer{border-top:1px solid #e2e8f0}.fx-footer-text{color:#94a3b8}.fx-pricing-card{background:#f8fafc;border:1px solid #e2e8f0}.fx-faq-item{background:#f8fafc}`,
		"acid":  `body{background:#000;color:#a3e635}.fx-nav{border-bottom:1px solid #1a2e05;background:rgba(0,0,0,.9)}.fx-nav-link{color:#86efac}.fx-sub{color:#4d7c0f}.fx-cta{background:#a3e635;color:#000;font-weight:800}.fx-stat-lbl{color:#365314}.fx-card{background:#0a0f00;border:1px solid #1a2e05}.fx-card-body{color:#365314}.fx-sect-body{color:#365314}.fx-form{background:#0a0f00;border:1px solid #1a2e05}.fx-label{color:#4d7c0f}.fx-input{background:#000;border:1px solid #1a2e05;color:#a3e635}.fx-btn{background:#a3e635;color:#000;font-weight:800}.fx-th{color:#365314;border-bottom:1px solid #1a2e05}.fx-footer{border-top:1px solid #1a2e05}.fx-footer-text{color:#1a2e05}.fx-pricing-card{background:#0a0f00;border:1px solid #1a2e05}.fx-faq-item{background:#0a0f00}`,
	}

	t, ok := themes[theme]
	if !ok { t = themes["dark"] }
	return base + t
}
