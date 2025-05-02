package presenter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"

	"github.com/olekukonko/tablewriter"
)

// TablePresenter は Presenter インターフェースを実装し、データをテーブル形式で出力します。
type TablePresenter struct {
	// Writer は出力先を指定します。nil の場合は os.Stdout がデフォルトで使用されます。
	Writer io.Writer
}

// ensureWriter は設定された Writer を返します。nil の場合は os.Stdout を返します。
func (p *TablePresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

// newTable は新しい tablewriter.Table インスタンスを作成し、設定します。
// これは元の main.go のロジックに基づいたヘルパー関数です。
func (p *TablePresenter) newTable() *tablewriter.Table {
	writer := p.ensureWriter()
	table := tablewriter.NewWriter(writer)
	table.SetAutoWrapText(false)       // テキストの自動折り返しを無効化
	table.SetAutoFormatHeaders(true)   // ヘッダーの自動フォーマットを有効化
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT) // ヘッダーの左寄せ
	table.SetAlignment(tablewriter.ALIGN_LEFT)     // セルの左寄せ
	table.SetCenterSeparator("")       // 中央の区切り文字をなしに
	table.SetColumnSeparator("")       // 列の区切り文字をなしに
	table.SetRowSeparator("")          // 行の区切り文字をなしに
	table.SetHeaderLine(false)         // ヘッダー下の線を非表示
	table.SetBorder(false)             // ボーダーを非表示
	table.SetTablePadding("\t")        // テーブルのパディングをタブに
	table.SetNoWhiteSpace(true)        // 不要な空白を削除
	return table
}

// PresentPainStatus は痛み予報ステータスをテーブル形式で表示します。
func (p *TablePresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	status := data.PainnoterateStatus
	// テーブルのタイトルを設定
	titleA := fmt.Sprintf("今のみんなの体調は? <%s>", status.AreaName)
	titleB := fmt.Sprintf("(集計時間: %s時-%s時台)", status.TimeStart, status.TimeEnd)

	table := p.newTable()
	table.SetHeader([]string{fmt.Sprintf("%s\n%s", titleA, titleB)}) // ヘッダーにタイトルを設定

	// 体調レベルに対応する絵文字とラベル、APIからの割合データ
	sicknessEmojis := []string{"😃", "😐", "😞", "🤯"}
	sicknessLabels := []string{"普通", "少し痛い", "痛い", "かなり痛い"}
	rates := []float64{status.RateNormal, status.RateLittle, status.RatePainful, status.RateBad}

	// 各体調レベルについて行を追加
	for i, emoji := range sicknessEmojis {
		rate := rates[i]
		// 割合が非負であることを確認してから絵文字を繰り返す
		repeatCount := 0
		if rate >= 0 {
			// 絵文字の繰り返し回数を割合に基づいて単純にスケーリング (rate / 2)
			repeatCount = int(rate / 2)
		}
		emojiRepeat := strings.Repeat(emoji, repeatCount)
		// 絵文字の繰り返しと割合(%)を表示
		table.Append([]string{fmt.Sprintf("%s %.0f%%", emojiRepeat, rate)})
	}

	// フッターに凡例を追加
	var footerParts []string
	for i, emoji := range sicknessEmojis {
		footerParts = append(footerParts, fmt.Sprintf("%s･･･%s", emoji, sicknessLabels[i]))
	}
	footerText := fmt.Sprintf("[%s]", strings.Join(footerParts, ", "))
	table.SetFooter([]string{footerText}) // 凡例のために SetFooter を使用

	table.Render() // テーブルを描画
	return nil
}

