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

// Client manages communication with the zutool and Otenki ASP APIs.
type Client struct {
	baseURL     string
	otenkiBaseURL string
	httpClient  *http.Client
	userAgent   string
}

// NewClient は新しいAPIクライアントを作成します。
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
		userAgent: defaultUserAgent,
	}
}

// doRequest はHTTPリクエストを実行し、共通のロジック（User-Agent設定、レスポンス読み込み、ステータスコードチェック、エラー処理）を処理するヘルパーメソッドです。
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの実行に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスボディの読み込みに失敗しました: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse models.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.ErrorMessage != "" {
			return nil, newAPIError(resp.StatusCode, string(body), errorResponse.ErrorMessage, nil)
		}
		return nil, newAPIError(resp.StatusCode, string(body), "", nil)
	}

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

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%s のリクエストに失敗しました: %w", path, err)
	}

	// --- Start: 200 OKレスポンス内のAPIエラーチェック ---
	// (このロジックは現状維持ですが、将来リファクタリングされる可能性があります)
	var errorResponse models.ErrorResponse
	_ = json.Unmarshal(body, &errorResponse) // エラーは無視

	if errorResponse.ErrorMessage != "" {
		return body, newAPIError(http.StatusOK, string(body), errorResponse.ErrorMessage, nil)
	}
	// --- End: 200 OKレスポンス内のAPIエラーチェック ---

	return body, nil
}

// GetPainStatus は痛み指数情報を取得します。
// setWeatherPointが指定されている場合、事前に地点設定APIを呼び出します。
func (c *Client) GetPainStatus(
	areaCode string,
	setWeatherPoint *string,
) (models.GetPainStatusResponse, error) {
	var result models.GetPainStatusResponse

	// 地点設定API呼び出しロジック
	if setWeatherPoint != nil && *setWeatherPoint != "" {
		setWeatherPointURL := fmt.Sprintf("%s/setweatherpoint/%s", c.baseURL, *setWeatherPoint)
		req, err := http.NewRequest("GET", setWeatherPointURL, nil)
		if err != nil {
			return result, fmt.Errorf("setweatherpointリクエストの作成に失敗しました: %w", err)
		}

		setBody, err := c.doRequest(req)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				return result, fmt.Errorf("地点コード '%s' が見つかりません (setweatherpoint failed with 404)", *setWeatherPoint)
			}
			return result, fmt.Errorf("setweatherpointリクエストに失敗しました: %w", err)
		}

		var setResp models.SetWeatherPointResponse
		if err := json.Unmarshal(setBody, &setResp); err != nil {
			return result, fmt.Errorf("setweatherpointレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(setBody))
		}
		if setResp.Response != "ok" {
			// TODO: より具体的なエラーを返すことを検討
			return result, fmt.Errorf("setweatherpointレスポンスが 'ok' ではありませんでした: %s", setResp.Response)
		}
	}

	// 痛み指数API呼び出し
	body, err := c._get("/getpainstatus", areaCode)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			return result, err // 埋め込みエラー
		}
		return result, fmt.Errorf("痛み指数情報の取得に失敗しました: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("GetPainStatusレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	return result, nil
}

// GetWeatherPoint は地点検索情報を取得します。
func (c *Client) GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error) {
	body, err := c._get("/getweatherpoint", keyword)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return models.GetWeatherPointResponse{}, err // APIエラーはそのまま返す
		}
		return models.GetWeatherPointResponse{}, fmt.Errorf("/getweatherpoint のリクエストに失敗しました: %w", err)
	}

	return parseWeatherPointResponse(body)
}

// GetWeatherStatus は詳細な気象状況を取得します。
func (c *Client) GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) {
	var result models.GetWeatherStatusResponse

	body, err := c._get("/getweatherstatus", cityCode)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusOK {
			return result, err // 埋め込みエラー
		}
		return result, fmt.Errorf("気象状況の取得に失敗しました: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("GetWeatherStatusレスポンスのアンマーシャルに失敗しました: %w, body: %s", err, string(body))
	}

	return result, nil
}

// GetOtenkiASP はOtenki ASPから気象情報を取得します。
func (c *Client) GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error) {
	params := url.Values{}
	params.Set("csid", "mmcm")
	params.Set("contents_id", "day_tenki--day_pre--hight_temp--low_temp--day_wind_v--day_wind_d--zutu_level_day--low_humidity")
	params.Set("duration_yohoushi", "7")
	params.Set("where", fmt.Sprintf("CHITEN_%s", cityCode))
	params.Set("json", "on")

	apiURL := fmt.Sprintf("%s/getElements", c.otenkiBaseURL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASP URLのパースに失敗しました: %w", err)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASPリクエストの作成に失敗しました: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return models.GetOtenkiASPResponse{}, fmt.Errorf("Otenki ASPリクエストに失敗しました: %w", err)
	}

	return parseOtenkiASPResponse(body)
}
