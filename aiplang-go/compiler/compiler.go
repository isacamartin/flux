// aiplang v2 Compiler
// Parses .flux source into an AppDef (models + api routes + pages)

package aiplangcompiler

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────
// AST
// ─────────────────────────────────────────────────────────────

type AppDef struct {
	Env        []EnvVar
	DB         *DBConfig
	Auth       *AuthConfig
	Cache      *CacheConfig
	Middleware []string
	Models     []Model
	APIs       []APIRoute
	Pages      []Page
}

type EnvVar struct {
	Name     string
	Default  string
	Required bool
}

type DBConfig struct {
	Driver string // postgres mysql sqlite mongo
	DSN    string // literal or $ENV_VAR
}

type AuthConfig struct {
	Provider string // jwt session google github
	Secret   string
	Expire   string
}

type CacheConfig struct {
	Driver string // redis memory
	URL    string
	TTL    int
}

// ── Model ──

type Model struct {
	Name   string
	Fields []ModelField
}

type ModelField struct {
	Name      string
	Type      string // uuid text int float bool timestamp json enum
	Modifiers []string
	EnumVals  []string
	Ref       string // foreign key model name
	Default   string
}

// ── API ──

type APIRoute struct {
	Method      string // GET POST PUT PATCH DELETE
	Path        string
	Guards      []string // auth admin owner
	Validate    []Validation
	QueryParams []QueryParam
	Body        []BodyOp
	Return      ReturnExpr
}

type Validation struct {
	Field string
	Rules []string // required email min=N max=N unique
}

type QueryParam struct {
	Name    string
	Default string
}

type BodyOp struct {
	Op      string // insert update delete find findBy check hash unique
	Model   string
	Fields  string
	Cond    string
	Binding string // $var to assign result to
}

type ReturnExpr struct {
	Expr   string
	Status int
}

// ── Page (same as aiplang v1) ──

type Page struct {
	ID          string
	Theme       string
	Route       string
	CustomTheme *CustomTheme
	State       map[string]string
	Queries     []LifecycleQuery
	Blocks      []Block
}

type CustomTheme struct {
	BG     string
	Text   string
	Accent string
}

type LifecycleQuery struct {
	Trigger  string
	Interval int
	Method   string
	Path     string
	Target   string
	Action   string
}

type Block struct {
	Kind     string
	RawLine  string
}

// ─────────────────────────────────────────────────────────────
// PARSER
// ─────────────────────────────────────────────────────────────

func Parse(src string) (*AppDef, error) {
	app := &AppDef{}

	// Split into global config section and page sections
	// Pages are delimited by --- on its own line
	lines := splitLines(src)

	// First pass: collect global directives and models/APIs
	// Second pass: collect pages

	var pagesSrc []string
	var globalLines []string
	var inPage bool
	var currentPageLines []string

	for _, line := range lines {
		if line == "---" {
			if inPage {
				pagesSrc = append(pagesSrc, strings.Join(currentPageLines, "\n"))
				currentPageLines = nil
			}
			inPage = false
			continue
		}

		if strings.HasPrefix(line, "%") {
			inPage = true
			currentPageLines = append(currentPageLines, line)
			continue
		}

		if inPage {
			currentPageLines = append(currentPageLines, line)
		} else {
			globalLines = append(globalLines, line)
		}
	}

	if inPage && len(currentPageLines) > 0 {
		pagesSrc = append(pagesSrc, strings.Join(currentPageLines, "\n"))
	}

	// Parse global config
	if err := parseGlobal(globalLines, app); err != nil {
		return nil, err
	}

	// Parse pages
	for _, ps := range pagesSrc {
		page, err := parsePage(ps)
		if err != nil {
			return nil, err
		}
		if page != nil {
			app.Pages = append(app.Pages, *page)
		}
	}

	return app, nil
}

