package admin

import (
	"fmt"
	"log/slog"
	"net/http"

	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	statsService *services.StatsService
}

func NewDashboardHandler(statsService *services.StatsService) *DashboardHandler {
	return &DashboardHandler{statsService: statsService}
}

// Index renders the admin dashboard with quick stats, recent bookings and pending reviews.
func (h *DashboardHandler) Index(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats(c.Request.Context())
	if err != nil {
		// Log the error but render with zero values to avoid a hard crash
		slog.Error(messages.LogDashboardLoadStatsFailed, "error", err)
		stats = &services.DashboardStats{}
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/dashboard.html", gin.H{
		"title":       messages.TitleAdminDashboard,
		"active_menu": "dashboard",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		// Stats widgets — pre-formatted so templates need no arithmetic
		"stats_cards": []gin.H{
			{
				"icon":  "bi-people",
				"label": messages.DashboardLabelTotalUsers,
				"value": fmt.Sprintf("%d", stats.TotalUsers),
				"color": "primary",
			},
			{
				"icon":  "bi-map",
				"label": messages.DashboardLabelActiveTours,
				"value": fmt.Sprintf("%d", stats.TotalTours),
				"color": "success",
			},
			{
				"icon":  "bi-calendar-check",
				"label": messages.DashboardLabelTodayBooking,
				"value": fmt.Sprintf("%d", stats.TodayBookings),
				"color": "warning",
			},
			{
				"icon":  "bi-graph-up",
				"label": messages.DashboardLabelMonthRevenue,
				"value": fmt.Sprintf("%.0f ₫", stats.MonthRevenue),
				"color": "info",
			},
		},

		"recent_bookings": stats.RecentBookings,
		"pending_reviews": stats.PendingReviews,
	})
}
