package commands

import (
	"fmt"
	"sort"
	"time"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/models"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// RunOtenkiAsp は 'otenki_asp' コマンドの実行ロジック（アプリケーションサービス）です。
func RunOtenkiAsp(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	cityArg := args[0]
	nFlag, _ := cmd.Flags().GetIntSlice("n")

	var cityCode string
	var cityName string
	var found bool
	if name, ok := models.ConfirmedOtenkiAspCityCodeMap[cityArg]; ok {
		cityCode = cityArg
		cityName = name
		found = true
	} else {
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			if name == cityArg {
				cityCode = code
				cityName = name
				found = true
				break
			}
		}
	}
	if !found {
		supportedValues := make([]string, 0, len(models.ConfirmedOtenkiAspCityCodeMap)*2)
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			supportedValues = append(supportedValues, code, name)
		}
		sort.Strings(supportedValues)
		return fmt.Errorf("無効な都市コードまたは都市名です: '%s' (サポートされている値: %v)", cityArg, supportedValues)
	}

	for _, n := range nFlag {
		// Otenki ASP は 0 (今日) から 6 (6日後) までをサポート
		if n < 0 || n > 6 {
			return fmt.Errorf("無効な日付オフセットです: %d (0 から 6 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag)

	res, err := client.GetOtenkiASP(cityCode)
	if err != nil {
		return fmt.Errorf("Otenki ASP データの取得に失敗しました: %w", err)
	}

	var availableDates []time.Time
	if len(res.Elements) > 0 && len(res.Elements[0].Records) > 0 {
		// 全ての要素が同じ日付キーを持つと仮定し、最初の要素から日付キーを取得
		for t := range res.Elements[0].Records {
			availableDates = append(availableDates, t)
		}
		sort.Slice(availableDates, func(i, j int) bool {
			return availableDates[i].Before(availableDates[j])
		})
	} else {
		fmt.Println("利用可能な日付データがありません。")
		return nil // 表示するものがないだけなので正常終了
	}

	targetDateIndices := make(map[int]bool)
	for _, n := range nFlag {
		targetDateIndices[n] = true
	}

	var targetDates []time.Time
	for i, date := range availableDates {
		if targetDateIndices[i] {
			targetDates = append(targetDates, date)
		}
	}

	// -n フラグが指定されなかった場合、または指定されたインデックスに対応する日付が
	// APIレスポンスに存在しなかった場合 (例: APIが7日分返さなかった)、targetDates は空になる可能性がある。
	// 現在の動作: targetDates が空の場合、プレゼンターはおそらく何も表示しないか、空のテーブルを表示する。
	// この動作を維持する。

	err = pres.PresentOtenkiASP(res, targetDates, cityName, cityCode)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
