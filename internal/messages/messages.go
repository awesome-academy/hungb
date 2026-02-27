package messages

// ── Generic / system
const (
	ErrInternalServer = "Đã xảy ra lỗi hệ thống. Vui lòng thử lại sau."
	ErrInvalidForm    = "Dữ liệu gửi lên không hợp lệ."
	ErrTryAgain       = "Đã có lỗi xảy ra. Vui lòng thử lại."
)

// ── Auth — Registration

const (
	TitleRegister = "Đăng ký tài khoản"

	ErrEmailTaken       = "Email đã được sử dụng. Vui lòng dùng email khác hoặc đăng nhập."
	ErrPasswordMismatch = "Mật khẩu xác nhận không khớp."

	MsgRegisterSuccess       = "Chào mừng %s! Tài khoản của bạn đã được tạo thành công."
	MsgRegisterAutoLoginFail = "Đăng ký thành công nhưng không thể đăng nhập tự động. Vui lòng đăng nhập."
)

// ── Auth — Login

const (
	TitleLogin = "Đăng nhập"

	ErrInvalidCredentials = "Email hoặc mật khẩu không đúng."
	ErrAccountNotActive   = "Tài khoản của bạn chưa được kích hoạt."
	ErrAccountBanned      = "Tài khoản của bạn đã bị khóa. Vui lòng liên hệ quản trị viên."
	ErrAdminMustUsePortal = "Vui lòng sử dụng trang đăng nhập dành cho quản trị viên."
	ErrCreateSessionFail  = "Đã có lỗi khi tạo phiên đăng nhập. Vui lòng thử lại."

	MsgLoginWelcomeBack = "Chào mừng %s quay trở lại!"
	MsgLogoutSuccess    = "Bạn đã đăng xuất thành công."
)

// ── Auth service internal text (sentinel + logs)
const (
	AuthErrAdminMustUsePortal = "admin must use admin portal"
	AuthErrAccountBanned      = "account banned"
	AuthErrAccountInactive    = "account inactive"

	LogRegisterCheckEmailExists = "register: check email exists"
	LogRegisterHashPassword     = "register: hash password"
	LogRegisterCreateUser       = "register: create user"
	LogRegisterUnexpectedError  = "register: unexpected error"
)

// ── Admin dashboard
const (
	TitleAdminDashboard = "Dashboard"

	DashboardLabelTotalUsers   = "Tổng người dùng"
	DashboardLabelActiveTours  = "Tour đang hoạt động"
	DashboardLabelTodayBooking = "Đặt tour hôm nay"
	DashboardLabelMonthRevenue = "Doanh thu tháng"

	LogDashboardLoadStatsFailed = "failed to load dashboard stats"
)

// ── Stats service internal text (wrapped error context)
const (
	ErrCtxCountUsers         = "count users"
	ErrCtxCountTours         = "count tours"
	ErrCtxCountTodayBookings = "count today bookings"
	ErrCtxSumMonthRevenue    = "sum month revenue"
	ErrCtxRecentBookings     = "recent bookings"
	ErrCtxPendingReviews     = "pending reviews"
)

// ── Form field labels (used in validation error messages)
const (
	FieldFullName        = "Họ tên"
	FieldEmail           = "Email"
	FieldPassword        = "Mật khẩu"
	FieldPasswordConfirm = "Xác nhận mật khẩu"
)

// ── Validation message templates (fmt.Sprintf, %s = field label / param)
const (
	ValRequired = "%s là bắt buộc."
	ValEmail    = "%s phải là địa chỉ email hợp lệ."
	ValMin      = "%s phải có ít nhất %s ký tự."
	ValMax      = "%s không được vượt quá %s ký tự."
	ValInvalid  = "%s không hợp lệ."
)

// ── App bootstrap / logging
const (
	FlagMigrateDescription = "Run database migration"
	FlagSeedDescription    = "Seed database with initial data"

	LogConfigurationLoaded = "configuration loaded"
	LogDatabaseConnFailed  = "database connection failed"
	LogMigrationFailed     = "migration failed"
	LogSeedingFailed       = "seeding failed"
	LogStartingServer      = "starting server"
	LogServerStartFailed   = "server failed to start"

	LogTemplateNotFound        = "template not found"
	LogTemplateWalkFailed      = "failed to walk templates directory"
	LogSharedTemplateReadFail  = "failed to read shared template"
	LogSharedTemplateParseFail = "failed to parse shared template"
	LogPageTemplateReadFail    = "failed to read page template"
	LogPageTemplateParseFail   = "failed to parse page template"
	LogLoadedPageTemplate      = "loaded page template"

	TemplateNotFoundText = "template not found: "
)
