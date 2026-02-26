package routes

import (
	"net/http"

	adminHandlers "sun-booking-tours/internal/handlers/admin"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/health", healthCheck)

	router.Use(middleware.LoadUser(db))

	setupPublicRoutes(router)
	setupAdminRoutes(router, db)
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

func setupAdminRoutes(router *gin.Engine, db *gorm.DB) {
	// Wire admin dependencies
	statsRepo := repository.NewStatsRepository(db)
	statsService := services.NewStatsService(statsRepo)
	dashboardHandler := adminHandlers.NewDashboardHandler(statsService)

	admin := router.Group("/admin")
	{
		admin.GET("/", redirectToDashboard)
		// TODO: admin login routes (no auth required)
	}

	adminAuth := admin.Group("/", middleware.RequireAdmin())
	{
		adminAuth.GET("/dashboard", dashboardHandler.Index)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func homePage(c *gin.Context) {
	flashSuccess, flashError := middleware.GetFlash(c)
	c.HTML(http.StatusOK, "public/pages/home.html", gin.H{
		"title":         "Trang chủ",
		"user":          middleware.GetCurrentUser(c),
		"csrf_token":    middleware.CSRFToken(c),
		"flash_success": flashSuccess,
		"flash_error":   flashError,
	})
}

func redirectToDashboard(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}
