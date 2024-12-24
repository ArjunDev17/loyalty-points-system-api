package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents a standardized success response structure.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorResponse represents a standardized error response structure.
type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Details string `json:"details"`
}

// WarningResponse represents a standardized warning response structure.
type WarningResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Warning APIWarning  `json:"warning"`
}

type APIWarning struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Details string `json:"details"`
}

// WriteSuccessResponse writes a success response to the client.
func WriteSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

// WriteErrorResponse writes an error response to the client.
func WriteErrorResponse(w http.ResponseWriter, code int, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := ErrorResponse{
		Success: false,
		Error:   apiErr,
	}
	json.NewEncoder(w).Encode(response)
}

// WriteWarningResponse writes a warning response to the client.
func WriteWarningResponse(w http.ResponseWriter, data interface{}, apiWarn APIWarning) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := WarningResponse{
		Success: true,
		Data:    data,
		Warning: apiWarn,
	}
	json.NewEncoder(w).Encode(response)
}

// // CreateUserHandler handles user creation and logs the action
// func CreateUserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	if r.Method != http.MethodPost {
// 		WriteErrorResponse(w, http.StatusMethodNotAllowed, APIError{
// 			Code:    "405",
// 			Msg:     "Method Not Allowed",

// // Example Usage
// // func ExampleHandler(w http.ResponseWriter, r *http.Request) {
// 	// On success:
// 	WriteSuccessResponse(w, []string{"item1", "item2"}, "Operation successful.")
//
// 	// On error:
// 	apiErr := APIError{
// 		Code:    "400",
// 		Msg:     "Bad Request",
// 		Details: "Invalid input data",
// 	}
// 	WriteErrorResponse(w, http.StatusBadRequest, apiErr)
//
// 	// On warning:
// 	apiWarn := APIWarning{
// 		Code:    "1001",
// 		Msg:     "Partial Data",
// 		Details: "Some data could not be retrieved.",
// 	}
// 	WriteWarningResponse(w, []string{"item1"}, apiWarn)
// }
