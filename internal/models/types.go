package models

import (
	"fmt"
	"regexp"
	"time"
)

// WeatherPoint represents a weather observation point.
type WeatherPoint struct {
	CityCode string `json:"city_code"`
	NameKata string `json:"name_kata"`
	Name     string `json:"name"`
}

// Validate checks if the WeatherPoint fields are valid.
func (w *WeatherPoint) Validate() error {
	// Note: Original regex `^\\d{5}$` seems incorrect for Go, using `^\d{5}$`.
	if matched, _ := regexp.MatchString(`^\d{5}$`, w.CityCode); !matched {
		return fmt.Errorf("CityCode must be a 5-digit number, got: %s", w.CityCode)
	}
	// Note: Original regex `^[\\uff61-\\uff9f]+$` seems incorrect for Go, using `^[\uff61-\uff9f]+$`.
	if matched, _ := regexp.MatchString(`^[\uff61-\uff9f]+$`, w.NameKata); !matched {
		return fmt.Errorf("NameKata must be in half-width katakana, got: %s", w.NameKata)
	}
	return nil
}

// WeatherPoints is a collection of WeatherPoint.
type WeatherPoints struct {
	Root []WeatherPoint `json:"root"`
}

// GetWeatherPointResponse represents the response structure for the weather point API.
type GetWeatherPointResponse struct {
	Result WeatherPoints `json:"result"`
}

// GetPainStatus represents the pain status data.
type GetPainStatus struct {
	AreaName    string  `json:"area_name"`
	TimeStart   string  `json:"time_start"` // Consider using time.Time if format is consistent
	TimeEnd     string  `json:"time_end"`   // Consider using time.Time if format is consistent
	RateNormal  float64 `json:"rate_0"`
	RateLittle  float64 `json:"rate_1"`
	RatePainful float64 `json:"rate_2"`
	RateBad     float64 `json:"rate_3"`
}

// Validate checks if the GetPainStatus rates are valid.
func (g *GetPainStatus) Validate() error {
	if g.RateNormal < 0 {
		return fmt.Errorf("RateNormal must be non-negative, got: %f", g.RateNormal)
	}
	if g.RateLittle < 0 {
		return fmt.Errorf("RateLittle must be non-negative, got: %f", g.RateLittle)
	}
	if g.RatePainful < 0 {
		return fmt.Errorf("RatePainful must be non-negative, got: %f", g.RatePainful)
	}
	if g.RateBad < 0 {
		return fmt.Errorf("RateBad must be non-negative, got: %f", g.RateBad)
	}
	return nil
}

// GetPainStatusResponse represents the response structure for the pain status API.
type GetPainStatusResponse struct {
	PainnoterateStatus GetPainStatus `json:"painnoterate_status"`
}

// WeatherStatusByTime represents weather status at a specific time.
type WeatherStatusByTime struct {
	Time          string          `json:"time"` // Changed from int to string to match JSON
	Weather       WeatherEnum     `json:"weather"`
	Temp          *string         `json:"temp"` // Changed from *float64 to *string to match JSON
	Pressure      string          `json:"pressure"` // Changed from float64 to string to match JSON ("1007.5")
	PressureLevel PressureLevelEnum `json:"pressure_level"`
}

// Validate checks if the WeatherStatusByTime fields are valid.
// Note: Time validation removed as type changed to string. Add parsing/validation if needed.
func (w *WeatherStatusByTime) Validate() error {
	// Add validation logic here if necessary, e.g., check if Time can be parsed to an int between 0-23
	return nil
}

// GetWeatherStatusResponse represents the response structure for the weather status API.
type GetWeatherStatusResponse struct {
	PlaceName        string                `json:"place_name"`
	PlaceID          string                `json:"place_id"`
	PrefecturesID    AreaEnum              `json:"prefectures_id"`
	DateTime         APIDateTime           `json:"dateTime"`
	Yesterday        []WeatherStatusByTime `json:"yesterday"`
	Today            []WeatherStatusByTime `json:"today"`
	Tomorrow         []WeatherStatusByTime `json:"tomorrow"`
	DayAfterTomorrow []WeatherStatusByTime `json:"dayaftertomorrow"`
}

// Validate checks if the GetWeatherStatusResponse PlaceID field is valid.
func (g *GetWeatherStatusResponse) Validate() error {
	// Note: Original regex `^\\d{3}$` seems incorrect for Go, using `^\d{3}$`.
	if matched, _ := regexp.MatchString(`^\d{3}$`, g.PlaceID); !matched {
		return fmt.Errorf("PlaceID must be a 3-digit number, got: %s", g.PlaceID)
	}
	return nil
}

// RawHead represents the head part of the raw Otenki ASP response.
type RawHead struct {
	ContentsID string      `json:"contentsId"`
	Title      string      `json:"title"`
	DateTime   APIDateTime `json:"dateTime"`
	Status     string      `json:"status"`
}

// RawProperty represents a property in the raw Otenki ASP response.
// Using interface{} as the structure is not fully defined/consistent.
type RawProperty struct {
	Property []interface{} `json:"property"`
}

// RawRecord represents a record in the raw Otenki ASP response.
type RawRecord struct {
	Record []RawProperty `json:"record"`
}

// RawElement represents an element in the raw Otenki ASP response.
type RawElement struct {
	Element []RawRecord `json:"element"`
}

// RawBody represents the body part of the raw Otenki ASP response.
type RawBody struct {
	Location RawElement `json:"location"`
}

// GetOtenkiASPRawResponse represents the raw response structure from the Otenki ASP API.
type GetOtenkiASPRawResponse struct {
	Head RawHead `json:"head"`
	Body RawBody `json:"body"`
}

// Element represents a processed element from the Otenki ASP response.
type Element struct {
	ContentID string                 `json:"content_id"`
	Title     string                 `json:"title"`
	Records   map[time.Time]interface{} `json:"records"` // Using time.Time as key, interface{} for value flexibility
}

// GetOtenkiASPResponse represents the processed response structure for the Otenki ASP API.
type GetOtenkiASPResponse struct {
	Status   string      `json:"status"`
	DateTime APIDateTime `json:"date_time"`
	Elements []Element   `json:"elements"`
}

// Validate for GetOtenkiASPResponse (currently does nothing as per original).
func (g *GetOtenkiASPResponse) Validate() error {
	return nil
}

// ErrorResponse represents a generic API error response.
type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// SetWeatherPointResponse represents the response structure for setting the weather point.
type SetWeatherPointResponse struct {
	Response string `json:"response"`
}