// PresentWeatherPoint は地点検索結果をテーブル形式で表示します。
func (p *TablePresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	table := p.newTable()
	headers := []string{"地域コード", "地域名"}
	// kata フラグが true の場合、ヘッダーに "地域カナ" を追加
	if kata {
		headers = append(headers, "地域カナ")
	}
	table.SetHeader(headers)
	// キャプションに検索キーワードを含める
	table.SetCaption(true, fmt.Sprintf("「%s」の検索結果", keyword))

	// 検索結果がない場合の処理
	if len(data.Result.Root) == 0 {
		fmt.Fprintf(p.ensureWriter(), "「%s」に一致する地域が見つかりませんでした。\n", keyword)
		return nil
	}

	// 各地点情報をテーブルの行として追加
	for _, weatherPoint := range data.Result.Root {
		row := []string{weatherPoint.CityCode, weatherPoint.Name}
		// kata フラグが true の場合、行にカタカナ名を追加
		if kata {
			row = append(row, weatherPoint.NameKata)
		}
		table.Append(row)
	}
	table.Render() // テーブルを描画
	return nil
}

// renderWeatherStatusSubTable は気象状況の12時間分のセグメントを描画します。
// これは元の main.go のロジックに基づいたヘルパー関数です。
// prevPressure は前の12時間セグメントの最後の気圧を保持し、矢印表示に使用します。
// 最後の気圧を返します。
func (p *TablePresenter) renderWeatherStatusSubTable(
	dayData []models.WeatherStatusByTime, // 表示する日の時間別データ
	startHour int,                        // このセグメントの開始時間 (0 または 12)
	prevPressure float64,                 // 前のセグメントの最後の気圧 (矢印表示用)
	title string,                         // テーブルのタイトル (最初のセグメントのみ)
) float64 { // このセグメントの最後の気圧を返す
	table := p.newTable()
	// タイトルが指定されていればキャプションとして設定
	if title != "" {
		table.SetCaption(true, title)
	}

	var headers []string
	dataLen := len(dayData)
	// データがない場合はメッセージを表示して終了
	if dataLen == 0 {
		fmt.Fprintf(p.ensureWriter(), "データがありません (%d時台)\n", startHour)
		return prevPressure // データがない場合は前の気圧をそのまま返す
	}

	// 12時間未満のデータしかなくても処理できるようにループを調整
	numHours := min(12, dataLen)
	// ヘッダー（時間）を作成
	for i := 0; i < numHours; i++ {
		headers = append(headers, strconv.Itoa(startHour+i))
	}
	table.SetHeader(headers)

	// 各行のデータを保持するスライス
	weathers := make([]string, numHours)
	temps := make([]string, numHours)
	pressures := make([]string, numHours)
	pressureLevels := make([]string, numHours)

	lastPressure := prevPressure // このセグメント内の最後の気圧を追跡
	for i := 0; i < numHours; i++ {
		byTime := dayData[i] // 現在の時間帯のデータ

		// 天気: 簡略化された天気コード (最初の桁) を使用して絵文字にマッピング
		weatherCodeInt, err := strconv.Atoi(string(byTime.Weather))
		emoji := "?" // デフォルトは "?"
		if err == nil {
			simplifiedCode := (weatherCodeInt / 100) * 100 // 100, 200, 300, 400 に丸める
			if e, ok := models.WeatherEmojiMap[simplifiedCode]; ok {
				emoji = e // マップにあれば絵文字を使用
			}
		}
		weathers[i] = emoji

		// 気温: 文字列から float に変換して表示
		if byTime.Temp != nil {
			tempFloat, err := strconv.ParseFloat(*byTime.Temp, 64)
			if err != nil {
				// 変換エラー処理
				temps[i] = "?℃" // エラーを示す
				fmt.Fprintf(os.Stderr, "警告: 温度 '%s' の数値変換に失敗しました: %v\n", *byTime.Temp, err)
			} else {
				temps[i] = fmt.Sprintf("%.1f℃", tempFloat) // 小数点以下1桁で表示
			}
		} else {
			temps[i] = "-℃" // 気温データがないことを示す
		}

		// 気圧: 文字列から float64 に変換して比較と表示に使用
		pressureFloat, err := strconv.ParseFloat(byTime.Pressure, 64)
		if err != nil {
			// 変換エラー処理 (例: ログ出力、デフォルト値使用)
			// 現状は 0 にデフォルト設定し、変換失敗時は矢印ロジックをスキップ
			pressureFloat = 0 // または lastPressure を使用？
			fmt.Fprintf(os.Stderr, "警告: 気圧 '%s' の数値変換に失敗しました: %v\n", byTime.Pressure, err)
		}

		// 気圧の変化を示す矢印を設定
		var arrow string
		// 初回ケースまたは変換エラーの場合
		if (lastPressure == 0 && i == 0 && startHour == 0) || err != nil {
			arrow = "→" // または ""
		} else if pressureFloat > lastPressure {
			arrow = "↗" // 上昇
		} else if pressureFloat < lastPressure {
			arrow = "↘" // 下降
		} else {
			arrow = "→" // 変化なし
		}
		// 変換後の float を使用して表示 (矢印 + 改行 + 気圧値)
		pressures[i] = fmt.Sprintf("%s\n%.1f", arrow, pressureFloat)
		// lastPressure を変換後の float で更新
		lastPressure = pressureFloat

		// 気圧レベル: 必要に応じて PressureLevelEnum をより説明的な表現にマッピングするか、そのまま使用
		pressureLevels[i] = string(byTime.PressureLevel) // 現状はコードをそのまま使用
	}

	// 各行のデータをテーブルに追加
	table.Append(weathers)
	table.Append(temps)
	table.Append(pressures)
	table.Append(pressureLevels)
	table.Render() // このサブテーブルを描画

	return lastPressure // このセグメントの最後の気圧を返す
}

