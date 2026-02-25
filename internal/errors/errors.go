package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

// Authentication & Authorization
var (
	ErrUnauthorized       = NewAppError(http.StatusUnauthorized, "unauthorized")
	ErrForbidden          = NewAppError(http.StatusForbidden, "forbidden")
	ErrInvalidCredentials = NewAppError(http.StatusUnauthorized, "invalid email or password")
	ErrSessionExpired     = NewAppError(http.StatusUnauthorized, "session expired")
	ErrUserNotActive      = NewAppError(http.StatusForbidden, "user account is not active")
	ErrCSRFTokenMismatch  = NewAppError(http.StatusBadRequest, "CSRF token mismatch")
)

// User
var (
	ErrUserNotFound      = NewAppError(http.StatusNotFound, "user not found")
	ErrUserAlreadyExists = NewAppError(http.StatusConflict, "user already exists")
	ErrEmailAlreadyTaken = NewAppError(http.StatusConflict, "email already taken")
)

// Tour
var (
	ErrTourNotFound     = NewAppError(http.StatusNotFound, "tour not found")
	ErrTourNotAvailable = NewAppError(http.StatusBadRequest, "tour is not available")
)

// Category
var (
	ErrCategoryNotFound = NewAppError(http.StatusNotFound, "category not found")
	ErrCategoryHasTours = NewAppError(http.StatusBadRequest, "category has associated tours")
)

// Booking
var (
	ErrBookingNotFound  = NewAppError(http.StatusNotFound, "booking not found")
	ErrBookingFull      = NewAppError(http.StatusBadRequest, "no available slots")
	ErrBookingCancelled = NewAppError(http.StatusBadRequest, "booking already cancelled")
	ErrBookingCompleted = NewAppError(http.StatusBadRequest, "booking already completed")
	ErrInvalidBooking   = NewAppError(http.StatusBadRequest, "invalid booking")
)

// Payment
var (
	ErrPaymentNotFound = NewAppError(http.StatusNotFound, "payment not found")
	ErrPaymentFailed   = NewAppError(http.StatusBadRequest, "payment failed")
)

// Review & Rating
var (
	ErrReviewNotFound  = NewAppError(http.StatusNotFound, "review not found")
	ErrRatingNotFound  = NewAppError(http.StatusNotFound, "rating not found")
	ErrAlreadyRated    = NewAppError(http.StatusConflict, "already rated this tour")
	ErrCommentNotFound = NewAppError(http.StatusNotFound, "comment not found")
)

// Authorization
var (
	ErrNotAuthorizedUpdate = NewAppError(http.StatusForbidden, "not authorized to update")
	ErrNotAuthorizedDelete = NewAppError(http.StatusForbidden, "not authorized to delete")
)

// Validation & Server
var (
	ErrValidationFailed    = NewAppError(http.StatusBadRequest, "validation failed")
	ErrInvalidInput        = NewAppError(http.StatusBadRequest, "invalid input")
	ErrInternalServerError = NewAppError(http.StatusInternalServerError, "internal server error")
	ErrFailedToFetch       = NewAppError(http.StatusInternalServerError, "failed to fetch data")
	ErrFailedToCreate      = NewAppError(http.StatusInternalServerError, "failed to create")
	ErrFailedToUpdate      = NewAppError(http.StatusInternalServerError, "failed to update")
	ErrFailedToDelete      = NewAppError(http.StatusInternalServerError, "failed to delete")
)

// --- Response helpers ---

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func RespondError(c *gin.Context, status int, message string, details ...interface{}) {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}
	c.JSON(status, ErrorResponse{
		Code:    status,
		Message: message,
		Details: detail,
	})
}

// RespondAppError checks if err is *AppError and responds accordingly;
// otherwise responds with a generic 500.
func RespondAppError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		RespondError(c, appErr.Status, appErr.Message)
		return
	}
	RespondError(c, http.StatusInternalServerError, ErrInternalServerError.Message)
}

// RespondPageError renders an HTML error page (for SSR).
func RespondPageError(c *gin.Context, status int, templateName string, message string) {
	c.HTML(status, templateName, gin.H{
		"error": message,
	})
}

// HandleBindError validates binding errors and responds with field-level details.
// Returns true if there was an error.
func HandleBindError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		RespondError(c, http.StatusBadRequest, "invalid request body format")
		return true
	}

	fields := make(map[string]string)
	for _, fe := range validationErrs {
		switch fe.Tag() {
		case "required":
			fields[fe.Field()] = "is required"
		case "email":
			fields[fe.Field()] = "must be a valid email"
		case "min":
			fields[fe.Field()] = "must be at least " + fe.Param() + " characters"
		case "max":
			fields[fe.Field()] = "must be at most " + fe.Param() + " characters"
		default:
			fields[fe.Field()] = "is invalid"
		}
	}

	RespondError(c, http.StatusBadRequest, ErrValidationFailed.Message, fields)
	return true
}

// IsDuplicateEntryError checks for PostgreSQL unique constraint violation (code 23505).
func IsDuplicateEntryError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Is checks if the given error matches a target *AppError.
func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message == target.Message
	}
	return false
}
