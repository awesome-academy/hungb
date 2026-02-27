package middleware

import (
	"log/slog"
	"net/http"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	sessionKeyUserID = "user_id"
	contextKeyUser   = "current_user"
)

func GetCurrentUser(c *gin.Context) *models.User {
	val, exists := c.Get(contextKeyUser)
	if !exists {
		return nil
	}
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func SetSessionUserID(c *gin.Context, userID uint) error {
	session := sessions.Default(c)
	session.Set(sessionKeyUserID, userID)
	return session.Save()
}

func ClearSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

// LoadUser loads the authenticated user from session into gin.Context.
// Does NOT block unauthenticated requests.
func LoadUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(sessionKeyUserID)
		if userID == nil {
			c.Next()
			return
		}

		id, ok := userID.(uint)
		if !ok {
			c.Next()
			return
		}

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			slog.Warn(appErrors.ErrUserNotFound.Message, "user_id", id)
			session.Delete(sessionKeyUserID)
			_ = session.Save()
			c.Next()
			return
		}

		if user.Status != constants.StatusActive {
			slog.Warn(appErrors.ErrUserNotActive.Message, "user_id", id, "status", user.Status)
			session.Delete(sessionKeyUserID)
			_ = session.Save()
			c.Next()
			return
		}

		c.Set(contextKeyUser, &user)
		c.Next()
	}
}

func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetCurrentUser(c) == nil {
			c.Redirect(http.StatusFound, constants.RouteLogin)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil || user.Role != constants.RoleAdmin {
			c.Redirect(http.StatusFound, constants.RouteAdminLogin)
			c.Abort()
			return
		}
		c.Next()
	}
}
