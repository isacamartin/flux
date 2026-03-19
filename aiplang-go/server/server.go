// aiplang v2 Server
// Executes compiled AppDef as a real HTTP server
// - Handles API routes with DB operations
// - Serves SSR frontend pages
// - JWT auth middleware
// - Auto-migration on startup

package aiplangserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	fc "github.com/isacamartin/aiplang/compiler"
)

// ── Server ────────────────────────────────────────────────────────

type Server struct {
	app    *fc.AppDef
	db     *gorm.DB
	mux    *http.ServeMux
	port   string
	jwt    *JWTConfig
}

type JWTConfig struct {
	Secret string
	Expire time.Duration
}

func New(app *fc.AppDef) (*Server, error) {
	s := &Server{
		app:  app,
		mux:  http.NewServeMux(),
		port: resolveEnv("PORT", "3000"),
	}

	// Setup DB
	if app.DB != nil {
		db, err := connectDB(app.DB)
		if err != nil {
			return nil, fmt.Errorf("db connect: %w", err)
		}
		s.db = db
		if err := autoMigrate(db, app.Models); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
		log.Printf("[aiplang] DB connected (%s)", app.DB.Driver)
	}

	// Setup auth
	if app.Auth != nil {
		s.jwt = &JWTConfig{
			Secret: resolveEnv(strings.TrimPrefix(app.Auth.Secret, "$"), ""),
			Expire: parseDuration(app.Auth.Expire),
		}
	}

	// Register API routes
	for _, route := range app.APIs {
		s.registerAPIRoute(route)
	}

	// Register frontend pages (SSR)
	for _, page := range app.Pages {
		s.registerPageRoute(page)
	}

	// Static files
	s.mux.HandleFunc("/static/", s.handleStatic)

	return s, nil
}

func (s *Server) Start() error {
	addr := ":" + s.port
	log.Printf("[aiplang] Server running on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.middlewareChain(s.mux))
}

// ── Middleware ─────────────────────────────────────────────────────

func (s *Server) middlewareChain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == "OPTIONS" { w.WriteHeader(204); return }

		// Request ID + logging
		start := time.Now()
		log.Printf("[aiplang] %s %s", r.Method, r.URL.Path)

		// Inject auth user into context if token present
		if s.jwt != nil {
			if token := extractBearerToken(r); token != "" {
				if user, err := s.parseJWT(token); err == nil {
					ctx := context.WithValue(r.Context(), contextKeyUser, user)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
		log.Printf("[aiplang] %s %s %dms", r.Method, r.URL.Path, time.Since(start).Milliseconds())
	})
}

// ── API Route Registration ─────────────────────────────────────────

func (s *Server) registerAPIRoute(route fc.APIRoute) {
	pattern := route.Path
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != route.Method {
			http.Error(w, "Method not allowed", 405)
			return
		}

		ctx := &RequestCtx{
			W:      w,
			R:      r,
			DB:     s.db,
			Server: s,
			Vars:   map[string]interface{}{},
		}

		// Extract URL params (:id → ctx.Vars["id"])
		extractURLParams(pattern, r.URL.Path, ctx)

		// Parse body
		if r.Method != "GET" && r.Method != "DELETE" {
			if err := json.NewDecoder(r.Body).Decode(&ctx.Body); err != nil {
				ctx.Body = map[string]interface{}{}
			}
		}

		// Parse query params
		for _, qp := range route.QueryParams {
			val := r.URL.Query().Get(qp.Name)
			if val == "" { val = qp.Default }
			ctx.Vars[qp.Name] = val
		}

		// Execute guards
		for _, guard := range route.Guards {
			if !s.checkGuard(guard, ctx) {
				return
			}
		}

		// Execute validations
		for _, v := range route.Validate {
			if !validateField(v, ctx) {
				return
			}
		}

		// Execute body operations
		for _, op := range route.Body {
			if !s.execBodyOp(op, ctx) {
				return
			}
		}

		// Return response
		s.execReturn(route.Return, ctx)
	})
}

// ── Request Context ────────────────────────────────────────────────

type RequestCtx struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *gorm.DB
	Server *Server
	Body   map[string]interface{}
	Vars   map[string]interface{}
	User   map[string]interface{}
}

func (c *RequestCtx) JSON(status int, v interface{}) {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(status)
	json.NewEncoder(c.W).Encode(v)
}

func (c *RequestCtx) Error(status int, msg string) {
	c.JSON(status, map[string]string{"error": msg})
}

// ── Guards ─────────────────────────────────────────────────────────

type contextKey string
const contextKeyUser = contextKey("user")

