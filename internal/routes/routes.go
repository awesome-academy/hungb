package routes

import (
	"net/http"

	"sun-booking-tours/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/health", healthCheck)

	router.Use(middleware.LoadUser(db))

	setupPublicRoutes(router)
	setupAdminRoutes(router)
}

func setupPublicRoutes(router *gin.Engine) {
	public := router.Group("/")
	{
		public.GET("/", homePage)
	}

	// Protected public routes (requires login)
	auth := public.Group("/", middleware.RequireLogin())
	{
		_ = auth // TODO: add protected public routes (profile, bookings, etc.)
	}
}

func setupAdminRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	{
		admin.GET("/", redirectToDashboard)
		// TODO: admin login routes (no auth required)
	}

	adminAuth := admin.Group("/", middleware.RequireAdmin())
	{
		adminAuth.GET("/dashboard", dashboardPage)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "public/pages/home.html", gin.H{
		"title":      "Trang chủ",
		"user":       middleware.GetCurrentUser(c),
		"csrf_token": middleware.CSRFToken(c),
	})
}

func redirectToDashboard(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func dashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/pages/dashboard.html", gin.H{
		"title":       "Dashboard",
		"active_menu": "dashboard",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),
	})
}
