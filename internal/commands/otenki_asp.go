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
	// 1. 引数とフラグを解析
	cityArg := args[0] // 都市コードまたは都市名
	nFlag, _ := cmd.Flags().GetIntSlice("n") // 表示する日付オフセット (0:今日, 1:明日, ...)

	// 都市コード/都市名のバリデーションと特定
	var cityCode string
	var cityName string
	var found bool
	// まず、引数が都市コードとしてマップに存在するか確認
	if name, ok := models.ConfirmedOtenkiAspCityCodeMap[cityArg]; ok {
		cityCode = cityArg
		cityName = name
		found = true
	} else {
		// 都市コードでない場合、都市名としてマップに存在するか確認
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			if name == cityArg {
				cityCode = code
				cityName = name
				found = true
				break
			}
		}
	}
	// 見つからなかった場合はエラー
	if !found {
		// サポートされている値をリストアップしてエラーメッセージに含める
		supportedValues := make([]string, 0, len(models.ConfirmedOtenkiAspCityCodeMap)*2)
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			supportedValues = append(supportedValues, code, name)
		}
		sort.Strings(supportedValues) // 見やすいようにソート
		return fmt.Errorf("無効な都市コードまたは都市名です: '%s' (サポートされている値: %v)", cityArg, supportedValues)
	}

	// -n フラグ (日付オフセット) のバリデーション
	for _, n := range nFlag {
		// Otenki ASP は 0 (今日) から 6 (6日後) までをサポート
		if n < 0 || n > 6 {
			return fmt.Errorf("無効な日付オフセットです: %d (0 から 6 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag) // 要求された日付をソート

	// 2. APIクライアント (リポジトリ) を呼び出して Otenki ASP データを取得
	res, err := client.GetOtenkiASP(cityCode)
	if err != nil {
		// API呼び出しエラーハンドリング
		return fmt.Errorf("Otenki ASP データの取得に失敗しました: %w", err)
	}

	// 3. フラグと利用可能なデータに基づいて表示対象の日付をフィルタリング
	var availableDates []time.Time
	// APIレスポンスにデータが存在するか確認
	if len(res.Elements) > 0 && len(res.Elements[0].Records) > 0 {
		// 全ての要素が同じ日付キーを持つと仮定し、最初の要素から日付キーを取得
		for t := range res.Elements[0].Records {
			availableDates = append(availableDates, t)
		}
		// 日付を時系列順にソート
		sort.Slice(availableDates, func(i, j int) bool {
			return availableDates[i].Before(availableDates[j])
		})
	} else {
		// 利用可能な日付データがない場合の処理
		fmt.Println("利用可能な日付データがありません。") // ユーザーに通知
		return nil // エラーではなく、正常終了とする (表示するものがないだけ)
	}

	// -n フラグで指定されたインデックスをマップに記録
	targetDateIndices := make(map[int]bool)
	for _, n := range nFlag {
		targetDateIndices[n] = true
	}

	// 利用可能な日付リスト (availableDates) から、フラグで指定されたインデックスの日付を選択
	var targetDates []time.Time
	for i, date := range availableDates {
		if targetDateIndices[i] { // i がフラグで指定されたインデックスと一致する場合
			targetDates = append(targetDates, date)
		}
	}

	// -n フラグが指定されなかった場合、または指定されたインデックスに対応する日付が
	// APIレスポンスに存在しなかった場合 (例: APIが7日分返さなかった)、targetDates は空になる可能性がある。
	// 現在の動作: targetDates が空の場合、プレゼンターはおそらく何も表示しないか、空のテーブルを表示する。
	// この動作を維持する。

	// 4. データをプレゼンターに渡して表示
	err = pres.PresentOtenkiASP(res, targetDates, cityName, cityCode)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
