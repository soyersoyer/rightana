package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/rightana/errors"
	"github.com/soyersoyer/rightana/service"
)

type createTokenT struct {
	NameOrEmail string `json:"name_or_email"`
	Password    string `json:"password"`
}

type createTokenOutT struct {
	ID                string `json:"id"`
	service.UserInfoT `json:"user_info"`
}

func createTokenE(w http.ResponseWriter, r *http.Request) error {
	var input createTokenT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	tokenID, user, err := service.CreateAuthToken(input.NameOrEmail, input.Password)
	if err != nil {
		return err
	}

	return respond(w, createTokenOutT{tokenID, service.UserInfoT{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Created: user.Created,
		IsAdmin: user.IsAdmin},
	})
}

var createToken = handleError(createTokenE)

func deleteTokenE(w http.ResponseWriter, r *http.Request) error {
	tokenID := chi.URLParam(r, "token")
	if err := service.DeleteAuthToken(tokenID); err != nil {
		return err
	}
	return respond(w, tokenID)
}

var deleteToken = handleError(deleteTokenE)
