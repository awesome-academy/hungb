package routes

import (
	"net/http"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/constants"
	adminHandlers "sun-booking-tours/internal/handlers/admin"
	publicHandlers "sun-booking-tours/internal/handlers/public"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	router.GET("/health", healthCheck)

	router.Use(middleware.LoadUser(db))

	userRepo := repository.NewUserRepository(db)
	socialAcctRepo := repository.NewSocialAccountRepository(db)
	authService := services.NewAuthService(userRepo, socialAcctRepo)

	setupPublicRoutes(router, authService, cfg)
	setupAdminRoutes(router, db, authService)
}

func setupPublicRoutes(router *gin.Engine, authService *services.AuthService, cfg *config.Config) {
	authHandler := publicHandlers.NewAuthHandler(authService)

	public := router.Group("/")
	{
		public.GET("/", homePage)
		public.GET("/register", authHandler.RegisterForm)
		public.POST("/register", authHandler.Register)
		public.GET("/login", authHandler.LoginForm)
		public.POST("/login", authHandler.Login)
		public.POST("/logout", authHandler.Logout)
	}

	// OAuth routes (exempt from CSRF — provider handles security via state param)
	oauthHandler := publicHandlers.NewOAuthHandler(authService, cfg)
	oauth := router.Group("/auth")
	{
		oauth.GET("/:provider", oauthHandler.Begin)
		oauth.GET("/:provider/callback", oauthHandler.Callback)
	}

	// Protected public routes (requires login)
	auth := public.Group("/", middleware.RequireLogin())
	{
		_ = auth // TODO: add protected public routes (profile, bookings, etc.)
	}
}

func setupAdminRoutes(router *gin.Engine, db *gorm.DB, authService *services.AuthService) {
	statsRepo := repository.NewStatsRepository(db)
	statsService := services.NewStatsService(statsRepo)
	dashboardHandler := adminHandlers.NewDashboardHandler(statsService)
	adminAuthHandler := adminHandlers.NewAdminAuthHandler(authService)

	admin := router.Group("/admin")
	{
		admin.GET("/", redirectToDashboard)
		admin.GET("/login", adminAuthHandler.LoginForm)
		admin.POST("/login", adminAuthHandler.Login)
		admin.POST("/logout", adminAuthHandler.Logout)
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
		"title":         messages.TitleHome,
		"user":          middleware.GetCurrentUser(c),
		"csrf_token":    middleware.CSRFToken(c),
		"flash_success": flashSuccess,
		"flash_error":   flashError,
	})
}

func redirectToDashboard(c *gin.Context) {
	c.Redirect(http.StatusFound, constants.RouteAdminDashboard)
}
