package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/models"
)

type createTokenT struct {
	Email    string
	Password string
}

type createTokenOutT struct {
	ID string `json:"id"`
}

func createTokenE(w http.ResponseWriter, r *http.Request) error {
	var input createTokenT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	tokenID, err := models.CreateAuthToken(input.Email, input.Password)
	if err != nil {
		return err
	}

	return respond(w, createTokenOutT{tokenID})
}

var createToken = handleError(createTokenE)

func deleteTokenE(w http.ResponseWriter, r *http.Request) error {
	tokenID := chi.URLParam(r, "token")
	if err := models.DeleteAuthToken(tokenID); err != nil {
		return err
	}
	return respond(w, tokenID)
}

var deleteToken = handleError(deleteTokenE)
