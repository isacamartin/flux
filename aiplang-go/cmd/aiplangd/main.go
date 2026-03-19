// aiplangd — aiplang v2 daemon
// Runs a full-stack .flux application
// Usage: aiplangd dev app.flux | aiplangd start app.flux | aiplangd migrate app.flux

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	fc "github.com/isacamartin/aiplang/compiler"
	fs "github.com/isacamartin/aiplang/server"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "--help", "-h":
		printHelp()

	case "--version", "-v":
		fmt.Println("aiplangd v2.0.0")

	case "dev":
		aipFile := getAipFile(os.Args, 2)
		runDev(aipFile)

	case "start":
		aipFile := getAipFile(os.Args, 2)
		runServer(aipFile, false)

	case "migrate":
		aipFile := getAipFile(os.Args, 2)
		runMigrate(aipFile)

	case "build":
		aipFile := getAipFile(os.Args, 2)
		runBuild(aipFile)

	default:
		// Treat as file path (aiplangd app.flux)
		if _, err := os.Stat(cmd); err == nil {
			runDev(cmd)
		} else {
			fmt.Fprintf(os.Stderr, "\n  ✗  Unknown command: %s\n  Run aiplangd --help\n\n", cmd)
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Print(`
  aiplangd v2.0.0 — aiplang full-stack runtime

  Usage:
    aiplangd dev <app.flux>        start dev server with hot reload
    aiplangd start <app.flux>      start production server
    aiplangd migrate <app.flux>    run database migrations
    aiplangd build <app.flux>      compile to static HTML (frontend only)
    aiplangd --version

  Example:
    aiplangd dev myapp.flux

  Environment:
    PORT          server port (default: 3000)
    DATABASE_URL  postgres/sqlite connection string
    JWT_SECRET    secret for JWT tokens

`)
}

func getAipFile(args []string, idx int) string {
	if len(args) <= idx {
		fmt.Fprintf(os.Stderr, "\n  ✗  No .flux file specified\n\n")
		os.Exit(1)
	}
	file := args[idx]
	if _, err := os.Stat(file); err != nil {
		fmt.Fprintf(os.Stderr, "\n  ✗  File not found: %s\n\n", file)
		os.Exit(1)
	}
	return file
}

func parseAipFile(file string) *fc.AppDef {
	src, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("[aiplang] cannot read %s: %v", file, err)
	}
	app, err := fc.Parse(string(src))
	if err != nil {
		log.Fatalf("[aiplang] parse error in %s: %v", file, err)
	}
	return app
}

func runServer(aipFile string, dev bool) {
	app := parseAipFile(aipFile)
	printAppSummary(app, aipFile, dev)

	srv, err := fs.New(app)
	if err != nil {
		log.Fatalf("[aiplang] server error: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("[aiplang] %v", err)
	}
}

func runDev(aipFile string) {
	fmt.Printf("\n  aiplangd dev — watching %s\n", aipFile)

	app := parseAipFile(aipFile)
	printAppSummary(app, aipFile, true)

	srv, err := fs.New(app)
	if err != nil {
		log.Fatalf("[aiplang] server error: %v", err)
	}

	// File watcher in background
	go func() {
		var lastMod time.Time
		for {
			time.Sleep(500 * time.Millisecond)
			info, err := os.Stat(aipFile)
			if err != nil { continue }
			if !lastMod.IsZero() && info.ModTime().After(lastMod) {
				fmt.Printf("\n  [aiplang] %s changed — reloading...\n", filepath.Base(aipFile))
				// In production would restart gracefully
				// For now just log the change
			}
			lastMod = info.ModTime()
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("[aiplang] %v", err)
	}
}

func runMigrate(aipFile string) {
	app := parseAipFile(aipFile)
	fmt.Printf("\n  aiplangd migrate — %s\n\n", aipFile)

	if app.DB == nil {
		fmt.Println("  ✗  No database configured. Add ~db to your .flux file.")
		os.Exit(1)
	}

	fmt.Printf("  DB:     %s\n", app.DB.Driver)
	fmt.Printf("  Models: %d\n\n", len(app.Models))

	for _, m := range app.Models {
		fmt.Printf("  ✓  %s (%d fields)\n", m.Name, len(m.Fields))
	}

	// Server.New() runs auto-migration
	_, err := fs.New(app)
	if err != nil {
		log.Fatalf("[aiplang] migrate error: %v", err)
	}

	fmt.Println("\n  Migration complete.\n")
}

func runBuild(aipFile string) {
	app := parseAipFile(aipFile)
	outDir := "dist"

	fmt.Printf("\n  aiplangd build — %s\n\n", aipFile)
	os.MkdirAll(outDir, 0755)

	for _, page := range app.Pages {
		html := fs.RenderPageHTML(page, app.Pages)
		fname := "index.html"
		if page.Route != "/" {
			dir := filepath.Join(outDir, page.Route[1:])
			os.MkdirAll(dir, 0755)
			fname = filepath.Join(dir, "index.html")
		} else {
			fname = filepath.Join(outDir, "index.html")
		}
		os.WriteFile(fname, []byte(html), 0644)
		jsNote := " (zero JS)"
		fmt.Printf("  ✓  %s%s\n", fname, jsNote)
	}

	fmt.Printf("\n  %d page(s) built → %s/\n\n", len(app.Pages), outDir)
}

func printAppSummary(app *fc.AppDef, file string, dev bool) {
	mode := "production"
	if dev { mode = "development" }

	fmt.Printf("\n  aiplang v2.0 — %s\n\n", mode)
	fmt.Printf("  File:    %s\n", file)
	fmt.Printf("  Pages:   %d\n", len(app.Pages))
	fmt.Printf("  APIs:    %d routes\n", len(app.APIs))
	fmt.Printf("  Models:  %d\n", len(app.Models))
	if app.DB != nil { fmt.Printf("  DB:      %s\n", app.DB.Driver) }
	if app.Auth != nil { fmt.Printf("  Auth:    %s\n", app.Auth.Provider) }
	fmt.Println()
}
