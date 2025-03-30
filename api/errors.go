package api

import (
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int
	Body       string
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error: %s (status: %d)", e.Message, e.StatusCode)
	}
	if e.Err != nil {
		return fmt.Sprintf("API error (status: %d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("API error (status: %d): %s", e.StatusCode, e.Body)
}

func newAPIError(statusCode int, body string, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Body:       body,
		Message:    message,
		Err:        err,
	}
}

func newNotFoundError(resource string, identifier string) *APIError {
	return newAPIError(
		http.StatusNotFound,
		"",
		fmt.Sprintf("%s '%s' が見つかりません", resource, identifier),
		nil,
	)
}
