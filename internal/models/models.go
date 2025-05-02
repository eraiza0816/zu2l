package models

import (
	"fmt"
	"strings"
	"time"
)

// APIDateTime は API からの特定の日時フォーマット "YYYY-MM-DD HH" を処理するためのカスタム型 (値オブジェクト) です。
type APIDateTime struct {
	time.Time // time.Time を埋め込む
}

// apiDateTimeLayout は APIDateTime のパースおよびフォーマットに使用される期待されるレイアウトを定義します。
const apiDateTimeLayout = "2006-01-02 15" // "YYYY-MM-DD HH" 形式

// UnmarshalJSON は APIDateTime の json.Unmarshaler インターフェースを実装します。
// 文字列 "YYYY-MM-DD HH" を time.Time オブジェクトにパースします。
func (adt *APIDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`) // ダブルクォートを除去
	// null または空文字列の場合の処理
	if s == "null" || s == "" {
		adt.Time = time.Time{} // ゼロ値に設定
		return nil
	}
	// まず "YYYY-MM-DD HH" 形式でのパースを試みる
	t, err := time.Parse(apiDateTimeLayout, s)
	if err != nil {
		// パース失敗した場合、時刻部分が欠落している可能性を考慮し、
		// 日付のみ ("YYYY-MM-DD") の形式でのパースを試みる
		// (一部のAPIレスポンスで必要になる可能性があるため。必要に応じて調整)
		t, errDate := time.Parse("2006-01-02", s)
		if errDate != nil {
			// 日付のみのパースも失敗した場合、元のパースエラーを返す
			return fmt.Errorf("APIDateTime %q のパースに失敗しました: %w", s, err)
		}
		// 日付のみのパースが成功した場合、その結果を使用する (時刻部分は 00:00:00 になる)
		adt.Time = t
		return nil
	}
	// "YYYY-MM-DD HH" 形式でのパースが成功した場合
	adt.Time = t
	return nil
}

// MarshalJSON は APIDateTime の json.Marshaler インターフェースを実装します。
// time.Time オブジェクトを "YYYY-MM-DD HH" 文字列形式にフォーマットします。
func (adt APIDateTime) MarshalJSON() ([]byte, error) {
	// Time がゼロ値の場合、"null" を返す
	if adt.Time.IsZero() {
		return []byte("null"), nil
	}
	// 指定されたレイアウトでフォーマットし、ダブルクォートで囲んだバイトスライスを返す
	return []byte(`"` + adt.Time.Format(apiDateTimeLayout) + `"`), nil
}