// PresentWeatherStatus は詳細な気象状況をテーブル形式 (12時間ごとのセグメント) で表示します。
func (p *TablePresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	var dayData []models.WeatherStatusByTime
	// コマンドロジックから渡された dayName に基づいて表示する日のデータを選択
	switch dayName {
	case "yesterday":
		dayData = data.Yesterday
	case "today":
		dayData = data.Today
	case "tomorrow":
		dayData = data.Tomorrow
	case "dayaftertomorrow":
		dayData = data.DayAfterTomorrow
	default:
		return fmt.Errorf("無効な dayName が提供されました: %s", dayName)
	}

	// 選択された日のデータがない場合の処理
	if len(dayData) == 0 {
		fmt.Fprintf(p.ensureWriter(), "%s のデータがありません。\n", dayName)
		return nil
	}
	// データが24時間分ない場合に警告を表示
	if len(dayData) < 24 {
		fmt.Fprintf(p.ensureWriter(), "警告: %s のデータが24時間分ありません (%d時間分)。利用可能なデータを表示します。\n", dayName, len(dayData))
	}

	// APIレスポンスの DateTime とオフセットに基づいて表示日付文字列を計算
	displayDate := data.DateTime.AddDate(0, 0, dayOffset).Format("2006-01-02") // YYYY-MM-DD 形式
	// 最初のサブテーブルのタイトルを作成
	title := fmt.Sprintf("<%s|%s>の気圧予報\n%s = %s",
		data.PlaceName, data.PlaceID, dayName, displayDate)

	// 最初の12時間 (0-11時) を描画
	// renderWeatherStatusSubTable を呼び出し、最後の気圧を prevPressure に保存
	prevPressure := p.renderWeatherStatusSubTable(dayData[0:min(12, len(dayData))], 0, 0, title)

	// データが12時間分以上存在する場合、次の12時間 (12-23時) を描画
	if len(dayData) > 12 {
		// 前のセグメントの最後の気圧 (prevPressure) を渡して呼び出し (タイトルはなし)
		p.renderWeatherStatusSubTable(dayData[12:min(24, len(dayData))], 12, prevPressure, "")
	}

	return nil
}

