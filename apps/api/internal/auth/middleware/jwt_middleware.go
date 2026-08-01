package middleware

import (
	"context"
	"net/http"
	"strings"

	"enterprise-core/api/internal/auth/token"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func JWTAuth(tokenManager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(
				r.Header.Get("Authorization"),
			)

			if authHeader == "" {
				http.Error(
					w,
					"missing authorization header",
					http.StatusUnauthorized,
				)
				return
			}

			parts := strings.Fields(authHeader)

			if len(parts) != 2 ||
				!strings.EqualFold(parts[0], "Bearer") {
				http.Error(
					w,
					"invalid authorization header",
					http.StatusUnauthorized,
				)
				return
			}

			accessToken := strings.TrimSpace(parts[1])

			if accessToken == "" {
				http.Error(
					w,
					"missing access token",
					http.StatusUnauthorized,
				)
				return
			}

			claims, err := tokenManager.ValidateAccessToken(
				accessToken,
			)
			if err != nil {
				http.Error(
					w,
					"invalid or expired token",
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())

			if !ok {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			if !strings.EqualFold(
				strings.TrimSpace(claims.Role),
				strings.TrimSpace(requiredRole),
			) {
				http.Error(
					w,
					"forbidden",
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(
	ctx context.Context,
) (*token.Claims, bool) {
	claims, ok := ctx.Value(
		claimsContextKey,
	).(*token.Claims)

	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

func ClaimsContextKeyForTest() interface{} {
	return claimsContextKey
}
