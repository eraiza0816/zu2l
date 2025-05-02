package commands

import (
	"fmt"
	"errors"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// RunWeatherPoint は 'weather_point' コマンドの実行ロジック（アプリケーションサービス）です。
func RunWeatherPoint(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. 引数とフラグを解析
	keyword := args[0] // 検索キーワード (地名など)
	kataFlag, _ := cmd.Flags().GetBool("kata") // --kata フラグの値を取得

	// 2. APIクライアント (リポジトリ) を呼び出して地点情報を検索
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		// APIエラーハンドリング (例: キーワードが見つからない場合など)
		var apiErr *api.APIError
		// エラーが APIError で、かつステータスコードが 404 (Not Found) かどうかを確認
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// 404 の場合、ユーザーに通知するだけでよいか？
			// または特定のエラーメッセージを返すか？ 現状は API エラーを返す。
			// プレゼンターに専用メソッドがないため、一旦メッセージを直接表示してみる。
			// TODO: プレゼンターインターフェースに汎用的な `PresentError` や `PresentMessage` の追加を検討。
			fmt.Printf("地域地点の検索に失敗しました: %s\n", err.Error()) // エラーを直接表示
			// nil を返して、cobra が再度エラーを表示しないようにする
			return nil
		}
		// その他のエラーの場合はラップして返す
		return fmt.Errorf("地域地点の検索中にエラーが発生しました: %w", err)
	}

	// 3. 取得したデータをプレゼンターに渡して表示
	err = pres.PresentWeatherPoint(res, kataFlag, keyword)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
