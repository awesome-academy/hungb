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

type BookingHandler struct {
	bookingService *services.BookingService
	tourService    *services.TourService
}

func NewBookingHandler(bookingService *services.BookingService, tourService *services.TourService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService, tourService: tourService}
}

func (h *BookingHandler) Form(c *gin.Context) {
	slug := c.Param("slug")

	tour, _, err := h.tourService.GetPublicTourBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, appErrors.ErrTourNotFound) {
			h.renderError(c, http.StatusNotFound, messages.ErrPublicTourNotFound)
			return
		}
		slog.Error(messages.LogBookingFormFailed, "slug", slug, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	if len(tour.Schedules) == 0 {
		middleware.SetFlashError(c, messages.ErrBookingNoSchedules)
		c.Redirect(http.StatusFound, fmt.Sprintf("/tours/%s", tour.Slug))
		return
	}

	selectedScheduleID, _ := strconv.ParseUint(c.Query("schedule_id"), 10, 64)
	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/booking_form.html", gin.H{
		"title":                messages.TitleBooking,
		"user":                 middleware.GetCurrentUser(c),
		"csrf_token":           middleware.CSRFToken(c),
		"nav_categories":       middleware.GetNavCategories(c),
		"flash_success":        flashSuccess,
		"flash_error":          flashError,
		"tour":                 tour,
		"selected_schedule_id": uint(selectedScheduleID),
	})
}

func (h *BookingHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	slug := c.Param("slug")

	tour, _, err := h.tourService.GetPublicTourBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, appErrors.ErrTourNotFound) {
			h.renderError(c, http.StatusNotFound, messages.ErrPublicTourNotFound)
			return
		}
		slog.Error(messages.LogBookingCreateFailed, "slug", slug, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	scheduleID, _ := strconv.ParseUint(c.PostForm("schedule_id"), 10, 64)
	numParticipants, _ := strconv.Atoi(c.PostForm("num_participants"))
	note := c.PostForm("note")

	if scheduleID == 0 || numParticipants < 1 {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, fmt.Sprintf("/tours/%s/book?schedule_id=%d", slug, scheduleID))
		return
	}

	booking, err := h.bookingService.CreateBooking(
		c.Request.Context(),
		user.ID,
		tour.ID,
		uint(scheduleID),
		numParticipants,
		tour.Price,
		note,
	)
	if err != nil {
		slog.Error(messages.LogBookingCreateFailed, "slug", slug, "schedule_id", scheduleID, "error", err)

		var appErr *appErrors.AppError
		errMsg := messages.ErrBookingCreateFail
		if errors.As(err, &appErr) {
			switch {
			case appErrors.Is(err, appErrors.ErrScheduleNotOpen):
				errMsg = messages.ErrBookingScheduleNotOpen
			case appErrors.Is(err, appErrors.ErrNotEnoughSlots):
				errMsg = messages.ErrBookingNotEnoughSlots
			case appErrors.Is(err, appErrors.ErrScheduleNotFound):
				errMsg = messages.ErrBookingScheduleNotOpen
			}
		}

		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, fmt.Sprintf("/tours/%s/book?schedule_id=%d", slug, scheduleID))
		return
	}

	fullBooking, err := h.bookingService.GetBooking(c.Request.Context(), booking.ID, user.ID)
	if err != nil {
		middleware.SetFlashSuccess(c, messages.MsgBookingSuccess)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RouteMyBookings, booking.ID))
		return
	}

	c.HTML(http.StatusOK, "public/pages/booking_confirmation.html", gin.H{
		"title":          messages.TitleBookingConfirmation,
		"user":           user,
		"nav_categories": middleware.GetNavCategories(c),
		"booking":        fullBooking,
	})
}

func (h *BookingHandler) MyList(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	bookings, total, err := h.bookingService.ListMyBookings(c.Request.Context(), user.ID, page, constants.DefaultPageLimit)
	if err != nil {
		slog.Error(messages.LogBookingListFailed, "user_id", user.ID, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	totalPages := max(1, (int(total)+constants.DefaultPageLimit-1)/constants.DefaultPageLimit)

	const pageWindow = 2
	winStart := page - pageWindow
	if winStart < 1 {
		winStart = 1
	}
	winEnd := page + pageWindow
	if winEnd > totalPages {
		winEnd = totalPages
	}
	pages := make([]int, 0, winEnd-winStart+1)
	for i := winStart; i <= winEnd; i++ {
		pages = append(pages, i)
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/my_bookings_list.html", gin.H{
		"title":          messages.TitleMyBookings,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"bookings":       bookings,
		"total":          total,
		"pagination": map[string]any{
			"Page":       page,
			"TotalPages": totalPages,
			"PrevPage":   max(1, page-1),
			"NextPage":   min(totalPages, page+1),
			"Pages":      pages,
		},
	})
}

func (h *BookingHandler) Detail(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	bookingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusBadRequest, messages.ErrBookingNotFound)
		return
	}

	booking, err := h.bookingService.GetBooking(c.Request.Context(), uint(bookingID), user.ID)
	if err != nil {
		if errors.Is(err, appErrors.ErrBookingNotFound) {
			h.renderError(c, http.StatusNotFound, messages.ErrBookingNotFound)
			return
		}
		slog.Error(messages.LogBookingGetFailed, "booking_id", bookingID, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/my_booking_detail.html", gin.H{
		"title":          messages.TitleMyBookingDetail,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"booking":        booking,
	})
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	bookingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusBadRequest, messages.ErrBookingNotFound)
		return
	}

	if err := h.bookingService.CancelBooking(c.Request.Context(), uint(bookingID), user.ID); err != nil {
		slog.Error(messages.LogBookingCancelFailed, "booking_id", bookingID, "error", err)

		errMsg := messages.ErrBookingCancelFail
		if errors.Is(err, appErrors.ErrBookingNotFound) {
			errMsg = messages.ErrBookingNotFound
		} else if errors.Is(err, appErrors.ErrBookingCannotCancel) {
			errMsg = messages.ErrBookingCannotCancel
		}

		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RouteMyBookings, bookingID))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgBookingCancelSuccess)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RouteMyBookings, bookingID))
}

func (h *BookingHandler) renderError(c *gin.Context, status int, message string) {
	c.HTML(status, "public/pages/error.html", gin.H{
		"title":          message,
		"status":         status,
		"message":        message,
		"user":           middleware.GetCurrentUser(c),
		"nav_categories": middleware.GetNavCategories(c),
	})
}
