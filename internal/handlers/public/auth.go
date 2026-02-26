package public

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterForm renders GET /register.
// Redirects to "/" if the user is already logged in.
func (h *AuthHandler) RegisterForm(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "public/pages/register.html", gin.H{
		"title":      messages.TitleRegister,
		"user":       nil,
		"csrf_token": middleware.CSRFToken(c),
	})
}

// Register handles POST /register.
func (h *AuthHandler) Register(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	var form services.RegisterForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/register.html", gin.H{
			"title":      messages.TitleRegister,
			"user":       nil,
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
			// Service already logged the root cause; show a safe generic message.
			errMsg = messages.ErrInternalServer
		default:
			// Unknown error — log it and show a safe fallback (no internal details).
			slog.ErrorContext(c.Request.Context(), "register: unexpected error", "error", err)
			errMsg = messages.ErrInternalServer
		}
		c.HTML(http.StatusUnprocessableEntity, "public/pages/register.html", gin.H{
			"title":      messages.TitleRegister,
			"user":       nil,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{errMsg},
			// Retain non-sensitive fields so the user doesn't retype everything.
			"form": form,
		})
		return
	}

	// Auto-login: store the new user's ID in session.
	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		middleware.SetFlashError(c, messages.MsgRegisterAutoLoginFail)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	middleware.SetFlashSuccess(c, fmt.Sprintf(messages.MsgRegisterSuccess, user.FullName))
	c.Redirect(http.StatusFound, "/")
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
