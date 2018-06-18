package service

import (
	"github.com/soyersoyer/rightana/db/db"
)

// UserInfoT is struct for clients, stores the user information
type UserInfoT struct {
	ID                  uint64 `json:"id"`
	Email               string `json:"email"`
	Name                string `json:"name"`
	Created             int64  `json:"created"`
	IsAdmin             bool   `json:"is_admin"`
	DisablePwChange     bool   `json:"disable_pw_change"`
	LimitCollections    bool   `json:"limit_collections"`
	CollectionLimit     uint32 `json:"collection_limit"`
	DisableUserDeletion bool   `json:"disable_user_deletion"`
	EmailVerified       bool   `json:"email_verified"`
	CollectionCount     int    `json:"collection_count"`
}

// GetUsers returns all user
func GetUsers() ([]UserInfoT, error) {
	users, err := db.GetUsers()
	if err != nil {
		return nil, ErrDB.Wrap(err)
	}
	userInfos := []UserInfoT{}
	for _, u := range users {
		collections, err := db.GetCollectionsOwnedByUser(u.ID)
		if err != nil {
			return nil, ErrDB.Wrap(err)
		}
		userInfos = append(userInfos, UserInfoT{
			u.ID,
			u.Email,
			u.Name,
			u.Created,
			u.IsAdmin,
			u.DisablePwChange,
			u.LimitCollections,
			u.CollectionLimit,
			u.DisableUserDeletion,
			u.EmailVerified,
			len(collections),
		})
	}
	return userInfos, nil
}

// GetUserInfo fetch an user by the user's email
func GetUserInfo(name string) (*UserInfoT, error) {
	user, err := db.GetUserByName(name)
	if err != nil {
		return nil, ErrUserNotExist.T(name).Wrap(err)
	}
	return &UserInfoT{
		user.ID,
		user.Email,
		user.Name,
		user.Created,
		user.IsAdmin,
		user.DisablePwChange,
		user.LimitCollections,
		user.CollectionLimit,
		user.DisableUserDeletion,
		user.EmailVerified,
		0,
	}, nil
}

// UserUpdateT is the struct for updating a user
type UserUpdateT struct {
	Name                string `json:"name"`
	Email               string `json:"email"`
	Password            string `json:"password"`
	IsAdmin             bool   `json:"is_admin"`
	DisablePwChange     bool   `json:"disable_pw_change"`
	LimitCollections    bool   `json:"limit_collections"`
	CollectionLimit     uint32 `json:"collection_limit"`
	DisableUserDeletion bool   `json:"disable_user_deletion"`
	EmailVerified       bool   `json:"email_verified"`
}

// UpdateUser updates a user with UserUpdateT struct
func UpdateUser(name string, input *UserUpdateT) error {
	user, err := db.GetUserByName(name)
	if err != nil {
		return ErrUserNotExist.T(name).Wrap(err)
	}

	if user.Name != input.Name {
		if !usernameCheck(input.Name) {
			return ErrInvalidUsername.T(input.Name)
		}
		_, err = db.GetUserByName(input.Name)
		if err != nil && err != db.ErrKeyNotExists {
			return ErrDB.T(input.Name).Wrap(err)
		}
		if err == nil {
			return ErrUserNameExist.T(input.Name)
		}

		user.Name = input.Name
	}

	if !emailCheck(input.Email) {
		return ErrInvalidEmail.T(input.Email)
	}
	user.Email = input.Email

	if input.Password != "" {
		if !passwordCheck(input.Password) {
			return ErrPasswordTooShort
		}
		hashedPass, err := hashPassword(input.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPass
	}

	user.IsAdmin = input.IsAdmin
	if input.IsAdmin == false {
		admins, err := db.GetAdminUsers()
		if err != nil {
			return err
		}
		if len(admins) == 1 && admins[0].Email == user.Email {
			return ErrUserIsTheLastAdmin
		}
	}

	user.DisablePwChange = input.DisablePwChange

	user.LimitCollections = input.LimitCollections
	user.CollectionLimit = input.CollectionLimit

	user.DisableUserDeletion = input.DisableUserDeletion

	user.EmailVerified = input.EmailVerified

	err = db.UpdateUser(user)
	if err != nil {
		return ErrDB.Wrap(err)
	}
	return nil
}

// CollectionInfoT is struct for clients, stores the user information
type CollectionInfoT struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnerName     string `json:"owner_name"`
	Created       int64  `json:"created"`
	TeammateCount int    `json:"teammate_count"`
}

// GetCollections returns all collection
func GetCollections() ([]CollectionInfoT, error) {
	collections, err := db.GetCollections()
	if err != nil {
		return nil, ErrDB.Wrap(err)
	}
	collectionInfos := []CollectionInfoT{}
	for _, c := range collections {
		user, err := db.GetUserByID(c.OwnerID)
		if err != nil {
			return nil, ErrDB.T(string(c.OwnerID)).Wrap(err)
		}
		collectionInfos = append(collectionInfos, CollectionInfoT{
			c.ID,
			c.Name,
			user.Name,
			c.Created,
			len(c.Teammates)})
	}
	return collectionInfos, nil
}