func (s *Server) checkGuard(guard string, ctx *RequestCtx) bool {
	switch guard {
	case "auth":
		user := ctx.R.Context().Value(contextKeyUser)
		if user == nil {
			ctx.Error(401, "Unauthorized")
			return false
		}
		ctx.User = user.(map[string]interface{})
		return true
	case "admin":
		if ctx.User == nil {
			ctx.Error(401, "Unauthorized")
			return false
		}
		if ctx.User["role"] != "admin" {
			ctx.Error(403, "Forbidden")
			return false
		}
		return true
	}
	return true
}

// ── Validation ─────────────────────────────────────────────────────

func validateField(v fc.Validation, ctx *RequestCtx) bool {
	val, _ := ctx.Body[v.Field]
	valStr := fmt.Sprintf("%v", val)

	for _, rule := range v.Rules {
		switch {
		case rule == "required":
			if val == nil || valStr == "" {
				ctx.Error(422, fmt.Sprintf("%s is required", v.Field))
				return false
			}
		case rule == "email":
			if !strings.Contains(valStr, "@") {
				ctx.Error(422, fmt.Sprintf("%s must be a valid email", v.Field))
				return false
			}
		case strings.HasPrefix(rule, "min="):
			var min int
			fmt.Sscanf(rule[4:], "%d", &min)
			if len(valStr) < min {
				ctx.Error(422, fmt.Sprintf("%s must be at least %d characters", v.Field, min))
				return false
			}
		case strings.HasPrefix(rule, "max="):
			var max int
			fmt.Sscanf(rule[4:], "%d", &max)
			if len(valStr) > max {
				ctx.Error(422, fmt.Sprintf("%s must be at most %d characters", v.Field, max))
				return false
			}
		}
	}
	return true
}

// ── Body Operations ────────────────────────────────────────────────

func (s *Server) execBodyOp(op fc.BodyOp, ctx *RequestCtx) bool {
	switch op.Op {

	case "hash":
		if val, ok := ctx.Body[op.Fields]; ok {
			hash, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%v", val)), 12)
			if err == nil {
				ctx.Body[op.Fields] = string(hash)
			}
		}

	case "check":
		// ~check password $body.password $user.password | 401
		plain := resolveVar(op.Binding, ctx)
		hashed := resolveVar(op.Model, ctx)
		err := bcrypt.CompareHashAndPassword(
			[]byte(fmt.Sprintf("%v", hashed)),
			[]byte(fmt.Sprintf("%v", plain)),
		)
		if err != nil {
			status := 401
			if op.Cond != "" { fmt.Sscanf(op.Cond, "%d", &status) }
			ctx.Error(status, "Invalid credentials")
			return false
		}

	case "unique":
		// ~unique User email $body.email | 409
		if s.db == nil { break }
		val := resolveVar(op.Binding, ctx)
		var count int64
		s.db.Table(toTableName(op.Model)).Where(fmt.Sprintf("%s = ?", op.Fields), val).Count(&count)
		if count > 0 {
			status := 409
			if op.Cond != "" { fmt.Sscanf(op.Cond, "%d", &status) }
			ctx.Error(status, fmt.Sprintf("%s already exists", op.Fields))
			return false
		}

	case "assign":
		// $user = User.findBy(email=$body.email)
		result := s.execExpression(op.Fields, ctx)
		varName := strings.TrimPrefix(op.Binding, "$")
		ctx.Vars[varName] = result

	case "insert":
		if s.db == nil { break }
		data := prepareInsert(ctx.Body)
		data["id"] = uuid.New().String()
		data["created_at"] = time.Now()
		data["updated_at"] = time.Now()
		result := s.db.Table(toTableName(op.Model)).Create(data)
		if result.Error != nil {
			ctx.Error(500, "Database error")
			return false
		}
		ctx.Vars["inserted"] = data

	case "update":
		if s.db == nil { break }
		id := ctx.Vars["id"]
		data := prepareInsert(ctx.Body)
		data["updated_at"] = time.Now()
		delete(data, "id")
		delete(data, "password") // don't update password via PUT unless explicitly
		s.db.Table(toTableName(op.Model)).Where("id = ?", id).Updates(data)
		// Return updated record
		var updated map[string]interface{}
		s.db.Table(toTableName(op.Model)).Where("id = ?", id).First(&updated)
		ctx.Vars["updated"] = updated

	case "delete":
		if s.db == nil { break }
		id := ctx.Vars["id"]
		s.db.Table(toTableName(op.Model)).Where("id = ?", id).Delete(nil)
	}

	return true
}

