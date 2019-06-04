package api

import (
	"context"
	"net/http"

	"github.com/soyersoyer/rightana/internal/service"
)

func getLoggedInUserCtx(ctx context.Context) *service.User {
	return ctx.Value(keyLoggedInUser).(*service.User)
}

func setLoggedInUserCtx(ctx context.Context, user *service.User) context.Context {
	return context.WithValue(ctx, keyLoggedInUser, user)
}

func loggedOnlyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			authToken := r.Header.Get("Authorization")

			userID, err := service.CheckAuthToken(authToken)
			if err != nil {
				return err
			}

			user, err := service.GetUserByID(userID)
			if err != nil {
				return err
			}
			ctx := setLoggedInUserCtx(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}
