package api

import (
	"fmt"
	"net/http"
)

// APIError represents an error returned by the zutool API or during the request process.
type APIError struct {
	StatusCode int    // HTTP status code
	Body       string // Raw response body
	Message    string // Specific error message from API response, if available
	Err        error  // Original underlying error, if any
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Message != "" {
		// Prefer the specific message from the API error response
		return fmt.Sprintf("API error: %s (status: %d)", e.Message, e.StatusCode)
	}
	if e.Err != nil {
		// Include underlying error if present
		return fmt.Sprintf("API error (status: %d): %v", e.StatusCode, e.Err)
	}
	// Fallback to status and raw body if no specific message or underlying error
	return fmt.Sprintf("API error (status: %d): %s", e.StatusCode, e.Body)
}

// newAPIError creates a new APIError instance.
// It's kept unexported as it's an internal helper.
func newAPIError(statusCode int, body string, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Body:       body,
		Message:    message,
		Err:        err,
	}
}

// Helper function to create a standard 404 Not Found error message often needed.
func newNotFoundError(resource string, identifier string) *APIError {
	return newAPIError(
		http.StatusNotFound,
		"",
		fmt.Sprintf("%s '%s' が見つかりません", resource, identifier),
		nil,
	)
}
