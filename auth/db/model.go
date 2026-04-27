package db

import (
	"github.com/xiehqing/infra/pkg/ormx"
	"time"
)

type User struct {
	ormx.DeleteAbleModel
	Username       string     `json:"username" gorm:"column:username;type:varchar(255);not null"`
	NickName       string     `json:"nickName" gorm:"column:nickName;type:varchar(255);"`
	Password       string     `json:"password" gorm:"column:password;type:varchar(255);not null"`
	Phone          string     `json:"phone" gorm:"column:phone;type:varchar(255);"`
	Email          string     `json:"email" gorm:"column:email;type:varchar(255);"`
	Avatar         string     `json:"avatar" gorm:"column:avatar;type:varchar(3000);"`
	Gender         Gender     `json:"gender" gorm:"column:gender;type:varchar(10);"`
	Birthday       string     `json:"birthday" gorm:"column:birthday;type:varchar(25);"`
	Signature      string     `json:"signature" gorm:"column:signature;type:varchar(5000);"`
	Status         UserStatus `json:"status" gorm:"column:status;type:tinyint;not null"`
	LastActiveTime *time.Time `json:"lastActiveTime" gorm:"column:last_active_time;type:dateTime;"`
}

func (u *User) TableName() string {
	return "users"
}

// WxUser 微信用户 - 小程序关联微信账号
type WxUser struct {
	ormx.BaseModel
	OpenId  string `json:"openId" gorm:"column:open_id;type:varchar(255);not null"`
	UnionId string `json:"unionId" gorm:"column:union_id;type:varchar(255);not null"`
	UserID  int64  `json:"userId" gorm:"column:user_id;type:bigint;not null"`
}

func (w *WxUser) TableName() string {
	return "wx_users"
}

// Role 角色
type Role struct {
	ormx.BaseModel
	Name     string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Code     string `json:"code" gorm:"column:code;type:varchar(255);not null"`
	IsAdmin  int    `json:"isAdmin" gorm:"column:is_admin;type:tinyint;not null"`
	Comment  string `json:"comment" gorm:"column:comment;type:varchar(255);"`
	TenantID int64  `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"` // 租户角色，若为0则为系统角色
}

func (r *Role) TableName() string {
	return "role"
}

func (r *Role) IsSystemRole() bool {
	return r.TenantID == 0
}

func (r *Role) IsTenantRole() bool {
	return r.TenantID != 0
}

// Operation 操作[包括但不限于，api,router]
type Operation struct {
	ormx.BaseModel
	Name        string        `json:"name" gorm:"column:name;type:varchar(50);not null"`
	DisplayName string        `json:"displayName" gorm:"column:display_name;type:varchar(50);not null"`
	Type        OperationType `json:"type" gorm:"column:type;type:tinyint;not null"`
}

func (o *Operation) TableName() string {
	return "operation"
}

// RoleOperation 角色操作关联
type RoleOperation struct {
	ormx.BaseModel
	RoleID      int64      `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role        *Role      `json:"role" gorm:"foreignKey:RoleID"`
	OperationID int64      `json:"operationId" gorm:"column:operation_id;type:bigint;not null"`
	Operation   *Operation `json:"operation" gorm:"foreignKey:OperationID"`
}

func (r *RoleOperation) TableName() string {
	return "role_operation"
}

// Tenant 租户
type Tenant struct {
	ormx.DeleteAbleModel
	Name    string `json:"name" gorm:"column:name;type:varchar(30);not null"`
	Code    string `json:"code" gorm:"column:code;type:varchar(50);not null"`
	DBName  string `json:"dbName" gorm:"column:dbName;type:varchar(100);not null"`
	Comment string `json:"comment" gorm:"column:comment;type:varchar(2000);"`
}

func (t *Tenant) TableName() string {
	return "tenant"
}

// UserTenant 用户租户关联
type UserTenant struct {
	ormx.BaseModel
	UserID   int64   `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	User     *User   `json:"user" gorm:"foreignKey:UserID"`
	TenantID int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"`
	Tenant   *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

func (ut *UserTenant) TableName() string {
	return "user_tenant"
}

// UserRole 用户角色关联
type UserRole struct {
	ormx.BaseModel
	UserID int64 `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	User   *User `json:"user" gorm:"foreignKey:UserID"`
	RoleID int64 `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role   *Role `json:"role" gorm:"foreignKey:RoleID"`
}

func (ur *UserRole) TableName() string {
	return "user_role"
}

// UserAuditLog 审计日志
type UserAuditLog struct {
	ormx.BaseModel
	UserID      int64  `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	Action      string `json:"action" gorm:"column:action;type:varchar(500);not null"`
	Description string `json:"description" gorm:"column:description;type:varchar(2500);not null"`
	IP          string `json:"ip" gorm:"column:ip;type:varchar(255);not null"`
	UserAgent   string `json:"userAgent" gorm:"column:user_agent;type:varchar(255);not null"`
	// ...
}

func (al *UserAuditLog) TableName() string {
	return "user_audit_log"
}

// SystemConfigs 系统配置
type SystemConfigs struct {
	ormx.BaseModel
	Key     string `json:"configKey" gorm:"column:config_key;type:varchar(255);not null"`
	Value   string `json:"configValue" gorm:"column:config_value;type:text;not null"`
	Comment string `json:"comment" gorm:"column:comment;type:varchar(2000);"`
}

func (s *SystemConfigs) TableName() string {
	return "system_configs"
}
