package public

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	appErrors "sun-booking-tours/internal/errors"
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
		"title":      "Đăng ký tài khoản",
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
			"title":      "Đăng ký tài khoản",
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
			errMsg = "Email đã được sử dụng. Vui lòng dùng email khác hoặc đăng nhập."
		default:
			errMsg = err.Error()
		}
		c.HTML(http.StatusUnprocessableEntity, "public/pages/register.html", gin.H{
			"title":      "Đăng ký tài khoản",
			"user":       nil,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{errMsg},
			// Retain non-sensitive fields so the user doesn't retype everything
			"form": form,
		})
		return
	}

	// Auto-login: store the new user's ID in session
	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		middleware.SetFlashError(c, "Đăng ký thành công nhưng không thể đăng nhập tự động. Vui lòng đăng nhập.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	middleware.SetFlashSuccess(c, "Chào mừng "+user.FullName+"! Tài khoản của bạn đã được tạo thành công.")
	c.Redirect(http.StatusFound, "/")
}

// LoginForm renders GET /login.
// Redirects to "/" if the user is already logged in.
func (h *AuthHandler) LoginForm(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "public/pages/login.html", gin.H{
		"title":      "Đăng nhập",
		"user":       nil,
		"csrf_token": middleware.CSRFToken(c),
	})
}

// Login handles POST /login.
func (h *AuthHandler) Login(c *gin.Context) {
	if middleware.GetCurrentUser(c) != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	var form services.LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/login.html", gin.H{
			"title":      "Đăng nhập",
			"user":       nil,
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
			middleware.SetFlashError(c, "Vui lòng sử dụng trang đăng nhập dành cho quản trị viên.")
			c.Redirect(http.StatusFound, "/admin/login")
			return
		case appErrors.Is(err, services.ErrAccountBanned):
			errMsg = "Tài khoản của bạn đã bị khóa. Vui lòng liên hệ quản trị viên."
		case appErrors.Is(err, services.ErrAccountInactive):
			errMsg = "Tài khoản của bạn chưa được kích hoạt."
		case appErrors.Is(err, appErrors.ErrInvalidCredentials):
			errMsg = "Email hoặc mật khẩu không đúng."
		default:
			errMsg = "Đã có lỗi xảy ra. Vui lòng thử lại."
		}
		c.HTML(http.StatusUnprocessableEntity, "public/pages/login.html", gin.H{
			"title":      "Đăng nhập",
			"user":       nil,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{errMsg},
			"form":       form,
		})
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		c.HTML(http.StatusInternalServerError, "public/pages/login.html", gin.H{
			"title":      "Đăng nhập",
			"user":       nil,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{"Đã có lỗi khi tạo phiên đăng nhập. Vui lòng thử lại."},
			"form":       form,
		})
		return
	}

	middleware.SetFlashSuccess(c, "Chào mừng "+user.FullName+" quay trở lại!")
	c.Redirect(http.StatusFound, "/")
}

// Logout handles GET /logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	_ = middleware.ClearSession(c)
	middleware.SetFlashSuccess(c, "Bạn đã đăng xuất thành công.")
	c.Redirect(http.StatusFound, "/")
}

// translateBindErrors converts go-playground/validator errors into Vietnamese messages.
func translateBindErrors(err error) []string {
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return []string{"Dữ liệu gửi lên không hợp lệ."}
	}

	fields := map[string]string{
		"FullName":        "Họ tên",
		"Email":           "Email",
		"Password":        "Mật khẩu",
		"PasswordConfirm": "Xác nhận mật khẩu",
	}

	msgs := make([]string, 0, len(valErrs))
	for _, fe := range valErrs {
		label := fe.Field()
		if vn, ok := fields[fe.Field()]; ok {
			label = vn
		}
		var msg string
		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf("%s là bắt buộc.", label)
		case "email":
			msg = fmt.Sprintf("%s phải là địa chỉ email hợp lệ.", label)
		case "min":
			msg = fmt.Sprintf("%s phải có ít nhất %s ký tự.", label, fe.Param())
		case "max":
			msg = fmt.Sprintf("%s không được vượt quá %s ký tự.", label, fe.Param())
		default:
			msg = fmt.Sprintf("%s không hợp lệ.", label)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}
