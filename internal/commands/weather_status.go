package commands

import (
	"fmt"
	"sort"
	"errors"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// RunWeatherStatus executes the logic for the 'weather_status' command.
func RunWeatherStatus(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. Parse arguments and flags
	cityCode := args[0]
	// TODO: Add validation for cityCode format (e.g., 3 digits)

	nFlag, _ := cmd.Flags().GetIntSlice("n")
	// Validate the -n flag values
	for _, n := range nFlag {
		if n < -1 || n > 2 {
			return fmt.Errorf("無効な日付オフセットです: %d (-1 から 2 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag) // Sort the days to display them chronologically

	// 2. APIクライアントを呼び出して気象状況を取得
	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		// APIエラーかどうかを判定
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			// APIエラーの場合は、エラー内容をそのまま返す
			return fmt.Errorf("気象状況の取得に失敗しました: %w", err)
		}
		// その他のエラーの場合は、予期せぬエラーとしてラップして返す
		return fmt.Errorf("気象状況の取得中に予期せぬエラーが発生しました: %w", err)
	}

	// 3. 指定された日付オフセットごとに結果を表示
	for _, n := range nFlag {
		var dayName string
		// 日付オフセット(n)に応じて表示用の日付名を設定
		switch n {
		case -1:
			dayName = "yesterday"
		case 0:
			dayName = "today"
		case 1:
			dayName = "tomorrow"
		case 2:
			dayName = "dayaftertomorrow"
		default:
			// バリデーションにより、このケースには到達しないはず
			return fmt.Errorf("内部エラー: 無効な日付オフセット %d", n)
		}

		// プレゼンターを呼び出して、該当日の気象状況を表示
		err = pres.PresentWeatherStatus(res, n, dayName)
		if err != nil {
			// 表示中にエラーが発生した場合、そのエラーを返す
			// TODO: 特定の日の表示エラーが発生しても、他の日の表示を続けるか検討
			return fmt.Errorf("%s の結果表示に失敗しました: %w", dayName, err)
		}
	}

	return nil
}
