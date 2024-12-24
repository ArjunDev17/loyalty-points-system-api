package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"loyalty-points-system-api/internal/models"
	"loyalty-points-system-api/internal/utils"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserHandler handles user creation and logs the action
func CreateUserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if len(req.Username) == 0 || len(req.Password) < 6 {
		http.Error(w, "Username and password are required, and password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Insert user into the database
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	result, err := db.Exec(query, req.Username, hashedPassword)
	if err != nil {
		// Check for MySQL-specific errors
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 { // Duplicate entry error
				http.Error(w, "Username already exists", http.StatusConflict)
				return
			}
		}

		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving new user ID", http.StatusInternalServerError)
		return
	}

	// Log the user creation action
	utils.LogAction(db, int(userID), "Create User", "New user created successfully")

	// Respond with success
	response := models.CreateUserResponse{
		Message: "User created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