func (s *Server) execExpression(expr string, ctx *RequestCtx) interface{} {
	if s.db == nil { return nil }

	// User.findBy(email=$body.email)
	if strings.Contains(expr, ".findBy(") {
		parts := strings.SplitN(expr, ".findBy(", 2)
		model := parts[0]
		args := strings.TrimSuffix(parts[1], ")")
		// Parse "field=value"
		kv := strings.SplitN(args, "=", 2)
		if len(kv) == 2 {
			field := strings.TrimSpace(kv[0])
			val := resolveVar(strings.TrimSpace(kv[1]), ctx)
			var result map[string]interface{}
			s.db.Table(toTableName(model)).Where(fmt.Sprintf("%s = ?", field), val).First(&result)
			return result
		}
	}

	// User.find($id) or User.find($params.id)
	if strings.Contains(expr, ".find(") {
		parts := strings.SplitN(expr, ".find(", 2)
		model := parts[0]
		idExpr := strings.TrimSuffix(parts[1], ")")
		id := resolveVar(idExpr, ctx)
		var result map[string]interface{}
		s.db.Table(toTableName(model)).Where("id = ?", id).First(&result)
		return result
	}

	// User.all(...)
	if strings.Contains(expr, ".all(") {
		parts := strings.SplitN(expr, ".all(", 2)
		model := parts[0]
		var results []map[string]interface{}
		query := s.db.Table(toTableName(model))
		// Parse args: limit=N offset=N order=field dir
		args := strings.TrimSuffix(parts[1], ")")
		for _, arg := range strings.Split(args, ",") {
			arg = strings.TrimSpace(arg)
			if strings.HasPrefix(arg, "limit=") {
				query = query.Limit(resolveInt(arg[6:], ctx, 20))
			} else if strings.HasPrefix(arg, "offset=") {
				query = query.Offset(resolveInt(arg[7:], ctx, 0))
			} else if strings.HasPrefix(arg, "order=") {
				query = query.Order(arg[6:])
			} else if strings.HasPrefix(arg, "where=") {
				query = query.Where(arg[6:])
			}
		}
		query.Find(&results)
		return results
	}

	// jwt($user)
	if strings.HasPrefix(expr, "jwt(") {
		varName := strings.TrimSuffix(strings.TrimPrefix(expr, "jwt($"), ")")
		user, _ := ctx.Vars[varName].(map[string]interface{})
		if user == nil { user = ctx.Body }
		if s.jwt != nil {
			token, _ := s.Server.generateJWT(user)
			return map[string]interface{}{"token": token, "user": sanitizeUser(user)}
		}
	}

	return resolveVar(expr, ctx)
}

func (s *Server) execReturn(ret fc.ReturnExpr, ctx *RequestCtx) {
	if ret.Status == 204 {
		ctx.W.WriteHeader(204)
		return
	}

	status := ret.Status
	if status == 0 { status = 200 }

	result := s.execExpression(ret.Expr, ctx)
	if result == nil {
		result = ctx.Vars["inserted"]
		if result == nil { result = ctx.Vars["updated"] }
	}

	ctx.JSON(status, result)
}

// ── JWT ────────────────────────────────────────────────────────────

func (s *Server) generateJWT(user map[string]interface{}) (string, error) {
	if s.jwt == nil { return "", fmt.Errorf("no jwt config") }
	claims := jwt.MapClaims{
		"id":    user["id"],
		"email": user["email"],
		"role":  user["role"],
		"exp":   time.Now().Add(s.jwt.Expire).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwt.Secret))
}

func (s *Server) parseJWT(tokenStr string) (map[string]interface{}, error) {
	if s.jwt == nil { return nil, fmt.Errorf("no jwt config") }
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwt.Secret), nil
	})
	if err != nil || !token.Valid { return nil, err }
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok { return nil, fmt.Errorf("invalid claims") }
	return map[string]interface{}(claims), nil
}

// ── Database ────────────────────────────────────────────────────────

func connectDB(cfg *fc.DBConfig) (*gorm.DB, error) {
	dsn := resolveEnv(strings.TrimPrefix(cfg.DSN, "$"), cfg.DSN)
	gormCfg := &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "[DB] ", 0), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Warn,
		}),
	}

	switch cfg.Driver {
	case "postgres":
		return gorm.Open(postgres.Open(dsn), gormCfg)
	case "sqlite":
		return gorm.Open(sqlite.Open(dsn), gormCfg)
	default:
		return gorm.Open(sqlite.Open("./aiplang.db"), gormCfg)
	}
}