func splitLines(src string) []string {
	var lines []string
	for _, l := range strings.Split(src, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}

func parseGlobal(lines []string, app *AppDef) error {
	i := 0
	for i < len(lines) {
		line := lines[i]

		switch {
		case strings.HasPrefix(line, "~env "):
			ev := parseEnvVar(line[5:])
			app.Env = append(app.Env, ev)

		case strings.HasPrefix(line, "~db "):
			app.DB = parseDB(line[4:])

		case strings.HasPrefix(line, "~auth "):
			app.Auth = parseAuth(line[6:])

		case strings.HasPrefix(line, "~cache "):
			app.Cache = parseCache(line[7:])

		case strings.HasPrefix(line, "~middleware "):
			app.Middleware = splitPipe(line[12:])

		case strings.HasPrefix(line, "model "):
			// Collect multi-line model block
			var modelLines []string
			modelLines = append(modelLines, line)
			for i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "model ") &&
				!strings.HasPrefix(lines[i+1], "api ") &&
				!strings.HasPrefix(lines[i+1], "~") {
				i++
				modelLines = append(modelLines, lines[i])
			}
			m, err := parseModel(modelLines)
			if err != nil {
				return err
			}
			app.Models = append(app.Models, m)

		case strings.HasPrefix(line, "api "):
			// Collect multi-line api block
			var apiLines []string
			apiLines = append(apiLines, line)
			for i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "api ") &&
				!strings.HasPrefix(lines[i+1], "model ") &&
				!strings.HasPrefix(lines[i+1], "~") &&
				!strings.HasPrefix(lines[i+1], "%") {
				i++
				apiLines = append(apiLines, lines[i])
			}
			route, err := parseAPI(apiLines)
			if err != nil {
				return err
			}
			app.APIs = append(app.APIs, route)
		}
		i++
	}
	return nil
}

func parseEnvVar(s string) EnvVar {
	parts := strings.Fields(s)
	ev := EnvVar{}
	for _, p := range parts {
		if p == "required" {
			ev.Required = true
		} else if strings.Contains(p, "=") {
			kv := strings.SplitN(p, "=", 2)
			ev.Name = kv[0]
			ev.Default = kv[1]
		} else {
			ev.Name = p
		}
	}
	return ev
}

func parseDB(s string) *DBConfig {
	parts := strings.Fields(s)
	db := &DBConfig{}
	if len(parts) >= 1 { db.Driver = parts[0] }
	if len(parts) >= 2 { db.DSN = parts[1] }
	return db
}

func parseAuth(s string) *AuthConfig {
	parts := strings.Fields(s)
	auth := &AuthConfig{}
	if len(parts) >= 1 { auth.Provider = parts[0] }
	if len(parts) >= 2 { auth.Secret = parts[1] }
	for _, p := range parts[2:] {
		if strings.HasPrefix(p, "expire=") {
			auth.Expire = p[7:]
		}
	}
	return auth
}

func parseCache(s string) *CacheConfig {
	parts := strings.Fields(s)
	c := &CacheConfig{TTL: 300}
	if len(parts) >= 1 { c.Driver = parts[0] }
	if len(parts) >= 2 { c.URL = parts[1] }
	for _, p := range parts {
		if strings.HasPrefix(p, "ttl=") {
			fmt.Sscanf(p[4:], "%d", &c.TTL)
		}
	}
	return c
}

func parseModel(lines []string) (Model, error) {
	m := Model{}
	if len(lines) == 0 {
		return m, nil
	}

	// First line: "model Name {" or "model Name { field : type : mod }"
	head := lines[0]
	bi := strings.Index(head, "{")
	if bi == -1 {
		m.Name = strings.TrimSpace(head[6:])
		return m, nil
	}
	m.Name = strings.TrimSpace(head[6:bi])

	// Collect field lines from all model lines
	var fieldLines []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "model ") || l == "{" || l == "}" {
			continue
		}
		// strip inline braces
		l = strings.TrimPrefix(l, "{")
		l = strings.TrimSuffix(l, "}")
		l = strings.TrimSpace(l)
		if l != "" {
			fieldLines = append(fieldLines, l)
		}
	}

	// Also parse single-line models: model User { id:uuid:pk | name:text:required }
	if len(lines) == 1 && bi != -1 {
		body := head[bi+1:strings.LastIndex(head, "}")]
		fieldLines = strings.Split(body, "|")
	}

	for _, fl := range fieldLines {
		fl = strings.TrimSpace(fl)
		if fl == "" { continue }
		field, err := parseModelField(fl)
		if err != nil { continue }
		m.Fields = append(m.Fields, field)
	}

	return m, nil
}

