package messages

// ── Generic / system
const (
	ErrInternalServer = "Đã xảy ra lỗi hệ thống. Vui lòng thử lại sau."
	ErrInvalidForm    = "Dữ liệu gửi lên không hợp lệ."
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
	ErrAccountNotActive   = "Tài khoản của bạn đã bị khóa. Vui lòng liên hệ hỗ trợ."
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
