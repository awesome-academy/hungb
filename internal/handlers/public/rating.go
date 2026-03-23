package public

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type RatingHandler struct {
	ratingService *services.RatingService
	tourService   *services.TourService
}

func NewRatingHandler(ratingService *services.RatingService, tourService *services.TourService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService, tourService: tourService}
}

func (h *RatingHandler) Rate(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	slug := c.Param("slug")
	redirectURL := fmt.Sprintf("%s/%s", constants.RoutePublicTours, slug)

	tour, _, err := h.tourService.GetPublicTourBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, appErrors.ErrTourNotFound) {
			middleware.SetFlashError(c, messages.ErrRatingTourNotFound)
		} else {
			slog.Error("failed to get public tour by slug", "slug", slug, "error", err)
			middleware.SetFlashError(c, messages.ErrRatingFail)
		}
		c.Redirect(http.StatusFound, constants.RoutePublicTours)
		return
	}

	score, err := strconv.Atoi(c.PostForm("score"))
	if err != nil || score < constants.RatingMinScore || score > constants.RatingMaxScore {
		middleware.SetFlashError(c, messages.ErrRatingInvalid)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	input := services.RatingInput{
		Score:   score,
		Comment: c.PostForm("comment"),
	}

	isNew, err := h.ratingService.RateOrUpdate(c.Request.Context(), user.ID, tour.ID, input)
	if err != nil {
		slog.Error(messages.LogRatingFailed, "tour_id", tour.ID, "user_id", user.ID, "error", err)
		errMsg := messages.ErrRatingFail
		if errors.Is(err, appErrors.ErrInvalidScore) {
			errMsg = messages.ErrRatingInvalid
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	if isNew {
		middleware.SetFlashSuccess(c, messages.MsgRatingSuccess)
	} else {
		middleware.SetFlashSuccess(c, messages.MsgRatingUpdated)
	}
	c.Redirect(http.StatusFound, redirectURL)
}
