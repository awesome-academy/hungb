package admin

import (
	"net/http"
	"net/url"
	"testing"

	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/services"
	"sun-booking-tours/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAdminAuthHandler() (*AdminAuthHandler, *testutil.MockUserRepo) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	authSvc := services.NewAuthService(nil, userRepo, socialRepo, nil, testutil.TestBaseURL)
	handler := NewAdminAuthHandler(authSvc)
	return handler, userRepo
}

var adminTemplates = []string{
	"admin/pages/login.html",
}

// --- LoginForm --------------------------------------------------------

func TestAdminLoginForm_ShowsPage(t *testing.T) {
	handler, _ := newAdminAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, adminTemplates...)
	r.GET("/admin/login", handler.LoginForm)

	w := testutil.MakeGetRequest(r, "/admin/login")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminLoginForm_RedirectsIfAlreadyAdmin(t *testing.T) {
	handler, _ := newAdminAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, adminTemplates...)

	r.GET("/admin/login", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewAdminUser())
		c.Next()
	}, handler.LoginForm)

	w := testutil.MakeGetRequest(r, "/admin/login")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminDashboard, w.Header().Get("Location"))
}

// --- Login POST -------------------------------------------------------

func TestAdminLogin_Success(t *testing.T) {
	handler, userRepo := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, adminTemplates...)
	r.POST("/admin/login", handler.Login)

	admin := testutil.NewAdminUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)

	form := url.Values{
		"email":    {testutil.TestAdminEmail},
		"password": {testutil.TestPassword},
	}
	w := testutil.MakePostFormRequest(r, "/admin/login", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminDashboard, w.Header().Get("Location"))
	userRepo.AssertExpectations(t)
}

func TestAdminLogin_InvalidForm(t *testing.T) {
	handler, _ := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, adminTemplates...)
	r.POST("/admin/login", handler.Login)

	w := testutil.MakePostFormRequest(r, "/admin/login", url.Values{})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAdminLogin_WrongCredentials(t *testing.T) {
	handler, userRepo := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, adminTemplates...)
	r.POST("/admin/login", handler.Login)

	admin := testutil.NewAdminUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)

	form := url.Values{
		"email":    {testutil.TestAdminEmail},
		"password": {"wrongpass"},
	}
	w := testutil.MakePostFormRequest(r, "/admin/login", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAdminLogin_NonAdminUser(t *testing.T) {
	handler, userRepo := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, adminTemplates...)
	r.POST("/admin/login", handler.Login)

	user := testutil.NewActiveUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	form := url.Values{
		"email":    {testutil.TestEmail},
		"password": {testutil.TestPassword},
	}
	w := testutil.MakePostFormRequest(r, "/admin/login", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAdminLogin_RedirectsIfAlreadyLoggedIn(t *testing.T) {
	handler, _ := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, adminTemplates...)

	r.POST("/admin/login", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewAdminUser())
		c.Next()
	}, handler.Login)

	form := url.Values{
		"email":    {"any@mail.com"},
		"password": {"any"},
	}
	w := testutil.MakePostFormRequest(r, "/admin/login", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminDashboard, w.Header().Get("Location"))
}

// --- Logout -----------------------------------------------------------

func TestAdminLogout_Redirects(t *testing.T) {
	handler, _ := newAdminAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	r.POST("/admin/logout", handler.Logout)

	w := testutil.MakePostFormRequest(r, "/admin/logout", url.Values{})

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminLogin, w.Header().Get("Location"))
}
