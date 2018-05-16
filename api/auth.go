package api

import (
	"context"
	"net/http"

	"github.com/soyersoyer/k20a/service"
)

func getUserIDCtx(ctx context.Context) uint64 {
	return ctx.Value(keyUserID).(uint64)
}

func setUserIDCtx(ctx context.Context, ID uint64) context.Context {
	return context.WithValue(ctx, keyUserID, ID)
}

func loggedOnlyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(handleError(
		func(w http.ResponseWriter, r *http.Request) error {
			authToken := r.Header.Get("Authorization")

			userID, err := service.CheckAuthToken(authToken)
			if err != nil {
				return err
			}

			ctx := setUserIDCtx(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		}))
}
