package public

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/services"
	"sun-booking-tours/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newPublicAuthHandler() (*AuthHandler, *testutil.MockUserRepo) {
	userRepo := new(testutil.MockUserRepo)
	socialRepo := new(testutil.MockSocialAccountRepo)
	authSvc := services.NewAuthService(nil, userRepo, socialRepo, nil, testutil.TestBaseURL)
	cfg := &config.Config{}
	handler := NewAuthHandler(authSvc, cfg)
	return handler, userRepo
}

var publicAuthTemplates = []string{
	"public/pages/login.html",
	"public/pages/register.html",
	"public/pages/register_success.html",
	"public/pages/verify_result.html",
}

// --- RegisterForm GET ------------------------------------------------

func TestRegisterForm_ShowsPage(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.GET("/register", handler.RegisterForm)

	w := testutil.MakeGetRequest(r, "/register")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterForm_RedirectsIfLoggedIn(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)

	r.GET("/register", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewActiveUser())
		c.Next()
	}, handler.RegisterForm)

	w := testutil.MakeGetRequest(r, "/register")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
}

// --- Register POST ---------------------------------------------------

func TestRegister_Success(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/register", handler.Register)

	userRepo.On("ExistsByEmail", mock.Anything, "new@example.com").Return(false, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*models.User)
			u.ID = 10
		}).
		Return(nil)

	form := url.Values{
		"full_name":        {"New User"},
		"email":            {"new@example.com"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	w := testutil.MakePostFormRequest(r, "/register", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
	userRepo.AssertExpectations(t)
}

func TestRegister_InvalidForm_MissingFields(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/register", handler.Register)

	w := testutil.MakePostFormRequest(r, "/register", url.Values{})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_EmailAlreadyTaken(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/register", handler.Register)

	userRepo.On("ExistsByEmail", mock.Anything, testutil.TestEmail).Return(true, nil)

	form := url.Values{
		"full_name":        {"User"},
		"email":            {testutil.TestEmail},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	w := testutil.MakePostFormRequest(r, "/register", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_PasswordMismatch(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/register", handler.Register)

	form := url.Values{
		"full_name":        {"User"},
		"email":            {"test@example.com"},
		"password":         {"password123"},
		"password_confirm": {"different"},
	}
	w := testutil.MakePostFormRequest(r, "/register", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRegister_RedirectsIfLoggedIn(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)

	r.POST("/register", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewActiveUser())
		c.Next()
	}, handler.Register)

	form := url.Values{
		"full_name":        {"User"},
		"email":            {"test@example.com"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	w := testutil.MakePostFormRequest(r, "/register", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
}

// --- LoginForm GET ---------------------------------------------------

func TestLoginForm_ShowsPage(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.GET("/login", handler.LoginForm)

	w := testutil.MakeGetRequest(r, "/login")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginForm_RedirectsIfLoggedIn(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)

	r.GET("/login", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewActiveUser())
		c.Next()
	}, handler.LoginForm)

	w := testutil.MakeGetRequest(r, "/login")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
}

// --- Login POST ------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/login", handler.Login)

	user := testutil.NewActiveUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	form := url.Values{
		"email":    {testutil.TestEmail},
		"password": {testutil.TestPassword},
	}
	w := testutil.MakePostFormRequest(r, "/login", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
	userRepo.AssertExpectations(t)
}

func TestLogin_InvalidForm(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/login", handler.Login)

	w := testutil.MakePostFormRequest(r, "/login", url.Values{})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestLogin_WrongCredentials(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/login", handler.Login)

	user := testutil.NewActiveUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	form := url.Values{
		"email":    {testutil.TestEmail},
		"password": {"wrong"},
	}
	w := testutil.MakePostFormRequest(r, "/login", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestLogin_BannedUser(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/login", handler.Login)

	user := testutil.NewBannedUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestEmail).Return(user, nil)

	form := url.Values{
		"email":    {testutil.TestEmail},
		"password": {testutil.TestPassword},
	}
	w := testutil.MakePostFormRequest(r, "/login", form)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestLogin_AdminRedirectsToAdminPortal(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.POST("/login", handler.Login)

	admin := testutil.NewAdminUser()
	userRepo.On("FindByEmail", mock.Anything, testutil.TestAdminEmail).Return(admin, nil)

	form := url.Values{
		"email":    {testutil.TestAdminEmail},
		"password": {testutil.TestPassword},
	}
	w := testutil.MakePostFormRequest(r, "/login", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteAdminLogin, w.Header().Get("Location"))
}

func TestLogin_RedirectsIfLoggedIn(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouterNoCSRF()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)

	r.POST("/login", func(c *gin.Context) {
		testutil.SetCurrentUser(c, testutil.NewActiveUser())
		c.Next()
	}, handler.Login)

	form := url.Values{
		"email":    {"any@mail.com"},
		"password": {"any"},
	}
	w := testutil.MakePostFormRequest(r, "/login", form)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
}

// --- VerifyEmail GET -------------------------------------------------

func TestVerifyEmail_Success(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.GET("/verify-email", handler.VerifyEmail)

	user := testutil.NewUnverifiedUser()
	userRepo.On("FindByVerifyToken", mock.Anything, "valid-verify-token").Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	w := testutil.MakeGetRequest(r, "/verify-email?token=valid-verify-token")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
	userRepo.AssertExpectations(t)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	handler, userRepo := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.GET("/verify-email", handler.VerifyEmail)

	userRepo.On("FindByVerifyToken", mock.Anything, "bad-token").
		Return(nil, fmt.Errorf("find: %w", fmt.Errorf("record not found")))

	w := testutil.MakeGetRequest(r, "/verify-email?token=bad-token")

	// The handler renders verify_result.html with success=false
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	testutil.LoadTestTemplates(r, publicAuthTemplates...)
	r.GET("/verify-email", handler.VerifyEmail)

	w := testutil.MakeGetRequest(r, "/verify-email")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Logout ----------------------------------------------------------

func TestLogout_Redirects(t *testing.T) {
	handler, _ := newPublicAuthHandler()
	r := testutil.SetupTestRouter()
	r.GET("/logout", handler.Logout)

	w := testutil.MakeGetRequest(r, "/logout")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, constants.RouteHome, w.Header().Get("Location"))
}

// --- translateBindErrors (unexported) ---------------------------------

func TestTranslateBindErrors_NonValidatorError(t *testing.T) {
	msgs := translateBindErrors(fmt.Errorf("random error"))
	assert.Len(t, msgs, 1)
}
