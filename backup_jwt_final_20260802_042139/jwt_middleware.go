package middleware

import (
	"context"
	"net/http"
	"strings"

	authservice "enterprise-core/api/internal/auth"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

type TokenValidator interface {
	ValidateToken(
		ctx context.Context,
		tokenString string,
	) (*authservice.Claims, error)
}

func JWTAuth(
	tokenValidator TokenValidator,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
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

			claims, err := tokenValidator.ValidateToken(
				r.Context(),
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

			if claims.Type != "access" {
				http.Error(
					w,
					"invalid access token",
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

func RequireRole(
	requiredRole string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			claims, ok := ClaimsFromContext(
				r.Context(),
			)

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

			next.ServeHTTP(
				w,
				r,
			)
		})
	}
}

func ClaimsFromContext(
	ctx context.Context,
) (*authservice.Claims, bool) {
	claims, ok := ctx.Value(
		claimsContextKey,
	).(*authservice.Claims)

	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

func ClaimsContextKeyForTest() interface{} {
	return claimsContextKey
}

func GetClaims(
	r *http.Request,
) *authservice.Claims {
	claims, _ := ClaimsFromContext(
		r.Context(),
	)

	return claims
}