func autoMigrate(db *gorm.DB, models []fc.Model) error {
	for _, m := range models {
		sql := buildCreateTable(m)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("[aiplang] migration warning for %s: %v", m.Name, err)
		}
	}
	return nil
}

func buildCreateTable(m fc.Model) string {
	table := toTableName(m.Name)
	var cols []string

	for _, f := range m.Fields {
		col := buildColumn(f)
		if col != "" { cols = append(cols, col) }
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)", table, strings.Join(cols, ",\n  "))
}

func buildColumn(f fc.ModelField) string {
	var sb strings.Builder
	sb.WriteString(toSnakeCase(f.Name))
	sb.WriteString(" ")

	// SQL type
	switch f.Type {
	case "uuid":   sb.WriteString("TEXT")
	case "text":   sb.WriteString("TEXT")
	case "int":    sb.WriteString("INTEGER")
	case "float":  sb.WriteString("REAL")
	case "bool":   sb.WriteString("INTEGER")
	case "timestamp": sb.WriteString("DATETIME")
	case "json":   sb.WriteString("TEXT")
	case "enum":   sb.WriteString("TEXT")
	default:       sb.WriteString("TEXT")
	}

	// Modifiers
	for _, mod := range f.Modifiers {
		switch mod {
		case "pk":       sb.WriteString(" PRIMARY KEY")
		case "required": sb.WriteString(" NOT NULL")
		case "unique":   sb.WriteString(" UNIQUE")
		}
	}
	if f.Default != "" {
		sb.WriteString(fmt.Sprintf(" DEFAULT '%s'", f.Default))
	}

	return sb.String()
}

// ── Page Routes (SSR) ───────────────────────────────────────────────

func (s *Server) registerPageRoute(page fc.Page) {
	route := page.Route
	if route == "/" {
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" { http.NotFound(w, r); return }
			s.servePage(page, w, r)
		})
	} else {
		s.mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			s.servePage(page, w, r)
		})
	}
}

func (s *Server) servePage(page fc.Page, w http.ResponseWriter, r *http.Request) {
	html := renderPageHTML(page, s.app.Pages)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(html))
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimPrefix(r.URL.Path, "/static/")
	switch file {
	case "aiplang-hydrate.js":
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, "./aiplang-hydrate.js")
	}
}

// ── Helpers ────────────────────────────────────────────────────────

func resolveEnv(name, def string) string {
	if v := os.Getenv(name); v != "" { return v }
	return def
}

func resolveVar(expr string, ctx *RequestCtx) interface{} {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "$body.") {
		field := expr[6:]
		return ctx.Body[field]
	}
	if strings.HasPrefix(expr, "$params.") || expr == "$id" {
		field := strings.TrimPrefix(strings.TrimPrefix(expr, "$params."), "$")
		return ctx.Vars[field]
	}
	if strings.HasPrefix(expr, "$") {
		return ctx.Vars[expr[1:]]
	}
	if strings.HasPrefix(expr, "$auth.") {
		field := expr[6:]
		if ctx.User != nil { return ctx.User[field] }
	}
	return expr
}

func resolveInt(expr string, ctx *RequestCtx, def int) int {
	val := resolveVar(expr, ctx)
	if val == nil { return def }
	var i int
	fmt.Sscanf(fmt.Sprintf("%v", val), "%d", &i)
	if i == 0 { return def }
	return i
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") { return h[7:] }
	// Also check cookie
	if c, err := r.Cookie("aiplang_token"); err == nil { return c.Value }
	return ""
}

func extractURLParams(pattern, path string, ctx *RequestCtx) {
	pParts := strings.Split(pattern, "/")
	uParts := strings.Split(path, "/")
	for i, p := range pParts {
		if strings.HasPrefix(p, ":") && i < len(uParts) {
			key := p[1:]
			ctx.Vars[key] = uParts[i]
		}
	}
}

func toTableName(model string) string {
	return toSnakeCase(model) + "s"
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' { result.WriteByte('_') }
		result.WriteRune(r | 0x20) // tolower
	}
	return result.String()
}

func prepareInsert(body map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range body { result[toSnakeCase(k)] = v }
	return result
}

func sanitizeUser(user map[string]interface{}) map[string]interface{} {
	safe := make(map[string]interface{})
	for k, v := range user {
		if k != "password" { safe[k] = v }
	}
	return safe
}

func parseDuration(s string) time.Duration {
	if strings.HasSuffix(s, "d") {
		var days int
		fmt.Sscanf(s, "%d", &days)
		return time.Duration(days) * 24 * time.Hour
	}
	d, _ := time.ParseDuration(s)
	if d == 0 { return 7 * 24 * time.Hour }
	return d
}
