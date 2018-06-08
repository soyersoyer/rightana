package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/rightana/service"
)

func createUserE(w http.ResponseWriter, r *http.Request) error {
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

var createUser = handleError(createUserE)

func setUserCtx(ctx context.Context, user *service.User) context.Context {
	return context.WithValue(ctx, keyUser, user)
}

func getUserCtx(ctx context.Context) *service.User {
	return ctx.Value(keyUser).(*service.User)
}

func userBaseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			name := chi.URLParam(r, "name")
			user, err := service.GetUserByName(name)
			if err != nil {
				return err
			}
			ctx := setUserCtx(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}

func userAccessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			userID := getUserIDCtx(r.Context())
			user := getUserCtx(r.Context())
			if user.ID != userID {
				return service.ErrAccessDenied
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
	user := getUserCtx(r.Context())
	var input updateUserPasswordT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	if err := service.ChangePassword(user, input.CurrentPassword, input.Password); err != nil {
		return err
	}

	return respond(w, "")
}

var updateUserPassword = handleError(updateUserPasswordE)

type deleteUserInputT struct {
	Password string
}

func deleteUserE(w http.ResponseWriter, r *http.Request) error {
	user := getUserCtx(r.Context())
	var input deleteUserInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	if err := service.DeleteUser(user, input.Password); err != nil {
		return err
	}
	return respond(w, user.Email)
}

var deleteUser = handleError(deleteUserE)
