package commands

import (
	"fmt"
	"errors"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/models" // models をインポート
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// ダミーの変数宣言で models パッケージが使用されていることを示す
var _ models.GetWeatherPointResponse

// runWeatherPointLogic は地点検索のコアロジックを担当します。
// 依存関係はインターフェースを通じて注入されます。
// この関数シグネチャは weather_point_test.go の想定と一致させます。
func runWeatherPointLogic(client ClientInterface, pres PresenterInterface, keyword string, kata bool) error {
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		// 404エラーの場合、APIクライアントは models.GetWeatherPointResponse{} とエラーを返す想定。
		// エラーハンドリングは呼び出し元か、より上位の層で行う。
		// ここではエラーをラップして返す。
		return fmt.Errorf("地域地点の検索に失敗しました: %w", err)
	}

	err = pres.PresentWeatherPoint(res, kata, keyword)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}
	return nil
}

// RunWeatherPoint は 'weather_point' コマンドの実行ロジック（アプリケーションサービス）です。
// cobra.Command から引数をパースし、コアロジック関数を呼び出します。
func RunWeatherPoint(apiClient *api.Client, actualPresenter presenter.Presenter, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("検索キーワードを指定してください")
	}
	keyword := args[0]
	kataFlag, _ := cmd.Flags().GetBool("kata")

	// apiClient は ClientInterface を満たし、actualPresenter は PresenterInterface を満たすと仮定
	// (pain_status.go と同様の前提)
	var pWrapper PresenterInterface
	// 型アサーションまたはアダプターパターンを使用
	// actualPresenter が presenter.Presenter 型で、PresenterInterface (PresentWeatherPoint を持つ) を満たすと仮定
	pWrapper, ok := actualPresenter.(PresenterInterface)
	if !ok {
		// 実際には、presenter.Presenter が PresenterInterface のメソッドセットを包含しているか、
		// 適切なアダプターを介して変換する必要がある。
		// ここでは、actualPresenter が PresenterInterface を満たさない場合にエラーとする。
		// ただし、pain_status.go で定義した PresenterInterface は PresentPainStatus と PresentWeatherPoint を持つため、
		// actualPresenter (presenter.Presenter型) がこれらのメソッドをすべて持っていればキャストは成功する。
		return fmt.Errorf("内部エラー: プレゼンターが期待されるインターフェースを満たしていません")
	}


	err := runWeatherPointLogic(apiClient, pWrapper, keyword, kataFlag)
	if err != nil {
		// 404 Not Found の場合の特別扱いをここで行うか検討
		// 例えば、エラーメッセージの内容で判断する
		// errors.As(err, &apiErr) && apiErr.StatusCode == 404 のようなチェックは
		// client.GetWeatherPoint が返すエラーの型に依存する。
		// runWeatherPointLogic から返るエラーは fmt.Errorf でラップされているため、
		// errors.Unwrap を使って元のエラーを取得する必要がある。
		originalErr := errors.Unwrap(err)
		var apiErr *api.APIError
		if errors.As(originalErr, &apiErr) && apiErr.StatusCode == 404 {
			// 404の場合はエラーメッセージを標準出力し、コマンドとしては成功扱い (nilを返す)
			// この動作はテストで別途検証するか、プレゼンターに責務を移すことを検討
			fmt.Printf("地域地点の検索に失敗しました: %s (該当する地点が見つかりませんでした)\n", originalErr.Error())
			return nil
		}
		return err // その他のエラーはそのまま返す
	}
	return nil
}
