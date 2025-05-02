package api_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/eraiza0816/zu2l/api"
)

// テスト用の新しいクライアントを作成するヘルパー関数
func newTestClient() *api.Client {
	// 実際のAPIに対してテストするためにデフォルトのURLとタイムアウトを使用
	// 将来的には、より制御されたテストのためにモックサーバーの使用を検討
	return api.NewClient("", "", 0)
}

// --- APIクライアントテスト ---

// TestGetPainStatus は GetPainStatus の正常系テストです。
func TestGetPainStatus(t *testing.T) {
	client := newTestClient()
	areaCode := "13" // 東京都
	_, err := client.GetPainStatus(areaCode, nil) // setWeatherPoint なしで呼び出し
	if err != nil {
		t.Errorf("client.GetPainStatus(%q) が失敗しました: %v", areaCode, err)
	}
	// 注: ここでレスポンス内容に対するさらなるチェックを追加できます。
}

// TestGetPainStatusInvalid は GetPainStatus の異常系テスト (無効なエリアコード) です。
func TestGetPainStatusInvalid(t *testing.T) {
	client := newTestClient()
	areaCode := "1" // 無効なコード
	_, err := client.GetPainStatus(areaCode, nil)
	if err == nil {
		t.Errorf("client.GetPainStatus(%q) は失敗するはずですが、nil エラーが返されました", areaCode)
		return
	}

	// エラーが期待される APIError 型であり、メッセージが期待通りかチェック
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		// 特定のエラーコード/メッセージはAPIの実装に依存する可能性があります
		expectedMsgPart := "存在しない都道府県コードです" // 元のテストに基づく部分的なチェック
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetPainStatus(%q) が予期しないメッセージで失敗しました: %q, 期待する部分文字列: %q", areaCode, apiErr.Message, expectedMsgPart)
		}
		// オプション: APIが一貫して特定のコードを返す場合は apiErr.StatusCode をチェック
		// if apiErr.StatusCode != expectedStatusCode { ... }
	} else {
		// 予期しないエラー型の場合
		t.Errorf("client.GetPainStatus(%q) が予期しないエラー型で失敗しました: %T, 期待する型: *api.APIError", areaCode, err)
	}
}

// TestGetPainStatusSetWeatherPoint は GetPainStatus で setWeatherPoint を指定した場合のテストです。
func TestGetPainStatusSetWeatherPoint(t *testing.T) {
	client := newTestClient()

	// テストケース1: 札幌
	cityCode1 := "01101"
	setWeatherPoint1 := cityCode1
	areaCode1 := cityCode1[:2] // "01" (北海道)
	expectedAreaName1 := "北海道"
	res1, err1 := client.GetPainStatus(areaCode1, &setWeatherPoint1)
	if err1 != nil {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) が失敗しました: %v", areaCode1, setWeatherPoint1, err1)
	} else if res1.PainnoterateStatus.AreaName != expectedAreaName1 {
		// レスポンスの AreaName が期待通りかチェック
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) が予期しない AreaName を返しました: %q, 期待値: %q", areaCode1, setWeatherPoint1, res1.PainnoterateStatus.AreaName, expectedAreaName1)
	}

	// テストケース2: 渋谷
	cityCode2 := "13113"
	setWeatherPoint2 := cityCode2
	areaCode2 := cityCode2[:2] // "13" (東京都)
	expectedAreaName2 := "東京都"
	res2, err2 := client.GetPainStatus(areaCode2, &setWeatherPoint2)
	if err2 != nil {
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) が失敗しました: %v", areaCode2, setWeatherPoint2, err2)
	} else if res2.PainnoterateStatus.AreaName != expectedAreaName2 {
		// レスポンスの AreaName が期待通りかチェック
		t.Errorf("client.GetPainStatus(%q, setWeatherPoint=%q) が予期しない AreaName を返しました: %q, 期待値: %q", areaCode2, setWeatherPoint2, res2.PainnoterateStatus.AreaName, expectedAreaName2)
	}
}

// TestGetWeatherPoint は GetWeatherPoint の正常系テストです。
func TestGetWeatherPoint(t *testing.T) {
	client := newTestClient()
	keyword := "神戸市"
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		t.Errorf("client.GetWeatherPoint(%q) が失敗しました: %v", keyword, err)
	}
	// 結果が空でないことを確認
	if len(res.Result.Root) == 0 {
		t.Errorf("client.GetWeatherPoint(%q) が空の結果を返しました。空でない結果を期待していました", keyword)
	}
}

// TestGetWeatherPointExtra は GetWeatherPoint で結果が少ない、または無い場合のテストです。
func TestGetWeatherPointExtra(t *testing.T) {
	client := newTestClient()
	keyword := "a" // ほとんど、または全く結果が返らないと予想されるキーワード
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		// APIの挙動によっては、ここでのエラーも許容される可能性があります
		t.Logf("client.GetWeatherPoint(%q) が期待通り失敗した可能性があります: %v", keyword, err)
		// 404 APIエラーかどうかを確認 (「結果なし」として有効な場合がある)
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// 404が「見つかりません」を意味する場合、これはパスと見なせる
			return
		}
		// 他のエラーであればテスト失敗
		t.Errorf("client.GetWeatherPoint(%q) が予期せず失敗しました: %v", keyword, err)
		return
	}
	// エラーがない場合、結果リストが空かどうかを確認 (元のPythonテストに基づく)
	if len(res.Result.Root) != 0 {
		t.Errorf("client.GetWeatherPoint(%q) が %d 件の結果を返しました。0件または404エラーを期待していました", keyword, len(res.Result.Root))
	}
}

