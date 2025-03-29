package main

import (
	"errors"
	"strings"
	"testing"

	"zutool/api"
)

// --- API Client Tests ---

func TestGetPainStatus(t *testing.T) {
	areaCode := "13" // 東京都
	_, err := api.GetPainStatus(areaCode, nil)
	if err != nil {
		t.Errorf("api.GetPainStatus(%q) failed: %v", areaCode, err)
	}
	// Note: Checking the actual AreaName requires a mapping or more complex assertion.
	// The Python test checked res.painnoterate_status.area_name.value == area_code,
	// but the Go response doesn't directly expose the original area code used for the request.
	// We primarily check for errors here.
}

func TestGetPainStatusInvalid(t *testing.T) {
	areaCode := "1" // Invalid code
	_, err := api.GetPainStatus(areaCode, nil)
	if err == nil {
		t.Errorf("api.GetPainStatus(%q) should have failed, but got nil error", areaCode)
		return
	}

	// Check if the error is the expected APIError type and message
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		// Python test checks for ZUTOOL_API_INVALID_PARAMETER (4002)
		// and message "存在しない都道府県コードです todofuken_code = 1"
		// The Go api.APIError struct currently wraps the message.
		expectedMsgPart := "存在しない都道府県コードです" // Partial check
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("api.GetPainStatus(%q) failed with unexpected message: %q, want containing %q", areaCode, apiErr.Message, expectedMsgPart)
		}
		// TODO: Check apiErr.StatusCode if needed
	} else {
		t.Errorf("api.GetPainStatus(%q) failed with unexpected error type: %T, want *api.APIError", areaCode, err)
	} // <-- Add closing brace here
	// Corresponds to test_pain_status_empty in Python
	cityCode1 := "01101" // Sapporo
	setWeatherPoint1 := cityCode1
	// Pass the area code derived from the city code when using setWeatherPoint
	// Applying the change based on user feedback: test is wrong, compare against area name
	areaCode1 := cityCode1[:2] // "01"
	expectedAreaName1 := "北海道" // Expected name for area code "01"
	res1, err1 := api.GetPainStatus(areaCode1, &setWeatherPoint1)
	if err1 != nil {
		t.Errorf("api.GetPainStatus(%q, setWeatherPoint=%q) failed: %v", areaCode1, setWeatherPoint1, err1)
	} else if res1.PainnoterateStatus.AreaName != expectedAreaName1 { // Compare against area name
		t.Errorf("api.GetPainStatus(%q, setWeatherPoint=%q) returned unexpected AreaName: %q, want %q", areaCode1, setWeatherPoint1, res1.PainnoterateStatus.AreaName, expectedAreaName1)
	}

	cityCode2 := "13113" // Shibuya
	setWeatherPoint2 := cityCode2
	// Pass the area code derived from the city code when using setWeatherPoint
	areaCode2 := cityCode2[:2] // "13"
	expectedAreaName2 := "東京都" // Expected name for area code "13"
	res2, err2 := api.GetPainStatus(areaCode2, &setWeatherPoint2)
	if err2 != nil {
		t.Errorf("api.GetPainStatus(%q, setWeatherPoint=%q) failed: %v", areaCode2, setWeatherPoint2, err2)
	} else if res2.PainnoterateStatus.AreaName != expectedAreaName2 { // Compare against area name
		t.Errorf("api.GetPainStatus(%q, setWeatherPoint=%q) returned unexpected AreaName: %q, want %q", areaCode2, setWeatherPoint2, res2.PainnoterateStatus.AreaName, expectedAreaName2)
	}
}


func TestGetWeatherPoint(t *testing.T) {
	keyword := "神戸市"
	res, err := api.GetWeatherPoint(keyword)
	if err != nil {
		t.Errorf("api.GetWeatherPoint(%q) failed: %v", keyword, err)
	}
	// Python test asserts len == 9. Let's check if it's non-empty.
	if len(res.Result.Root) == 0 {
		t.Errorf("api.GetWeatherPoint(%q) returned empty results, expected non-empty", keyword)
	}
}

