package commands

import (
	"bytes"
	"errors"
	// "flag" // No longer needed for urfave/cli context
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// "github.com/urfave/cli/v2" // No longer needed for urfave/cli context

	"github.com/eraiza0816/zu2l/internal/models"
	// "zutool/internal/presenter" // We use commands.PresenterInterface. This might also need updating if presenter is used directly.
)

// MockClient is a mock type for the commands.ClientInterface type
type MockClient struct {
	mock.Mock
}

// GetPainStatus is a mock method
func (m *MockClient) GetPainStatus(areaCode string, setWeatherPoint *string) (models.GetPainStatusResponse, error) {
	args := m.Called(areaCode, setWeatherPoint)
	return args.Get(0).(models.GetPainStatusResponse), args.Error(1)
}

// GetWeatherPoint is a mock method (added for weather_point)
func (m *MockClient) GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error) {
	args := m.Called(keyword)
	return args.Get(0).(models.GetWeatherPointResponse), args.Error(1)
}

// GetWeatherStatus is a mock method (added for weather_status)
func (m *MockClient) GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) {
	args := m.Called(cityCode)
	return args.Get(0).(models.GetWeatherStatusResponse), args.Error(1)
}

// Ensure MockClient implements commands.ClientInterface
var _ ClientInterface = (*MockClient)(nil)

// MockPresenter is a mock type for the commands.PresenterInterface type
type MockPresenter struct {
	mock.Mock
	Output *bytes.Buffer // Kept for potential future use, though not directly used by PresentPainStatus mock
}

func NewMockPresenter() *MockPresenter {
	return &MockPresenter{Output: new(bytes.Buffer)}
}

// PresentPainStatus is a mock method
func (m *MockPresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	args := m.Called(data)
	return args.Error(0)
}

// PresentWeatherPoint is a mock method (added for weather_point)
func (m *MockPresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	args := m.Called(data, kata, keyword)
	return args.Error(0)
}

// PresentWeatherStatus is a mock method (added for weather_status)
func (m *MockPresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	args := m.Called(data, dayOffset, dayName)
	return args.Error(0)
}

// Ensure MockPresenter implements commands.PresenterInterface
var _ PresenterInterface = (*MockPresenter)(nil)

// Note: runPainStatusLogic is now defined in pain_status.go, so we are testing that directly.
// The ClientInterface and PresenterInterface are also defined in pain_status.go.

func TestRunPainStatusLogic_Success(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := NewMockPresenter()

	areaCode := "130010" // Example area code for Tokyo
	var weatherPoint *string = nil
	expectedResponse := models.GetPainStatusResponse{
		PainnoterateStatus: models.GetPainStatus{
			AreaName:    "東京",
			TimeStart:   "00", // Example data
			TimeEnd:     "03", // Example data
			RateNormal:  0.7,  // Example data
			RateLittle:  0.2,  // Example data
			RatePainful: 0.05, // Example data
			RateBad:     0.05, // Example data
		},
	}

	// Setup expectations
	mockClient.On("GetPainStatus", areaCode, weatherPoint).Return(expectedResponse, nil)
	mockPresenter.On("PresentPainStatus", expectedResponse).Return(nil)

	err := runPainStatusLogic(mockClient, mockPresenter, areaCode, weatherPoint)
	assert.NoError(t, err)

	// Verify that the expected methods were called
	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunPainStatusLogic_ClientError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := NewMockPresenter() // Presenter won't be called if client fails

	areaCode := "130010"
	var weatherPoint *string = nil
	clientError := errors.New("API error")

	mockClient.On("GetPainStatus", areaCode, weatherPoint).Return(models.GetPainStatusResponse{}, clientError)
	// Presenter.PresentPainStatus should not be called

	err := runPainStatusLogic(mockClient, mockPresenter, areaCode, weatherPoint)

	assert.Error(t, err)
	assert.EqualError(t, err, "痛み予報の取得に失敗しました: API error") // Error is wrapped

	mockClient.AssertExpectations(t)
	mockPresenter.AssertNotCalled(t, "PresentPainStatus", mock.Anything)
}

func TestRunPainStatusLogic_PresenterError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := NewMockPresenter()

	areaCode := "130010"
	var weatherPoint *string = nil
	expectedResponse := models.GetPainStatusResponse{
		PainnoterateStatus: models.GetPainStatus{
			AreaName: "東京", // Minimal valid data
		},
	}
	presenterError := errors.New("presenter failed")

	mockClient.On("GetPainStatus", areaCode, weatherPoint).Return(expectedResponse, nil)
	mockPresenter.On("PresentPainStatus", expectedResponse).Return(presenterError)

	err := runPainStatusLogic(mockClient, mockPresenter, areaCode, weatherPoint)

	assert.Error(t, err)
	assert.EqualError(t, err, "結果の表示に失敗しました: presenter failed") // Error is wrapped

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunPainStatusLogic_WithWeatherPoint(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := NewMockPresenter()

	areaCode := "270000" // Example Osaka
	weatherPointValue := "osaka_city_hall"
	weatherPoint := &weatherPointValue
	expectedResponse := models.GetPainStatusResponse{
		PainnoterateStatus: models.GetPainStatus{
			AreaName: "大阪", // GetPainStatus has AreaName
			// Add other fields if necessary for this test case
		},
	}

	// Setup expectations for GetPainStatus with weatherPoint
	mockClient.On("GetPainStatus", areaCode, weatherPoint).Return(expectedResponse, nil)
	mockPresenter.On("PresentPainStatus", expectedResponse).Return(nil)

	err := runPainStatusLogic(mockClient, mockPresenter, areaCode, weatherPoint)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}
