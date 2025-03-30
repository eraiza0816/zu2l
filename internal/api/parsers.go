package api

import (
	"encoding/json"
	"fmt"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"
)

// parseWeatherPointResponse parses the raw response body for the GetWeatherPoint endpoint.
// It handles the nested JSON string within the "result" field.
func parseWeatherPointResponse(body []byte) (models.GetWeatherPointResponse, error) {
	var finalResult models.GetWeatherPointResponse

	// Define a temporary struct to capture the initial response where "result" is a string
	type tempResp struct {
		Result string `json:"result"`
	}
	var tempResult tempResp

	// Unmarshal the raw body into the temporary struct
	if err := json.Unmarshal(body, &tempResult); err != nil {
		// Check for API error structure within the body first
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// Return an APIError if the body represents a known API error format
			// Using StatusOK because the HTTP request itself might have been successful
			return finalResult, newAPIError(200, string(body), errorResponse.ErrorMessage, nil)
		}
		// Otherwise, return the unmarshalling error for the temp structure
		return finalResult, fmt.Errorf("failed to unmarshal initial response for /getweatherpoint: %w, body: %s", err, string(body))
	}

	// Now, unmarshal the "result" string into the actual WeatherPoints struct
	var weatherPoints models.WeatherPoints
	var points []models.WeatherPoint
	if errUnmarshalArray := json.Unmarshal([]byte(tempResult.Result), &points); errUnmarshalArray != nil {
		// Handle cases like `{"result":"[]"}` correctly, resulting in an empty points slice.
		// If the result string is not a valid JSON array, return an error.
		if tempResult.Result != "[]" { // Allow empty array string "[]"
			return finalResult, fmt.Errorf("failed to unmarshal nested result string for /getweatherpoint: %w, result string: %s", errUnmarshalArray, tempResult.Result)
		}
		// If it was "[]", points will be nil or empty, which is fine.
	}
	weatherPoints.Root = points
	finalResult.Result = weatherPoints

	return finalResult, nil
}

// parseOtenkiASPResponse parses the raw response body from the Otenki ASP API.
func parseOtenkiASPResponse(body []byte) (models.GetOtenkiASPResponse, error) {
	var rawResponse models.GetOtenkiASPRawResponse
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		// Check for API error structure first
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			return models.GetOtenkiASPResponse{}, newAPIError(200, string(body), errorResponse.ErrorMessage, nil)
		}
		return models.GetOtenkiASPResponse{}, fmt.Errorf("failed to unmarshal otenki asp raw response: %w, body: %s", err, string(body))
	}

	// Convert raw response to structured response
	response := models.GetOtenkiASPResponse{
		Status:   rawResponse.Head.Status,
		DateTime: rawResponse.Head.DateTime,
		Elements: make([]models.Element, 0, len(rawResponse.Body.Location.Element)),
	}

	// Iterate through each RawRecord in the Element slice
	for _, rawElem := range rawResponse.Body.Location.Element { // rawElem is of type models.RawRecord
		if len(rawElem.Record) < 2 {
			fmt.Printf("Skipping raw element with insufficient records (need header + data): %+v\n", rawElem.Record)
			continue
		}

		headerRecord := rawElem.Record[0]
		if len(headerRecord.Property) < 2 {
			fmt.Printf("Skipping element with invalid header property structure (need ContentID + Title): %+v\n", headerRecord.Property)
			continue
		}
		headerProps := headerRecord.Property
		contentID, okID := headerProps[0].(string)
		title, okTitle := headerProps[1].(string)
		if !okID || !okTitle {
			fmt.Printf("Skipping element with invalid type for ContentID or Title in header: %+v\n", headerProps)
			continue
		}

		elem := models.Element{
			ContentID: contentID,
			Title:     title,
			Records:   make(map[time.Time]interface{}),
		}

		for _, rawDataProperty := range rawElem.Record[1:] {
			if len(rawDataProperty.Property) < 2 {
				fmt.Printf("Skipping invalid data property structure (need time + value): %+v\n", rawDataProperty.Property)
				continue
			}
			dataProps := rawDataProperty.Property
			timeStr, okTime := dataProps[0].(string)
			value := dataProps[1]

			if !okTime {
				fmt.Printf("Skipping data record with non-string time: %v\n", dataProps[0])
				continue
			}

			// Parse the timestamp (handle multiple formats)
			var t time.Time
			var err error
			parsed := false
			formats := []string{time.RFC3339, "2006-01-02T15:04:05", "20060102"}
			for _, format := range formats {
				t, err = time.Parse(format, timeStr)
				if err == nil {
					parsed = true
					break
				}
			}

			if !parsed {
				fmt.Printf("Skipping data record with unparseable time '%s': tried formats %v\n", timeStr, formats)
				continue
			}
			elem.Records[t] = value
		}
		response.Elements = append(response.Elements, elem)
	}

	return response, nil
}
