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
	ErrTourNotFound        = NewAppError(http.StatusNotFound, "tour not found")
	ErrTourNotAvailable    = NewAppError(http.StatusBadRequest, "tour is not available")
	ErrTourHasBookings     = NewAppError(http.StatusBadRequest, "tour has active bookings")
	ErrScheduleNotFound    = NewAppError(http.StatusNotFound, "schedule not found")
	ErrScheduleHasBookings = NewAppError(http.StatusBadRequest, "schedule has bookings")
)

const (
	ErrCtxTourFindAll            = "find all tours"
	ErrCtxTourCount              = "count tours"
	ErrCtxTourFindByID           = "find tour by id"
	ErrCtxTourFindBySlug         = "find tour by slug"
	ErrCtxTourCheckSlugExists    = "check tour slug exists"
	ErrCtxTourCheckSlugExcluding = "check tour slug exists excluding"
	ErrCtxTourCreate             = "create tour"
	ErrCtxTourUpdate             = "update tour"
	ErrCtxTourDelete             = "delete tour"
	ErrCtxTourHasActiveBookings  = "check tour has active bookings"
	ErrCtxTourReplaceCategories  = "replace tour categories"
)

const (
	ErrCtxTourServiceList               = "list tours"
	ErrCtxTourServiceGet                = "get tour"
	ErrCtxTourServiceCreateCheckSlug    = "create tour check slug"
	ErrCtxTourServiceCreate             = "create tour"
	ErrCtxTourServiceUpdateFind         = "update tour find"
	ErrCtxTourServiceUpdateCheckSlug    = "update tour check slug"
	ErrCtxTourServiceUpdate             = "update tour"
	ErrCtxTourServiceDeleteFind         = "delete tour find"
	ErrCtxTourServiceDeleteCheckBooks   = "delete tour check bookings"
	ErrCtxTourServiceDelete             = "delete tour"
	ErrCtxTourServiceValidateCategories = "validate category ids"
)

const (
	ErrMsgTourTitleRequired       = "Tên tour là bắt buộc."
	ErrMsgTourTitleDuplicate      = "Tour với tên tương tự đã tồn tại."
	ErrMsgTourPricePositive       = "Giá tour phải lớn hơn 0."
	ErrMsgTourDurationPositive    = "Số ngày tour phải lớn hơn 0."
	ErrMsgTourMaxParticipants     = "Số người tham gia tối đa phải lớn hơn 0."
	ErrMsgTourInvalidStatus       = "Trạng thái tour không hợp lệ."
	ErrMsgTourCannotDeleteBooking = "Không thể xóa tour đang có booking."
	ErrMsgTourCategoryNotFound    = "Một hoặc nhiều danh mục không tồn tại."
)

const (
	ErrCtxScheduleFindByTour  = "find schedules by tour"
	ErrCtxScheduleFindByID    = "find schedule by id"
	ErrCtxScheduleCreate      = "create schedule"
	ErrCtxScheduleUpdate      = "update schedule"
	ErrCtxScheduleDelete      = "delete schedule"
	ErrCtxScheduleHasBookings = "check schedule has bookings"
)

const (
	ErrCtxScheduleServiceList        = "list schedules"
	ErrCtxScheduleServiceGet         = "get schedule"
	ErrCtxScheduleServiceCreate      = "create schedule"
	ErrCtxScheduleServiceUpdateFind  = "update schedule find"
	ErrCtxScheduleServiceUpdate      = "update schedule"
	ErrCtxScheduleServiceDeleteFind  = "delete schedule find"
	ErrCtxScheduleServiceDeleteCheck = "delete schedule check bookings"
	ErrCtxScheduleServiceDelete      = "delete schedule"
)

const (
	ErrMsgScheduleReturnNotBeforeDepart = "Ngày về không được trước ngày khởi hành."
	ErrMsgSchedulePriceOverridePositive = "Giá ghi đè phải lớn hơn 0."
	ErrMsgScheduleSlotsPositive         = "Số chỗ phải lớn hơn 0."
	ErrMsgScheduleInvalidStatus         = "Trạng thái lịch trình không hợp lệ."
	ErrMsgScheduleCannotDeleteBooking   = "Không thể xóa lịch trình đang có booking."
	ErrMsgScheduleTourNotFound          = "Tour không tồn tại."
	ErrMsgScheduleDepartureDateReq      = "Ngày khởi hành là bắt buộc."
	ErrMsgScheduleReturnDateReq         = "Ngày về là bắt buộc."
)

