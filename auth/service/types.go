package service

import (
	"github.com/xiehqing/infra/auth/db"
	"github.com/xiehqing/infra/pkg/cryptox"
)

// cryptoPass 加密密码
func cryptoPass(raw, salt string) string {
	s, _ := cryptox.NewMd5().Encrypt(salt + "<-*Jaime*->" + raw)
	return s
}

type User struct {
	Details    *BaseUserInfo   `json:"details"`
	Permission *UserPermission `json:"permission"`
}

type BaseUserInfo struct {
	ID             int64         `json:"id"`
	Username       string        `json:"username"`
	NickName       string        `json:"nickName"`
	Email          string        `json:"email"`
	Phone          string        `json:"phone"`
	Password       string        `json:"-"`
	Avatar         string        `json:"avatar"`
	Gender         db.Gender     `json:"gender"`
	Birthday       string        `json:"birthday"`
	Signature      string        `json:"signature"`
	LastActiveTime string        `json:"lastActiveTime"`
	Status         db.UserStatus `json:"status"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
}

// convertUserDetails 转换为用户详情
func convertUserDetails(user *db.User) *BaseUserInfo {
	details := &BaseUserInfo{
		ID:        user.ID,
		Username:  user.Username,
		NickName:  user.NickName,
		Email:     user.Email,
		Phone:     user.Phone,
		Password:  user.Password,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Signature: user.Signature,
		Status:    user.Status,
	}
	if user.LastActiveTime != nil {
		details.LastActiveTime = user.LastActiveTime.Format("2006-01-02 15:04:05")
	}
	if user.CreatedAt != nil {
		details.CreatedAt = user.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt != nil {
		details.UpdatedAt = user.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return details
}

type RolePermission struct {
	RoleID     int64    `json:"roleId"`
	Role       *Role    `json:"role"`
	Operations []string `json:"operations"`
}

type TenantPermission struct {
	TenantID int64             `json:"tenantId"`
	Tenant   *Tenant           `json:"tenant"`
	Roles    []*RolePermission `json:"roles"`
}

type Role struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	IsAdmin  bool   `json:"isAdmin"`
	Comment  string `json:"comment"`
	TenantID int64  `json:"tenantId"` // 租户角色，若为0则为系统角色
}

func (r *RolePermission) IsAdmin() bool {
	return r.Role != nil && r.Role.IsAdmin
}

type Tenant struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	DBName  string `json:"dbName"`
	Comment string `json:"comment"`
}

func convertToRole(role *db.Role) *Role {
	return &Role{
		ID:      role.ID,
		Name:    role.Name,
		Code:    role.Code,
		IsAdmin: role.IsAdmin == 1,
		Comment: role.Comment,
	}
}

func convertToTenant(tenant *db.Tenant) *Tenant {
	return &Tenant{
		ID:      tenant.ID,
		Name:    tenant.Name,
		Code:    tenant.Code,
		DBName:  tenant.DBName,
		Comment: tenant.Comment,
	}
}

// IsAdmin 是否是管理员
func (t *TenantPermission) IsAdmin() bool {
	var isAdmin bool
	for _, role := range t.Roles {
		if role.IsAdmin() {
			isAdmin = true
			break
		}
	}
	return isAdmin
}

type UserPermission struct {
	UserID            int64               `json:"userId"`
	SystemPermissions []*RolePermission   `json:"systemPermissions"`
	TenantPermissions []*TenantPermission `json:"tenantPermissions"`
}

func (up *UserPermission) IsSystemAdmin() bool {
	var isAdmin = false
	for _, r := range up.SystemPermissions {
		if r.IsAdmin() {
			isAdmin = true
			break
		}
	}
	return isAdmin
}

// buildUserPermission 构建用户权限
func buildUserPermission(userId int64,
	userRoles []*db.UserRole,
	userTenants []*db.UserTenant,
	roleOperations []*db.RoleOperation) *UserPermission {
	// 角色权限
	var roleOptsMap = make(map[int64][]*db.RoleOperation)
	for _, ro := range roleOperations {
		roleOptsMap[ro.RoleID] = append(roleOptsMap[ro.RoleID], ro)
	}
	var filterUserRoles []*db.UserRole
	var filterUserTenants []*db.UserTenant
	for _, ur := range userRoles {
		if ur.UserID == userId {
			filterUserRoles = append(filterUserRoles, ur)
		}
	}
	for _, ut := range userTenants {
		if ut.UserID == userId {
			filterUserTenants = append(filterUserTenants, ut)
		}
	}
	// 租户角色
	var tenantPermissionMap = make(map[int64]map[int64]*RolePermission) // key : tenantId  value: map[roleId]rolePermission
	var systemPermissionMap = make(map[int64]*RolePermission)
	for _, ur := range filterUserRoles {
		if ur.Role != nil && ur.Role.IsSystemRole() {
			var ops []string
			if ros, ok := roleOptsMap[ur.Role.ID]; ok {
				for _, ro := range ros {
					ops = append(ops, ro.Operation.Name)
				}
			}
			// 系统级权限
			if _, ok := systemPermissionMap[ur.Role.ID]; !ok {
				systemPermissionMap[ur.Role.ID] = &RolePermission{
					RoleID:     ur.Role.ID,
					Role:       convertToRole(ur.Role),
					Operations: ops,
				}
			}
		} else if ur.Role != nil && ur.Role.IsTenantRole() {
			var ops []string
			if ros, ok := roleOptsMap[ur.Role.ID]; ok {
				for _, ro := range ros {
					ops = append(ops, ro.Operation.Name)
				}
			}
			// 组合权限 是否存在此租户的信息
			if ute, ok := tenantPermissionMap[ur.Role.TenantID]; ok {
				// 租户级权限，租户下是否存在此角色
				if _, exist := ute[ur.Role.ID]; !exist {
					tenantPermissionMap[ur.Role.TenantID][ur.Role.ID] = &RolePermission{
						RoleID:     ur.Role.ID,
						Role:       convertToRole(ur.Role),
						Operations: ops,
					}
				}
			} else {
				tenantPermissionMap[ur.Role.TenantID] = make(map[int64]*RolePermission)
				tenantPermissionMap[ur.Role.TenantID][ur.Role.ID] = &RolePermission{
					RoleID:     ur.Role.ID,
					Role:       convertToRole(ur.Role),
					Operations: ops,
				}
			}
		}
	}
	var systemPermissions = make([]*RolePermission, 0)
	for _, sp := range systemPermissionMap {
		systemPermissions = append(systemPermissions, sp)
	}
	var tenantPermissions = make([]*TenantPermission, 0)
	for _, ut := range filterUserTenants {
		var roles = make([]*RolePermission, 0)
		if ute, ok := tenantPermissionMap[ut.TenantID]; ok {
			for _, rp := range ute {
				roles = append(roles, rp)
			}
		}
		tenantPermissions = append(tenantPermissions, &TenantPermission{
			TenantID: ut.TenantID,
			Tenant:   convertToTenant(ut.Tenant),
			Roles:    roles,
		})
	}
	return &UserPermission{
		UserID:            userId,
		SystemPermissions: systemPermissions,
		TenantPermissions: tenantPermissions,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	NickName  string    `json:"nickName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	Avatar    string    `json:"avatar"`
	Gender    db.Gender `json:"gender"`
	Birthday  string    `json:"birthday"`
	Signature string    `json:"signature"`
}