func parseModelField(s string) (ModelField, error) {
	// format: name : type : modifier modifier | enum:val1,val2 | default=val
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return ModelField{}, fmt.Errorf("invalid field: %s", s)
	}

	f := ModelField{
		Name: strings.TrimSpace(parts[0]),
		Type: strings.TrimSpace(parts[1]),
	}

	if len(parts) >= 3 {
		mods := strings.Fields(parts[2])
		for _, mod := range mods {
			if strings.HasPrefix(mod, "default=") {
				f.Default = mod[8:]
			} else if strings.HasPrefix(mod, "ref ") {
				f.Ref = mod[4:]
			} else {
				f.Modifiers = append(f.Modifiers, mod)
			}
		}
	}

	// Enum values embedded in type: "enum:val1,val2"
	if strings.HasPrefix(f.Type, "enum") && strings.Contains(f.Type, ":") {
		typeParts := strings.SplitN(f.Type, ":", 2)
		f.Type = typeParts[0]
		f.EnumVals = strings.Split(typeParts[1], ",")
	} else if len(parts) >= 4 {
		// enum values in 4th section
		f.EnumVals = strings.Split(strings.TrimSpace(parts[3]), ",")
	}

	return f, nil
}

func parseAPI(lines []string) (APIRoute, error) {
	route := APIRoute{Return: ReturnExpr{Status: 200}}
	if len(lines) == 0 {
		return route, nil
	}

	// First line: "api METHOD /path {"
	head := lines[0]
	bi := strings.Index(head, "{")
	var headPart string
	if bi != -1 {
		headPart = strings.TrimSpace(head[4:bi])
	} else {
		headPart = strings.TrimSpace(head[4:])
	}

	hParts := strings.Fields(headPart)
	if len(hParts) >= 1 { route.Method = hParts[0] }
	if len(hParts) >= 2 { route.Path = hParts[1] }

	// Parse body lines
	var bodyLines []string
	if bi != -1 && strings.Contains(head, "}") {
		// single line api
		body := head[bi+1:strings.LastIndex(head, "}")]
		bodyLines = strings.Split(body, "|")
	} else {
		for _, l := range lines[1:] {
			l = strings.TrimSpace(l)
			if l == "{" || l == "}" { continue }
			bodyLines = append(bodyLines, l)
		}
	}

	for _, bl := range bodyLines {
		bl = strings.TrimSpace(bl)
		if bl == "" { continue }
		parseAPILine(bl, &route)
	}

	return route, nil
}

func parseAPILine(line string, route *APIRoute) {
	switch {
	case strings.HasPrefix(line, "~guard "):
		route.Guards = splitPipe(line[7:])

	case strings.HasPrefix(line, "~validate "):
		for _, v := range splitPipe(line[10:]) {
			parts := strings.Fields(v)
			if len(parts) >= 2 {
				route.Validate = append(route.Validate, Validation{
					Field: parts[0],
					Rules: parts[1:],
				})
			}
		}

	case strings.HasPrefix(line, "~query "):
		for _, q := range splitPipe(line[7:]) {
			q = strings.TrimSpace(q)
			if strings.Contains(q, "=") {
				kv := strings.SplitN(q, "=", 2)
				route.QueryParams = append(route.QueryParams, QueryParam{Name: kv[0], Default: kv[1]})
			} else {
				route.QueryParams = append(route.QueryParams, QueryParam{Name: q})
			}
		}

	case strings.HasPrefix(line, "~hash "):
		route.Body = append(route.Body, BodyOp{Op: "hash", Fields: line[6:]})

	case strings.HasPrefix(line, "~check "):
		parts := strings.Fields(line[7:])
		op := BodyOp{Op: "check"}
		if len(parts) >= 1 { op.Fields = parts[0] }
		if len(parts) >= 2 { op.Binding = parts[1] }
		if len(parts) >= 3 { op.Model = parts[2] }
		if len(parts) >= 4 { op.Cond = parts[3] } // status code
		route.Body = append(route.Body, op)

	case strings.HasPrefix(line, "~unique "):
		parts := strings.Fields(line[8:])
		op := BodyOp{Op: "unique"}
		if len(parts) >= 1 { op.Model = parts[0] }
		if len(parts) >= 2 { op.Fields = parts[1] }
		if len(parts) >= 3 { op.Binding = parts[2] }
		if len(parts) >= 4 { op.Cond = parts[3] }
		route.Body = append(route.Body, op)

	case strings.HasPrefix(line, "insert "):
		parts := strings.Fields(line[7:])
		op := BodyOp{Op: "insert"}
		if len(parts) >= 1 { op.Model = strings.TrimSuffix(parts[0], "(") }
		route.Body = append(route.Body, op)

	case strings.HasPrefix(line, "update "):
		parts := strings.Fields(line[7:])
		op := BodyOp{Op: "update"}
		if len(parts) >= 1 { op.Model = parts[0] }
		route.Body = append(route.Body, op)

	case strings.HasPrefix(line, "delete "):
		parts := strings.Fields(line[7:])
		op := BodyOp{Op: "delete"}
		if len(parts) >= 1 { op.Model = parts[0] }
		route.Body = append(route.Body, op)

	case strings.HasPrefix(line, "return "):
		ret := line[7:]
		parts := strings.Fields(ret)
		route.Return.Expr = parts[0]
		if len(parts) >= 2 {
			fmt.Sscanf(parts[1], "%d", &route.Return.Status)
		}

	default:
		// Variable assignment: $user = Model.findBy(...)
		if strings.Contains(line, "=") && strings.HasPrefix(line, "$") {
			parts := strings.SplitN(line, "=", 2)
			op := BodyOp{
				Op:      "assign",
				Binding: strings.TrimSpace(parts[0]),
				Fields:  strings.TrimSpace(parts[1]),
			}
			route.Body = append(route.Body, op)
		}
	}
}

