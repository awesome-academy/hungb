package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Health check endpoint (no auth required)
	router.GET("/health", healthCheck)

	// Public Site Routes
	setupPublicRoutes(router)

	// Admin Site Routes
	setupAdminRoutes(router)
}

// setupPublicRoutes configures routes for the public site (guest + authenticated users)
func setupPublicRoutes(router *gin.Engine) {
	public := router.Group("/")
	{
		// Home page
		public.GET("/", homePage)

	}
	// TODO: Add middleware
}

func setupAdminRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	{
		// Redirect /admin to /admin/dashboard
		admin.GET("/", redirectToDashboard)

	}

	// TODO: Add middleware
	adminAuth := admin.Group("/")
	{
		// Dashboard
		adminAuth.GET("/dashboard", dashboardPage)

	}
}

// TODO: Implement handlers
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "public/pages/home.html", gin.H{
		"title": "Trang chủ",
		"user":  nil,
	})
}

func redirectToDashboard(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func dashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/pages/dashboard.html", gin.H{
		"title": "Dashboard",
		"user":  nil,
	})
}
