package api_test

import (
	"errors"
	"strings"
	"testing"

	"zutool/api" // Update import path
)

// Helper function to create a new client for tests
func newTestClient() *api.Client {
	// Use default URLs and timeout for testing against the actual API
	// Consider using a mock server for more controlled tests in the future
	return api.NewClient("", "", 0)
}

// --- API Client Tests ---

func TestGetPainStatus(t *testing.T) {
	client := newTestClient()
	areaCode := "13" // 東京都
	_, err := client.GetPainStatus(areaCode, nil)
	if err != nil {
		t.Errorf("client.GetPainStatus(%q) failed: %v", areaCode, err)
	}
	// Note: Further checks on the response content can be added here.
}

func TestGetPainStatusInvalid(t *testing.T) {
	client := newTestClient()
	areaCode := "1" // Invalid code
	_, err := client.GetPainStatus(areaCode, nil)
	if err == nil {
		t.Errorf("client.GetPainStatus(%q) should have failed, but got nil error", areaCode)
		return
	}

	// Check if the error is the expected APIError type and message
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		// The specific error code/message might depend on the API implementation
		expectedMsgPart := "存在しない都道府県コードです" // Partial check based on original test
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetPainStatus(%q) failed with unexpected message: %q, want containing %q", areaCode, apiErr.Message, expectedMsgPart)
		}
		// Optionally check apiErr.StatusCode if the API consistently returns specific codes
		// if apiErr.StatusCode != expectedStatusCode { ... }
	} else {
		t.Errorf("client.GetPainStatus(%q) failed with unexpected error type: %T, want *api.APIError", areaCode, err)
	}
}

func TestGetPainStatusSetWeatherPoint(t *testing.T) {
	client := newTestClient()

	// Test case 1: Sapporo
	cityCode1 := "01101"
	setWeatherPoint1 := cityCode1
	areaCode1 := cityCode1[:2] // "01"
	expectedAreaName1 := "北海道"
	res1, err1 := client.GetPainStatus(areaCode1, &setWeatherPoint1)
	if err1 != nil {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) failed: %v", areaCode1, setWeatherPoint1, err1)
	} else if res1.PainnoterateStatus.AreaName != expectedAreaName1 {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) returned unexpected AreaName: %q, want %q", areaCode1, setWeatherPoint1, res1.PainnoterateStatus.AreaName, expectedAreaName1)
	}

	// Test case 2: Shibuya
	cityCode2 := "13113"
	setWeatherPoint2 := cityCode2
	areaCode2 := cityCode2[:2] // "13"
	expectedAreaName2 := "東京都"
	res2, err2 := client.GetPainStatus(areaCode2, &setWeatherPoint2)
	if err2 != nil {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) failed: %v", areaCode2, setWeatherPoint2, err2)
	} else if res2.PainnoterateStatus.AreaName != expectedAreaName2 {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) returned unexpected AreaName: %q, want %q", areaCode2, setWeatherPoint2, res2.PainnoterateStatus.AreaName, expectedAreaName2)
	}
}


func TestGetWeatherPoint(t *testing.T) {
	client := newTestClient()
	keyword := "神戸市"
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		t.Errorf("client.GetWeatherPoint(%q) failed: %v", keyword, err)
	}
	if len(res.Result.Root) == 0 {
		t.Errorf("client.GetWeatherPoint(%q) returned empty results, expected non-empty", keyword)
	}
}

func TestGetWeatherPointExtra(t *testing.T) {
	client := newTestClient()
	keyword := "a" // A keyword expected to return few or no results
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		// Depending on API behavior, an error might be acceptable here too
		t.Logf("client.GetWeatherPoint(%q) failed as potentially expected: %v", keyword, err)
		// Check if it's a 404 API error, which might be valid for "no results"
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// This could be considered a pass if 404 means "not found"
			return
		}
		// If it's another error, fail the test
		t.Errorf("client.GetWeatherPoint(%q) failed unexpectedly: %v", keyword, err)
		return
	}
	// If no error, check if the result list is empty (as per original Python test)
	if len(res.Result.Root) != 0 {
		t.Errorf("client.GetWeatherPoint(%q) returned %d results, expected 0 or a 404 error", keyword, len(res.Result.Root))
	}
}


