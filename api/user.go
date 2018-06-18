package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/soyersoyer/rightana/service"
)

func registerUserE(w http.ResponseWriter, r *http.Request) error {
	var input service.CreateUserT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	user, err := service.RegisterUser(&input)
	if err != nil {
		return err
	}

	return respond(w, user.Email)
}

var registerUser = handleError(registerUserE)

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
			loggedInUser := getLoggedInUserCtx(r.Context())
			user := getUserCtx(r.Context())
			if user.ID != loggedInUser.ID && !loggedInUser.IsAdmin {
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

func sendVerifyEmailE(w http.ResponseWriter, r *http.Request) error {
	user := getUserCtx(r.Context())
	if err := service.SendVerifyEmail(user); err != nil {
		return err
	}
	return respond(w, user.Email)
}

var sendVerifyEmail = handleError(sendVerifyEmailE)

type verifyEmailInputT struct {
	VerificationKey string
}

func verifyEmailE(w http.ResponseWriter, r *http.Request) error {
	user := getUserCtx(r.Context())
	var input verifyEmailInputT
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}
	if err := service.VerifyEmail(user, input.VerificationKey); err != nil {
		return err
	}
	return respond(w, user.Email)
}

var verifyEmail = handleError(verifyEmailE)

type sendResetPasswordInput struct {
	Email string
}

func sendResetPasswordE(w http.ResponseWriter, r *http.Request) error {
	var input sendResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}

	user, err := service.GetUserByEmail(input.Email)
	if err != nil {
		return err
	}

	err = service.SendResetPassword(user)
	if err != nil {
		return err
	}
	return respond(w, user.Email)
}

var sendResetPassword = handleError(sendResetPasswordE)

type resetPasswordInput struct {
	ResetKey string
	Password string
}

func resetPasswordE(w http.ResponseWriter, r *http.Request) error {
	user := getUserCtx(r.Context())
	var input resetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return service.ErrInputDecodeFailed.Wrap(err)
	}
	if err := service.ChangePasswordWithResetKey(user, input.ResetKey, input.Password); err != nil {
		return err
	}
	return respond(w, user.Email)
}

var resetPassword = handleError(resetPasswordE)
