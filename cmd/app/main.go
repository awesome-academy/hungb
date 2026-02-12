package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// CLI flags for database operations
	migrateFlag := flag.Bool("migrate", false, "Run database migration")
	seedFlag := flag.Bool("seed", false, "Seed database with initial data")
	flag.Parse()

	//TODO: Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()
	slog.Info("configuration loaded",
		"port", cfg.Port,
		"gin_mode", cfg.GinMode,
		"db_host", cfg.DBHost,
		"db_name", cfg.DBName,
	)

	// Connect to database
	db := config.ConnectDB(cfg)

	// Handle CLI flags: migrate and/or seed, then exit
	if *migrateFlag || *seedFlag {
		if *migrateFlag {
			if err := models.Migrate(db); err != nil {
				slog.Error("migration failed", "error", err)
				os.Exit(1)
			}
		}
		if *seedFlag {
			if err := models.Seed(db); err != nil {
				slog.Error("seeding failed", "error", err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}

	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// In development, reload templates on each request for hot reload
	// In production, load once at startup for performance
	if cfg.GinMode == gin.DebugMode {
		r.Use(func(c *gin.Context) {
			r.SetHTMLTemplate(loadTemplates("templates"))
			c.Next()
		})
	} else {
		tmpl := loadTemplates("templates")
		r.SetHTMLTemplate(tmpl)
	}

	r.Static("/static", "./static")

	// Setup all routes
	routes.SetupRoutes(r)

	// Start server
	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// loadTemplates recursively loads all .html template files and registers them
// with paths relative to the base directory (e.g., "public/pages/home.html").
func loadTemplates(baseDir string) *template.Template {
	tmpl := template.New("")

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = tmpl.New(relPath).Parse(string(data))
		if err != nil {
			slog.Error("failed to parse template", "path", relPath, "error", err)
			return err
		}
		slog.Debug("loaded template", "name", relPath)
		return nil
	})

	if err != nil {
		slog.Error("failed to load templates", "error", err)
		os.Exit(1)
	}

	return tmpl
}
