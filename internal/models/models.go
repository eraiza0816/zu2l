package models

import (
	"fmt"
	"strings"
	"time"
)

// APIDateTime は API からの特定の日時フォーマット "YYYY-MM-DD HH" を処理するためのカスタム型 (値オブジェクト) です。
type APIDateTime struct {
	time.Time
}

// apiDateTimeLayout は APIDateTime のパースおよびフォーマットに使用される期待されるレイアウトを定義します。
const apiDateTimeLayout = "2006-01-02 15"

// UnmarshalJSON は APIDateTime の json.Unmarshaler インターフェースを実装します。
// 文字列 "YYYY-MM-DD HH" を time.Time オブジェクトにパースします。
func (adt *APIDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		adt.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(apiDateTimeLayout, s)
	if err != nil {
		// パース失敗した場合、日付のみ ("YYYY-MM-DD") の形式でのパースを試みる
		// (一部のAPIレスポンスで必要になる可能性があるため)
		t, errDate := time.Parse("2006-01-02", s)
		if errDate != nil {
			return fmt.Errorf("APIDateTime %q のパースに失敗しました: %w", s, err)
		}
		adt.Time = t
		return nil
	}
	adt.Time = t
	return nil
}

// MarshalJSON は APIDateTime の json.Marshaler インターフェースを実装します。
// time.Time オブジェクトを "YYYY-MM-DD HH" 文字列形式にフォーマットします。
func (adt APIDateTime) MarshalJSON() ([]byte, error) {
	if adt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + adt.Time.Format(apiDateTimeLayout) + `"`), nil
}

// NewString は文字列へのポインタを返すヘルパー関数です。
// JSONのオプショナルな文字列フィールドなどで役立ちます。
func NewString(s string) *string {
	return &s
}