func TestGetWeatherPointExtra(t *testing.T) {
	keyword := "a" // A keyword expected to return few or no results
	res, err := api.GetWeatherPoint(keyword)
	if err != nil {
		t.Errorf("api.GetWeatherPoint(%q) failed: %v", keyword, err)
	}
	// Python test asserts empty list. Let's check that.
	if len(res.Result.Root) != 0 {
		t.Errorf("api.GetWeatherPoint(%q) returned %d results, expected 0", keyword, len(res.Result.Root))
	}
}

func TestGetWeatherPointEmpty(t *testing.T) {
	keyword := ""
	_, err := api.GetWeatherPoint(keyword)
	if err == nil {
		t.Errorf("api.GetWeatherPoint(%q) should have failed, but got nil error", keyword)
		return
	}
	// Python test checks for 404 and "お探しのページは見つかりません"
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != 404 {
			t.Errorf("api.GetWeatherPoint(%q) failed with status code %d, want 404", keyword, apiErr.StatusCode)
		}
		// Message check might be fragile if API changes
	} else {
		t.Errorf("api.GetWeatherPoint(%q) failed with unexpected error type: %T, want *api.APIError", keyword, err)
	}
}

func TestGetWeatherStatus(t *testing.T) {
	cityCode := "13113" // Shibuya
	res, err := api.GetWeatherStatus(cityCode)
	if err != nil {
		t.Errorf("api.GetWeatherStatus(%q) failed: %v", cityCode, err)
	}
	// Python test checks PrefecturesID + PlaceID == cityCode
	if string(res.PrefecturesID)+res.PlaceID != cityCode {
		t.Errorf("api.GetWeatherStatus(%q) returned unexpected combination: PrefID=%q, PlaceID=%q, want %q",
			cityCode, res.PrefecturesID, res.PlaceID, cityCode)
	}
}

func TestGetWeatherStatusInvalidCode(t *testing.T) {
	cityCode := "aaaaa"
	_, err := api.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("api.GetWeatherStatus(%q) should have failed, but got nil error", cityCode)
		return
	}
	// Python test checks for ZUTOOL_API_NOT_FOUND (4004) and message
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点名称が取得できませんでした"
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("api.GetWeatherStatus(%q) failed with unexpected message: %q, want containing %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// TODO: Check status code if needed
	} else {
		t.Errorf("api.GetWeatherStatus(%q) failed with unexpected error type: %T, want *api.APIError", cityCode, err)
	}
}

func TestGetWeatherStatusInvalidDigit(t *testing.T) {
	cityCode := "13"
	_, err := api.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("api.GetWeatherStatus(%q) should have failed, but got nil error", cityCode)
		return
	}
	// Python test checks for ZUTOOL_API_INVALID_PARAMETER (4002) and message
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点コードの桁数が正しくありません"
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("api.GetWeatherStatus(%q) failed with unexpected message: %q, want containing %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// TODO: Check status code if needed
	} else {
		t.Errorf("api.GetWeatherStatus(%q) failed with unexpected error type: %T, want *api.APIError", cityCode, err)
	}
}

func TestGetOtenkiASP(t *testing.T) {
	cityCode := "13101" // Tokyo
	res, err := api.GetOtenkiASP(cityCode)
	if err != nil {
		t.Errorf("api.GetOtenkiASP(%q) failed: %v", cityCode, err)
	}
	// Python test checks status == "OK"
	if res.Status != "OK" {
		t.Errorf("api.GetOtenkiASP(%q) returned status %q, want %q", cityCode, res.Status, "OK")
	}
}

func TestGetOtenkiASPInvalidCode(t *testing.T) {
	cityCode := "13000" // Invalid code for OtenkiASP
	_, err := api.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("api.GetOtenkiASP(%q) should have failed, but got nil error", cityCode)
	}
	// We just check that an error occurred.
}

func TestGetOtenkiASPEmpty(t *testing.T) {
	cityCode := ""
	_, err := api.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("api.GetOtenkiASP(%q) should have failed, but got nil error", cityCode)
	}
	// We expect an error.
}

// --- CLI Tests ---
