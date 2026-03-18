package admin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service *services.BookingService
}

func NewBookingHandler(service *services.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	filter := repository.BookingFilter{
		Status: c.Query("status"),
		Page:   page,
		Limit:  constants.DefaultPageLimit,
	}
	if uid, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		filter.UserID = uint(uid)
	}
	if tid, err := strconv.ParseUint(c.Query("tour_id"), 10, 64); err == nil {
		filter.TourID = uint(tid)
	}
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

	bookings, total, err := h.service.ListAllBookings(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogAdminBookingListFailed, "error", err)
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

	c.HTML(http.StatusOK, "admin/pages/bookings_list.html", gin.H{
		"title":       messages.TitleAdminBookings,
		"active_menu": "bookings",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"bookings":    bookings,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"filter":      filter,
	})
}

func (h *BookingHandler) Confirm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminBookingNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	if err := h.service.AdminConfirmBooking(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminBookingConfirmFailed, "booking_id", id, "error", err)
		errMsg := messages.ErrAdminBookingConfirmFail
		if errors.Is(err, appErrors.ErrBookingNotFound) {
			errMsg = messages.ErrAdminBookingNotFound
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminBookingConfirmed)
	c.Redirect(http.StatusFound, constants.RouteAdminBookings)
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminBookingNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	if err := h.service.AdminCancelBooking(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminBookingCancelFailed, "booking_id", id, "error", err)
		errMsg := messages.ErrAdminBookingCancelFail
		if errors.Is(err, appErrors.ErrBookingNotFound) {
			errMsg = messages.ErrAdminBookingNotFound
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminBookingCancelled)
	c.Redirect(http.StatusFound, constants.RouteAdminBookings)
}

func (h *BookingHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminBookingNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	if err := h.service.AdminCompleteBooking(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminBookingCompleteFailed, "booking_id", id, "error", err)
		errMsg := messages.ErrAdminBookingCompleteFail
		if errors.Is(err, appErrors.ErrBookingNotFound) {
			errMsg = messages.ErrAdminBookingNotFound
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteAdminBookings)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminBookingCompleted)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?status=%s", constants.RouteAdminBookings, constants.BookingStatusCompleted))
}
