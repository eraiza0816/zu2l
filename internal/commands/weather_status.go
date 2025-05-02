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
	cityCode := args[0]
	// TODO: Add validation for cityCode format (e.g., 3 digits)

	nFlag, _ := cmd.Flags().GetIntSlice("n")
	for _, n := range nFlag {
		if n < -1 || n > 2 {
			return fmt.Errorf("無効な日付オフセットです: %d (-1 から 2 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag)

	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("気象状況の取得に失敗しました: %w", err)
		}
		return fmt.Errorf("気象状況の取得中に予期せぬエラーが発生しました: %w", err)
	}

	for _, n := range nFlag {
		var dayName string
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

		err = pres.PresentWeatherStatus(res, n, dayName)
		if err != nil {
			// TODO: 特定の日の表示エラーが発生しても、他の日の表示を続けるか検討
			return fmt.Errorf("%s の結果表示に失敗しました: %w", dayName, err)
		}
	}

	return nil
}