func parsePage(src string) (*Page, error) {
	lines := splitLines(src)
	if len(lines) == 0 { return nil, nil }

	page := &Page{
		ID:    "page",
		Theme: "dark",
		Route: "/",
		State: map[string]string{},
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "%"):
			parts := strings.Fields(line[1:])
			if len(parts) > 0 { page.ID = parts[0] }
			if len(parts) > 2 { page.Route = parts[2] }
			if len(parts) > 1 {
				t := parts[1]
				if strings.Contains(t, "#") {
					colors := strings.Split(t, ",")
					page.Theme = "custom"
					ct := &CustomTheme{BG: colors[0]}
					if len(colors) > 1 { ct.Text = colors[1] }
					if len(colors) > 2 { ct.Accent = colors[2] }
					page.CustomTheme = ct
				} else {
					page.Theme = t
				}
			}

		case strings.HasPrefix(line, "@") && strings.Contains(line, "="):
			eq := strings.Index(line, "=")
			page.State[strings.TrimSpace(line[1:eq])] = strings.TrimSpace(line[eq+1:])

		case strings.HasPrefix(line, "~"):
			q := parseLifecycleQuery(line[1:])
			if q != nil { page.Queries = append(page.Queries, *q) }

		default:
			page.Blocks = append(page.Blocks, Block{Kind: blockKind(line), RawLine: line})
		}
	}

	return page, nil
}

func parseLifecycleQuery(s string) *LifecycleQuery {
	parts := strings.Fields(s)
	ai := -1
	for i, p := range parts {
		if p == "=>" { ai = i; break }
	}
	q := &LifecycleQuery{}
	switch parts[0] {
	case "mount":
		q.Trigger = "mount"
		if len(parts) > 1 { q.Method = parts[1] }
		if len(parts) > 2 { q.Path = parts[2] }
		if ai != -1 { q.Action = strings.Join(parts[ai+1:], " ") } else if len(parts) > 3 { q.Target = parts[3] }
	case "interval":
		q.Trigger = "interval"
		fmt.Sscanf(parts[1], "%d", &q.Interval)
		if len(parts) > 2 { q.Method = parts[2] }
		if len(parts) > 3 { q.Path = parts[3] }
		if ai != -1 { q.Action = strings.Join(parts[ai+1:], " ") } else if len(parts) > 4 { q.Target = parts[4] }
	default:
		return nil
	}
	return q
}

func blockKind(line string) string {
	bi := strings.Index(line, "{")
	if bi == -1 { return "unknown" }
	head := strings.TrimSpace(line[:bi])
	if m := strings.TrimRight(head, "0123456789"); m != head { return m }
	return head
}

func splitPipe(s string) []string {
	parts := strings.Split(s, "|")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" { out = append(out, p) }
	}
	return out
}
