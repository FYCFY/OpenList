package op

import (
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/singleflight"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

var userG singleflight.Group[*model.User]
var guestUser *model.User
var adminUser *model.User

func GetAdmin() (*model.User, error) {
	if adminUser == nil {
		user, err := db.GetUserByRole(model.ADMIN)
		if err != nil {
			return nil, err
		}
		adminUser = user
	}
	return adminUser, nil
}

func GetGuest() (*model.User, error) {
	if guestUser == nil {
		user, err := db.GetUserByRole(model.GUEST)
		if err != nil {
			return nil, err
		}
		guestUser = user
	}
	return guestUser, nil
}

func GetUserByRole(role int) (*model.User, error) {
	return db.GetUserByRole(role)
}

func GetUserByName(username string) (*model.User, error) {
	if username == "" {
		return nil, errs.EmptyUsername
	}
	if user, exists := Cache.GetUser(username); exists {
		if isUserExpired(user, time.Now().UTC()) {
			_ = db.DeleteUserById(user.ID)
			Cache.DeleteUser(username)
			return nil, errors.New("user expired")
		}
		return user, nil
	}
	user, err, _ := userG.Do(username, func() (*model.User, error) {
		_user, err := db.GetUserByName(username)
		if err != nil {
			return nil, err
		}
		if isUserExpired(_user, time.Now().UTC()) {
			_ = db.DeleteUserById(_user.ID)
			return nil, errors.New("user expired")
		}
		Cache.SetUser(username, _user)
		return _user, nil
	})
	return user, err
}

func GetUserById(id uint) (*model.User, error) {
	u, err := db.GetUserById(id)
	if err != nil {
		return nil, err
	}
	if isUserExpired(u, time.Now().UTC()) {
		_ = db.DeleteUserById(u.ID)
		Cache.DeleteUser(u.Username)
		return nil, errors.New("user expired")
	}
	return u, nil
}

func GetUsers(pageIndex, pageSize int) (users []model.User, count int64, err error) {
	return db.GetUsers(pageIndex, pageSize)
}

func CreateUser(u *model.User) error {
	u.BasePath = utils.FixAndCleanPath(u.BasePath)
	if u.ExpiresAt != nil {
		t := u.ExpiresAt.UTC()
		u.ExpiresAt = &t
	}
	return db.CreateUser(u)
}

func BindUserIP(u *model.User, ip string) error {
	u.WebdavBindIP = &ip
	Cache.DeleteUser(u.Username)
	return db.UpdateUser(u)
}

func ClearUserBindIP(id uint) error {
	u, err := db.GetUserById(id)
	if err != nil {
		return err
	}
	u.WebdavBindIP = nil
	Cache.DeleteUser(u.Username)
	return db.UpdateUser(u)
}

func DeleteUserById(id uint) error {
	old, err := db.GetUserById(id)
	if err != nil {
		return err
	}
	if old.IsAdmin() || old.IsGuest() {
		return errs.DeleteAdminOrGuest
	}
	Cache.DeleteUser(old.Username)
	if err := DeleteSharingsByCreatorId(id); err != nil {
		return errors.WithMessage(err, "failed to delete user's sharings")
	}
	return db.DeleteUserById(id)
}

func UpdateUser(u *model.User) error {
	old, err := db.GetUserById(u.ID)
	if err != nil {
		return err
	}
	if u.IsAdmin() {
		adminUser = nil
	}
	if u.IsGuest() {
		guestUser = nil
	}
	Cache.DeleteUser(old.Username)
	u.BasePath = utils.FixAndCleanPath(u.BasePath)
	if u.ExpiresAt != nil {
		t := u.ExpiresAt.UTC()
		u.ExpiresAt = &t
	}
	return db.UpdateUser(u)
}

func Cancel2FAByUser(u *model.User) error {
	u.OtpSecret = ""
	return UpdateUser(u)
}

func Cancel2FAById(id uint) error {
	user, err := db.GetUserById(id)
	if err != nil {
		return err
	}
	return Cancel2FAByUser(user)
}

func DelUserCache(username string) error {
	user, err := GetUserByName(username)
	if err != nil {
		return err
	}
	if user.IsAdmin() {
		adminUser = nil
	}
	if user.IsGuest() {
		guestUser = nil
	}
	Cache.DeleteUser(username)
	return nil
}
