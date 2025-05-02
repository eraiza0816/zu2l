package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/eraiza0816/zu2l/internal/models"
)

const (
	defaultBaseURL         = "https://zutool.jp/api"
	defaultOtenkiASPBaseURL = "https://ap.otenki.com/OtenkiASP/asp"
	defaultTimeout         = 10 * time.Second
	defaultUserAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
)

// Client は zutool API および Otenki ASP API との通信を管理するリポジトリです。
// 外部データソースから集約を取得する役割を担います。
type Client struct {
	baseURL     string // zutool API のベースURL
	otenkiBaseURL string // Otenki ASP API のベースURL
	httpClient  *http.Client
	userAgent   string
}

// NewClient は新しいAPIクライアント（リポジトリ実装）を作成します。
// baseURL または otenkiBaseURL が空文字列の場合、デフォルト値が使用されます。
// timeout がゼロの場合、デフォルトのタイムアウト値が使用されます。
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

// doRequest はHTTPリクエストを実行し、共通のロジック（User-Agent設定、レスポンス読み込み、ステータスコードチェック、エラー処理）を処理するヘルパーメソッドです。
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", c.userAgent) // Client構造体のuserAgentを使用
	resp, err := c.httpClient.Do(req) // Client構造体のhttpClientを使用
	if err != nil {
		// ネットワークエラーまたは接続エラー
		return nil, fmt.Errorf("リクエストの実行に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// レスポンスボディの読み込みエラー
		return nil, fmt.Errorf("レスポンスボディの読み込みに失敗しました: %w", err)
	}

	// ステータスコードが200 OKでない場合のエラーハンドリング
	if resp.StatusCode != http.StatusOK {
		var errorResponse models.ErrorResponse
		// zutool固有のエラー形式のデコードを試みる
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			// デコード成功し、エラーメッセージがあればAPIErrorを生成して返す
			return nil, newAPIError(resp.StatusCode, string(body), errorResponse.ErrorMessage, nil)
		}
		// デコード失敗、またはエラーメッセージがない場合は、汎用のAPIErrorを生成して返す
		return nil, newAPIError(resp.StatusCode, string(body), "", nil)
	}

	// 成功した場合はレスポンスボディを返す
	return body, nil
}

// _get はプライマリAPI (baseURL) から生のレスポンスボディを取得するプライベートヘルパーメソッドです。
// 200 OKレスポンス内に埋め込まれたAPIエラーもチェックします。
func (c *Client) _get(path string, param string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s%s/%s", c.baseURL, path, param)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s のリクエスト作成に失敗しました: %w", path, err)
	}

	// doRequestメソッドを呼び出してリクエストを実行
	body, err := c.doRequest(req)
	if err != nil {
		// doRequestがエラー（例: 非200ステータス）を返した場合、そのまま返す
		return nil, fmt.Errorf("%s のリクエストに失敗しました: %w", path, err)
	}

	// --- Start: 200 OKレスポンス内のAPIエラーチェック ---
	// (このロジックは現状維持ですが、将来リファクタリングされる可能性があります)
	var errorResponse models.ErrorResponse
	// まずエラー構造体へのアンマーシャルを試みる
	// 有効なデータの場合、ErrorResponseへのアンマーシャル失敗は想定されるため、エラーは無視する
	_ = json.Unmarshal(body, &errorResponse)

	// ErrorResponseへのアンマーシャルが成功し、かつエラーメッセージが存在する場合、
	// APIは200 OKを返したが、内部的にエラーを示していることを意味する
	if errorResponse.ErrorMessage != "" {
		// ボディと特定のAPIErrorを一緒に返す
		return body, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
	}
	// --- End: 200 OKレスポンス内のAPIエラーチェック ---

	// 埋め込みエラーがない場合は、呼び出し元がアンマーシャルするために生のボディを返す
	return body, nil
}

