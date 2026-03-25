package public

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService          *services.AuthService
	googleOAuthEnabled   bool
	facebookOAuthEnabled bool
	twitterOAuthEnabled  bool
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		googleOAuthEnabled:   cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "",
		facebookOAuthEnabled: cfg.FBClientID != "" && cfg.FBClientSecret != "",
		twitterOAuthEnabled:  cfg.TwitterAPIKey != "" && cfg.TwitterAPISecret != "",
	}
}

func (h *AuthHandler) RegisterForm(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, constants.RouteHome)
		return
	}
	flashSuccess, flashError := middleware.GetFlash(c)
	c.HTML(http.StatusOK, "public/pages/register.html", gin.H{
		"title":                messages.TitleRegister,
		"csrf_token":           middleware.CSRFToken(c),
		"flash_success":        flashSuccess,
		"flash_error":          flashError,
		"GoogleOAuthEnabled":   h.googleOAuthEnabled,
		"FacebookOAuthEnabled": h.facebookOAuthEnabled,
		"TwitterOAuthEnabled":  h.twitterOAuthEnabled,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, constants.RouteHome)
		return
	}

	var form services.RegisterForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/register.html", gin.H{
			"title":      messages.TitleRegister,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     translateBindErrors(err),
			"form":       form,
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &form)
	if err != nil {
		var errMsg string
		switch {
		case appErrors.Is(err, appErrors.ErrEmailAlreadyTaken):
			errMsg = messages.ErrEmailTaken
		case appErrors.Is(err, appErrors.ErrInternalServerError):
			errMsg = messages.ErrInternalServer
		default:
			slog.ErrorContext(c.Request.Context(), messages.LogRegisterUnexpectedError, "error", err)
			errMsg = messages.ErrInternalServer
		}
		c.HTML(http.StatusUnprocessableEntity, "public/pages/register.html", gin.H{
			"title":      messages.TitleRegister,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{errMsg},
			"form":       form,
		})
		return
	}

	if h.authService.EmailVerificationRequired() {
		c.HTML(http.StatusOK, "public/pages/register_success.html", gin.H{
			"title": messages.TitleRegisterCheckEmail,
			"email": user.Email,
		})
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLoginSetSessionFailed, "error", err)
		middleware.SetFlashError(c, messages.MsgRegisterAutoLoginFail)
		c.Redirect(http.StatusFound, constants.RouteLogin)
		return
	}

	middleware.SetFlashSuccess(c, fmt.Sprintf(messages.MsgRegisterSuccess, user.FullName))
	c.Redirect(http.StatusFound, constants.RouteHome)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	user, err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		var errMsg string
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			errMsg = appErr.Message
		} else {
			slog.ErrorContext(c.Request.Context(), messages.LogVerifyEmailFailed, "error", err)
			errMsg = messages.ErrVerifyFailed
		}
		c.HTML(http.StatusBadRequest, "public/pages/verify_result.html", gin.H{
			"title":   messages.TitleVerifyEmail,
			"success": false,
			"error":   errMsg,
		})
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLoginSetSessionFailed, "error", err)
	}

	c.HTML(http.StatusOK, "public/pages/verify_result.html", gin.H{
		"title":   messages.TitleVerifyEmail,
		"success": true,
		"message": fmt.Sprintf(messages.MsgVerifySuccess, user.FullName),
		"user":    user,
	})
}

func (h *AuthHandler) LoginForm(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, constants.RouteHome)
		return
	}
	flashSuccess, flashError := middleware.GetFlash(c)
	c.HTML(http.StatusOK, "public/pages/login.html", gin.H{
		"title":               messages.TitleLogin,
		"csrf_token":          middleware.CSRFToken(c),
		"flash_success":       flashSuccess,
		"flash_error":         flashError,
		"GoogleAuthEnabled":   h.googleOAuthEnabled,
		"FacebookAuthEnabled": h.facebookOAuthEnabled,
		"TwitterAuthEnabled":  h.twitterOAuthEnabled,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, constants.RouteHome)
		return
	}

	var form services.LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/login.html", gin.H{
			"title":      messages.TitleLogin,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     translateBindErrors(err),
			"form":       form,
		})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), &form)
	if err != nil {
		var errMsg string
		switch {
		case appErrors.Is(err, services.ErrAdminMustUsePortal):
			middleware.SetFlashError(c, messages.ErrAdminMustUsePortal)
			c.Redirect(http.StatusFound, constants.RouteAdminLogin)
			return
		case appErrors.Is(err, services.ErrAccountBanned):
			errMsg = messages.ErrAccountBanned
		case appErrors.Is(err, services.ErrAccountInactive):
			errMsg = messages.ErrAccountNotActive
		case appErrors.Is(err, appErrors.ErrInvalidCredentials):
			errMsg = messages.ErrInvalidCredentials
		case appErrors.Is(err, appErrors.ErrInternalServerError):
			errMsg = messages.ErrInternalServer
		default:
			slog.ErrorContext(c.Request.Context(), messages.LogLoginUnexpectedError, "error", err)
			errMsg = messages.ErrTryAgain
		}
		c.HTML(http.StatusUnprocessableEntity, "public/pages/login.html", gin.H{
			"title":      messages.TitleLogin,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{errMsg},
			"form":       form,
		})
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLoginSetSessionFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "public/pages/login.html", gin.H{
			"title":      messages.TitleLogin,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{messages.ErrCreateSessionFail},
			"form":       form,
		})
		return
	}

	middleware.SetFlashSuccess(c, fmt.Sprintf(messages.MsgLoginWelcomeBack, user.FullName))
	c.Redirect(http.StatusFound, constants.RouteHome)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if err := middleware.ClearSession(c); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLogoutClearSessionFailed, "error", err)
		middleware.ExpireSessionCookie(c)
	}
	middleware.SetFlashSuccess(c, messages.MsgLogoutSuccess)
	c.Redirect(http.StatusFound, constants.RouteHome)
}

// translateBindErrors converts go-playground/validator errors into Vietnamese messages.
func translateBindErrors(err error) []string {
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return []string{messages.ErrInvalidForm}
	}

	fieldLabels := map[string]string{
		"FullName":        messages.FieldFullName,
		"Email":           messages.FieldEmail,
		"Password":        messages.FieldPassword,
		"PasswordConfirm": messages.FieldPasswordConfirm,
	}

	msgs := make([]string, 0, len(valErrs))
	for _, fe := range valErrs {
		label := fe.Field()
		if vn, ok := fieldLabels[fe.Field()]; ok {
			label = vn
		}
		var msg string
		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf(messages.ValRequired, label)
		case "email":
			msg = fmt.Sprintf(messages.ValEmail, label)
		case "min":
			msg = fmt.Sprintf(messages.ValMin, label, fe.Param())
		case "max":
			msg = fmt.Sprintf(messages.ValMax, label, fe.Param())
		default:
			msg = fmt.Sprintf(messages.ValInvalid, label)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}
