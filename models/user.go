package models

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/soyersoyer/k20a/config"
	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/errors"
)

var (
	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type User = db.User

func CreateUser(email string, password string) (*User, error) {
	if !config.ActualConfig.EnableRegistration {
		return nil, errors.RegistrationDisabled
	}

	if !emailCheck(email) {
		return nil, errors.InvalidEmail.T(email)
	}
	if !passwordCheck(password) {
		return nil, errors.PasswordTooShort
	}
	hashedPass, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &db.User{
		Email:    email,
		Password: hashedPass,
	}
	if err := db.InsertUser(user); err != nil {
		if err == db.ErrKeyExists {
			return nil, errors.UserExist.T(email)
		}
		return nil, errors.DBError.Wrap(err, user)
	}

	return user, nil
}

func ChangePassword(user *User, currentPassword string, password string) error {
	if !passwordCheck(password) {
		return errors.PasswordTooShort
	}
	if err := compareHashAndPassword(user.Password, currentPassword); err != nil {
		return errors.PasswordNotMatch
	}
	hashedPass, err := hashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hashedPass

	if err := db.UpdateUser(user); err != nil {
		return errors.DBError.Wrap(err, user)
	}

	return nil
}

func DeleteUser(user *User, password string) error {
	if err := compareHashAndPassword(user.Password, password); err != nil {
		return errors.PasswordNotMatch
	}
	if err := db.DeleteUser(user); err != nil {
		return errors.DBError.Wrap(err, user)
	}
	return nil
}

func GetUserByEmail(email string) (*User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, errors.UserNotExist.T(email).Wrap(err)
	}
	return user, nil
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
