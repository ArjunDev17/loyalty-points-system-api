package middleware

import (
	"context"
	"net/http"
	"strings"

	response "loyalty-points-system-api/internal/reponse"
	"loyalty-points-system-api/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware validates the JWT token and extracts the user_id
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
				Code:    "401",
				Msg:     "Unauthorized",
				Details: "Missing Authorization header",
			})
			return
		}

		// Check if the token is in the format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
				Code:    "401",
				Msg:     "Unauthorized",
				Details: "Invalid Authorization header format",
			})
			return
		}

		// Validate the token
		claims, err := utils.ValidateToken(tokenParts[1])
		if err != nil {
			response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
				Code:    "401",
				Msg:     "Unauthorized",
				Details: "Invalid or expired token",
			})
			return
		}

		// Extract the user_id (or username) from the token claims
		username := claims.Username
		if username == "" {
			response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
				Code:    "401",
				Msg:     "Unauthorized",
				Details: "Invalid token payload",
			})
			return
		}

		// Add the user_id/username to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, username)

		// Pass the updated context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
