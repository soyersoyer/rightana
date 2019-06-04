package service

import (
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/soyersoyer/rightana/internal/db"
)

// AuthToken is the db's authToken struct
type AuthToken = db.AuthToken

// CreateAuthToken creates an AuthToken
func CreateAuthToken(nameOrEmail string, password string) (string, *User, error) {
	var user *User
	var err error
	if strings.Contains(nameOrEmail, "@") {
		user, err = db.GetUserByEmail(nameOrEmail)
	} else {
		user, err = db.GetUserByName(nameOrEmail)
	}
	if err != nil || user == nil {
		return "", nil, ErrUserNotExist.T(nameOrEmail)
	}
	if err := compareHashAndPassword(user.Password, password); err != nil {
		return "", nil, ErrPasswordNotMatch
	}
	token := db.AuthToken{
		ID:      uuid.Must(uuid.NewV4()).String(),
		OwnerID: user.ID,
	}
	if err := db.InsertAuthToken(&token); err != nil {
		return "", nil, ErrDB.Wrap(err, token)
	}
	return token.ID, user, nil
}

// DeleteAuthToken deletes an AuthToken
func DeleteAuthToken(tokenID string) error {
	if err := db.DeleteAuthToken(tokenID); err != nil {
		if err == db.ErrKeyNotExists {
			return ErrAuthtokenNotExist.T(tokenID)
		}
		return ErrDB.Wrap(err, tokenID)
	}
	return nil
}

// CheckAuthToken check whether the AuthToken is valid
func CheckAuthToken(tokenID string) (uint64, error) {
	token, err := getAuthToken(tokenID)
	if err != nil {
		return 0, ErrAuthtokenExpired
	}

	expiryTime := time.Unix(0, token.Created).Add(time.Duration(token.TTL) * time.Second)
	if expiryTime.Before(time.Now()) {
		DeleteAuthToken(tokenID)
		return 0, ErrAuthtokenExpired
	}
	return token.OwnerID, nil
}

func getAuthToken(tokenID string) (*AuthToken, error) {
	token, err := db.GetAuthToken(tokenID)
	if err != nil {
		if err == db.ErrKeyNotExists {
			return nil, ErrAuthtokenNotExist.T(tokenID)
		}
		return nil, ErrDB.Wrap(err, tokenID)
	}
	return token, nil
}
