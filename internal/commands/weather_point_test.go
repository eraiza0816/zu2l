package commands

import (
	"errors" // bytes は削除
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/eraiza0816/zu2l/internal/models"
)

// MockClientForWeatherPoint は weather_point コマンド用の ClientInterface のモックです。
// GetWeatherPoint メソッドのみをモックします。
// pain_status_test.go の MockClient とは別に定義するか、共通の MockClient を拡張します。
// ここでは、GetWeatherPoint のみを必要とするため、新しいモックまたは適合するインターフェースを想定します。
// 簡単のため、既存の MockClient が GetWeatherPoint も持つと仮定して進めますが、
// 本来はインターフェースを適切に分離・定義すべきです。
// (pain_status_test.go の MockClient は GetPainStatus のみでした)
// ここでは、テスト対象の runWeatherPointLogic が ClientInterface を期待し、
// その ClientInterface が GetWeatherPoint を持つと仮定します。

// MockPresenterForWeatherPoint は weather_point コマンド用の PresenterInterface のモックです。
// PresentWeatherPoint メソッドのみをモックします。
// (pain_status_test.go の MockPresenter は PresentPainStatus のみでした)
// ここでは、テスト対象の runWeatherPointLogic が PresenterInterface を期待し、
// その PresenterInterface が PresentWeatherPoint を持つと仮定します。

// runWeatherPointLogic は weather_point.go で定義されているため、ここでは削除します。

func TestRunWeatherPointLogic_Success(t *testing.T) {
	mockClient := new(MockClient) // pain_status_test.go の MockClient を再利用 (GetWeatherPoint を追加する必要あり)
	mockPresenter := new(MockPresenter) // pain_status_test.go の MockPresenter を再利用 (PresentWeatherPoint を追加する必要あり)

	keyword := "東京"
	kata := false
	expectedResponse := models.GetWeatherPointResponse{
		Result: models.WeatherPoints{
			Root: []models.WeatherPoint{
				{CityCode: "130010", NameKata: "トウキョウ", Name: "東京"},
			},
		},
	}

	mockClient.On("GetWeatherPoint", keyword).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherPoint", expectedResponse, kata, keyword).Return(nil)

	err := runWeatherPointLogic(mockClient, mockPresenter, keyword, kata)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunWeatherPointLogic_Success_Kata(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	keyword := "トウキョウ"
	kata := true // kata フラグが true
	expectedResponse := models.GetWeatherPointResponse{
		Result: models.WeatherPoints{
			Root: []models.WeatherPoint{
				{CityCode: "130010", NameKata: "トウキョウ", Name: "東京"},
			},
		},
	}

	mockClient.On("GetWeatherPoint", keyword).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherPoint", expectedResponse, kata, keyword).Return(nil)

	err := runWeatherPointLogic(mockClient, mockPresenter, keyword, kata)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunWeatherPointLogic_ClientError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	keyword := "存在しない場所"
	kata := false
	clientError := errors.New("API error from client")

	mockClient.On("GetWeatherPoint", keyword).Return(models.GetWeatherPointResponse{}, clientError)

	err := runWeatherPointLogic(mockClient, mockPresenter, keyword, kata)
	assert.Error(t, err)
	assert.EqualError(t, err, "地域地点の検索に失敗しました: API error from client")

	mockClient.AssertExpectations(t)
	mockPresenter.AssertNotCalled(t, "PresentWeatherPoint", mock.Anything, mock.Anything, mock.Anything)
}

func TestRunWeatherPointLogic_PresenterError(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	keyword := "東京"
	kata := false
	expectedResponse := models.GetWeatherPointResponse{
		Result: models.WeatherPoints{
			Root: []models.WeatherPoint{{Name: "東京"}},
		},
	}
	presenterError := errors.New("presenter failed to present")

	mockClient.On("GetWeatherPoint", keyword).Return(expectedResponse, nil)
	mockPresenter.On("PresentWeatherPoint", expectedResponse, kata, keyword).Return(presenterError)

	err := runWeatherPointLogic(mockClient, mockPresenter, keyword, kata)
	assert.Error(t, err)
	assert.EqualError(t, err, "結果の表示に失敗しました: presenter failed to present")

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}

func TestRunWeatherPointLogic_NoResults(t *testing.T) {
	mockClient := new(MockClient)
	mockPresenter := new(MockPresenter)

	keyword := "非常に稀な地名"
	kata := false
	// APIが結果なしの場合に空のレスポンスとnilエラーを返すことを想定
	emptyResponse := models.GetWeatherPointResponse{
		Result: models.WeatherPoints{
			Root: []models.WeatherPoint{}, // 空の配列
		},
	}

	mockClient.On("GetWeatherPoint", keyword).Return(emptyResponse, nil)
	mockPresenter.On("PresentWeatherPoint", emptyResponse, kata, keyword).Return(nil) // プレゼンターは結果なしを適切に処理すると期待

	err := runWeatherPointLogic(mockClient, mockPresenter, keyword, kata)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockPresenter.AssertExpectations(t)
}
