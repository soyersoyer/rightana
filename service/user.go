package service

import (
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/soyersoyer/rightana/config"
	"github.com/soyersoyer/rightana/db/db"
)

var (
	emailRegexp    = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	usernameRegexp = regexp.MustCompile("^[a-z0-9.]+$")
)

// User is the the database user struct
type User = db.User

// CreateUserT is a struct for the clients to creating a user
type CreateUserT struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// CreateUser can create an user
func CreateUser(input *CreateUserT) (*User, error) {
	if !config.ActualConfig.EnableRegistration {
		return nil, ErrRegistrationDisabled
	}

	if !usernameCheck(input.Name) {
		return nil, ErrInvalidUsername.T(input.Name)
	}
	if !emailCheck(input.Email) {
		return nil, ErrInvalidEmail.T(input.Email)
	}
	if !passwordCheck(input.Password) {
		return nil, ErrPasswordTooShort
	}

	_, err := db.GetUserByEmail(input.Email)
	if err != nil && err != db.ErrKeyNotExists {
		return nil, ErrDB.T(input.Email).Wrap(err)
	}
	if err == nil {
		return nil, ErrUserEmailExist.T(input.Email)
	}

	_, err = db.GetUserByName(input.Name)
	if err != nil && err != db.ErrKeyNotExists {
		return nil, ErrDB.T(input.Name).Wrap(err)
	}
	if err == nil {
		return nil, ErrUserNameExist.T(input.Name)
	}

	hashedPass, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	isFirstUser, err := isFirstUser()
	if err != nil {
		return nil, err
	}

	user := &db.User{
		Email:    input.Email,
		Name:     input.Name,
		Password: hashedPass,
		Created:  time.Now().UnixNano(),
		IsAdmin:  isFirstUser,
	}

	if err := db.InsertUser(user); err != nil {
		return nil, ErrDB.Wrap(err, user)
	}

	return user, nil
}

// ChangePassword change a user's password when the password match
func ChangePassword(user *User, currentPassword string, password string) error {
	if !passwordCheck(password) {
		return ErrPasswordTooShort
	}
	if user.DisablePwChange {
		return ErrPasswordChangeDisabled
	}
	if err := compareHashAndPassword(user.Password, currentPassword); err != nil {
		return ErrPasswordNotMatch
	}
	hashedPass, err := hashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hashedPass

	if err := db.UpdateUser(user); err != nil {
		return ErrDB.Wrap(err, user)
	}

	return nil
}

// ChangePasswordForce change a user's password
func ChangePasswordForce(user *User, password string) error {
	if !passwordCheck(password) {
		return ErrPasswordTooShort
	}
	hashedPass, err := hashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hashedPass

	if err := db.UpdateUser(user); err != nil {
		return ErrDB.Wrap(err, user)
	}

	return nil
}

// DeleteUser deletes a user when the password patch
func DeleteUser(user *User, password string) error {
	if user.DisableUserDeletion {
		return ErrUserDeletionDisabled
	}
	if user.IsAdmin {
		admins, err := db.GetAdminUsers()
		if err != nil {
			return err
		}
		if len(admins) == 1 && admins[0].Email == user.Email {
			return ErrUserIsTheLastAdmin
		}
	}
	if err := compareHashAndPassword(user.Password, password); err != nil {
		return ErrPasswordNotMatch
	}
	if err := db.DeleteUser(user); err != nil {
		return ErrDB.Wrap(err, user)
	}
	return nil
}

// GetUserByEmail fetch an user by the user's email
func GetUserByEmail(email string) (*User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotExist.T(email).Wrap(err)
	}
	return user, nil
}

// GetUserByName fetch an user by the user's name
func GetUserByName(name string) (*User, error) {
	user, err := db.GetUserByName(name)
	if err != nil {
		return nil, ErrUserNotExist.T(name).Wrap(err)
	}
	return user, nil
}

// GetUserByID fetch an user by the user's email
func GetUserByID(ID uint64) (*User, error) {
	user, err := db.GetUserByID(ID)
	if err != nil {
		return nil, ErrUserNotExist.T(string(ID)).Wrap(err)
	}
	return user, nil
}

func isFirstUser() (bool, error) {
	userCount, err := db.CountUsers()
	if err != nil {
		return false, ErrDB.Wrap(err)
	}
	return userCount == 0, nil
}

func usernameCheck(name string) bool {
	return usernameRegexp.MatchString(name)
}

func emailCheck(email string) bool {
	return emailRegexp.MatchString(email)
}

func passwordCheck(password string) bool {
	return len(password) >= 8
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func compareHashAndPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
