package service

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/soyersoyer/rightana/db/db"
	"github.com/soyersoyer/rightana/errors"
)

// AuthToken is the db's authToken struct
type AuthToken = db.AuthToken

// CreateAuthToken creates an AuthToken
func CreateAuthToken(email string, password string) (string, *User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil || user == nil {
		return "", nil, errors.UserNotExist.T(email)
	}
	if err := compareHashAndPassword(user.Password, password); err != nil {
		return "", nil, errors.PasswordNotMatch
	}
	token := db.AuthToken{
		ID:      uuid.Must(uuid.NewV4()).String(),
		OwnerID: user.ID,
	}
	if err := db.InsertAuthToken(&token); err != nil {
		return "", nil, errors.DBError.Wrap(err, token)
	}
	return token.ID, user, nil
}

// DeleteAuthToken deletes an AuthToken
func DeleteAuthToken(tokenID string) error {
	if err := db.DeleteAuthToken(tokenID); err != nil {
		if err == db.ErrKeyNotExists {
			return errors.AuthtokenNotExist.T(tokenID)
		}
		return errors.DBError.Wrap(err, tokenID)
	}
	return nil
}

// CheckAuthToken check whether the AuthToken is valid
func CheckAuthToken(tokenID string) (uint64, error) {
	token, err := getAuthToken(tokenID)
	if err != nil {
		return 0, errors.AuthtokenExpired
	}

	expiryTime := time.Unix(0, token.Created).Add(time.Duration(token.TTL) * time.Second)
	if expiryTime.Before(time.Now()) {
		DeleteAuthToken(tokenID)
		return 0, errors.AuthtokenExpired
	}
	return token.OwnerID, nil
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
