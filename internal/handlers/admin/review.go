package admin

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service *services.ReviewService
}

func NewReviewHandler(service *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	filter := repository.ReviewFilter{
		Status:  c.Query("status"),
		Type:    c.Query("type"),
		Keyword: c.Query("keyword"),
		Page:    page,
		Limit:   constants.DefaultPageLimit,
	}
	if uid, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		filter.UserID = uint(uid)
	}

	reviews, total, err := h.service.AdminListReviews(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogAdminReviewListFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "admin/pages/error.html", gin.H{
			"status":  500,
			"message": messages.ErrInternalServer,
		})
		return
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/reviews_list.html", gin.H{
		"title":       messages.TitleAdminReviews,
		"active_menu": "reviews",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"reviews":     reviews,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"filter":      filter,
	})
}

func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminReviewNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminReviews)
		return
	}

	if err := h.service.AdminApproveReview(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminReviewApproveFailed, "review_id", id, "error", err)
		errMsg := messages.ErrAdminReviewApproveFail
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			errMsg = messages.ErrAdminReviewNotFound
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteAdminReviews)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminReviewApproved)
	c.Redirect(http.StatusFound, constants.RouteAdminReviews)
}

func (h *ReviewHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminReviewNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminReviews)
		return
	}

	if err := h.service.AdminRejectReview(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminReviewRejectFailed, "review_id", id, "error", err)
		errMsg := messages.ErrAdminReviewRejectFail
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			errMsg = messages.ErrAdminReviewNotFound
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteAdminReviews)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminReviewRejected)
	c.Redirect(http.StatusFound, constants.RouteAdminReviews)
}
