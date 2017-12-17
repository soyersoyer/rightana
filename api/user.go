package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/k20a/errors"
	"github.com/soyersoyer/k20a/models"
)

type createUserT struct {
	Email    string
	Password string
}

func createUserE(w http.ResponseWriter, r *http.Request) error {
	var input createUserT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	user, err := models.CreateUser(input.Email, input.Password)
	if err != nil {
		return err
	}

	return respond(w, user.Email)
}

var createUser = handleError(createUserE)

func SetUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, "user", user)
}

func GetUser(ctx context.Context) *models.User {
	return ctx.Value("user").(*models.User)
}

func userBaseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			email := chi.URLParam(r, "email")
			user, err := models.GetUserByEmail(email)
			if err != nil {
				return err
			}
			ctx := SetUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}

func userAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userEmail := GetUserEmail(r.Context())
			user := GetUser(r.Context())
			if user.Email != userEmail {
				return errors.AccessDenied
			}
			next.ServeHTTP(w, r)
			return nil
		}))
}

type updateUserPasswordT struct {
	CurrentPassword string
	Password        string
}

func updateUserPasswordE(w http.ResponseWriter, r *http.Request) error {
	user := GetUser(r.Context())
	var input updateUserPasswordT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	if err := models.ChangePassword(user, input.CurrentPassword, input.Password); err != nil {
		return err
	}

	return respond(w, user.Email)
}

var updateUserPassword = handleError(updateUserPasswordE)

type deleteUserInputT struct {
	Password string
}

func deleteUserE(w http.ResponseWriter, r *http.Request) error {
	user := GetUser(r.Context())
	var input deleteUserInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errors.InputDecodeFailed.Wrap(err)
	}

	if err := models.DeleteUser(user, input.Password); err != nil {
		return err
	}
	return respond(w, user.Email)
}

var deleteUser = handleError(deleteUserE)
