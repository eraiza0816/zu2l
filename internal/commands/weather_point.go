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
	keyword := args[0]
	kataFlag, _ := cmd.Flags().GetBool("kata")

	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		var apiErr *api.APIError
		// 404 Not Found の場合は特別扱い
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// TODO: プレゼンターインターフェースに汎用的な `PresentError` や `PresentMessage` の追加を検討。
			fmt.Printf("地域地点の検索に失敗しました: %s\n", err.Error())
			// nil を返して、cobra が再度エラーを表示しないようにする
			return nil
		}
		return fmt.Errorf("地域地点の検索中にエラーが発生しました: %w", err)
	}

	err = pres.PresentWeatherPoint(res, kataFlag, keyword)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
