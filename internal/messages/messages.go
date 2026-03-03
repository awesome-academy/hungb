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

	LogLoginFindUser            = "login: find user by email failed"
	LogLoginUnexpectedError     = "login: unexpected error"
	LogLoginSetSessionFailed    = "login: set session failed"
	LogLogoutClearSessionFailed = "logout: clear session failed"

	LogAdminLoginFindUser   = "admin login: find user by email failed"
	LogAdminLoginUnexpected = "admin login: unexpected error"
)

const (
	TitleHome = "Trang chủ"
)

// ── Admin Auth
const (
	TitleAdminLogin = "Đăng nhập"

	ErrAdminNoAccess        = "Bạn không có quyền truy cập trang admin."
	ErrAdminAccountBanned   = "Tài khoản quản trị đã bị khóa."
	ErrAdminAccountInactive = "Tài khoản quản trị chưa được kích hoạt."
	ErrAdminInvalidCreds    = "Email hoặc mật khẩu không đúng."
	ErrAdminCreateSession   = "Không thể tạo phiên đăng nhập. Vui lòng thử lại."

	MsgAdminLoginWelcome = "Chào mừng %s!"
	MsgAdminLogout       = "Bạn đã đăng xuất khỏi trang admin."
)

// ── OAuth
const (
	ErrOAuthBegin    = "Không thể bắt đầu đăng nhập mạng xã hội."
	ErrOAuthCallback = "Đăng nhập mạng xã hội thất bại. Vui lòng thử lại."
	ErrOAuthBanned   = "Tài khoản đã bị khóa. Vui lòng liên hệ quản trị viên."
	ErrOAuthInactive = "Tài khoản chưa được kích hoạt."
	ErrOAuthAdmin    = "Tài khoản quản trị không thể đăng nhập bằng mạng xã hội."

	MsgOAuthLoginSuccess = "Đăng nhập thành công! Chào mừng %s."

	LogOAuthBeginFailed    = "oauth: begin auth failed"
	LogOAuthCallbackFailed = "oauth: complete auth failed"
	LogOAuthLoginFailed    = "oauth: login/register failed"
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

// ── Profile
const (
	TitleProfile     = "Hồ sơ cá nhân"
	TitleProfileEdit = "Chỉnh sửa hồ sơ"

	MsgProfileUpdateSuccess = "Cập nhật hồ sơ thành công."
	ErrProfileUpdateFailed  = "Không thể cập nhật hồ sơ. Vui lòng thử lại."

	LogProfileLoadFailed   = "profile: load user failed"
	LogProfileUpdateFailed = "profile: update user failed"
)

// ── Bank Accounts
const (
	TitleBankAccounts    = "Tài khoản ngân hàng"
	TitleBankAccountAdd  = "Thêm tài khoản ngân hàng"
	TitleBankAccountEdit = "Chỉnh sửa tài khoản ngân hàng"

	MsgBankAccountCreated    = "Thêm tài khoản ngân hàng thành công."
	MsgBankAccountUpdated    = "Cập nhật tài khoản ngân hàng thành công."
	MsgBankAccountDeleted    = "Xóa tài khoản ngân hàng thành công."
	MsgBankAccountSetDefault = "Đã đặt tài khoản mặc định."

	ErrBankAccountNotFound   = "Tài khoản ngân hàng không tồn tại."
	ErrBankAccountForbidden  = "Bạn không có quyền thao tác tài khoản này."
	ErrBankAccountCreateFail = "Không thể thêm tài khoản ngân hàng."
	ErrBankAccountUpdateFail = "Không thể cập nhật tài khoản ngân hàng."
	ErrBankAccountDeleteFail = "Không thể xóa tài khoản ngân hàng."

	LogBankAccountLoadFailed       = "bank_account: load failed"
	LogBankAccountCreateFailed     = "bank_account: create failed"
	LogBankAccountUpdateFailed     = "bank_account: update failed"
	LogBankAccountDeleteFailed     = "bank_account: delete failed"
	LogBankAccountSetDefaultFailed = "bank_account: set default failed"
)

// ── Form field labels (used in validation error messages)
const (
	FieldFullName        = "Họ tên"
	FieldEmail           = "Email"
	FieldPassword        = "Mật khẩu"
	FieldPasswordConfirm = "Xác nhận mật khẩu"
	FieldPhone           = "Số điện thoại"
	FieldAvatarURL       = "URL ảnh đại diện"
	FieldBankName        = "Tên ngân hàng"
	FieldAccountNumber   = "Số tài khoản"
	FieldAccountHolder   = "Tên chủ tài khoản"
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
	LogTemplateRelPathFail     = "failed to compute relative template path"
	LogSharedTemplateReadFail  = "failed to read shared template"
	LogSharedTemplateParseFail = "failed to parse shared template"
	LogPageTemplateReadFail    = "failed to read page template"
	LogPageTemplateParseFail   = "failed to parse page template"
	LogLoadedPageTemplate      = "loaded page template"

	TemplateNotFoundText = "template not found: "
)
