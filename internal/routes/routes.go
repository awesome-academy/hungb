package routes

import (
	"net/http"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/constants"
	adminHandlers "sun-booking-tours/internal/handlers/admin"
	publicHandlers "sun-booking-tours/internal/handlers/public"
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
	emailService := services.NewEmailService(cfg)
	authService := services.NewAuthService(db, userRepo, socialAcctRepo, emailService, cfg.BaseURL)
	catRepo := repository.NewCategoryRepository(db)
	tourRepo := repository.NewTourRepository(db)

	setupPublicRoutes(router, db, authService, cfg, userRepo, catRepo, tourRepo)
	setupAdminRoutes(router, db, authService, catRepo)
}

func setupPublicRoutes(router *gin.Engine, db *gorm.DB, authService *services.AuthService, cfg *config.Config, userRepo repository.UserRepo, catRepo repository.CategoryRepo, tourRepo repository.TourRepo) {
	authHandler := publicHandlers.NewAuthHandler(authService, cfg)

	profileService := services.NewProfileService(userRepo)
	profileHandler := publicHandlers.NewProfileHandler(profileService)

	bankAccountRepo := repository.NewBankAccountRepository(db)
	bankAccountService := services.NewBankAccountService(db, bankAccountRepo)
	bankAccountHandler := publicHandlers.NewBankAccountHandler(bankAccountService)

	categoryService := services.NewCategoryService(catRepo)
	tourService := services.NewTourService(tourRepo, catRepo)
	ratingRepo := repository.NewRatingRepository(db)
	ratingService := services.NewRatingService(ratingRepo, tourRepo)
	publicTourHandler := publicHandlers.NewPublicTourHandler(tourService, categoryService, ratingService)
	ratingHandler := publicHandlers.NewRatingHandler(ratingService, tourService)
	homeHandler := publicHandlers.NewHomeHandler(tourService)

	scheduleRepo := repository.NewScheduleRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingService := services.NewBookingService(db, bookingRepo, scheduleRepo)
	bookingHandler := publicHandlers.NewBookingHandler(bookingService, tourService)

	reviewRepo := repository.NewReviewRepository(db)
	likeRepo := repository.NewReviewLikeRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reviewService := services.NewReviewService(db, reviewRepo, likeRepo, commentRepo)
	reviewHandler := publicHandlers.NewReviewHandler(reviewService)

	public := router.Group("/")
	public.Use(middleware.LoadCategories(catRepo))
	{
		public.GET("/", homeHandler.Index)
		public.GET("/tours", publicTourHandler.List)
		public.GET("/tours/:slug", publicTourHandler.Detail)
		public.GET("/register", authHandler.RegisterForm)
		public.POST("/register", authHandler.Register)
		public.GET("/login", authHandler.LoginForm)
		public.POST("/login", authHandler.Login)
		public.POST("/logout", authHandler.Logout)
		public.GET(constants.RouteVerifyEmail, authHandler.VerifyEmail)

		public.GET("/reviews", reviewHandler.List)
		public.GET("/reviews/:id", reviewHandler.Detail)
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

		auth.GET("/tours/:slug/book", bookingHandler.Form)
		auth.POST("/tours/:slug/book", bookingHandler.Create)
		auth.POST("/tours/:slug/rate", ratingHandler.Rate)
		auth.GET("/my/bookings", bookingHandler.MyList)
		auth.GET("/my/bookings/:id", bookingHandler.Detail)
		auth.POST("/my/bookings/:id/cancel", bookingHandler.Cancel)

		auth.GET("/reviews/create", reviewHandler.CreateForm)
		auth.POST("/reviews/create", reviewHandler.Create)
		auth.POST("/reviews/:id/like", reviewHandler.ToggleLike)
		auth.POST("/reviews/:id/comments", reviewHandler.AddComment)
		auth.POST("/comments/:id/reply", reviewHandler.ReplyComment)
		auth.POST("/comments/:id/delete", reviewHandler.DeleteComment)
		auth.GET("/my/reviews", reviewHandler.MyList)
		auth.GET("/my/reviews/:id/edit", reviewHandler.EditForm)
		auth.POST("/my/reviews/:id/edit", reviewHandler.Update)
		auth.POST("/my/reviews/:id/delete", reviewHandler.Delete)
	}
}

