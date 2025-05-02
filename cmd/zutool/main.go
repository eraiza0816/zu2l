package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/eraiza0816/zu2l/api"
	"github.com/eraiza0816/zu2l/internal/commands"
	"github.com/eraiza0816/zu2l/internal/presenter"
)

func main() {
	// APIクライアントを一度だけインスタンス化
	// 現状はURLとタイムアウトにデフォルト値を使用
	// TODO: 将来的にフラグや設定ファイルで設定可能にすることを検討
	apiClient := api.NewClient("", "", 0)

	// フラグに基づいて適切なプレゼンターを作成するヘルパー関数を定義
	getPresenter := func(cmd *cobra.Command) presenter.Presenter {
		jsonOutput, _ := cmd.Flags().GetBool("json") // --json フラグの値を取得
		if jsonOutput {
			// --json が指定されていれば JSONPresenter を使用
			return &presenter.JSONPresenter{Writer: os.Stdout}
		}
		// デフォルトは TablePresenter を使用
		return &presenter.TablePresenter{Writer: os.Stdout}
	}

	// ルートコマンドの定義
	rootCmd := &cobra.Command{
		Use:   "zutool",
		Short: "zutool <https://zutool.jp/> から情報を取得します",
		Long:  "zutool.jp から天気や痛み予報の情報を取得するコマンドラインツールです。",
		// ルートコマンド自体が実行された場合はヘルプを表示
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// pain_status サブコマンドの定義
	painStatusCommand := &cobra.Command{
		Use:     "pain_status [area_code]", // area_code は必須引数
		Aliases: []string{"ps"},            // エイリアス "ps"
		Short:   "都道府県別の痛み予報を取得します",
		Long:    "指定された都道府県コードの痛み予報を取得して表示します。",
		Args:    cobra.ExactArgs(1), // 引数がちょうど1つであることを要求
		// コマンド実行時の処理 (エラーを返す可能性があるため RunE を使用)
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd) // フラグに基づいてプレゼンターを取得
			// internal/commands のハンドラ関数を呼び出す
			return commands.RunPainStatus(apiClient, pres, cmd, args)
		},
	}
	// pain_status コマンドに --set_weather_point フラグを追加
	painStatusCommand.Flags().StringP("set_weather_point", "s", "", "地点コード (例: '13113') を指定して地域固有の予報を取得")
	rootCmd.AddCommand(painStatusCommand) // ルートコマンドにサブコマンドを追加

	// weather_point サブコマンドの定義
	weatherPointCommand := &cobra.Command{
		Use:     "weather_point [keyword]", // keyword は必須引数
		Aliases: []string{"wp"},            // エイリアス "wp"
		Short:   "気象観測地点を検索します",
		Long:    "指定されたキーワード (例: 都市名) に基づいて気象観測地点を検索します。",
		Args:    cobra.ExactArgs(1), // 引数がちょうど1つであることを要求
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd) // フラグに基づいてプレゼンターを取得
			// RunWeatherPoint が --kata フラグにアクセスできるように cmd を渡す
			return commands.RunWeatherPoint(apiClient, pres, cmd, args)
		},
	}
	// weather_point コマンドに --kata フラグを追加
	weatherPointCommand.Flags().BoolP("kata", "k", false, "出力テーブルにカタカナ名を含める")
	rootCmd.AddCommand(weatherPointCommand) // ルートコマンドにサブコマンドを追加

	// weather_status サブコマンドの定義
	weatherStatusCommand := &cobra.Command{
		Use:     "weather_status [city_code]", // city_code は必須引数
		Aliases: []string{"ws"},               // エイリアス "ws"
		Short:   "都市別の詳細な気象状況を取得します",
		Long:    "指定された都市コードの詳細な気象状況 (気温、気圧など) を取得して表示します。",
		Args:    cobra.ExactArgs(1), // 引数がちょうど1つであることを要求
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd) // フラグに基づいてプレゼンターを取得
			// RunWeatherStatus が --n フラグにアクセスできるように cmd を渡す
			return commands.RunWeatherStatus(apiClient, pres, cmd, args)
		},
	}
	// weather_status コマンドに --n フラグを追加
	weatherStatusCommand.Flags().IntSliceP("n", "n", []int{0}, "表示する日のオフセット番号 (-1 から 2) を指定 (複数指定可)")
	rootCmd.AddCommand(weatherStatusCommand) // ルートコマンドにサブコマンドを追加

	// otenki_asp サブコマンドの定義
	otenkiAspCommand := &cobra.Command{
		Use:     "otenki_asp [city_code]", // city_code は必須引数
		Aliases: []string{"oa"},           // エイリアス "oa"
		Short:   "Otenki ASP から気象情報を取得します",
		Long:    "特定の主要都市コードについて、Otenki ASP サービスから様々な天気予報要素 (天気、気温、風など) を取得します。",
		Args:    cobra.ExactArgs(1), // 引数がちょうど1つであることを要求
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd) // フラグに基づいてプレゼンターを取得
			// RunOtenkiAsp が --n フラグにアクセスできるように cmd を渡す
			return commands.RunOtenkiAsp(apiClient, pres, cmd, args)
		},
	}
	// otenki_asp コマンドに --n フラグを追加
	otenkiAspCommand.Flags().IntSliceP("n", "n", []int{0, 1, 2, 3, 4, 5, 6}, "表示する予報日のオフセット番号 (0 から 6) を指定 (複数指定可)")
	rootCmd.AddCommand(otenkiAspCommand) // ルートコマンドにサブコマンドを追加

	// --json フラグを全てのコマンドで利用可能な永続フラグとして定義
	rootCmd.PersistentFlags().BoolP("json", "j", false, "結果をJSON形式で出力する")

	// ルートコマンドを実行
	if err := rootCmd.Execute(); err != nil {
		// Cobra が既にエラーを出力するため、ここでは何もしない
		os.Exit(1) // エラーコード 1 で終了
	}
}
