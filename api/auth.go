package api

import (
	"context"
	"net/http"

	"github.com/soyersoyer/k20a/service"
)

func getUserEmailCtx(ctx context.Context) string {
	return ctx.Value(keyUserEmail).(string)
}

func setUserEmailCtx(ctx context.Context, userEmail string) context.Context {
	return context.WithValue(ctx, keyUserEmail, userEmail)
}

func loggedOnlyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			authToken := r.Header.Get("Authorization")

			userEmail, err := service.CheckAuthToken(authToken)
			if err != nil {
				return err
			}

			ctx := setUserEmailCtx(r.Context(), userEmail)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}
