package db

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Queries struct {
	db *gorm.DB
}

func NewQueries(db *gorm.DB) *Queries {
	return &Queries{db: db}
}

// GetUsers 根据条件获取用户列表
func (q *Queries) GetUsers(where string, args ...interface{}) ([]*User, error) {
	var users []*User
	err := q.db.Where(where, args...).Find(&users).Error
	if err != nil {
		return nil, errors.WithMessage(err, "auth.queries: get users failed")
	}
	return users, nil
}

// GetUserByID 根据用户id获取用户
func (q *Queries) GetUserByID(id int64) (*User, error) {
	if id <= 0 {
		return nil, errors.New("auth.queries: user id is invalid")
	}
	users, err := q.GetUsers("id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// GetUserByUsername 根据用户名获取用户
func (q *Queries) GetUserByUsername(username string) (*User, error) {
	if username == "" {
		return nil, errors.New("auth.queries: username is invalid")
	}
	users, err := q.GetUsers("username = ?", username)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// GetConfigValue 根据配置key获取配置值
func (q *Queries) GetConfigValue(configKey string) (string, error) {
	var lst []string
	err := q.db.Model(&SystemConfigs{}).Where("config_key = ?", configKey).Pluck("config_value", &lst).Error
	if err != nil {
		return "", errors.WithMessage(err, "auth.queries: get system config value failed")
	}
	if len(lst) == 0 {
		return "", nil
	}
	return lst[0], nil
}

// GetUserByIDs 根据用户id列表获取用户列表
func (q *Queries) GetUserByIDs(ids []int64) ([]*User, error) {
	return q.GetUsers("id in (?)", ids)
}

// GetUserRoles 根据用户id列表获取用户角色列表
func (q *Queries) GetUserRoles(userIds []int64) ([]*UserRole, error) {
	var userRoles []*UserRole
	err := q.db.Where("user_id in (?)", userIds).Preload("Role").Find(&userRoles).Error
	if err != nil {
		return nil, errors.WithMessage(err, "auth.queries: get user roles failed")
	}
	return userRoles, nil
}

// GetRoleOperations 根据角色id列表获取角色操作列表
func (q *Queries) GetRoleOperations() ([]*RoleOperation, error) {
	var roleOperations []*RoleOperation
	err := q.db.Preload("Operation").Find(&roleOperations).Error
	if err != nil {
		return nil, errors.WithMessage(err, "auth.queries: get role operations failed")
	}
	return roleOperations, nil
}

// GetUserTenants 根据用户id列表获取用户租户列表
func (q *Queries) GetUserTenants(userIds []int64) ([]*UserTenant, error) {
	var userTenants []*UserTenant
	err := q.db.Where("user_id in (?)", userIds).Preload("Tenant").Find(&userTenants).Error
	if err != nil {
		return nil, errors.WithMessage(err, "auth.queries: get user tenants failed")
	}
	return userTenants, nil
}
