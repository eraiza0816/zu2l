package commands

import (
	"errors"
	"fmt" // Import fmt for Sprintf
	"testing"
	// "time" // time is not used

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/eraiza0816/zu2l/internal/models"
)

// runWeatherStatusLogic は、weather_status.go にあるべきコアロジック関数（仮定）
// このシグネチャは weather_status.go の実装に合わせる必要があります。
// ddd_doc.md によると、Presenter は dayOffset と dayName も引数に取ります。
// コマンドライン引数から dayOffset (today, tomorrow など) を解釈し、
// それを数値の offset と表示用の dayName に変換するロジックがどこかにあるはずです。
// ここでは、runWeatherStatusLogic が cityCode と、解決済みの dayOffset, dayName を受け取ると仮定します。
// runWeatherStatusLogic は weather_status.go で定義されているため、ここでは削除します。

func TestRunWeatherStatusLogic_Success_Today(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	cityCode := "130010" // Tokyo
	dayOffset := 0       // Today
	dayName := "today"   // Changed to match weather_status.go
	expectedResponse := models.GetWeatherStatusResponse{
		PlaceName: "東京",
		PlaceID:   "131", // Example PlaceID, ddd_doc says 3 digits
		Today: []models.WeatherStatusByTime{
			{Time: "0", Weather: models.Sunny, Temp: models.NewString("15.0"), Pressure: "1010", PressureLevel: models.Normal},
		},
	}

	mockClient.On("GetWeatherStatus", cityCode).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherStatus", expectedResponse, dayOffset, dayName).Return(nil)

	err := runWeatherStatusLogic(mockClient, mockPresenter, cityCode, dayOffset, dayName)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunWeatherStatusLogic_Success_Tomorrow(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	cityCode := "130010"
	dayOffset := 1 // Tomorrow
	dayName := "tomorrow" // Changed to match weather_status.go
	expectedResponse := models.GetWeatherStatusResponse{
		PlaceName: "東京",
		Tomorrow: []models.WeatherStatusByTime{
			{Time: "0", Weather: models.Cloudy, Temp: models.NewString("18.0"), Pressure: "1008", PressureLevel: models.SlightAlert},
		},
	}

	mockClient.On("GetWeatherStatus", cityCode).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherStatus", expectedResponse, dayOffset, dayName).Return(nil)

	err := runWeatherStatusLogic(mockClient, mockPresenter, cityCode, dayOffset, dayName)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunWeatherStatusLogic_ClientError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	cityCode := "999999" // Invalid city code
	dayOffset := 0
	dayName := "today"   // Changed to match weather_status.go
	clientError := errors.New("API client failed")

	mockClient.On("GetWeatherStatus", cityCode).Return(models.GetWeatherStatusResponse{}, clientError)

	err := runWeatherStatusLogic(mockClient, mockPresenter, cityCode, dayOffset, dayName)
	assert.Error(t, err)
	expectedErrorMessage := fmt.Sprintf("気象状況の取得に失敗しました (%s): %s", cityCode, clientError.Error())
	assert.EqualError(t, err, expectedErrorMessage)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertNotCalled(t, "PresentWeatherStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestRunWeatherStatusLogic_PresenterError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	cityCode := "130010"
	dayOffset := 0
	dayName := "today"   // Changed to match weather_status.go
	expectedResponse := models.GetWeatherStatusResponse{PlaceName: "東京"}
	presenterError := errors.New("presenter display failed")

	mockClient.On("GetWeatherStatus", cityCode).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherStatus", expectedResponse, dayOffset, dayName).Return(presenterError)

	err := runWeatherStatusLogic(mockClient, mockPresenter, cityCode, dayOffset, dayName)
	assert.Error(t, err)
	expectedErrorMessage := fmt.Sprintf("%s の結果表示に失敗しました: %s", dayName, presenterError.Error())
	assert.EqualError(t, err, expectedErrorMessage)


	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

// Helper function to create a pointer to a string, useful for nullable fields like Temp.
// This should ideally be in a shared utility or models package if used frequently.
// For now, defining it here for models.NewString usage.
// func NewString(s string) *string { return &s }
// models.NewString が internal/models にあると仮定。なければ上記を有効化。
