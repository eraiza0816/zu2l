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
	// 1. 引数とフラグを解析・検証
	areaArg := args[0] // 地域コード(2桁数字)または地域名
	var areaCode string

	// 引数が数字のコードか、地域名かを判定
	if _, err := strconv.Atoi(areaArg); err == nil && len(areaArg) == 2 { // 2桁の数字であれば地域コードとみなす (基本的な検証)
		areaCode = areaArg
	} else {
		// 数字でなければ、地域名として AreaCodeMap でコードを検索
		code, ok := models.AreaCodeMap[areaArg]
		if !ok {
			// マップに存在しない場合は無効な引数
			// TODO: 有効な地域名/コードを提示するか、リスト表示することを検討
			return fmt.Errorf("無効な地域コードまたは地域名です: %s", areaArg)
		}
		areaCode = code // 見つかったコードを使用
	}

	// オプションの --set_weather_point フラグを取得
	setWeatherPointFlag, _ := cmd.Flags().GetString("set_weather_point")
	var setWeatherPoint *string // ポインタ型にして、フラグが指定されなかった場合に nil を渡せるようにする
	if setWeatherPointFlag != "" {
		// TODO: 必要であれば地点コードのフォーマット検証 (例: 5桁の数字) を追加
		setWeatherPoint = &setWeatherPointFlag // フラグが指定されていれば、その値のアドレスを渡す
	}

	// 2. APIクライアント (リポジトリ) を呼び出してデータを取得
	res, err := client.GetPainStatus(areaCode, setWeatherPoint)
	if err != nil {
		// APIクライアントからのエラーをそのまま返す
		return fmt.Errorf("痛み予報の取得に失敗しました: %w", err)
	}

	// 3. 取得したデータをプレゼンターに渡して表示
	err = pres.PresentPainStatus(res)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
