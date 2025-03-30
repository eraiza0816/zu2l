package models

import (
	"fmt"
	"strings"
	"time"
)

// APIDateTime is a custom type for handling the specific date-time format "YYYY-MM-DD HH" from the API.
type APIDateTime struct {
	time.Time
}

// apiDateTimeLayout defines the expected layout for parsing and formatting APIDateTime.
const apiDateTimeLayout = "2006-01-02 15"

// UnmarshalJSON implements the json.Unmarshaler interface for APIDateTime.
// It parses the string "YYYY-MM-DD HH" into a time.Time object.
func (adt *APIDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" { // Handle null or empty string
		adt.Time = time.Time{} // Set to zero value
		return nil
	}
	t, err := time.Parse(apiDateTimeLayout, s)
	if err != nil {
		// Attempt to parse just the date if time is missing (e.g., "2023-10-26")
		// This might be needed for some API responses, adjust if necessary.
		t, errDate := time.Parse("2006-01-02", s)
		if errDate != nil {
			// Return the original parsing error if date-only parsing also fails
			return fmt.Errorf("failed to parse APIDateTime %q: %w", s, err)
		}
		// If date-only parsing succeeds, use the resulting time (time part will be 00:00:00)
		adt.Time = t
		return nil
	}
	adt.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface for APIDateTime.
// It formats the time.Time object back into the "YYYY-MM-DD HH" string format.
func (adt APIDateTime) MarshalJSON() ([]byte, error) {
	if adt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + adt.Time.Format(apiDateTimeLayout) + `"`), nil
}