// GetPainStatus は zutool API から痛み指数情報の集約 (GetPainStatusResponse) を取得します。
// setWeatherPointが指定されている場合、事前に地点設定APIを呼び出します。
func (c *Client) GetPainStatus(
	areaCode string, // エリアコード
	setWeatherPoint *string, // 地点コード (オプション)
) (models.GetPainStatusResponse, error) {
	var result models.GetPainStatusResponse

	// --- Start: 地点設定API呼び出しロジック ---
	if setWeatherPoint != nil && *setWeatherPoint != "" {
		setWeatherPointURL := fmt.Sprintf("%s/setweatherpoint/%s", c.baseURL, *setWeatherPoint)
		req, err := http.NewRequest("GET", setWeatherPointURL, nil)
		if err != nil {
			return result, fmt.Errorf("setweatherpointリクエストの作成に失敗しました: %w", err)
		}

		// 地点設定APIを呼び出す
		setBody, err := c.doRequest(req)
		if err != nil {
			// doRequestからのAPIErrorかどうかを確認
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				// setweatherpointの404は特定のエラーとして扱う
				return result, fmt.Errorf("地点コード '%s' が見つかりません (setweatherpoint failed with 404)", *setWeatherPoint)
			}
			// それ以外のエラーはラップして返す
			return result, fmt.Errorf("setweatherpointリクエストに失敗しました: %w", err)
		}

		// 地点設定APIのレスポンスを解析
		var setResp models.SetWeatherPointResponse
		if err := json.Unmarshal(setBody, &setResp); err != nil {
			return result, fmt.Errorf("setweatherpointレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(setBody))
		}
		if setResp.Response != "ok" {
			// レスポンスが "ok" でない場合はエラーとする
			// TODO: より具体的なエラーを返すことを検討
			return result, fmt.Errorf("setweatherpointレスポンスが 'ok' ではありませんでした: %s", setResp.Response)
		}
	}
	// --- End: 地点設定API呼び出しロジック ---

	// 痛み指数APIを呼び出す (_getを使用)
	body, err := c._get("/getpainstatus", areaCode)
	if err != nil {
		// _getが返したエラーが埋め込みAPIError (status 200) かどうかを確認
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			// 埋め込みエラーの場合はそのまま返す
			return result, err
		}
		// それ以外のエラーはラップして返す
		return result, fmt.Errorf("痛み指数情報の取得に失敗しました: %w", err)
	}

	// 正常なレスポンスボディを結果構造体にアンマーシャル
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("GetPainStatusレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	return result, nil
}

// GetWeatherPoint は zutool API から地点検索情報の集約 (GetWeatherPointResponse) を取得します。
func (c *Client) GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error) {
	// 内部の_getメソッドを使用して生のボディを取得
	body, err := c._get("/getweatherpoint", keyword)
	if err != nil {
		// doRequest (via _get) からのAPIErrorかどうかを確認
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// 既知のAPIエラー（例：空キーワードでの404）の場合はそのまま返す
			// レスポンス構造体のゼロ値とエラーを返す
			return models.GetWeatherPointResponse{}, err
		}
		// それ以外のエラーはラップして返す
		return models.GetWeatherPointResponse{}, fmt.Errorf("/getweatherpoint のリクエストに失敗しました: %w", err)
	}

	// 専用のパーサー関数を呼び出す
	return parseWeatherPointResponse(body)
}

// GetWeatherStatus は zutool API から詳細な気象状況の集約 (GetWeatherStatusResponse) を取得します。
func (c *Client) GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) {
	var result models.GetWeatherStatusResponse

	// 内部の_getメソッドを使用して生のボディを取得
	body, err := c._get("/getweatherstatus", cityCode)
	if err != nil {
		// _getが返したエラーが埋め込みAPIError (status 200) かどうかを確認
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			// 埋め込みエラーの場合はそのまま返す
			return result, err
		}
		// それ以外のエラーはラップして返す
		return result, fmt.Errorf("気象状況の取得に失敗しました: %w", err)
	}

	// 正常なレスポンスボディを結果構造体にアンマーシャル
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("GetWeatherStatusレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	return result, nil
}

// GetOtenkiASP は Otenki ASP API から気象情報の集約 (GetOtenkiASPResponse) を取得します。
func (c *Client) GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error) {
	params := url.Values{}
	params.Set("csid", "mmcm") // 現状維持
	params.Set("contents_id", "day_tenki--day_pre--hight_temp--low_temp--day_wind_v--day_wind_d--zutu_level_day--low_humidity") // 現状維持
	params.Set("duration_yohoushi", "7")
	params.Set("where", fmt.Sprintf("CHITEN_%s", cityCode))
	params.Set("json", "on") // 現状維持

	apiURL := fmt.Sprintf("%s/getElements", c.otenkiBaseURL) // Client構造体のotenkiBaseURLを使用
	u, err := url.Parse(apiURL)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASP URLのパースに失敗しました: %w", err)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASPリクエストの作成に失敗しました: %w", err)
	}

	// ClientのdoRequestを使用してリクエストを実行
	body, err := c.doRequest(req)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASPリクエストに失敗しました: %w", err)
	}

	// 専用のパーサー関数を呼び出す
	return parseOtenkiASPResponse(body)
}
