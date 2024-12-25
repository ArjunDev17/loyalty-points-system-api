package handlers

import (
	"database/sql"
	"encoding/json"
	"log"

	"net/http"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("GetAllUsersHandler: Fetching all users.")

	// Query to fetch id and username only
	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Slice to hold users
	var users []User

	// Iterate through rows and scan into User struct
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			log.Printf("Error scanning user row: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating rows: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Encode users to JSON and send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Error encoding users to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
