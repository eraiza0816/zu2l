package api

import (
	"fmt"
	"net/http"
)

// APIError は API 通信に関連するエラーを表すカスタムエラー型です。
type APIError struct {
	StatusCode int    // HTTPステータスコード
	Body       string // レスポンスボディ (エラー時)
	Message    string // APIから返された、または内部で生成された特定のエラーメッセージ
	Err        error  // ラップされた元のエラー (例: ネットワークエラー)
}

// Error は error インターフェースを実装し、APIError の内容に基づいたエラーメッセージ文字列を返します。
// Message, Err, Body の順で利用可能な情報を使用してメッセージを構築します。
func (e *APIError) Error() string {
	if e.Message != "" {
		// Message があればそれを使用
		return fmt.Sprintf("APIエラー: %s (ステータス: %d)", e.Message, e.StatusCode)
	}
	if e.Err != nil {
		// Err があればそれを使用
		return fmt.Sprintf("APIエラー (ステータス: %d): %v", e.StatusCode, e.Err)
	}
	// Message も Err もなければ Body を使用
	return fmt.Sprintf("APIエラー (ステータス: %d): %s", e.StatusCode, e.Body)
}

// newAPIError は新しい APIError インスタンスを作成するヘルパー関数です。
func newAPIError(statusCode int, body string, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Body:       body,
		Message:    message,
		Err:        err,
	}
}

// newNotFoundError は指定されたリソースが見つからない場合の 404 Not Found エラーを生成するヘルパー関数です。
func newNotFoundError(resource string, identifier string) *APIError {
	return newAPIError(
		http.StatusNotFound, // ステータスコードは 404
		"",                  // ボディは空
		fmt.Sprintf("%s '%s' が見つかりません", resource, identifier),
		nil,
	)
}
