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

	// フラグに基づいて適切なプレゼンターを作成するヘルパー関数
	getPresenter := func(cmd *cobra.Command) presenter.Presenter {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return &presenter.JSONPresenter{Writer: os.Stdout}
		}
		return &presenter.TablePresenter{Writer: os.Stdout}
	}

	rootCmd := &cobra.Command{
		Use:   "zutool",
		Short: "zutool <https://zutool.jp/> から情報を取得します",
		Long:  "zutool.jp から天気や痛み予報の情報を取得するコマンドラインツールです。",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	painStatusCommand := &cobra.Command{
		Use:     "pain_status [area_code]",
		Aliases: []string{"ps"},
		Short:   "都道府県別の痛み予報を取得します",
		Long:    "指定された都道府県コードの痛み予報を取得して表示します。",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd)
			return commands.RunPainStatus(apiClient, pres, cmd, args)
		},
	}
	painStatusCommand.Flags().StringP("set_weather_point", "s", "", "地点コード (例: '13113') を指定して地域固有の予報を取得")
	rootCmd.AddCommand(painStatusCommand)

	weatherPointCommand := &cobra.Command{
		Use:     "weather_point [keyword]",
		Aliases: []string{"wp"},
		Short:   "気象観測地点を検索します",
		Long:    "指定されたキーワード (例: 都市名) に基づいて気象観測地点を検索します。",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd)
			// RunWeatherPoint が --kata フラグにアクセスできるように cmd を渡す
			return commands.RunWeatherPoint(apiClient, pres, cmd, args)
		},
	}
	weatherPointCommand.Flags().BoolP("kata", "k", false, "出力テーブルにカタカナ名を含める")
	rootCmd.AddCommand(weatherPointCommand)

	weatherStatusCommand := &cobra.Command{
		Use:     "weather_status [city_code]",
		Aliases: []string{"ws"},
		Short:   "都市別の詳細な気象状況を取得します",
		Long:    "指定された都市コードの詳細な気象状況 (気温、気圧など) を取得して表示します。",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd)
			// RunWeatherStatus が --n フラグにアクセスできるように cmd を渡す
			return commands.RunWeatherStatus(apiClient, pres, cmd, args)
		},
	}
	weatherStatusCommand.Flags().IntSliceP("n", "n", []int{0}, "表示する日のオフセット番号 (-1 から 2) を指定 (複数指定可)")
	rootCmd.AddCommand(weatherStatusCommand)

	otenkiAspCommand := &cobra.Command{
		Use:     "otenki_asp [city_code]",
		Aliases: []string{"oa"},
		Short:   "Otenki ASP から気象情報を取得します",
		Long:    "特定の主要都市コードについて、Otenki ASP サービスから様々な天気予報要素 (天気、気温、風など) を取得します。",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pres := getPresenter(cmd)
			// RunOtenkiAsp が --n フラグにアクセスできるように cmd を渡す
			return commands.RunOtenkiAsp(apiClient, pres, cmd, args)
		},
	}
	otenkiAspCommand.Flags().IntSliceP("n", "n", []int{0, 1, 2, 3, 4, 5, 6}, "表示する予報日のオフセット番号 (0 から 6) を指定 (複数指定可)")
	rootCmd.AddCommand(otenkiAspCommand)

	rootCmd.PersistentFlags().BoolP("json", "j", false, "結果をJSON形式で出力する")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
