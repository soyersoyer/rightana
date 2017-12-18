package models

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/soyersoyer/k20a/db/db"
	"github.com/soyersoyer/k20a/errors"
)

type AuthToken = db.AuthToken

func CreateAuthToken(email string, password string) (string, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil || user == nil {
		return "", errors.UserNotExist.T(email)
	}
	if err := compareHashAndPassword(user.Password, password); err != nil {
		return "", errors.PasswordNotMatch
	}
	token := db.AuthToken{
		ID:         uuid.NewV4().String(),
		OwnerEmail: email,
	}
	if err := db.InsertAuthToken(&token); err != nil {
		return "", errors.DBError.Wrap(err, token)
	}
	return token.ID, nil
}

func DeleteAuthToken(tokenID string) error {
	if err := db.DeleteAuthToken(tokenID); err != nil {
		if err == db.ErrKeyNotExists {
			return errors.AuthtokenNotExist.T(tokenID)
		}
		return errors.DBError.Wrap(err, tokenID)
	}
	return nil
}

func CheckAuthToken(tokenID string) (string, error) {
	token, err := getAuthToken(tokenID)
	if err != nil {
		return "", errors.AuthtokenExpired
	}

	expiryTime := time.Unix(0, token.Created).Add(time.Duration(token.TTL) * time.Second)
	if expiryTime.Before(time.Now()) {
		DeleteAuthToken(tokenID)
		return "", errors.AuthtokenExpired
	}
	return token.OwnerEmail, nil
}

func getAuthToken(tokenID string) (*AuthToken, error) {
	token, err := db.GetAuthToken(tokenID)
	if err != nil {
		if err == db.ErrKeyNotExists {
			return nil, errors.AuthtokenNotExist.T(tokenID)
		}
		return nil, errors.DBError.Wrap(err, tokenID)
	}
	return token, nil
}