// PresentOtenkiASP は Otenki ASP の気象情報を元の複数列形式で表示します。
// ユーザーの理想的なフォーマットに基づいて手動でヘッダー行を出力します。
func (p *TablePresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	table := p.newTable() // テーブル設定は維持
	// キャプションを設定
	table.SetCaption(true, fmt.Sprintf("<%s|%s>の天気情報", cityName, cityCode))

	// 表示する要素がない場合はメッセージを表示して終了
	if len(data.Elements) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示する天気情報要素がありません。\n")
		return nil
	}

	// --- 成功した複数列表示状態からのデータ処理ロジック ---
	// 要素のソート (オプション、コメントアウトを維持)
	// sort.Slice(data.Elements, func(i, j int) bool { ... })

	// 表示対象の日付 (targetDates) をソート
	sort.Slice(targetDates, func(i, j int) bool {
		return targetDates[i].Before(targetDates[j])
	})

	// 表示対象の日付がない場合はメッセージを表示して終了
	if len(targetDates) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示対象の日付が指定されていません。\n")
		return nil
	}

	// tablewriter のヘッダーではなく、理想的なヘッダー文字列を手動で出力
	idealHeaderString := "日付\t天気\t降水確率\t最高気温\t最低気温\t最大風速\t最大風速時風向\t気圧予報レベル\t最小湿度"
	fmt.Fprintln(p.ensureWriter(), idealHeaderString)

	// 各対象日について行を作成
	for _, targetDate := range targetDates {
		// 最初の列は日付 (MM/DD形式)
		row := []string{targetDate.Format("01/02")}

		// 各要素 (天気、降水確率など) をループして列の値を取得
		// これらが日付の後ろの列になる
		for _, element := range data.Elements {
			value, ok := element.Records[targetDate] // 対象日のデータを取得
			valueStr := "-" // データがない場合のデフォルト値
			if ok {
				// パーサーが正しい型 (string または float64) を提供することを前提としたフォーマットロジック
				switch v := value.(type) {
				case string:
					// 天気コードの絵文字マッピングを試みる
					if element.ContentID == "day_tenki" { // 特定の ContentID をチェック
						weatherCodeInt, err := strconv.Atoi(v)
						if err == nil {
							simplifiedCode := (weatherCodeInt / 100) * 100 // 100の位に丸める
							if emoji, okEmoji := models.WeatherEmojiMap[simplifiedCode]; okEmoji {
								valueStr = emoji // 絵文字を使用
							} else {
								valueStr = v // 絵文字がなければコードを使用
							}
						} else {
							valueStr = v // 整数コードでない場合 (例: 天気テキスト) は元の文字列を使用
						}
					} else {
						// 他の要素 (風向、レベルなど) は文字列を直接使用
						valueStr = v
					}
				case float64:
					// float のフォーマット (降水確率, 気温, 風速, 湿度)
					// 降水確率, 湿度, レベル, 風向が小数部分を持たない場合は整数としてフォーマット
					if element.ContentID == "day_pre" || element.ContentID == "low_humidity" || element.ContentID == "zutu_level_day" || element.ContentID == "day_wind_d" {
						if v == float64(int(v)) { // 整数かどうかチェック
							valueStr = strconv.Itoa(int(v))
						} else {
							valueStr = fmt.Sprintf("%.1f", v) // 小数点がある場合は維持
						}
					} else {
						// 気温、風速は小数点以下1桁でフォーマット
						valueStr = fmt.Sprintf("%.1f", v)
					}
				default:
					// パーサーからの予期しない型に対するフォールバック
					fmt.Fprintf(os.Stderr, "警告: ContentID '%s' の予期しない値の型: %T, 値: %v\n", element.ContentID, v, v)
					valueStr = fmt.Sprintf("%v", v) // そのまま文字列化
				}
			}
			// 処理した値を列として追加
			row = append(row, valueStr)
		}
		// 完成した行をテーブルに追加
		table.Append(row)
	}

	// テーブルを描画 (手動でヘッダーを出力したので、テーブル自体のヘッダーはなし)
	table.Render()
	return nil
}

// min ヘルパー関数 (既に存在しないかインポートされていない場合)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// コンパイル時チェック: TablePresenter が Presenter インターフェースを実装していることを保証します。
var _ Presenter = (*TablePresenter)(nil)
