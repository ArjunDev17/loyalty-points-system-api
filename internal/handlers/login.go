package handlers

import (
	"database/sql"
	"encoding/json"
	"loyalty-points-system-api/config"
	"loyalty-points-system-api/internal/models"
	utils "loyalty-points-system-api/internal/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles user login and logs the action
func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, cfg *config.Config) {
	// Parse the request body
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve user from the database
	var user models.User
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		http.Error(w, "Could not generate access token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.Username)
	if err != nil {
		http.Error(w, "Could not generate refresh token", http.StatusInternalServerError)
		return
	}

	// Store refresh token in the database
	_, err = db.Exec("UPDATE users SET refresh_token = ? WHERE id = ?", refreshToken, user.ID)
	if err != nil {
		http.Error(w, "Could not store refresh token", http.StatusInternalServerError)
		return
	}

	// Log the login action
	utils.LogAction(db, user.ID, "Login", "User logged in successfully")

	// Respond with tokens
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
