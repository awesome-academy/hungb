package public

import (
	"fmt"
	"log/slog"
	"net/http"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/twitterv2"

	gsessions "github.com/gorilla/sessions"
)

type OAuthHandler struct {
	authService *services.AuthService
}

func NewOAuthHandler(authService *services.AuthService, cfg *config.Config) *OAuthHandler {
	initOAuthProviders(cfg)
	return &OAuthHandler{authService: authService}
}

func initOAuthProviders(cfg *config.Config) {
	// Gothic stores OAuth state in gorilla/sessions.
	// Use the same secret as the app session for consistency.
	gothic.Store = gsessions.NewCookieStore([]byte(cfg.SessionSecret))

	var providers []goth.Provider

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		providers = append(providers, google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.BaseURL+"/auth/google/callback",
			"email", "profile",
		))
	}

	if cfg.FBClientID != "" && cfg.FBClientSecret != "" {
		providers = append(providers, facebook.New(
			cfg.FBClientID,
			cfg.FBClientSecret,
			cfg.BaseURL+"/auth/facebook/callback",
			"email",
		))
	}

	if cfg.TwitterAPIKey != "" && cfg.TwitterAPISecret != "" {
		providers = append(providers, twitterv2.New(
			cfg.TwitterAPIKey,
			cfg.TwitterAPISecret,
			cfg.BaseURL+"/auth/twitterv2/callback",
		))
	}

	if len(providers) > 0 {
		goth.UseProviders(providers...)
	}
}

func (h *OAuthHandler) Begin(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Set("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogOAuthCallbackFailed,
			"provider", provider, "error", err)
		middleware.SetFlashError(c, messages.ErrOAuthCallback)
		c.Redirect(http.StatusFound, constants.RouteLogin)
		return
	}

	user, err := h.authService.OAuthLogin(
		c.Request.Context(),
		provider,
		gothUser.UserID,
		gothUser.Email,
		gothUser.Name,
		gothUser.AvatarURL,
	)
	if err != nil {
		var errMsg string
		switch {
		case appErrors.Is(err, services.ErrAccountBanned):
			errMsg = messages.ErrOAuthBanned
		case appErrors.Is(err, services.ErrAccountInactive):
			errMsg = messages.ErrOAuthInactive
		case appErrors.Is(err, services.ErrAdminMustUsePortal):
			errMsg = messages.ErrOAuthAdmin
		default:
			slog.ErrorContext(c.Request.Context(), messages.LogOAuthLoginFailed,
				"provider", provider, "error", err)
			errMsg = messages.ErrOAuthCallback
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteLogin)
		return
	}

	if err := middleware.SetSessionUserID(c, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), messages.LogLoginSetSessionFailed, "error", err)
		middleware.SetFlashError(c, messages.ErrCreateSessionFail)
		c.Redirect(http.StatusFound, constants.RouteLogin)
		return
	}

	middleware.SetFlashSuccess(c, fmt.Sprintf(messages.MsgOAuthLoginSuccess, user.FullName))
	c.Redirect(http.StatusFound, constants.RouteHome)
}
