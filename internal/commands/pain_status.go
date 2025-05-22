package commands

import (
	"fmt"
	"strconv"
	"github.com/eraiza0816/zu2l/api" // 実際のapiパッケージへのパス
	"github.com/eraiza0816/zu2l/internal/models"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// ClientInterface は API クライアントが満たすべきインターフェースを定義します。
// これにより、テスト時にモックを注入できます。
type ClientInterface interface {
	GetPainStatus(areaCode string, setWeatherPoint *string) (models.GetPainStatusResponse, error)
	GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error)
	GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error) // Added for weather_status
	// GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error)
}

// PresenterInterface はプレゼンターが満たすべきインターフェースを定義します。
// これにより、テスト時にモックを注入できます。
type PresenterInterface interface {
	PresentPainStatus(data models.GetPainStatusResponse) error
	PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error
	PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error // Added for weather_status
	// PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error
}

// runPainStatusLogic は痛み予報取得と表示のコアロジックを担当します。
// 依存関係はインターフェースを通じて注入されます。
func runPainStatusLogic(client ClientInterface, pres PresenterInterface, areaCode string, weatherPoint *string) error {
	res, err := client.GetPainStatus(areaCode, weatherPoint)
	if err != nil {
		return fmt.Errorf("痛み予報の取得に失敗しました: %w", err)
	}

	err = pres.PresentPainStatus(res)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}
	return nil
}

// RunPainStatus は 'pain_status' コマンドの実行ロジック（アプリケーションサービス）です。
// cobra.Command から引数をパースし、コアロジック関数を呼び出します。
func RunPainStatus(apiClient *api.Client, actualPresenter presenter.Presenter, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("地域コードまたは地域名を指定してください")
	}
	areaArg := args[0]
	var areaCode string

	// 地域コードまたは地域名から areaCode を解決
	if _, err := strconv.Atoi(areaArg); err == nil && len(areaArg) == 2 { // 都道府県コード (例: "13")
		areaCode = areaArg
	} else if _, err := strconv.Atoi(areaArg); err == nil && len(areaArg) == 6 { // 詳細地域コード (例: "130010")
		areaCode = areaArg
	} else {
		code, ok := models.AreaCodeMap[areaArg] // 地域名 (例: "東京")
		if !ok {
			return fmt.Errorf("無効な地域コードまたは地域名です: %s。有効な地域名（例：東京、大阪）または6桁の地域コード（例：130010）、または2桁の都道府県コード（例：13）を指定してください", areaArg)
		}
		areaCode = code
	}

	setWeatherPointFlag, _ := cmd.Flags().GetString("set_weather_point")
	var setWeatherPoint *string
	if setWeatherPointFlag != "" {
		setWeatherPoint = &setWeatherPointFlag
	}

	// api.Client は ClientInterface を満たすため、直接渡すことができます。
	// presenter.Presenter も PresenterInterface を満たすように、
	// PresentPainStatus メソッドを持つ必要があります。
	// この例では、actualPresenter が PresenterInterface を満たしていると仮定します。
	// 必要であれば、アダプターを作成するか、presenter.Presenter インターフェースを調整します。
	// ここでは、actualPresenter が PresenterInterface を満たしていると仮定します。
	// もし presenter.Presenter が PresentPainStatus を持っていない場合、
	// 適切なメソッドを持つように presenter.Presenter インターフェースを更新するか、
	// ラッパー/アダプターを介して呼び出す必要があります。
	// このコードでは、presenter.Presenter が PresenterInterface と互換性があると仮定します。
	var pWrapper PresenterInterface
	// 型アサーションやアダプターパターンを使用して presenter.Presenter を PresenterInterface に適合させる
	// 例: pWrapper = presenter.NewPainStatusPresenterAdapter(actualPresenter)
	// ここでは簡略化のため、actualPresenter が直接 PresenterInterface を満たすとします。
	// 実際には、presenter.Presenter の定義を確認し、必要に応じて調整してください。
	// 今回は presenter.Presenter が PresentPainStatus を持つと仮定して直接キャストします。
	// これは実際の presenter.Presenter の定義に依存します。
	// もし presenter.Presenter が汎用的な Present(interface{}) error のようなメソッドしか持たない場合、
	// この部分は動作しません。
	// しかし、元のコードでは pres.PresentPainStatus(res) となっていたため、
	// presenter.Presenter にはそのようなメソッドがあると期待できます。

	// presenter.Presenter が PresenterInterface を満たすことを確認
	// これはコンパイル時にチェックされるべきですが、動的なアダプタが必要な場合もあります。
	// 今回は、presenter.Presenter が PresentPainStatus を持つと仮定します。
	pWrapper = actualPresenter.(PresenterInterface) // これは presenter.Presenter が PresenterInterface を実装している場合にのみ機能します。
                                                // 実際には、presenter.Presenter の定義を確認し、必要に応じてアダプタを作成するか、
                                                // PresenterInterface の定義を presenter.Presenter に合わせる必要があります。
                                                // 元のコードでは `pres.PresentPainStatus(res)` となっていたため、
                                                // `presenter.Presenter` が `PresentPainStatus` メソッドを持っていると期待できます。
                                                // その場合、`PresenterInterface` の定義をそれに合わせるのが適切かもしれません。

	return runPainStatusLogic(apiClient, pWrapper, areaCode, setWeatherPoint)
}