// Category
var (
	ErrCategoryNotFound = NewAppError(http.StatusNotFound, "category not found")
	ErrCategoryHasTours = NewAppError(http.StatusBadRequest, "category has associated tours")
)

// Bank Account
var (
	ErrBankAccountNotFound = NewAppError(http.StatusNotFound, "bank account not found")
)

const (
	ErrCtxCategoryFindAll            = "find all categories"
	ErrCtxCategoryFindAllParents     = "find parent categories"
	ErrCtxCategoryFindByID           = "find category by id"
	ErrCtxCategoryFindBySlug         = "find category by slug"
	ErrCtxCategoryCheckSlugExists    = "check slug exists"
	ErrCtxCategoryCheckSlugExcluding = "check slug exists excluding"
	ErrCtxCategoryCreate             = "create category"
	ErrCtxCategoryUpdate             = "update category"
	ErrCtxCategoryDelete             = "delete category"
	ErrCtxCategoryHasTours           = "check category has tours"
	ErrCtxCategoryHasChildren        = "check category has children"
	ErrCtxCategoryGetDescendantIDs   = "get descendant ids"
	ErrCtxCategoryCountByIDs         = "count categories by ids"
)

// Category Service error context (used in fmt.Errorf wrapping)
const (
	ErrCtxCategoryServiceListCategories       = "list categories"
	ErrCtxCategoryServiceAllFlat              = "all flat categories"
	ErrCtxCategoryServiceGetCategory          = "get category"
	ErrCtxCategoryServiceCreateCheckSlug      = "create category check slug"
	ErrCtxCategoryServiceCreate               = "create category"
	ErrCtxCategoryServiceUpdateFindCategory   = "update category find"
	ErrCtxCategoryServiceUpdateCheckSlug      = "update category check slug"
	ErrCtxCategoryServiceUpdateGetDescendants = "update category get descendants"
	ErrCtxCategoryServiceUpdate               = "update category"
	ErrCtxCategoryServiceDeleteFindCategory   = "delete category find"
	ErrCtxCategoryServiceDeleteCheckChildren  = "delete category check children"
	ErrCtxCategoryServiceDeleteCheckTours     = "delete category check tours"
	ErrCtxCategoryServiceDelete               = "delete category"
)

// Category Service validation error messages (user-facing)
const (
	ErrMsgCategoryNameRequired             = "Tên danh mục là bắt buộc."
	ErrMsgCategoryNameDuplicate            = "Danh mục với tên tương tự đã tồn tại."
	ErrMsgCategoryParentNotFound           = "Danh mục cha không tồn tại."
	ErrMsgCategoryParentMustBeRoot         = "Chỉ hỗ trợ danh mục con cấp 2 (danh mục cha phải là cấp gốc)."
	ErrMsgCategorySelfParent               = "Danh mục không thể là cha của chính nó."
	ErrMsgCategoryChildAsParent            = "Không thể chọn danh mục con làm danh mục cha."
	ErrMsgCategoryCannotDeleteWithChildren = "Không thể xóa danh mục có danh mục con. Hãy xóa danh mục con trước."
)

// Bank Account error context (used in fmt.Errorf wrapping)
const (
	ErrCtxBankAccountFindByID     = "find bank account by id"
	ErrCtxBankAccountFindByUser   = "find bank accounts by user"
	ErrCtxBankAccountCount        = "count bank accounts"
	ErrCtxBankAccountCreate       = "create bank account"
	ErrCtxBankAccountUpdate       = "update bank account"
	ErrCtxBankAccountDelete       = "delete bank account"
	ErrCtxBankAccountClearDefault = "clear default bank accounts"
	ErrCtxBankAccountSetDefault   = "set default bank account"
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
		"status":  status,
		"message": message,
		"error":   message,
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
