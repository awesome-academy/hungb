package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nguyenhungb/sun-booking-tours/internal/config"
)

func main() {
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

	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	tmpl := loadTemplates("templates")
	r.SetHTMLTemplate(tmpl)

	r.Static("/static", "./static")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public Site Routes
	public := r.Group("/")
	{
		public.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "public/pages/home.html", gin.H{
				"title": "Trang chủ",
				"user":  nil,
			})
		})
	}

	// Admin Site Routes
	admin := r.Group("/admin")
	{
		admin.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/admin/dashboard")
		})
		admin.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/pages/dashboard.html", gin.H{
				"title": "Dashboard",
				"user":  nil,
			})
		})
	}

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

		// Get relative path from baseDir for template name
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes for consistent template names
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
