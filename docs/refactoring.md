# Zutool Go CLI リファクタリング計画

## 目的

現在のGo言語で実装されたZutool CLIアプリケーションのコードベースをリファクタリングし、保守性、可読性、テスト容易性を向上させる。

## リファクタリング方針

以下の点を中心にリファクタリングを実施する。

1.  **関心の分離**: コマンドロジック、APIクライアント、データモデル、表示ロジックを明確に分離する。
2.  **APIクライアントの改善**: APIリクエスト/レスポンス処理の簡略化とエラーハンドリングの統一。
3.  **表示ロジックの抽象化**: テーブル表示とJSON表示のロジックを分離・共通化する。
4.  **テストの拡充**: 各コンポーネントに対するユニットテストを追加・改善する。

## 具体的なタスク

### 1. パッケージ構成の見直し

*   以下のパッケージ構成を導入する。
    *   `cmd/zutool/`: `main.go` を配置（エントリーポイント）
    *   `internal/commands/`: 各サブコマンドのロジックを実装するパッケージ (`pain_status.go`, `weather_point.go` など)
    *   `internal/api/`: 既存の `api` パッケージの内容を配置し、改善を加える (`client.go`, `errors.go`, `parsers.go` など)
    *   `internal/models/`: 既存の `models` パッケージの内容を配置 (`types.go`, `constants.go` など)
    *   `internal/presenter/`: 表示ロジック（テーブル、JSON）を実装するパッケージ (`table.go`, `json.go`, `presenter.go`)
    *   `internal/config/`: 設定関連（将来的な拡張用、現状は不要かも）
    *   `pkg/`: 公開API（もしライブラリとして提供する場合。CLIのみなら不要）

### 2. `main.go` のスリム化 (`cmd/zutool/main.go`)

*   `main.go` は `cobra` のセットアップと各コマンドの登録のみを行うようにする。
*   各コマンドの `RunE` 関数は `internal/commands` パッケージ内の関数を呼び出すように変更する。
*   `newTable` や `renderWeatherStatusTable` などの表示関連ロジックを削除する。
*   グローバル変数 `jsonFlag` を削除し、各コマンドのハンドラにフラグ情報を渡すようにする。

### 3. コマンドロジックの分離 (`internal/commands/`)

*   `main.go` 内の `runPainStatus`, `runWeatherPoint`, `runWeatherStatus`, `runOtenkiAsp` 関数を `internal/commands` パッケージに移動する。
    *   例: `internal/commands/pain_status.go` に `RunPainStatus(cmd *cobra.Command, args []string) error` を実装。
*   各コマンド関数は以下の責務を持つようにする。
    1.  引数とフラグのパースとバリデーション。
    2.  `internal/api` のクライアントを呼び出してデータを取得。
    3.  取得したデータを `internal/presenter` に渡して結果を表示（JSONまたはテーブル）。
*   地域コード/都市コードと名前の変換ロジックを `internal/models` または `internal/commands` 内のヘルパー関数にまとめる。

### 4. APIクライアントの改善 (`internal/api/`)

*   `APIClient` 構造体を定義し、`baseURL`, `httpClient` をフィールドとして持つようにする。
    ```go
    package api

    import (
        "net/http"
        "time"
    )

    type Client struct {
        baseURL     string
        otenkiBaseURL string // Otenki ASP 用
        httpClient  *http.Client
    }

    func NewClient(baseURL, otenkiBaseURL string, timeout time.Duration) *Client {
        return &Client{
            baseURL:     baseURL,
            otenkiBaseURL: otenkiBaseURL,
            httpClient: &http.Client{
                Timeout: timeout,
            },
        }
    }
    // ... 他のメソッド
    ```
