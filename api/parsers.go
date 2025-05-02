package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"
)

// parseWeatherPointResponse は /getweatherpoint エンドポイントの生のレスポンスボディを解析します。
// "result" フィールド内のネストされたJSON文字列を処理します。
func parseWeatherPointResponse(body []byte) (models.GetWeatherPointResponse, error) {
	var finalResult models.GetWeatherPointResponse

	// "result" フィールドが文字列である初期レスポンスをキャプチャするための一時的な構造体を定義
	type tempResp struct {
		Result string `json:"result"`
	}
	var tempResult tempResp

	if err := json.Unmarshal(body, &tempResult); err != nil {
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// HTTPリクエスト自体は成功している可能性があるため、ステータスコード200を使用
			return finalResult, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
		}
		return finalResult, fmt.Errorf("/getweatherpoint の初期レスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	var weatherPoints models.WeatherPoints
	var points []models.WeatherPoint
	if errUnmarshalArray := json.Unmarshal([]byte(tempResult.Result), &points); errUnmarshalArray != nil {
		// `{"result":"[]"}` のようなケースを正しく処理し、空のpointsスライスにする
		if tempResult.Result != "[]" { // 空の配列文字列 "[]" は許可
			return finalResult, fmt.Errorf("/getweatherpoint のネストされたresult文字列のアンマーシャルに失敗しました: %w, result string: %s", errUnmarshalArray, tempResult.Result)
		}
		// "[]" だった場合、pointsはnilまたは空になり、これは問題ない
	}
	weatherPoints.Root = points
	finalResult.Result = weatherPoints

	return finalResult, nil
}

// parseOtenkiASPResponse は Otenki ASP API からの生のレスポンスボディを解析します。
func parseOtenkiASPResponse(body []byte) (models.GetOtenkiASPResponse, error) {
	var rawResponse models.GetOtenkiASPRawResponse
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// APIエラー形式であれば、APIErrorを返す (ステータスコードは200とする)
			return models.GetOtenkiASPResponse{}, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
		}
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASP 生レスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	// 生レスポンスをより扱いやすい構造化レスポンス (GetOtenkiASPResponse) に変換
	response := models.GetOtenkiASPResponse{
		Status:   rawResponse.Head.Status,
		DateTime: rawResponse.Head.DateTime,
		Elements: make([]models.Element, 0, len(rawResponse.Body.Location.Element)), // 事前に容量確保
	}

	for _, rawElem := range rawResponse.Body.Location.Element {
		// 各要素はヘッダーレコードとデータレコードを持つ必要がある
		if len(rawElem.Record) < 2 {
			fmt.Printf("レコード数が不足しているため、生の要素をスキップします (ヘッダー + データが必要です): %+v\n", rawElem.Record)
			continue
		}

		headerRecord := rawElem.Record[0]
		if len(headerRecord.Property) < 2 {
			fmt.Printf("ヘッダープロパティの構造が無効なため、要素をスキップします (ContentID + Titleが必要です): %+v\n", headerRecord.Property)
			continue
		}
		headerProps := headerRecord.Property
		contentID, okID := headerProps[0].(string)
		title, okTitle := headerProps[1].(string)
		if !okID || !okTitle {
			fmt.Printf("ヘッダーのContentIDまたはTitleの型が無効なため、要素をスキップします: %+v\n", headerProps)
			continue
		}

		elem := models.Element{
			ContentID: contentID,
			Title:     title,
			Records:   make(map[time.Time]interface{}), // 時刻をキーとするマップ
		}

		for _, rawDataProperty := range rawElem.Record[1:] {
			// 各データプロパティは時刻と値を持つ必要がある
			if len(rawDataProperty.Property) < 2 {
				fmt.Printf("データプロパティの構造が無効なため、スキップします (時刻 + 値が必要です): %+v\n", rawDataProperty.Property)
				continue
			}
			dataProps := rawDataProperty.Property
			timeStr, okTime := dataProps[0].(string) // 1番目が時刻文字列のはず
			value := dataProps[1]                   // 2番目が値のはず

			if !okTime {
				fmt.Printf("時刻が文字列でないデータレコードをスキップします: %v\n", dataProps[0])
				continue
			}

			// タイムスタンプ文字列を time.Time 型にパース (複数のフォーマットを試行)
			var t time.Time
			var err error
			parsed := false
			// 想定されるフォーマットのリスト
			formats := []string{time.RFC3339, "2006-01-02T15:04:05", "20060102"}
			for _, format := range formats {
				t, err = time.Parse(format, timeStr)
				if err == nil {
					parsed = true
					break
				}
			}

			if !parsed {
				fmt.Printf("時刻 '%s' をパースできないため、データレコードをスキップします: 試行したフォーマット %v\n", timeStr, formats)
				continue
			}
			elem.Records[t] = value
		}
		response.Elements = append(response.Elements, elem)
	}

	return response, nil
}
