package handlers

import (
	"database/sql"
	"encoding/json"
	"loyalty-points-system-api/config"
	utils "loyalty-points-system-api/internal/utils"
	"net/http"
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, cfg *config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Generate a new access token
	accessToken, err := utils.GenerateAccessToken(claims.Username)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	// Respond with the new access token
	response := map[string]string{
		"access_token": accessToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
