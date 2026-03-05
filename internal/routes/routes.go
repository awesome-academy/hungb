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
	authService := services.NewAuthService(db, userRepo, socialAcctRepo)

	setupPublicRoutes(router, db, authService, cfg, userRepo)
	setupAdminRoutes(router, db, authService)
}

func setupPublicRoutes(router *gin.Engine, db *gorm.DB, authService *services.AuthService, cfg *config.Config, userRepo repository.UserRepo) {
	authHandler := publicHandlers.NewAuthHandler(authService, cfg)

	// Profile & Bank Account services
	profileService := services.NewProfileService(userRepo)
	profileHandler := publicHandlers.NewProfileHandler(profileService)

	bankAccountRepo := repository.NewBankAccountRepository(db)
	bankAccountService := services.NewBankAccountService(db, bankAccountRepo)
	bankAccountHandler := publicHandlers.NewBankAccountHandler(bankAccountService)

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
		auth.GET("/profile", profileHandler.Show)
		auth.GET("/profile/edit", profileHandler.Edit)
		auth.POST("/profile/edit", profileHandler.Update)

		auth.GET("/bank-accounts", bankAccountHandler.List)
		auth.GET("/bank-accounts/create", bankAccountHandler.CreateForm)
		auth.POST("/bank-accounts/create", bankAccountHandler.Create)
		auth.GET("/bank-accounts/:id/edit", bankAccountHandler.EditForm)
		auth.POST("/bank-accounts/:id/edit", bankAccountHandler.Update)
		auth.POST("/bank-accounts/:id/delete", bankAccountHandler.Delete)
		auth.POST("/bank-accounts/:id/set-default", bankAccountHandler.SetDefault)
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
