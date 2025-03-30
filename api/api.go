package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"zutool/internal/models" // Keep internal import for models
)

const (
	defaultBaseURL         = "https://zutool.jp/api"
	defaultOtenkiASPBaseURL = "https://ap.otenki.com/OtenkiASP/asp"
	defaultTimeout         = 10 * time.Second
	defaultUserAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
)

// Client manages communication with the zutool and Otenki ASP APIs.
type Client struct {
	baseURL     string
	otenkiBaseURL string
	httpClient  *http.Client
	userAgent   string
}

// NewClient creates a new API client.
// If baseURL or otenkiBaseURL are empty strings, default values are used.
// If timeout is zero, a default timeout is used.
func NewClient(baseURL, otenkiBaseURL string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if otenkiBaseURL == "" {
		otenkiBaseURL = defaultOtenkiASPBaseURL
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &Client{
		baseURL:     baseURL,
		otenkiBaseURL: otenkiBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: defaultUserAgent, // Use default UA for now
	}
}

// doRequest is a helper method to execute HTTP requests and handle common logic.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", c.userAgent) // Use userAgent from Client struct
	resp, err := c.httpClient.Do(req) // Use httpClient from Client struct
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

// _get is a private helper method to fetch raw response body from the primary API (baseURL).
// It also checks for embedded API errors within a 200 OK response.
func (c *Client) _get(path string, param string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s%s/%s", c.baseURL, path, param)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
	}

	body, err := c.doRequest(req) // Call the doRequest method
	if err != nil {
		// If doRequest returned an error (e.g., non-200 status), return it directly.
		return nil, fmt.Errorf("request failed for %s: %w", path, err)
	}

	// --- Start: Check for API error within 200 OK response ---
	// (This logic remains the same for now, but might be refactored later)
	var errorResponse models.ErrorResponse
	// Try unmarshalling into the error structure first.
	// We ignore the error here because failing to unmarshal into ErrorResponse is expected for valid data.
	_ = json.Unmarshal(body, &errorResponse)

	// If unmarshalling into ErrorResponse succeeded AND we got an error message,
	// it means the API returned 200 OK but signaled an error internally.
	if errorResponse.ErrorMessage != "" {
		// Return the body along with the specific APIError.
		return body, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
	}
	// --- End: Check for API error within 200 OK response ---

	// If no embedded error, return the raw body for the caller to unmarshal.
	return body, nil
}

// GetPainStatus fetches pain status information.
func (c *Client) GetPainStatus( // Added receiver (c *Client)
	areaCode string,
	setWeatherPoint *string,
) (models.GetPainStatusResponse, error) {
	var result models.GetPainStatusResponse

	// --- Start: setWeatherPoint logic ---
	if setWeatherPoint != nil && *setWeatherPoint != "" {
		setWeatherPointURL := fmt.Sprintf("%s/setweatherpoint/%s", c.baseURL, *setWeatherPoint)
		req, err := http.NewRequest("GET", setWeatherPointURL, nil)
		if err != nil {
			return result, fmt.Errorf("failed to create setweatherpoint request: %w", err)
		}

		setBody, err := c.doRequest(req) // Use client's doRequest
		if err != nil {
			// Check if it's an APIError from doRequest
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				// Treat 404 for setweatherpoint as a specific error
				return result, fmt.Errorf("地点コード '%s' が見つかりません (setweatherpoint failed with 404)", *setWeatherPoint)
			}
			// Otherwise, wrap the generic error
			return result, fmt.Errorf("setweatherpoint request failed: %w", err)
		}


		var setResp models.SetWeatherPointResponse
		if err := json.Unmarshal(setBody, &setResp); err != nil {
			return result, fmt.Errorf("failed to unmarshal setweatherpoint response: %w, body: %s", err, string(setBody))
		}
		if setResp.Response != "ok" {
			// Consider returning a more specific error if the response is not "ok"
			return result, fmt.Errorf("setweatherpoint response was not 'ok': %s", setResp.Response)
		}
	}
	// --- End: Moved setWeatherPoint logic here ---

	// --- End: setWeatherPoint logic ---

	// Now make the main request using the internal _get method
	body, err := c._get("/getpainstatus", areaCode)
	if err != nil {
		// Check if the error returned by _get was the embedded APIError (status 200)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			// It was the embedded error, return it directly
			return result, err
		}
		// Otherwise, wrap the error from _get
		return result, fmt.Errorf("failed to get pain status body: %w", err)
	}

	// Unmarshal the valid response body into the result struct
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal GetPainStatus response: %w, body: %s", err, string(body))
	}

	return result, nil
}


// GetWeatherPoint fetches weather point information.
func (c *Client) GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error) {
	// Use the internal _get method to fetch the raw body
	body, err := c._get("/getweatherpoint", keyword)
	if err != nil {
		// Check if it's an APIError from doRequest (via _get)
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// If it's a known API error (like 404 for empty keyword), return it directly
			// Return zero value of the response struct along with the error
			return models.GetWeatherPointResponse{}, err
		}
		// Otherwise, wrap the error
		return models.GetWeatherPointResponse{}, fmt.Errorf("request failed for /getweatherpoint: %w", err)
	}

	// Call the dedicated parser function
	return parseWeatherPointResponse(body)
}


// GetWeatherStatus fetches detailed weather status.
func (c *Client) GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) {
	var result models.GetWeatherStatusResponse

	body, err := c._get("/getweatherstatus", cityCode)
	if err != nil {
		// Check if the error returned by _get was the embedded APIError (status 200)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			// It was the embedded error, return it directly
			return result, err
		}
		// Otherwise, wrap the error from _get
		return result, fmt.Errorf("failed to get weather status body: %w", err)
	}

	// Unmarshal the valid response body into the result struct
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal GetWeatherStatus response: %w, body: %s", err, string(body))
	}

	return result, nil
}

// GetOtenkiASP fetches weather information from Otenki ASP.
func (c *Client) GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error) {
	params := url.Values{}
	params.Set("csid", "mmcm") // Keep as is for now
	params.Set("contents_id", "day_tenki--day_pre--hight_temp--low_temp--day_wind_v--day_wind_d--zutu_level_day--low_humidity") // Keep as is
	params.Set("duration_yohoushi", "7")
	params.Set("where", fmt.Sprintf("CHITEN_%s", cityCode))
	params.Set("json", "on") // Keep as is

	apiURL := fmt.Sprintf("%s/getElements", c.otenkiBaseURL) // Use otenkiBaseURL from Client struct
	u, err := url.Parse(apiURL)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("failed to parse otenki asp url: %w", err)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("failed to create otenki asp request: %w", err)
	}

	body, err := c.doRequest(req) // Use client's doRequest
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("otenki asp request failed: %w", err)
	}

	// Call the dedicated parser function
	return parseOtenkiASPResponse(body)
}
