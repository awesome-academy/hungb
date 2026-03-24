package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"sun-booking-tours/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// SetupTestRouter returns a Gin engine configured with session and CSRF middleware for testing.
func SetupTestRouter() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte(TestSessionSecret))
	r.Use(sessions.Sessions("test_session", store))
	r.Use(csrf.Middleware(csrf.Options{
		Secret: TestSessionSecret,
		ErrorFunc: func(c *gin.Context) {
			c.String(http.StatusBadRequest, "csrf error")
			c.Abort()
		},
	}))
	return r
}

// SetupTestRouterNoCSRF returns a router without CSRF, useful for POST tests
// where CSRF tokens are hard to obtain.
func SetupTestRouterNoCSRF() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte(TestSessionSecret))
	r.Use(sessions.Sessions("test_session", store))
	// Set a dummy csrfSecret so CSRFToken() doesn't panic
	r.Use(func(c *gin.Context) {
		c.Set("csrfSecret", TestSessionSecret)
		c.Set("csrfToken", "test-csrf-token")
		c.Next()
	})
	return r
}

// SetCurrentUser stores a user into the gin.Context so that
// middleware.GetCurrentUser(c) returns it during tests.
func SetCurrentUser(c *gin.Context, user *models.User) {
	c.Set("current_user", user)
}

// MakeGetRequest creates and performs a GET request, returning the recorder.
func MakeGetRequest(router *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(w, req)
	return w
}

// MakePostFormRequest creates and performs a POST request with form data, returning the recorder.
func MakePostFormRequest(router *gin.Engine, path string, formData url.Values) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	return w
}

// MakeRequestWithCookies performs a request forwarding cookies from a previous response.
// Useful for testing session-dependent flows.
func MakeRequestWithCookies(router *gin.Engine, method, path string, prevResp *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	for _, c := range prevResp.Result().Cookies() {
		req.AddCookie(c)
	}
	router.ServeHTTP(w, req)
	return w
}
