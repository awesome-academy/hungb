package admin

import (
	"encoding/json"
	"errors"
	"fmt"
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

type TourHandler struct {
	service    *services.TourService
	catService *services.CategoryService
}

func NewTourHandler(service *services.TourService, catService *services.CategoryService) *TourHandler {
	return &TourHandler{service: service, catService: catService}
}

func (h *TourHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	filter := repository.TourFilter{
		Status: c.Query("status"),
		Search: c.Query("search"),
		Page:   page,
		Limit:  constants.DefaultPageLimit,
	}
	if catID, err := strconv.ParseUint(c.Query("category_id"), 10, 64); err == nil {
		filter.CategoryID = uint(catID)
	}

	tours, total, err := h.service.ListTours(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogAdminTourListFailed, "error", err)
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

	categories, _ := h.catService.AllFlatCategories(c.Request.Context())

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/tours_list.html", gin.H{
		"title":       messages.TitleAdminTours,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"tours":       tours,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"filter":      filter,
		"categories":  categories,
	})
}

func (h *TourHandler) CreateForm(c *gin.Context) {
	categories, _ := h.catService.AllFlatCategories(c.Request.Context())
	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/tour_form.html", gin.H{
		"title":       messages.TitleAdminTourCreate,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"categories": categories,
		"is_edit":    false,
		"form_url":   constants.RouteAdminTourCreate,
	})
}

func (h *TourHandler) Create(c *gin.Context) {
	var form services.TourForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, constants.RouteAdminTourCreate)
		return
	}

	if err := h.service.CreateTour(c.Request.Context(), &form); err != nil {
		slog.Error(messages.LogAdminTourCreateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminTourCreateFail)
		}
		c.Redirect(http.StatusFound, constants.RouteAdminTourCreate)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminTourCreated)
	c.Redirect(http.StatusFound, constants.RouteAdminTours)
}

func (h *TourHandler) EditForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	tour, err := h.service.GetTour(c.Request.Context(), uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminTourNotFound,
		})
		return
	}

	categories, _ := h.catService.AllFlatCategories(c.Request.Context())

	var selectedCatIDs []uint
	for _, cat := range tour.Categories {
		selectedCatIDs = append(selectedCatIDs, cat.ID)
	}

	var imageURLs []string
	if len(tour.Images) > 0 {
		_ = json.Unmarshal(tour.Images, &imageURLs)
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/tour_form.html", gin.H{
		"title":       messages.TitleAdminTourEdit,
		"active_menu": "tours",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"categories":       categories,
		"is_edit":          true,
		"tour":             tour,
		"form_url":         fmt.Sprintf(constants.RouteAdminTourEdit, id),
		"selected_cat_ids": selectedCatIDs,
		"image_urls":       imageURLs,
	})
}

func (h *TourHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	var form services.TourForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourEdit, id))
		return
	}

	if err := h.service.UpdateTour(c.Request.Context(), uint(id), &form); err != nil {
		slog.Error(messages.LogAdminTourUpdateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminTourUpdateFail)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminTourEdit, id))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminTourUpdated)
	c.Redirect(http.StatusFound, constants.RouteAdminTours)
}

func (h *TourHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	if err := h.service.DeleteTour(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminTourDeleteFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminTourDeleteFail)
		}
		c.Redirect(http.StatusFound, constants.RouteAdminTours)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminTourDeleted)
	c.Redirect(http.StatusFound, constants.RouteAdminTours)
}
