package service

import (
	"github.com/soyersoyer/rightana/db/db"
	"github.com/soyersoyer/rightana/errors"
)

// UserInfoT is struct for clients, stores the user information
type UserInfoT struct {
	ID              uint64 `json:"id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Created         int64  `json:"created"`
	IsAdmin         bool   `json:"is_admin"`
	DisablePwChange bool   `json:"disable_pw_change"`
	CollectionCount int    `json:"collection_count"`
}

// GetUsers returns all user
func GetUsers() ([]UserInfoT, error) {
	users, err := db.GetUsers()
	if err != nil {
		return nil, errors.DBError.Wrap(err)
	}
	userInfos := []UserInfoT{}
	for _, u := range users {
		collections, err := db.GetCollectionsOwnedByUser(u.ID)
		if err != nil {
			return nil, errors.DBError.Wrap(err)
		}
		userInfos = append(userInfos, UserInfoT{u.ID, u.Email, u.Name, u.Created, u.IsAdmin, u.DisablePwChange, len(collections)})
	}
	return userInfos, nil
}

// GetUserInfo fetch an user by the user's email
func GetUserInfo(name string) (*UserInfoT, error) {
	user, err := db.GetUserByName(name)
	if err != nil {
		return nil, errors.UserNotExist.T(name).Wrap(err)
	}
	return &UserInfoT{
		user.ID,
		user.Email,
		user.Name,
		user.Created,
		user.IsAdmin,
		user.DisablePwChange,
		0,
	}, nil
}

// UserUpdateT is the struct for updating a user
type UserUpdateT struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	DisablePwChange bool   `json:"disable_pw_change"`
}

// UpdateUser updates a user with UserUpdateT struct
func UpdateUser(name string, input *UserUpdateT) error {
	user, err := db.GetUserByName(name)
	if err != nil {
		return errors.UserNotExist.T(name).Wrap(err)
	}

	if name != input.Name {
		if !usernameCheck(input.Name) {
			return errors.InvalidUsername.T(input.Name)
		}
		_, err = db.GetUserByName(input.Name)
		if err != nil && err != db.ErrKeyNotExists {
			return errors.DBError.T(input.Name).Wrap(err)
		}
		if err == nil {
			return errors.UserNameExist.T(input.Name)
		}

		user.Name = input.Name
	}

	if !emailCheck(input.Email) {
		return errors.InvalidEmail.T(input.Email)
	}
	user.Email = input.Email

	if input.Password != "" {
		if !passwordCheck(input.Password) {
			return errors.PasswordTooShort
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
			return errors.AccessDenied
		}
	}

	user.DisablePwChange = input.DisablePwChange

	err = db.UpdateUser(user)
	if err != nil {
		return errors.DBError.Wrap(err)
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
		return nil, errors.DBError.Wrap(err)
	}
	collectionInfos := []CollectionInfoT{}
	for _, c := range collections {
		user, err := db.GetUserByID(c.OwnerID)
		if err != nil {
			return nil, errors.DBError.T(string(c.OwnerID)).Wrap(err)
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
