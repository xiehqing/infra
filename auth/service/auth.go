package service

import (
	"github.com/pkg/errors"
	"github.com/xiehqing/infra/auth/db"
	"github.com/xiehqing/infra/pkg/cryptox"
	"github.com/xiehqing/infra/pkg/logs"
	"github.com/xiehqing/infra/pkg/ormx"
)

type AuthService struct {
	queries *db.Queries
}

func NewAuthService(queries *db.Queries) *AuthService {
	return &AuthService{
		queries: queries,
	}
}

// Auth 认证
func (s *AuthService) Auth(username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, errors.New("auth.Auth: username or password is empty")
	}
	pwdAesKey, err := s.queries.GetConfigValue(string(db.PwdAesKey))
	if err != nil {
		return nil, errors.WithMessage(err, "auth.Auth: get pwd salt failed")
	}
	authPwd := s.decryptPass(password, pwdAesKey, false)
	user, err := s.queries.GetUserByUsername(username)
	if err != nil {
		return nil, errors.WithMessage(err, "auth.Auth: get user by username failed")
	}
	if user == nil {
		return nil, errors.New("auth.Auth: user not exist")
	}
	if user.Status == db.UserStatusLocked {
		return nil, errors.New("auth.Auth: user is locked")
	}
	loginPwd := cryptoPass(authPwd, pwdAesKey)
	if loginPwd != user.Password {
		return nil, errors.New("auth.Auth: password is incorrect")
	}
	permission, err := s.Permission(user.ID)
	if err != nil {
		return nil, err
	}
	var userPerm *UserPermission
	if len(permission) > 0 {
		userPerm = permission[0]
	}
	return &User{
		Details:    convertUserDetails(user),
		Permission: userPerm,
	}, nil
}

// decryptPass 解密密码
func (s *AuthService) decryptPass(password, aesKey string, safe bool) string {
	authPwd := password
	if aesKey != "" {
		aes := cryptox.NewAes(cryptox.WithSafe(safe), cryptox.WithKey(aesKey))
		decrypted, err := aes.Decrypt(password)
		if err != nil {
			logs.Error("auth.decryptPass: decrypt password failed：%v", err)
		} else {
			authPwd = string(decrypted)
		}
	}
	return authPwd
}

// Permission 获取用户权限
func (s *AuthService) Permission(userId int64) ([]*UserPermission, error) {
	userIds := []int64{userId}
	users, err := s.queries.GetUserByIDs(userIds)
	if err != nil {
		return nil, err
	}
	userRoles, err := s.queries.GetUserRoles(userIds)
	if err != nil {
		return nil, err
	}
	userTenants, err := s.queries.GetUserTenants(userIds)
	if err != nil {
		return nil, err
	}
	roleOperations, err := s.queries.GetRoleOperations()
	if err != nil {
		return nil, err
	}
	var userPermissions = make([]*UserPermission, 0)
	for _, user := range users {
		userPermissions = append(userPermissions, buildUserPermission(user.ID, userRoles, userTenants, roleOperations))
	}
	return userPermissions, nil
}

// Permissions 获取用户权限
func (s *AuthService) Permissions(userIds []int64) ([]*UserPermission, error) {
	users, err := s.queries.GetUserByIDs(userIds)
	if err != nil {
		return nil, err
	}
	userRoles, err := s.queries.GetUserRoles(userIds)
	if err != nil {
		return nil, err
	}
	userTenants, err := s.queries.GetUserTenants(userIds)
	if err != nil {
		return nil, err
	}
	roleOperations, err := s.queries.GetRoleOperations()
	if err != nil {
		return nil, err
	}
	var userPermissions = make([]*UserPermission, 0)
	for _, user := range users {
		userPermissions = append(userPermissions, buildUserPermission(user.ID, userRoles, userTenants, roleOperations))
	}
	return userPermissions, nil
}

// Info 获取用户信息
func (s *AuthService) Info(userId int64) (*BaseUserInfo, error) {
	user, err := s.queries.GetUserByID(userId)
	if err != nil {
		return nil, err
	}
	return convertUserDetails(user), nil
}

func (s *AuthService) Users() ([]*User, error) {
	users, err := s.queries.GetUsers("username != ?", "admin")
	if err != nil {
		return nil, err
	}
	userIds := ormx.GetObjIDs(users)
	permissions, err := s.Permissions(userIds)
	if err != nil {
		return nil, err
	}
	var permMap = make(map[int64]*UserPermission)
	for _, perm := range permissions {
		permMap[perm.UserID] = perm
	}
	var infos = make([]*User, 0)
	for _, user := range users {
		var perm = permMap[user.ID]
		infos = append(infos, &User{
			Details:    convertUserDetails(user),
			Permission: perm,
		})
	}
	return infos, nil
}
