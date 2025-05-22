package commands

import (
	"fmt"
	"sort"
	// "errors" // errors パッケージは現在直接使用されていないためコメントアウト
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/models" // models をインポート
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// ダミーの変数宣言で models パッケージが使用されていることを示す (weather_point.go と同様)
var _ models.GetWeatherStatusResponse

// runWeatherStatusLogic は指定された cityCode と単一の dayOffset/dayName に対する
// 天気予報の取得と表示のコアロジックを担当します。
func runWeatherStatusLogic(client ClientInterface, pres PresenterInterface, cityCode string, dayOffset int, dayName string) error {
	// この関数内でAPI呼び出しを行う (テストの想定に合わせる)
	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		return fmt.Errorf("気象状況の取得に失敗しました (%s): %w", cityCode, err)
	}

	err = pres.PresentWeatherStatus(res, dayOffset, dayName)
	if err != nil {
		return fmt.Errorf("%s の結果表示に失敗しました: %w", dayName, err)
	}
	return nil
}

// RunWeatherStatus は 'weather_status' コマンドの実行ロジック（アプリケーションサービス）です。
func RunWeatherStatus(apiClient *api.Client, actualPresenter presenter.Presenter, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("都市コードを指定してください")
	}
	cityCode := args[0]
	// TODO: Add validation for cityCode format (e.g., 3 digits from ddd_doc, or 6 digits for zutool API?)

	nFlag, _ := cmd.Flags().GetIntSlice("n")
	if len(nFlag) == 0 {
		// -n が指定されなかった場合のデフォルト動作 (例: 今日のみ表示)
		// ここでは、-n が必須であるか、またはデフォルト値を設定するかを明確にする必要がある。
		// 元のコードでは -n がないとループが実行されない。
		// ユーザーフレンドリーにするならデフォルト値（例: 今日）を設定するか、エラーにする。
		// ここでは、少なくとも1つの -n が必要だと仮定し、なければエラーとするか、
		// あるいはデフォルトで今日 [0] を設定する。
		// 今回は、指定がなければ何もしない（元のコードに近い挙動）か、エラーとする。
		// わかりやすさのため、指定がない場合はエラーとする。
		return fmt.Errorf("-n オプションで表示する日付オフセットを指定してください (例: -n 0 で今日)")
	}

	for _, n := range nFlag {
		if n < -1 || n > 2 {
			return fmt.Errorf("無効な日付オフセットです: %d (-1 から 2 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag) // ユーザーが順不同で指定しても、昇順で処理する

	var pWrapper PresenterInterface
	pWrapper, ok := actualPresenter.(PresenterInterface)
	if !ok {
		return fmt.Errorf("内部エラー: プレゼンターが期待されるインターフェースを満たしていません")
	}

	// API呼び出しは各 runWeatherStatusLogic 内で行われるため、ここではレスポンスをキャッシュしない。
	// (テストの runWeatherStatusLogic の想定に合わせるため)
	// もしAPI呼び出しを1回にしたい場合は、runWeatherStatusLogic の責務を変える必要がある。

	for _, n := range nFlag {
		var dayName string
		// dayName の決定ロジックは presenter 側にある方が適切かもしれないが、
		// PresenterInterface のシグネチャに合わせるため、ここで決定する。
		// presenter.Presenter の PresentWeatherStatus が dayName を引数に取るため。
		switch n {
		case -1:
			dayName = "yesterday" // 実際の表示名に合わせて調整 (例: "昨日")
		case 0:
			dayName = "today"     // (例: "今日")
		case 1:
			dayName = "tomorrow"  // (例: "明日")
		case 2:
			dayName = "dayaftertomorrow" // (例: "明後日")
		default:
			// バリデーションにより、このケースには到達しないはず
			return fmt.Errorf("内部エラー: 無効な日付オフセット %d", n)
		}

		// runWeatherStatusLogic を呼び出す
		err := runWeatherStatusLogic(apiClient, pWrapper, cityCode, n, dayName)
		if err != nil {
			// TODO: 特定の日のエラーが発生しても、他の日の処理を続けるか、ここで全体をエラーとするか。
			// 現在は最初のエラーで全体が終了する。
			// 個々のエラーを収集して最後にまとめて報告するなども考えられる。
			// ユーザー体験としては、エラーがあっても処理を続けられる方が良い場合もある。
			// ここではシンプルに最初のエラーでリターンする。
			return err
		}
	}

	return nil
}
