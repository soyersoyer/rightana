package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/soyersoyer/rightana/service"
)

func adminAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userID := getUserIDCtx(r.Context())
			user, err := service.GetUserByID(userID)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				return service.ErrAccessDenied
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
	name := chi.URLParam(r, "name")
	user, err := service.GetUserInfo(name)
	if err != nil {
		return err
	}
	return respond(w, user)
}

var getUserInfo = handleError(getUserInfoE)

func updateUserE(w http.ResponseWriter, r *http.Request) error {
	var input service.UserUpdateT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	name := chi.URLParam(r, "name")
	if err := service.UpdateUser(name, &input); err != nil {
		return err
	}

	return respond(w, name)
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
