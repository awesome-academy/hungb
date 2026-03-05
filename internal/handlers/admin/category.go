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
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) List(c *gin.Context) {
	trees, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		slog.Error(messages.LogAdminCategoryListFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "admin/pages/error.html", gin.H{
			"status":  500,
			"message": messages.ErrInternalServer,
		})
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/categories_list.html", gin.H{
		"title":       messages.TitleAdminCategories,
		"active_menu": "categories",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"categories": trees,
	})
}

func (h *CategoryHandler) CreateForm(c *gin.Context) {
	parents, err := h.service.AllFlatCategories(c.Request.Context())
	if err != nil {
		slog.Error(messages.LogAdminCategoryListFailed, "error", err)
	}

	rootCats := filterRootCategories(parents, 0)
	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/category_form.html", gin.H{
		"title":       messages.TitleAdminCategoryCreate,
		"active_menu": "categories",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"parents":  rootCats,
		"is_edit":  false,
		"form_url": constants.RouteAdminCategoryCreate,
	})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var form services.CategoryForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, constants.RouteAdminCategoryCreate)
		return
	}

	if err := h.service.CreateCategory(c.Request.Context(), &form); err != nil {
		slog.Error(messages.LogAdminCategoryCreateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminCategoryCreateFail)
		}
		c.Redirect(http.StatusFound, constants.RouteAdminCategoryCreate)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminCategoryCreated)
	c.Redirect(http.StatusFound, constants.RouteAdminCategories)
}

func (h *CategoryHandler) EditForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	cat, err := h.service.GetCategory(c.Request.Context(), uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminCategoryNotFound,
		})
		return
	}

	parents, _ := h.service.AllFlatCategories(c.Request.Context())
	rootCats := filterRootCategories(parents, uint(id))

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/category_form.html", gin.H{
		"title":       messages.TitleAdminCategoryEdit,
		"active_menu": "categories",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"parents":  rootCats,
		"is_edit":  true,
		"category": cat,
		"form_url": fmt.Sprintf(constants.RouteAdminCategoryEdit, id),
	})
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	var form services.CategoryForm
	if err := c.ShouldBind(&form); err != nil {
		middleware.SetFlashError(c, messages.ErrInvalidForm)
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminCategoryEdit, id))
		return
	}

	if err := h.service.UpdateCategory(c.Request.Context(), uint(id), &form); err != nil {
		slog.Error(messages.LogAdminCategoryUpdateFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminCategoryUpdateFail)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf(constants.RouteAdminCategoryEdit, id))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminCategoryUpdated)
	c.Redirect(http.StatusFound, constants.RouteAdminCategories)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/pages/error.html", gin.H{
			"status":  400,
			"message": messages.ErrInvalidForm,
		})
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), uint(id)); err != nil {
		slog.Error(messages.LogAdminCategoryDeleteFailed, "error", err)
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			middleware.SetFlashError(c, appErr.Message)
		} else {
			middleware.SetFlashError(c, messages.ErrAdminCategoryDeleteFail)
		}
		c.Redirect(http.StatusFound, constants.RouteAdminCategories)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminCategoryDeleted)
	c.Redirect(http.StatusFound, constants.RouteAdminCategories)
}

type parentOption struct {
	ID   uint
	Name string
}

func filterRootCategories(cats []models.Category, excludeID uint) []parentOption {
	var opts []parentOption
	for _, c := range cats {
		if c.ParentID == nil && c.ID != excludeID {
			opts = append(opts, parentOption{ID: c.ID, Name: c.Name})
		}
	}
	return opts
}
