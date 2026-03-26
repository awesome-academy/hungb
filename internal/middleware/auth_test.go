package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("test_session", store))
	return r
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	return db
}

// --- GetCurrentUser ---------------------------------------------------

func TestGetCurrentUser_NoUser(t *testing.T) {
	r := setupTestRouter()
	r.GET("/test", func(c *gin.Context) {
		user := GetCurrentUser(c)
		assert.Nil(t, user)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCurrentUser_WithUser(t *testing.T) {
	r := setupTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("current_user", &models.User{ID: 1, FullName: "Test"})
		user := GetCurrentUser(c)
		require.NotNil(t, user)
		assert.Equal(t, uint(1), user.ID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
}

func TestGetCurrentUser_WrongType(t *testing.T) {
	r := setupTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("current_user", "not-a-user")
		user := GetCurrentUser(c)
		assert.Nil(t, user)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
}

// --- RequireLogin -----------------------------------------------------

func TestRequireLogin_NoUser_Redirects(t *testing.T) {
	r := setupTestRouter()
	r.GET("/protected", RequireLogin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteLogin, w.Header().Get("Location"))
}

func TestRequireLogin_WithUser_Passes(t *testing.T) {
	r := setupTestRouter()
	r.GET("/protected", func(c *gin.Context) {
		c.Set("current_user", &models.User{ID: 1, Role: constants.RoleUser})
		c.Next()
	}, RequireLogin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- RequireAdmin -----------------------------------------------------

func TestRequireAdmin_NoUser_Redirects(t *testing.T) {
	r := setupTestRouter()
	r.GET("/admin/test", RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminLogin, w.Header().Get("Location"))
}

func TestRequireAdmin_RegularUser_Redirects(t *testing.T) {
	r := setupTestRouter()
	r.GET("/admin/test", func(c *gin.Context) {
		c.Set("current_user", &models.User{ID: 1, Role: constants.RoleUser})
		c.Next()
	}, RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminLogin, w.Header().Get("Location"))
}

func TestRequireAdmin_AdminUser_Passes(t *testing.T) {
	r := setupTestRouter()
	r.GET("/admin/test", func(c *gin.Context) {
		c.Set("current_user", &models.User{ID: 1, Role: constants.RoleAdmin})
		c.Next()
	}, RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- SetSessionUserID / ClearSession ----------------------------------

func TestSetSessionUserID_And_ClearSession(t *testing.T) {
	r := setupTestRouter()

	var sessionSaved bool
	r.GET("/set", func(c *gin.Context) {
		err := SetSessionUserID(c, 42)
		require.NoError(t, err)
		sessionSaved = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/set", nil)
	r.ServeHTTP(w, req)

	assert.True(t, sessionSaved)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify session can be cleared
	r.GET("/clear", func(c *gin.Context) {
		err := ClearSession(c)
		require.NoError(t, err)
		c.Status(http.StatusOK)
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/clear", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

// --- LoadUser ---------------------------------------------------------

func TestLoadUser_NoSession_Passes(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter()

	r.GET("/test", LoadUser(db), func(c *gin.Context) {
		assert.Nil(t, GetCurrentUser(c))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoadUser_ValidUser_SetsContext(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:    "test@example.com",
		FullName: "Test",
		Role:     constants.RoleUser,
		Status:   constants.StatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	r := setupTestRouter()

	// First: set session
	r.GET("/login", func(c *gin.Context) {
		_ = SetSessionUserID(c, user.ID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	r.ServeHTTP(w, req)

	// Second: use LoadUser middleware with the session cookie
	r.GET("/check", LoadUser(db), func(c *gin.Context) {
		u := GetCurrentUser(c)
		if u != nil {
			c.String(http.StatusOK, u.Email)
		} else {
			c.Status(http.StatusUnauthorized)
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/check", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "test@example.com", w2.Body.String())
}

func TestLoadUser_InactiveUser_ClearsSession(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:    "inactive@example.com",
		FullName: "Inactive",
		Role:     constants.RoleUser,
		Status:   constants.StatusInactive,
	}
	require.NoError(t, db.Create(user).Error)

	r := setupTestRouter()

	r.GET("/login", func(c *gin.Context) {
		_ = SetSessionUserID(c, user.ID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	r.ServeHTTP(w, req)

	r.GET("/check", LoadUser(db), func(c *gin.Context) {
		u := GetCurrentUser(c)
		if u != nil {
			c.Status(http.StatusOK)
		} else {
			c.Status(http.StatusUnauthorized)
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/check", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestLoadUser_UserNotInDB_ClearsSession(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter()

	r.GET("/login", func(c *gin.Context) {
		_ = SetSessionUserID(c, uint(9999))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	r.ServeHTTP(w, req)

	r.GET("/check", LoadUser(db), func(c *gin.Context) {
		u := GetCurrentUser(c)
		if u != nil {
			c.Status(http.StatusOK)
		} else {
			c.Status(http.StatusUnauthorized)
		}
	})

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/check", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

// --- Flash helpers (exported via auth.go but used broadly) ---

func TestFlash_SetSuccessAndGet(t *testing.T) {
	r := setupTestRouter()

	r.GET("/set", func(c *gin.Context) {
		SetFlashSuccess(c, "ok!")
		c.Status(http.StatusOK)
	})
	r.GET("/get", func(c *gin.Context) {
		s, e := GetFlash(c)
		assert.Equal(t, "ok!", s)
		assert.Empty(t, e)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/set", nil)
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/get", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)
}

func TestFlash_SetErrorAndGet(t *testing.T) {
	r := setupTestRouter()

	r.GET("/set", func(c *gin.Context) {
		SetFlashError(c, "fail!")
		c.Status(http.StatusOK)
	})
	r.GET("/get", func(c *gin.Context) {
		s, e := GetFlash(c)
		assert.Empty(t, s)
		assert.Equal(t, "fail!", e)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/set", nil)
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/get", nil)
	for _, c := range w.Result().Cookies() {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)
}

func TestFlash_EmptyWhenNoneSet(t *testing.T) {
	r := setupTestRouter()

	r.GET("/get", func(c *gin.Context) {
		s, e := GetFlash(c)
		assert.Empty(t, s)
		assert.Empty(t, e)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/get", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
