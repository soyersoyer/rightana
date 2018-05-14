package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/service"
)

func adminAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userEmail := getUserEmailCtx(r.Context())
			user, err := service.GetUserByEmail(userEmail)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				return errors.AccessDenied
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

func getUsersE(w http.ResponseWriter, r *http.Request) error {
	users, err := service.GetUsers()
	if err != nil {
		return err
	}

	return respond(w, users)
}

var getUsers = handleError(getUsersE)

func getUserInfoE(w http.ResponseWriter, r *http.Request) error {
	email := chi.URLParam(r, "email")
	user, err := service.GetUserInfo(email)
	if err != nil {
		return err
	}
	return respond(w, user)
}

var getUserInfo = handleError(getUserInfoE)

func updateUserE(w http.ResponseWriter, r *http.Request) error {
	var input service.UserUpdateT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	email := chi.URLParam(r, "email")
	if err := service.UpdateUser(email, &input); err != nil {
		return err
	}

	return respond(w, email)
}

var updateUser = handleError(updateUserE)

func getCollectionsE(w http.ResponseWriter, r *http.Request) error {
	collections, err := service.GetCollections()
	if err != nil {
		return err
	}

	return respond(w, collections)
}

var getCollections = handleError(getCollectionsE)
