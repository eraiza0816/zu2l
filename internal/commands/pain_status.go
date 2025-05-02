package commands

import (
	"fmt"
	"strconv"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/models"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// RunPainStatus は 'pain_status' コマンドの実行ロジック（アプリケーションサービス）です。
// APIクライアントを使用して痛み予報データを取得し、プレゼンターを使用して表示します。
func RunPainStatus(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	areaArg := args[0]
	var areaCode string

	if _, err := strconv.Atoi(areaArg); err == nil && len(areaArg) == 2 {
		areaCode = areaArg
	} else {
		code, ok := models.AreaCodeMap[areaArg]
		if !ok {
			// TODO: 有効な地域名/コードを提示するか、リスト表示することを検討
			return fmt.Errorf("無効な地域コードまたは地域名です: %s", areaArg)
		}
		areaCode = code
	}

	setWeatherPointFlag, _ := cmd.Flags().GetString("set_weather_point")
	var setWeatherPoint *string
	if setWeatherPointFlag != "" {
		// TODO: 必要であれば地点コードのフォーマット検証 (例: 5桁の数字) を追加
		setWeatherPoint = &setWeatherPointFlag
	}

	res, err := client.GetPainStatus(areaCode, setWeatherPoint)
	if err != nil {
		return fmt.Errorf("痛み予報の取得に失敗しました: %w", err)
	}

	err = pres.PresentPainStatus(res)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
