package db

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusNormal UserStatus = 1
	UserStatusLocked UserStatus = 2
)

// Gender 性别
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Secret Gender = "secret"
)

// OperationType 操作类型
type OperationType string

const (
	OperationTypeOfAPI    OperationType = "api"
	OperationTypeOfRouter OperationType = "router"
)

type SystemConfigKey string

const (
	LoginFailCount       SystemConfigKey = "login_fail_count"         // 登录失败次数,格式 `300 5`
	PwdAesKey            SystemConfigKey = "pwd_aes_key"              // 密码加密盐, 若不为空，则认为开启aes认证，为空则认为关闭
	TenantDBNamePrefix   SystemConfigKey = "tenant_db_name_prefix"    // 租户数据库名称前缀
	LoginTokenPrefix     SystemConfigKey = "login_token_store_prefix" // 登录token存储的前缀
	LoginUserErrorPrefix SystemConfigKey = "login_user_error_prefix"  // 登录用户错误信息存储的前缀
)
