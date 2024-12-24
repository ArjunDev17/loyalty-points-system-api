package handlers

import (
	"database/sql"
	"encoding/json"
	"loyalty-points-system-api/config"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	utils "loyalty-points-system-api/internal/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles user login and logs the action
func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, cfg *config.Config) {
	// Parse the request body
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Request Body",
			Details: "Failed to decode JSON body",
		})
		return
	}

	// Retrieve user from the database
	var user models.User
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
			Code:    "401",
			Msg:     "Unauthorized",
			Details: "Invalid username or password",
		})
		return
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
			Code:    "401",
			Msg:     "Unauthorized",
			Details: "Invalid username or password",
		})
		return
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Could not generate access token",
		})
		return
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.Username)
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Could not generate refresh token",
		})
		return
	}

	// Store refresh token in the database
	_, err = db.Exec("UPDATE users SET refresh_token = ? WHERE id = ?", refreshToken, user.ID)
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Could not store refresh token",
		})
		return
	}

	// Log the login action
	utils.LogAction(db, user.ID, "Login", "User logged in successfully")

	// Respond with tokens
	response.WriteSuccessResponse(w, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, "Login successful")
}
