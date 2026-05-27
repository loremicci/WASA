package api

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "userId"

func (rt *_router) AuthMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify user exists
		_, err := rt.db.GetUserByID(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), userContextKey, token)
		next(w, r.WithContext(ctx), ps)
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) string {
	val := r.Context().Value(userContextKey)
	if val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
