package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"loyalty-points-system-api/internal/handlers"
	"loyalty-points-system-api/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

func TestCreateUserHandler(t *testing.T) {
	// Set up a test database connection
	db, err := sql.Open("mysql", "test_user:test_password@tcp(localhost:3306)/loyalty_test_db")
	if err != nil {
		t.Fatalf("Could not connect to test database: %v", err)
	}
	defer db.Close()

	// Ensure the test database is clean
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Could not clean test database: %v", err)
	}

	// Prepare the request payload
	reqBody := models.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonReqBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create an HTTP request
	req, err := http.NewRequest(http.MethodPost, "/create-user", bytes.NewBuffer(jsonReqBody))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUserHandler(w, r, db)
	})
	handler.ServeHTTP(rr, req)

	// Check the status code
	if rr.Code != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Check the response body
	expected := `{"message":"User created successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Verify the user was inserted into the database
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", reqBody.Username).Scan(&userCount)
	if err != nil {
		t.Fatalf("Failed to query test database: %v", err)
	}
	if userCount != 1 {
		t.Errorf("Expected 1 user in the database, got %d", userCount)
	}
}