// TestGetWeatherPointEmpty は GetWeatherPoint で空のキーワードを指定した場合のテストです。
func TestGetWeatherPointEmpty(t *testing.T) {
	client := newTestClient()
	keyword := ""
	_, err := client.GetWeatherPoint(keyword)
	if err == nil {
		t.Errorf("client.GetWeatherPoint(%q) は失敗するはずですが、nil エラーが返されました", keyword)
		return
	}
	// 特定の APIError をチェック (元のテストに基づき 404 を期待)
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != 404 {
			t.Errorf("client.GetWeatherPoint(%q) がステータスコード %d で失敗しました。404 を期待していました", keyword, apiErr.StatusCode)
		}
		// メッセージチェックは追加可能ですが、壊れやすい可能性があります
		// expectedMsgPart := "お探しのページは見つかりません"
		// if !strings.Contains(apiErr.Message, expectedMsgPart) { ... }
	} else {
		t.Errorf("client.GetWeatherPoint(%q) が予期しないエラー型で失敗しました: %T, 期待する型: *api.APIError", keyword, err)
	}
}

// TestGetWeatherStatus は GetWeatherStatus の正常系テストです。
func TestGetWeatherStatus(t *testing.T) {
	client := newTestClient()
	cityCode := "13113" // 渋谷
	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		t.Errorf("client.GetWeatherStatus(%q) が失敗しました: %v", cityCode, err)
	}
	// PrefecturesID と PlaceID を結合したものが元の cityCode と一致するかチェック
	// 注: models.PrefecturesID の型が異なる場合 (例: int)、比較を調整する必要があります。
	if string(res.PrefecturesID)+res.PlaceID != cityCode {
		t.Errorf("client.GetWeatherStatus(%q) が予期しない組み合わせを返しました: PrefID=%q, PlaceID=%q, 期待値: %q",
			cityCode, res.PrefecturesID, res.PlaceID, cityCode)
	}
}

// TestGetWeatherStatusInvalidCode は GetWeatherStatus の異常系テスト (無効な形式のコード) です。
func TestGetWeatherStatusInvalidCode(t *testing.T) {
	client := newTestClient()
	cityCode := "aaaaa" // 無効な形式のコード
	_, err := client.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("client.GetWeatherStatus(%q) は失敗するはずですが、nil エラーが返されました", cityCode)
		return
	}
	// APIError と、可能性として特定のメッセージ/コードをチェック
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点名称が取得できませんでした" // 元のテストに基づく
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetWeatherStatus(%q) が予期しないメッセージで失敗しました: %q, 期待する部分文字列: %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// オプション: 既知であればステータスコードをチェック
	} else {
		t.Errorf("client.GetWeatherStatus(%q) が予期しないエラー型で失敗しました: %T, 期待する型: *api.APIError", cityCode, err)
	}
}

// TestGetWeatherStatusInvalidDigit は GetWeatherStatus の異常系テスト (桁数が不正なコード) です。
func TestGetWeatherStatusInvalidDigit(t *testing.T) {
	client := newTestClient()
	cityCode := "13" // 桁数が足りないコード
	_, err := client.GetWeatherStatus(cityCode)
	if err == nil {
		t.Errorf("client.GetWeatherStatus(%q) は失敗するはずですが、nil エラーが返されました", cityCode)
		return
	}
	// APIError と、可能性として特定のメッセージ/コードをチェック
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		expectedMsgPart := "地点コードの桁数が正しくありません" // 元のテストに基づく
		if !strings.Contains(apiErr.Message, expectedMsgPart) {
			t.Errorf("client.GetWeatherStatus(%q) が予期しないメッセージで失敗しました: %q, 期待する部分文字列: %q", cityCode, apiErr.Message, expectedMsgPart)
		}
		// オプション: 既知であればステータスコードをチェック
	} else {
		t.Errorf("client.GetWeatherStatus(%q) が予期しないエラー型で失敗しました: %T, 期待する型: *api.APIError", cityCode, err)
	}
}

// TestGetOtenkiASP は GetOtenkiASP の正常系テストです。
func TestGetOtenkiASP(t *testing.T) {
	client := newTestClient()
	cityCode := "13101" // 東京
	res, err := client.GetOtenkiASP(cityCode)
	if err != nil {
		t.Errorf("client.GetOtenkiASP(%q) が失敗しました: %v", cityCode, err)
	}
	// レスポンスの Status フィールドをチェック
	if res.Status != "OK" { // models.GetOtenkiASPResponse に Status フィールドがあると仮定
		t.Errorf("client.GetOtenkiASP(%q) がステータス %q を返しました。期待値: %q", cityCode, res.Status, "OK")
	}
}

// TestGetOtenkiASPInvalidCode は GetOtenkiASP の異常系テスト (無効なコード) です。
func TestGetOtenkiASPInvalidCode(t *testing.T) {
	client := newTestClient()
	cityCode := "13000" // 元のテストに基づき、OtenkiASP にとって無効なコード
	_, err := client.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("client.GetOtenkiASP(%q) は失敗するはずですが、nil エラーが返されました", cityCode)
	}
	// 必要であれば、エラー型 (例: *api.APIError) に対するさらなるチェックを追加できます。
}

// TestGetOtenkiASPEmpty は GetOtenkiASP で空のコードを指定した場合のテストです。
func TestGetOtenkiASPEmpty(t *testing.T) {
	client := newTestClient()
	cityCode := ""
	_, err := client.GetOtenkiASP(cityCode)
	if err == nil {
		t.Errorf("client.GetOtenkiASP(%q) は失敗するはずですが、nil エラーが返されました", cityCode)
	}
	// 必要であれば、エラー型に対するさらなるチェックを追加できます。
}
