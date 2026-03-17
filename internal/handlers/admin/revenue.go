package admin

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type RevenueHandler struct {
	service *services.RevenueService
}

func NewRevenueHandler(service *services.RevenueService) *RevenueHandler {
	return &RevenueHandler{service: service}
}

func (h *RevenueHandler) Index(c *gin.Context) {
	filter := parseRevenueFilter(c)

	stats, err := h.service.GetRevenueStats(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogAdminRevenueLoadFailed, "error", err)
		stats = &services.RevenueStats{}
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/revenue.html", gin.H{
		"title":       messages.TitleAdminRevenue,
		"active_menu": "revenue",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"stats":  stats,
		"filter": filter,
	})
}

func parseRevenueFilter(c *gin.Context) repository.RevenueFilter {
	var filter repository.RevenueFilter

	if df := c.Query("date_from"); df != "" {
		if t, err := time.ParseInLocation("2006-01-02", df, time.Local); err == nil {
			filter.DateFrom = t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.ParseInLocation("2006-01-02", dt, time.Local); err == nil {
			filter.DateTo = t.Add(24*time.Hour - time.Second)
		}
	}
	if tid, err := strconv.ParseUint(c.Query("tour_id"), 10, 64); err == nil {
		filter.TourID = uint(tid)
	}

	return filter
}
