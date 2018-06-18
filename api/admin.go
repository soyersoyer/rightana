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
			user := getLoggedInUserCtx(r.Context())
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

func createUserAdminE(w http.ResponseWriter, r *http.Request) error {
	var input service.CreateUserT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	user, err := service.CreateUser(&input)
	if err != nil {
		return err
	}

	return respond(w, user.Email)
}

var createUserAdmin = handleError(createUserAdminE)

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

func deleteUserAdminE(w http.ResponseWriter, r *http.Request) error {
	name := chi.URLParam(r, "name")
	if err := service.DeleteUserByAdmin(name); err != nil {
		return err
	}
	return respond(w, name)
}

var deleteUserAdmin = handleError(deleteUserAdminE)

func getCollectionsE(w http.ResponseWriter, r *http.Request) error {
	collections, err := service.GetCollections()
	if err != nil {
		return err
	}

	return respond(w, collections)
}

var getCollections = handleError(getCollectionsE)
