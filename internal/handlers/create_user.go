package handlers

import (
	"database/sql"
	"encoding/json"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	"loyalty-points-system-api/internal/utils"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserHandler handles user creation and logs the action
// CreateUserHandler handles user creation and logs the action
func CreateUserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		response.WriteErrorResponse(w, http.StatusMethodNotAllowed, response.APIError{
			Code:    "405",
			Msg:     "Method Not Allowed",
			Details: "Only POST method is allowed",
		})
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Request Body",
			Details: "Failed to decode JSON body",
		})
		return
	}

	// Validate input
	if len(req.Username) == 0 || len(req.Password) < 6 {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Input",
			Details: "Username is required and password must be at least 6 characters long",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to hash the password",
		})
		return
	}

	// Insert user into the database
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	result, err := db.Exec(query, req.Username, hashedPassword)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 { // Duplicate entry error
				response.WriteErrorResponse(w, http.StatusConflict, response.APIError{
					Code:    "409",
					Msg:     "Conflict",
					Details: "Username already exists",
				})
				return
			}
		}
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to insert user into database",
		})
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to retrieve new user ID",
		})
		return
	}

	// Log the user creation action
	utils.LogAction(db, int(userID), "Create User", "New user created successfully")

	// Respond with success
	response.WriteSuccessResponse(w, map[string]interface{}{
		"user_id": userID,
	}, "User created successfully")
}
