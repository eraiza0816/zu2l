package api

import (
	"encoding/json"
	"errors" // Added for errors.As
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"zutool/models"
)

// Constants for API interaction (Exported if needed by other packages, otherwise keep unexported)
const (
	baseURL         = "https://zutool.jp/api"                     // zutool API base URL - unexported
	otenkiASPBaseURL = "https://ap.otenki.com/OtenkiASP/asp" // Otenki ASP API base URL - unexported
	timeout         = 10 * time.Second                          // Default HTTP timeout - unexported
	ua              = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36" // User-Agent string - unexported
	OtenkiASPBaseURL = "https://ap.otenki.com/OtenkiASP/asp" // Otenki ASP API base URL
	Timeout         = 10 * time.Second                          // Default HTTP timeout
	UA              = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36" // User-Agent string
)

var httpClient = &http.Client{
	Timeout: timeout,
}

// APIError represents an error returned by the zutool API or during the request process.
// Exported for type checking in tests.
type APIError struct {
	StatusCode int    // HTTP status code
	Body       string // Raw response body
	Message    string // Specific error message from API response, if available
	Err        error  // Original underlying error, if any
}

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

func newAPIError(statusCode int, body string, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Body:       body,
		Message:    message,
		Err:        err,
	}
}

func doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", ua)
	resp, err := httpClient.Do(req)
	if err != nil {
		// Network or connection error
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Error reading response body
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Non-200 status code
		var errorResponse models.ErrorResponse
		// Try decoding zutool's specific error format
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			return nil, newAPIError(resp.StatusCode, string(body), errorResponse.ErrorMessage, nil)
		}
		// Fallback to generic API error if decoding fails
		return nil, newAPIError(resp.StatusCode, string(body), "", nil)
	}

	return body, nil
}

func _get[T any](
	path string,
	param string,
	setWeatherPoint *string,
) (T, error) {
	var result T

	if path == "/getpainstatus" && setWeatherPoint != nil && *setWeatherPoint != "" {
		setWeatherPointURL := fmt.Sprintf("%s/setweatherpoint/%s", baseURL, *setWeatherPoint)
		req, err := http.NewRequest("GET", setWeatherPointURL, nil)
		if err != nil {
			return result, fmt.Errorf("failed to create setweatherpoint request: %w", err)
		}

		setBody, err := doRequest(req)
		if err != nil {
			return result, fmt.Errorf("setweatherpoint request failed: %w", err)
		}

		var setResp models.SetWeatherPointResponse
		if err := json.Unmarshal(setBody, &setResp); err != nil {
			return result, fmt.Errorf("failed to unmarshal setweatherpoint response: %w, body: %s", err, string(setBody))
		}
		if setResp.Response != "ok" {
			return result, fmt.Errorf("setweatherpoint response was not 'ok': %s", setResp.Response)
		}
	}

	apiURL := fmt.Sprintf("%s%s/%s", baseURL, path, param)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return result, fmt.Errorf("failed to create main request for %s: %w", path, err)
	}

	body, err := doRequest(req)
	if err != nil {
		// If doRequest returned an error (e.g., non-200 status), return it directly.
		return result, fmt.Errorf("main request failed for %s: %w", path, err)
	}

	// --- Start: Check for API error within 200 OK response ---
	var errorResponse models.ErrorResponse
	// Try unmarshalling into the error structure first.
	// We ignore the error here because failing to unmarshal into ErrorResponse is expected for valid data.
	_ = json.Unmarshal(body, &errorResponse)

	// If unmarshalling succeeded AND we got an error message, it means the API returned 200 OK but signaled an error.
	if errorResponse.ErrorMessage != "" {
		// Create an APIError, using StatusOK because the HTTP request itself was successful.
		// The actual error condition is indicated by the ErrorMessage.
		return result, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
	}
	// --- End: Check for API error within 200 OK response ---


	// If it wasn't an error structure, proceed to unmarshal into the expected result type T.
	if err := json.Unmarshal(body, &result); err != nil {
		// This error occurs if the body is valid JSON but doesn't match the structure of T.
		return result, fmt.Errorf("failed to unmarshal successful response body into target type for %s: %w, body: %s", path, err, string(body))
	}

	// Successfully unmarshalled into the target type T.
	return result, nil
}

func GetPainStatus(
	areaCode string,
	setWeatherPoint *string,
) (models.GetPainStatusResponse, error) {
	return _get[models.GetPainStatusResponse](
		"/getpainstatus",
		areaCode,
		setWeatherPoint,
	)
}

func GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error) {
	var finalResult models.GetWeatherPointResponse // The final struct to return

	apiURL := fmt.Sprintf("%s/getweatherpoint/%s", baseURL, keyword)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return finalResult, fmt.Errorf("failed to create request for /getweatherpoint: %w", err)
	}

	body, err := doRequest(req)
	if err != nil {
		// Check if it's an APIError from doRequest
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// If it's a known API error (like 404 for empty keyword), return it directly
			return finalResult, err
		}
		// Otherwise, wrap the error
		return finalResult, fmt.Errorf("request failed for /getweatherpoint: %w", err)
	}

	// Define a temporary struct to capture the initial response where "result" is a string
	type tempResp struct {
		Result string `json:"result"`
	}
	var tempResult tempResp

	// Unmarshal the raw body into the temporary struct
	if err := json.Unmarshal(body, &tempResult); err != nil {
		return finalResult, fmt.Errorf("failed to unmarshal initial response for /getweatherpoint: %w, body: %s", err, string(body))
	}

	// Now, unmarshal the "result" string into the actual WeatherPoints struct
	var weatherPoints models.WeatherPoints
	// The API returns a JSON array string `[{...}, {...}]` which needs to be unmarshalled into `weatherPoints.Root`
	var points []models.WeatherPoint
	if errUnmarshalArray := json.Unmarshal([]byte(tempResult.Result), &points); errUnmarshalArray != nil {
		// If unmarshalling into an array fails, return the error
		// This handles cases like `{"result":"[]"}` correctly as well, resulting in an empty points slice.
		return finalResult, fmt.Errorf("failed to unmarshal nested result string for /getweatherpoint: %w, result string: %s", errUnmarshalArray, tempResult.Result)
	}
	// Assign the unmarshalled points slice to the Root field
	weatherPoints.Root = points

	// Assign the successfully unmarshalled WeatherPoints to the final result
	finalResult.Result = weatherPoints

	// Optional validation can be added here if needed
	// for _, point := range finalResult.Result.Root {
	// 	if err := point.Validate(); err != nil {
	// 		fmt.Printf("Warning: validation failed for weather point %s: %v\n", point.CityCode, err)
	// 	}
	// }

	return finalResult, nil
}


func GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) {
	return _get[models.GetWeatherStatusResponse](
		"/getweatherstatus",
		cityCode,
		nil,
	)
}

func GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error) {
	params := url.Values{}
	params.Set("csid", "mmcm")
	params.Set("contents_id", "day_tenki--day_pre--hight_temp--low_temp--day_wind_v--day_wind_d--zutu_level_day--low_humidity")
	params.Set("duration_yohoushi", "7")
	params.Set("where", fmt.Sprintf("CHITEN_%s", cityCode))
	params.Set("json", "on")

	apiURL := fmt.Sprintf("%s/getElements", otenkiASPBaseURL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("failed to parse otenki asp url: %w", err)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("failed to create otenki asp request: %w", err)
	}

	body, err := doRequest(req)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("otenki asp request failed: %w", err)
	}

	var rawResponse models.GetOtenkiASPRawResponse
	if err := json.Unmarshal(body, &rawResponse); err != nil {
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
		// Check if the Record slice within RawRecord is long enough (header + data)
		if len(rawElem.Record) < 2 {
			fmt.Printf("Skipping raw element with insufficient records (need header + data): %+v\n", rawElem.Record)
			continue
		}

		// Process the header record (first RawProperty in the Record slice)
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

		// Initialize the structured Element
		elem := models.Element{
			ContentID: contentID,
			Title:     title,
			Records:   make(map[time.Time]interface{}),
		}

		// Process the data records (remaining RawProperty items in the Record slice)
		for _, rawDataProperty := range rawElem.Record[1:] {
			// Check if the Property slice within RawProperty is long enough (time + value)
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

			// Parse the timestamp
			t, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				// Try parsing without timezone offset if RFC3339 fails (common in some APIs)
				t, err = time.Parse("2006-01-02T15:04:05", timeStr)
				if err != nil {
					fmt.Printf("Skipping data record with unparseable time '%s': %v\n", timeStr, err)
					continue
				}
			}
			elem.Records[t] = value
		}
		// Add the processed element to the response
		response.Elements = append(response.Elements, elem)
	}

	return response, nil
}
