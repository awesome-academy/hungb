package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/database"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func main() {
	// CLI flags for database operations
	migrateFlag := flag.Bool("migrate", false, messages.FlagMigrateDescription)
	seedFlag := flag.Bool("seed", false, messages.FlagSeedDescription)
	flag.Parse()

	//TODO: Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()
	slog.Info(messages.LogConfigurationLoaded,
		"port", cfg.Port,
		"gin_mode", cfg.GinMode,
		"db_host", cfg.DBHost,
		"db_name", cfg.DBName,
	)

	// Connect to database
	db, err := config.ConnectDB(cfg)
	if err != nil {
		slog.Error(messages.LogDatabaseConnFailed, "error", err)
		os.Exit(1)
	}

	// Handle CLI flags: migrate and/or seed, then exit
	if *migrateFlag || *seedFlag {
		if *migrateFlag {
			if err := database.Migrate(db); err != nil {
				slog.Error(messages.LogMigrationFailed, "error", err)
				os.Exit(1)
			}
		}
		if *seedFlag {
			if err := database.Seed(db); err != nil {
				slog.Error(messages.LogSeedingFailed, "error", err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}

	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// In development, use a renderer that reloads on each request.
	// In production, load once at startup for performance.
	if cfg.GinMode == gin.DebugMode {
		r.HTMLRender = &devRenderer{baseDir: "templates"}
	} else {
		r.HTMLRender = loadTemplates("templates")
	}

	r.Static("/static", "./static")

	middleware.SetupSession(r, cfg.SessionSecret)
	r.Use(middleware.CSRFMiddleware(cfg.SessionSecret))

	routes.SetupRoutes(r, db)

	// Start server
	addr := ":" + cfg.Port
	slog.Info(messages.LogStartingServer, "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error(messages.LogServerStartFailed, "error", err)
		os.Exit(1)
	}
}

// multiRenderer holds one *template.Template set per page.
// Each set contains all shared layouts/partials plus one page file,
// so {{define "content"}} blocks never collide across pages.
type multiRenderer struct {
	sets map[string]*template.Template
}

func (mr *multiRenderer) Instance(name string, data any) render.Render {
	t, ok := mr.sets[name]
	if !ok {
		slog.Warn(messages.LogTemplateNotFound, "name", name)
		t = template.Must(template.New(name).Parse(messages.TemplateNotFoundText + name))
	}
	return render.HTML{Template: t, Name: name, Data: data}
}

// devRenderer reloads all templates on every request (debug mode only).
// Unlike mutating gin.Engine.HTMLRender in middleware, this is safe under
// concurrent requests because loadTemplates returns a fresh snapshot and
// no shared mutable state is touched.
type devRenderer struct {
	baseDir string
}

func (dr *devRenderer) Instance(name string, data any) render.Render {
	fresh := loadTemplates(dr.baseDir).(*multiRenderer)
	return fresh.Instance(name, data)
}

func loadTemplates(baseDir string) render.HTMLRender {
	var sharedFiles []string
	var pageFiles []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
			return err
		}
		rel, relErr := filepath.Rel(baseDir, path)
		if relErr != nil {
			slog.Error(messages.LogTemplateRelPathFail, "base", baseDir, "path", path, "error", relErr)
			os.Exit(1)
		}
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "/pages/") {
			pageFiles = append(pageFiles, path)
		} else {
			sharedFiles = append(sharedFiles, path)
		}
		return nil
	})
	if err != nil {
		slog.Error(messages.LogTemplateWalkFailed, "error", err)
		os.Exit(1)
	}

	sets := make(map[string]*template.Template, len(pageFiles))

	for _, pagePath := range pageFiles {
		rel, relErr := filepath.Rel(baseDir, pagePath)
		if relErr != nil {
			slog.Error(messages.LogTemplateRelPathFail, "base", baseDir, "path", pagePath, "error", relErr)
			os.Exit(1)
		}
		rel = filepath.ToSlash(rel)

		// Only add template helpers that are safe by design.
		// Avoid raw HTML helpers (e.g. safeHTML) unless the input is
		// guaranteed to be sanitised — they disable auto-escaping
		// and are an XSS vector if applied to user-controlled content.
		t := template.New("").Funcs(template.FuncMap{})

		for _, sf := range sharedFiles {
			sfRel, sfRelErr := filepath.Rel(baseDir, sf)
			if sfRelErr != nil {
				slog.Error(messages.LogTemplateRelPathFail, "base", baseDir, "path", sf, "error", sfRelErr)
				os.Exit(1)
			}
			sfRel = filepath.ToSlash(sfRel)
			data, err := os.ReadFile(sf)
			if err != nil {
				slog.Error(messages.LogSharedTemplateReadFail, "path", sfRel, "error", err)
				os.Exit(1)
			}
			if _, err = t.New(sfRel).Parse(string(data)); err != nil {
				slog.Error(messages.LogSharedTemplateParseFail, "path", sfRel, "error", err)
				os.Exit(1)
			}
		}

		data, err := os.ReadFile(pagePath)
		if err != nil {
			slog.Error(messages.LogPageTemplateReadFail, "path", rel, "error", err)
			os.Exit(1)
		}
		if _, err = t.New(rel).Parse(string(data)); err != nil {
			slog.Error(messages.LogPageTemplateParseFail, "path", rel, "error", err)
			os.Exit(1)
		}

		sets[rel] = t
		slog.Debug(messages.LogLoadedPageTemplate, "name", rel)
	}

	return &multiRenderer{sets: sets}
}
