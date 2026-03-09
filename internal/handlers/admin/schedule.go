package admin

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

type ScheduleHandler struct {
	service     *services.ScheduleService
	tourService *services.TourService
}

func NewScheduleHandler(service *services.ScheduleService, tourService *services.TourService) *ScheduleHandler {
	return &ScheduleHandler{service: service, tourService: tourService}
}

func (h *ScheduleHandler) List(c *gin.Context) {
	tourID, err := strconv.ParseUint(c.Param("tour_id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	tour, err := h.tourService.GetTour(c.Request.Context(), uint(tourID))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminTourNotFound,
		})
		return
	}

	schedules, err := h.service.ListByTour(c.Request.Context(), uint(tourID))
	if err != nil {
		slog.Error(messages.LogAdminScheduleListFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "admin/pages/error.html", gin.H{
			"status":  500,
			"message": messages.ErrInternalServer,
		})
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/schedules_list.html", gin.H{
		"title":       messages.TitleAdminSchedules,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"tour":      tour,
		"schedules": schedules,
	})
}

func (h *ScheduleHandler) CreateForm(c *gin.Context) {
	tourID, err := strconv.ParseUint(c.Param("tour_id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	tour, err := h.tourService.GetTour(c.Request.Context(), uint(tourID))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminTourNotFound,
		})
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/schedule_form.html", gin.H{
		"title":       messages.TitleAdminScheduleCreate,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"tour":     tour,
		"is_edit":  false,
		"form_url": fmt.Sprintf(constants.RouteAdminTourScheduleCreate, tourID),
	})
}

func (h *ScheduleHandler) Create(c *gin.Context) {
	tourID, err := strconv.ParseUint(c.Param("tour_id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	var form services.ScheduleForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourScheduleCreate, tourID))
		return
	}
	form.TourID = uint(tourID)

	if err := h.service.CreateSchedule(c.Request.Context(), &form); err != nil {
		slog.Error(messages.LogAdminScheduleCreateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminScheduleCreateFail)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourScheduleCreate, tourID))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminScheduleCreated)
	c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourSchedules, tourID))
}

func (h *ScheduleHandler) EditForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	schedule, err := h.service.GetSchedule(c.Request.Context(), uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminScheduleNotFound,
		})
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/schedule_form.html", gin.H{
		"title":       messages.TitleAdminScheduleEdit,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"tour":     schedule.Tour,
		"schedule": schedule,
		"is_edit":  true,
		"form_url": fmt.Sprintf(constants.RouteAdminScheduleEdit, id),
	})
}

func (h *ScheduleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	schedule, err := h.service.GetSchedule(c.Request.Context(), uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminScheduleNotFound,
		})
		return
	}

	var form services.ScheduleForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminScheduleEdit, id))
		return
	}
	form.TourID = schedule.TourID

	if err := h.service.UpdateSchedule(c.Request.Context(), uint(id), &form); err != nil {
		slog.Error(messages.LogAdminScheduleUpdateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminScheduleUpdateFail)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminScheduleEdit, id))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminScheduleUpdated)
	c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourSchedules, schedule.TourID))
}

func (h *ScheduleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	schedule, err := h.service.GetSchedule(c.Request.Context(), uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminScheduleNotFound,
		})
		return
	}

	tourID := schedule.TourID

	if err := h.service.DeleteSchedule(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminScheduleDeleteFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminScheduleDeleteFail)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourSchedules, tourID))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminScheduleDeleted)
	c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourSchedules, tourID))
}
