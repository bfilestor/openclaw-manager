package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

type contextKey string

const userContextKey contextKey = "auth_user"

type UserContext struct {
	UserID string
	Role   user.Role
	JTI    string
}

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	v := ctx.Value(userContextKey)
	u, ok := v.(*UserContext)
	return u, ok
}

// WithUserContext 用于在测试或内部调用时注入用户上下文。
func WithUserContext(ctx context.Context, u *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func AuthMiddleware(jwtSvc *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r.Header.Get("Authorization"))
			if err != nil {
				middleware.WriteAppError(w, middleware.NewUnauthorized())
				return
			}

			claims, err := jwtSvc.VerifyAccessToken(token)
			if err != nil {
				switch {
				case errors.Is(err, ErrTokenExpired):
					middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_EXPIRED", Message: "token expired", StatusCode: http.StatusUnauthorized})
				case errors.Is(err, ErrTokenRevoked):
					middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_REVOKED", Message: "token revoked", StatusCode: http.StatusUnauthorized})
				case errors.Is(err, ErrTokenInvalid):
					middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_INVALID", Message: "token invalid", StatusCode: http.StatusUnauthorized})
				default:
					middleware.WriteAppError(w, middleware.NewUnauthorized())
				}
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &UserContext{UserID: claims.Subject, Role: user.Role(claims.Role), JTI: claims.ID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(minRole user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := GetUserContext(r.Context())
			if !ok || u == nil {
				middleware.WriteAppError(w, middleware.NewUnauthorized())
				return
			}
			if roleWeight(u.Role) < roleWeight(minRole) {
				middleware.WriteAppError(w, middleware.NewForbidden(string(minRole)))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func roleWeight(r user.Role) int {
	switch r {
	case user.RoleAdmin:
		return 3
	case user.RoleOperator:
		return 2
	case user.RoleViewer:
		return 1
	case user.RoleUser:
		return 1
	default:
		return 1
	}
}

func extractBearerToken(v string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(v), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", ErrTokenMissing
	}
	return strings.TrimSpace(parts[1]), nil
}
