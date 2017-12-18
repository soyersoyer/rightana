package api

import (
	"context"
	"net/http"

	"github.com/soyersoyer/k20a/models"
)

func GetUserEmail(ctx context.Context) string {
	return ctx.Value(keyUserEmail).(string)
}

func SetUserEmail(ctx context.Context, userEmail string) context.Context {
	return context.WithValue(ctx, keyUserEmail, userEmail)
}

func LoggedOnlyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			authToken := r.Header.Get("Authorization")

			userEmail, err := models.CheckAuthToken(authToken)
			if err != nil {
				return err
			}

			ctx := SetUserEmail(r.Context(), userEmail)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}