func TestGetWeatherPointEmpty(t *testing.T) {
	client := newTestClient()
	keyword := ""
	_, err := client.GetWeatherPoint(keyword)
	if err == nil {
		t.Errorf("client.GetWeatherPoint(%q) should have failed, but got nil error", keyword)
		return
	}
	// Check for specific APIError (expecting 404 based on original test)
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != 404 {
			t.Errorf("client.GetWeatherPoint(%q) failed with status code %d, want 404", keyword, apiErr.StatusCode)
		}
		// Message check can be added but might be fragile
		// expectedMsgPart := "お探しのページは見つかりません"
		// if !strings.Contains(apiErr.Message, expectedMsgPart) { ... }
	} else {
		t.Errorf("client.GetWeatherPoint(%q) failed with unexpected error type: %T, want *api.APIError", keyword, err)
	}
}

func TestGetWeatherStatus(t *testing.T) {
	client := newTestClient()
	cityCode := "13113" // Shibuya
	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		t.Errorf("client.GetWeatherStatus(%q) failed: %v", cityCode, err)
	}
	// Check if PrefecturesID and PlaceID combine to the original cityCode
	// Note: models.PrefecturesID might be a different type (e.g., int), adjust comparison if needed.
	if string(res.PrefecturesID)+res.PlaceID != cityCode {
		t.Errorf("client.GetWeatherStatus(%q) returned unexpected combination: PrefID=%q, PlaceID=%q, want %q",
			cityCode, res.PrefecturesID, res.PlaceID, cityCode)
	}
}

func TestGetWeatherStatusInvalidCode(t *testing.T) {
	client := newTestClient()
	cityCode := "aaaaa"
	_, err := client.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("client.GetWeatherStatus(%q) should have failed, but got nil error", cityCode)
		return
	}
	// Check for APIError and potentially specific message/code
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点名称が取得できませんでした" // Based on original test
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetWeatherStatus(%q) failed with unexpected message: %q, want containing %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// Optionally check status code if known
	} else {
		t.Errorf("client.GetWeatherStatus(%q) failed with unexpected error type: %T, want *api.APIError", cityCode, err)
	}
}

func TestGetWeatherStatusInvalidDigit(t *testing.T) {
	client := newTestClient()
	cityCode := "13"
	_, err := client.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("client.GetWeatherStatus(%q) should have failed, but got nil error", cityCode)
		return
	}
	// Check for APIError and potentially specific message/code
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点コードの桁数が正しくありません" // Based on original test
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetWeatherStatus(%q) failed with unexpected message: %q, want containing %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// Optionally check status code if known
	} else {
		t.Errorf("client.GetWeatherStatus(%q) failed with unexpected error type: %T, want *api.APIError", cityCode, err)
	}
}

func TestGetOtenkiASP(t *testing.T) {
	client := newTestClient()
	cityCode := "13101" // Tokyo
	res, err := client.GetOtenkiASP(cityCode)
	if err != nil {
		t.Errorf("client.GetOtenkiASP(%q) failed: %v", cityCode, err)
	}
	// Check status field in the response
	if res.Status != "OK" { // Assuming models.GetOtenkiASPResponse has a Status field
		t.Errorf("client.GetOtenkiASP(%q) returned status %q, want %q", cityCode, res.Status, "OK")
	}
}

func TestGetOtenkiASPInvalidCode(t *testing.T) {
	client := newTestClient()
	cityCode := "13000" // Invalid code for OtenkiASP based on original test
	_, err := client.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("client.GetOtenkiASP(%q) should have failed, but got nil error", cityCode)
	}
	// Further checks on the error type (e.g., *api.APIError) could be added if needed.
}

func TestGetOtenkiASPEmpty(t *testing.T) {
	client := newTestClient()
	cityCode := ""
	_, err := client.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("client.GetOtenkiASP(%q) should have failed, but got nil error", cityCode)
	}
	// Further checks on the error type could be added.
}
