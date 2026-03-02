package admin

import (
	"fmt"
	"log/slog"
	"net/http"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminAuthHandler struct {
	authService *services.AuthService
}

func NewAdminAuthHandler(authService *services.AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

func (h *AdminAuthHandler) LoginForm(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user != nil && user.Role == constants.RoleAdmin {
		c.Redirect(http.StatusFound, constants.RouteAdminDashboard)
		return
	}
	c.HTML(http.StatusOK, "admin/pages/login.html", gin.H{
		"title":      messages.TitleAdminLogin,
		"csrf_token": middleware.CSRFToken(c),
	})
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user != nil && user.Role == constants.RoleAdmin {
		c.Redirect(http.StatusFound, constants.RouteAdminDashboard)
		return
	}

	var form services.LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "admin/pages/login.html", gin.H{
			"title":      messages.TitleAdminLogin,
			"csrf_token": middleware.CSRFToken(c),
			"error":      messages.ErrInvalidForm,
			"form":       form,
		})
		return
	}

	user, err := h.authService.AdminLogin(c.Request.Context(), &form)
	if err != nil {
		var errMsg string
		switch {
		case appErrors.Is(err, appErrors.ErrForbidden):
			errMsg = messages.ErrAdminNoAccess
		case appErrors.Is(err, services.ErrAccountBanned):
			errMsg = messages.ErrAdminAccountBanned
		case appErrors.Is(err, services.ErrAccountInactive):
			errMsg = messages.ErrAdminAccountInactive
		case appErrors.Is(err, appErrors.ErrInvalidCredentials):
			errMsg = messages.ErrAdminInvalidCreds
		default:
			slog.ErrorContext(c.Request.Context(), messages.LogAdminLoginUnexpected, "error", err)
			errMsg = messages.ErrInternalServer
		}
		c.HTML(http.StatusUnprocessableEntity, "admin/pages/login.html", gin.H{
			"title":      messages.TitleAdminLogin,
			"csrf_token": middleware.CSRFToken(c),
			"error":      errMsg,
			"form":       form,
		})
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLoginSetSessionFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "admin/pages/login.html", gin.H{
			"title":      messages.TitleAdminLogin,
			"csrf_token": middleware.CSRFToken(c),
			"error":      messages.ErrAdminCreateSession,
			"form":       form,
		})
		return
	}

	middleware.SetFlashSuccess(c, fmt.Sprintf(messages.MsgAdminLoginWelcome, user.FullName))
	c.Redirect(http.StatusFound, constants.RouteAdminDashboard)
}

func (h *AdminAuthHandler) Logout(c *gin.Context) {
	if err := middleware.ClearSession(c); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLogoutClearSessionFailed, "error", err)
		middleware.ExpireSessionCookie(c)
	}
	middleware.SetFlashSuccess(c, messages.MsgAdminLogout)
	c.Redirect(http.StatusFound, constants.RouteAdminLogin)
}
