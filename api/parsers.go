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

	// 生のボディを一時的な構造体にアンマーシャル
	if err := json.Unmarshal(body, &tempResult); err != nil {
		// まず、ボディ内にAPIエラー構造体が存在するかチェック
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// ボディが既知のAPIエラー形式を表す場合、APIErrorを返す
			// HTTPリクエスト自体は成功している可能性があるため、ステータスコード200を使用
			return finalResult, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
		}
		// それ以外の場合は、一時構造体のアンマーシャルエラーを返す
		return finalResult, fmt.Errorf("/getweatherpoint の初期レスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	// "result" 文字列を実際の WeatherPoints 構造体にアンマーシャル
	var weatherPoints models.WeatherPoints
	var points []models.WeatherPoint
	if errUnmarshalArray := json.Unmarshal([]byte(tempResult.Result), &points); errUnmarshalArray != nil {
		// `{"result":"[]"}` のようなケースを正しく処理し、空のpointsスライスにする
		// result文字列が有効なJSON配列でない場合はエラーを返す
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
	// まず生のレスポンス全体を GetOtenkiASPRawResponse 構造体にアンマーシャル
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		// アンマーシャル失敗時に、APIエラー構造体かどうかをチェック
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// APIエラー形式であれば、APIErrorを返す (ステータスコードは200とする)
			return models.GetOtenkiASPResponse{}, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
		}
		// それ以外のアンマーシャルエラー
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASP 生レスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	// 生レスポンスをより扱いやすい構造化レスポンス (GetOtenkiASPResponse) に変換
	response := models.GetOtenkiASPResponse{
		Status:   rawResponse.Head.Status,
		DateTime: rawResponse.Head.DateTime,
		Elements: make([]models.Element, 0, len(rawResponse.Body.Location.Element)), // 事前に容量確保
	}

	// Element スライス内の各 RawRecord を反復処理
	for _, rawElem := range rawResponse.Body.Location.Element { // rawElem は models.RawRecord 型
		// 各要素はヘッダーレコードとデータレコードを持つ必要がある
		if len(rawElem.Record) < 2 {
			fmt.Printf("レコード数が不足しているため、生の要素をスキップします (ヘッダー + データが必要です): %+v\n", rawElem.Record)
			continue // 次の要素へ
		}

		// ヘッダーレコードから ContentID と Title を抽出
		headerRecord := rawElem.Record[0]
		if len(headerRecord.Property) < 2 {
			fmt.Printf("ヘッダープロパティの構造が無効なため、要素をスキップします (ContentID + Titleが必要です): %+v\n", headerRecord.Property)
			continue // 次の要素へ
		}
		headerProps := headerRecord.Property
		contentID, okID := headerProps[0].(string)
		title, okTitle := headerProps[1].(string)
		if !okID || !okTitle {
			fmt.Printf("ヘッダーのContentIDまたはTitleの型が無効なため、要素をスキップします: %+v\n", headerProps)
			continue // 次の要素へ
		}

		// 変換後の Element 構造体を初期化
		elem := models.Element{
			ContentID: contentID,
			Title:     title,
			Records:   make(map[time.Time]interface{}), // 時刻をキーとするマップ
		}

		// データレコード (ヘッダー以降) を反復処理して Records マップを構築
		for _, rawDataProperty := range rawElem.Record[1:] {
			// 各データプロパティは時刻と値を持つ必要がある
			if len(rawDataProperty.Property) < 2 {
				fmt.Printf("データプロパティの構造が無効なため、スキップします (時刻 + 値が必要です): %+v\n", rawDataProperty.Property)
				continue // 次のデータプロパティへ
			}
			dataProps := rawDataProperty.Property
			timeStr, okTime := dataProps[0].(string) // 1番目が時刻文字列のはず
			value := dataProps[1]                   // 2番目が値のはず

			if !okTime {
				fmt.Printf("時刻が文字列でないデータレコードをスキップします: %v\n", dataProps[0])
				continue // 次のデータプロパティへ
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
					parsed = true // パース成功
					break
				}
			}

			if !parsed {
				// いずれのフォーマットでもパースできなかった場合
				fmt.Printf("時刻 '%s' をパースできないため、データレコードをスキップします: 試行したフォーマット %v\n", timeStr, formats)
				continue // 次のデータプロパティへ
			}
			// パース成功したら、時刻をキーとして値をマップに格納
			elem.Records[t] = value
		}
		// 完成した Element をレスポンススライスに追加
		response.Elements = append(response.Elements, elem)
	}

	return response, nil
}
