package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	flashKeySuccess = "flash_success"
	flashKeyError   = "flash_error"
)

// SetFlashSuccess stores a success flash message in the session.
func SetFlashSuccess(c *gin.Context, msg string) {
	session := sessions.Default(c)
	session.AddFlash(msg, flashKeySuccess)
	_ = session.Save()
}

// SetFlashError stores an error flash message in the session.
func SetFlashError(c *gin.Context, msg string) {
	session := sessions.Default(c)
	session.AddFlash(msg, flashKeyError)
	_ = session.Save()
}

// GetFlash reads and clears flash messages from the session (single-use).
// Returns (successMsg, errorMsg).
func GetFlash(c *gin.Context) (string, string) {
	session := sessions.Default(c)

	var success, errMsg string

	if flashes := session.Flashes(flashKeySuccess); len(flashes) > 0 {
		if msg, ok := flashes[0].(string); ok {
			success = msg
		}
	}

	if flashes := session.Flashes(flashKeyError); len(flashes) > 0 {
		if msg, ok := flashes[0].(string); ok {
			errMsg = msg
		}
	}

	_ = session.Save()
	return success, errMsg
}
