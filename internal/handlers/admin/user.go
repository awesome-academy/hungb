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
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *services.AdminUserService
}

func NewUserHandler(service *services.AdminUserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	filter := repository.UserFilter{
		Role:      c.Query("role"),
		Status:    c.Query("status"),
		Keyword:   c.Query("keyword"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Page:      page,
		Limit:     constants.DefaultPageLimit,
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogAdminUserListFailed, "error", err)
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

	c.HTML(http.StatusOK, "admin/pages/users_list.html", gin.H{
		"title":       messages.TitleAdminUsers,
		"active_menu": "users",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"users":       users,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
		"filter":      filter,
	})
}

func (h *UserHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
			"status":  404,
			"message": messages.ErrAdminUserNotFound,
		})
		return
	}

	target, err := h.service.GetUserDetail(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			c.HTML(http.StatusNotFound, "admin/pages/error.html", gin.H{
				"status":  404,
				"message": messages.ErrAdminUserNotFound,
			})
			return
		}
		slog.Error(messages.LogAdminUserDetailFailed, "user_id", id, "error", err)
		c.HTML(http.StatusInternalServerError, "admin/pages/error.html", gin.H{
			"status":  500,
			"message": messages.ErrInternalServer,
		})
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "admin/pages/user_detail.html", gin.H{
		"title":       messages.TitleAdminUserDetail,
		"active_menu": "users",
		"user":        middleware.GetCurrentUser(c),
		"csrf_token":  middleware.CSRFToken(c),

		"flash_success": flashSuccess,
		"flash_error":   flashError,

		"target": target,
	})
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrAdminUserNotFound)
		c.Redirect(http.StatusFound, constants.RouteAdminUsers)
		return
	}

	status := c.PostForm("status")
	backURL := fmt.Sprintf(constants.RouteAdminUserDetail, id)

	allowedStatuses := map[string]bool{
		constants.StatusActive:   true,
		constants.StatusInactive: true,
		constants.StatusBanned:   true,
	}
	if !allowedStatuses[status] {
		middleware.SetFlashError(c, messages.ErrAdminUserUpdateFail)
		c.Redirect(http.StatusFound, backURL)
		return
	}

	adminUser := middleware.GetCurrentUser(c)
	if err := h.service.UpdateUserStatus(c.Request.Context(), adminUser, uint(id), status); err != nil {
		errMsg := messages.ErrAdminUserUpdateFail
		switch {
		case errors.Is(err, appErrors.ErrCannotBanSelf):
			errMsg = messages.ErrAdminUserCannotBanSelf
		case errors.Is(err, appErrors.ErrCannotBanAdmin):
			errMsg = messages.ErrAdminUserCannotBanAdmin
		case errors.Is(err, appErrors.ErrUserNotFound):
			errMsg = messages.ErrAdminUserNotFound
		default:
			slog.Error(messages.LogAdminUserStatusFailed, "user_id", id, "status", status, "error", err)
		}

		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, backURL)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgAdminUserStatusUpdated)
	c.Redirect(http.StatusFound, backURL)
}
