package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	jwtpkg "github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	"github.com/google/uuid"
)

type contextKey int

const (
	ctxUserID contextKey = iota
	ctxUserRole
)

func ctxGetUserID(r *http.Request) uuid.UUID {
	id, _ := r.Context().Value(ctxUserID).(uuid.UUID)
	return id
}

func ctxGetRole(r *http.Request) entity.Role {
	role, _ := r.Context().Value(ctxUserRole).(entity.Role)
	return role
}

func AuthMiddleware(tokenManager *jwtpkg.TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
			return
		}
		claims, err := tokenManager.ValidateToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxUserRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...entity.Role) func(http.Handler) http.Handler {
	allowed := make(map[entity.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := ctxGetRole(r)
			if _, ok := allowed[role]; !ok {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
