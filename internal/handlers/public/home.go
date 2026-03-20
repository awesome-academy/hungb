package public

import (
	"log/slog"
	"net/http"

	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	tourService *services.TourService
}

func NewHomeHandler(tourService *services.TourService) *HomeHandler {
	return &HomeHandler{tourService: tourService}
}

func (h *HomeHandler) Index(c *gin.Context) {
	ctx := c.Request.Context()

	featured, err := h.tourService.GetFeaturedTours(ctx, constants.HomeFeaturedLimit)
	if err != nil {
		slog.Error(messages.LogHomeFeaturedFailed, "error", err)
	}

	latest, err := h.tourService.GetLatestTours(ctx, constants.HomeLatestLimit)
	if err != nil {
		slog.Error(messages.LogHomeLatestFailed, "error", err)
	}

	flashSuccess, flashError := middleware.GetFlash(c)
	c.HTML(http.StatusOK, "public/pages/home.html", gin.H{
		"title":          messages.TitleHome,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"nav_categories": middleware.GetNavCategories(c),
		"featured_tours": featured,
		"new_tours":      latest,
	})
}