*   `doRequest` 関数を `Client` の非公開メソッド (`doRequest`) として整理する。
*   `_get` 関数を削除し、各APIエンドポイントに対応する公開メソッドを `Client` に実装する (`GetPainStatus`, `GetWeatherPoint`, `GetWeatherStatus`, `GetOtenkiASP`)。
*   `GetPainStatus` メソッド内で `setWeatherPoint` のAPI呼び出しを行うようにする。
*   `GetWeatherPoint` のネストされたJSONパース処理を、`internal/api/parsers.go` 内の非公開ヘルパー関数 (`parseWeatherPointResponse`) に分離する。
*   `GetOtenkiASP` のレスポンス変換処理を、`internal/api/parsers.go` 内の非公開ヘルパー関数 (`parseOtenkiASPResponse`) に分離する。
*   エラーハンドリングを統一する。`APIError` 型 (`internal/api/errors.go` に定義) を引き続き使用し、エラーの生成とラップの方法を一貫させる。200 OKレスポンス内のエラーチェックも継続する。
*   定数 (`baseURL`, `timeout`, `ua`) は基本的に非公開とし、`NewClient` 関数などで設定できるようにする。`UA` は `doRequest` 内でヘッダーに設定する。

### 5. データモデルの整理 (`internal/models/`)

*   構造体定義 (`WeatherPoint`, `GetPainStatusResponse` など) を `types.go` にまとめる。
*   定数マップ (`AreaCodeMap`, `WeatherEmojiMap` など) とEnum型定義を `constants.go` にまとめる。
*   `Validate` メソッドの必要性を再評価し、使用箇所で呼び出すか、不要なら削除する。`GetWeatherPointResponse` 内のコメントアウトは削除する。
*   `APIDateTime` のカスタムマーシャリングは現状維持。

### 6. 表示ロジックの抽象化 (`internal/presenter/`)

*   `Presenter` インターフェースを定義する (`internal/presenter/presenter.go`)。
    ```go
    package presenter

    import (
        "time" // time.Time を使うためインポート
        "github.com/eraiza0816/zu2l/internal/models"
    )

    type Presenter interface {
        PresentPainStatus(data models.GetPainStatusResponse) error
        PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error
        PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error
        PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error
    }
    ```
*   テーブル表示用の `TablePresenter` (`internal/presenter/table.go`) と JSON表示用の `JSONPresenter` (`internal/presenter/json.go`) を実装する。
*   `TablePresenter` は `tablewriter` を利用する。`newTable` の設定ロジックを共通化するヘルパー関数を作成する。各 `Present*` メソッド内でテーブルの構築とレンダリングを行う。
*   `JSONPresenter` は `encoding/json` を利用して整形済みJSONを出力する。
*   `commands` パッケージは、`jsonFlag` の値に応じて適切な `Presenter` (インターフェース経由で) を選択して使用する。

### 7. テストの追加・改善

*   `internal/api`: `net/http/httptest` や `httpmock` などを使用してHTTPリクエストをモックし、APIクライアントの各メソッドをテストする。エラーケース（ネットワークエラー、APIエラー、パースエラー）も網羅する。
*   `internal/commands`: `Presenter` と `APIClient` をモック（インターフェース化推奨）し、コマンドのロジック（引数処理、API呼び出し、Presenter呼び出し）をテストする。
*   `internal/presenter`: 各 `Present*` メソッドが期待される出力（テーブル文字列、JSON文字列）を生成するかテストする。出力結果を `golden file` として保存し比較する手法も有効。
*   `internal/models`: `Validate` メソッドやカスタムマーシャリング/アンマーシャリングをテストする。

## 段階的な進め方（推奨）

1.  新しいパッケージ構成を作成し、既存コードを対応するパッケージに移動する（ビルドエラーが出る状態でもOK）
2.  `internal/models` を整理し、ビルドが通るようにする
3.  `internal/api` のリファクタリング（`Client` 構造体導入、メソッド分割、パーサー分離）とテストを行う
4.  `internal/presenter` を実装し、テストを行う
5.  `internal/commands` を実装し、`api` と `presenter` を利用するように変更し、テストを行う
6.  `cmd/zutool/main.go` を修正し、`commands` を呼び出すようにする
7.  全体的な動作確認と統合テスト（手動または自動）