func setupAdminRoutes(router *gin.Engine, db *gorm.DB, authService *services.AuthService, catRepo repository.CategoryRepo) {
	statsRepo := repository.NewStatsRepository(db)
	statsService := services.NewStatsService(statsRepo)
	dashboardHandler := adminHandlers.NewDashboardHandler(statsService)
	adminAuthHandler := adminHandlers.NewAdminAuthHandler(authService)

	categoryService := services.NewCategoryService(catRepo)
	categoryHandler := adminHandlers.NewCategoryHandler(categoryService)

	tourRepo := repository.NewTourRepository(db)
	tourService := services.NewTourService(tourRepo, catRepo)
	tourHandler := adminHandlers.NewTourHandler(tourService, categoryService)

	scheduleRepo := repository.NewScheduleRepository(db)
	scheduleService := services.NewScheduleService(scheduleRepo, tourRepo)
	scheduleHandler := adminHandlers.NewScheduleHandler(scheduleService, tourService)

	bookingRepo := repository.NewBookingRepository(db)
	bookingService := services.NewBookingService(db, bookingRepo, scheduleRepo)
	adminBookingHandler := adminHandlers.NewBookingHandler(bookingService)

	revenueRepo := repository.NewRevenueRepository(db)
	revenueService := services.NewRevenueService(revenueRepo)
	revenueHandler := adminHandlers.NewRevenueHandler(revenueService)

	reviewRepo := repository.NewReviewRepository(db)
	likeRepo := repository.NewReviewLikeRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	adminReviewService := services.NewReviewService(db, reviewRepo, likeRepo, commentRepo)
	adminReviewHandler := adminHandlers.NewReviewHandler(adminReviewService)

	userRepo := repository.NewUserRepository(db)
	adminUserService := services.NewAdminUserService(userRepo)
	adminUserHandler := adminHandlers.NewUserHandler(adminUserService)

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

		adminAuth.GET("/categories", categoryHandler.List)
		adminAuth.GET("/categories/create", categoryHandler.CreateForm)
		adminAuth.POST("/categories/create", categoryHandler.Create)
		adminAuth.GET("/categories/:id/edit", categoryHandler.EditForm)
		adminAuth.POST("/categories/:id/edit", categoryHandler.Update)
		adminAuth.POST("/categories/:id/delete", categoryHandler.Delete)

		adminAuth.GET("/tours", tourHandler.List)
		adminAuth.GET("/tours/create", tourHandler.CreateForm)
		adminAuth.POST("/tours/create", tourHandler.Create)
		adminAuth.GET("/tours/:id/edit", tourHandler.EditForm)
		adminAuth.POST("/tours/:id/edit", tourHandler.Update)
		adminAuth.POST("/tours/:id/delete", tourHandler.Delete)

		adminAuth.GET("/tours/:id/schedules", scheduleHandler.List)
		adminAuth.GET("/tours/:id/schedules/create", scheduleHandler.CreateForm)
		adminAuth.POST("/tours/:id/schedules/create", scheduleHandler.Create)
		adminAuth.GET("/schedules/:id/edit", scheduleHandler.EditForm)
		adminAuth.POST("/schedules/:id/edit", scheduleHandler.Update)
		adminAuth.POST("/schedules/:id/delete", scheduleHandler.Delete)

		adminAuth.GET("/bookings", adminBookingHandler.List)
		adminAuth.POST("/bookings/:id/confirm", adminBookingHandler.Confirm)
		adminAuth.POST("/bookings/:id/cancel", adminBookingHandler.Cancel)
		adminAuth.POST("/bookings/:id/complete", adminBookingHandler.Complete)

		adminAuth.GET("/revenue", revenueHandler.Index)

		adminAuth.GET("/reviews", adminReviewHandler.List)
		adminAuth.POST("/reviews/:id/approve", adminReviewHandler.Approve)
		adminAuth.POST("/reviews/:id/reject", adminReviewHandler.Reject)

		adminAuth.GET("/users", adminUserHandler.List)
		adminAuth.GET("/users/:id", adminUserHandler.Detail)
		adminAuth.POST("/users/:id/status", adminUserHandler.UpdateStatus)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func redirectToDashboard(c *gin.Context) {
	c.Redirect(http.StatusFound, constants.RouteAdminDashboard)
}
