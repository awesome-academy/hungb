package middleware

import (
	appErrors "sun-booking-tours/internal/errors"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func CSRFMiddleware(secret string) gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: secret,
		ErrorFunc: func(c *gin.Context) {
			c.String(appErrors.ErrCSRFTokenMismatch.Status, appErrors.ErrCSRFTokenMismatch.Message)
			c.Abort()
		},
	})
}

func CSRFToken(c *gin.Context) string {
	return csrf.GetToken(c)
}
