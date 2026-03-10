package constants

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBanned   = "banned"
)

const (
	RouteHome     = "/"
	RouteLogin    = "/login"
	RouteRegister = "/register"
	RouteLogout   = "/logout"
)

const (
	RoutePublicTours = "/tours"
)

const (
	RouteProfile     = "/profile"
	RouteProfileEdit = "/profile/edit"
)

const (
	RouteBankAccounts      = "/bank-accounts"
	RouteBankAccountCreate = "/bank-accounts/create"
)

const (
	RouteAdminRoot           = "/admin"
	RouteAdminLogin          = "/admin/login"
	RouteAdminLogout         = "/admin/logout"
	RouteAdminDashboard      = "/admin/dashboard"
	RouteAdminCategories     = "/admin/categories"
	RouteAdminCategoryCreate = "/admin/categories/create"
	RouteAdminCategoryEdit   = "/admin/categories/%d/edit"
	RouteAdminCategoryDelete = "/admin/categories/%d/delete"
)

const (
	RouteAdminTours      = "/admin/tours"
	RouteAdminTourCreate = "/admin/tours/create"
	RouteAdminTourEdit   = "/admin/tours/%d/edit"
	RouteAdminTourDelete = "/admin/tours/%d/delete"
)

const (
	RouteAdminTourSchedules      = "/admin/tours/%d/schedules"
	RouteAdminTourScheduleCreate = "/admin/tours/%d/schedules/create"
	RouteAdminScheduleEdit       = "/admin/schedules/%d/edit"
	RouteAdminScheduleDelete     = "/admin/schedules/%d/delete"
)

const (
	TourStatusDraft    = "draft"
	TourStatusActive   = "active"
	TourStatusInactive = "inactive"
)

const (
	ScheduleStatusOpen      = "open"
	ScheduleStatusFull      = "full"
	ScheduleStatusCancelled = "cancelled"
)

const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
)

const DefaultPageLimit = 10
