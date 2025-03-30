## DDDモデル (Go版)

このドキュメントは、`zutool` Go アプリケーションにおけるドメイン駆動設計 (DDD) の要素を記述します。

*   **エンティティ (Entities)**: 識別子を持ち、状態が変化するオブジェクト。
    *   `WeatherPoint` (`internal/models/types.go`): 天気予報の地点情報。`CityCode` が識別子となりうる。
    *   `GetPainStatus` (`internal/models/types.go`): 特定エリア・時間の痛み予報ステータス。エリアと時間が複合的な識別子となりうる。
    *   `WeatherStatusByTime` (`internal/models/types.go`): 特定地点・時間の天気情報。地点と時間が複合的な識別子となりうる。
    *   `Element` (`internal/models/types.go`): Otenki ASP API から取得した特定のコンテンツ要素（例: 天気、気温）。`ContentID` が識別子となりうる。

*   **値オブジェクト (Value Objects)**: 識別子を持たず、属性によって定義されるオブジェクト。不変であることが多い。
    *   `APIDateTime` (`internal/models/models.go`): API 特有の "YYYY-MM-DD HH" 形式の日時。
    *   `AreaEnum` (`internal/models/constants.go`): 都道府県コードを表す Enum。
    *   `PressureLevelEnum` (`internal/models/constants.go`): 気圧レベルを表す Enum。
    *   `WeatherEnum` (`internal/models/constants.go`): 天気コードを表す Enum。

*   **集約 (Aggregates)**: 関連するエンティティと値オブジェクトをまとめた単位。集約ルートを通じてのみ外部からアクセスされる。
    *   `GetWeatherPointResponse` (`internal/models/types.go`): `WeatherPoint` エンティティのリスト (`Root`) を含む集約。このレスポンス自体が集約ルート。(`WeatherPoints` 構造体も `internal/models/types.go` に定義)
    *   `GetPainStatusResponse` (`internal/models/types.go`): `GetPainStatus` エンティティ (`PainnoterateStatus`) を含む集約。このレスポンス自体が集約ルート。
    *   `GetWeatherStatusResponse` (`internal/models/types.go`): `WeatherStatusByTime` エンティティのリスト (`Yesterday`, `Today`, `Tomorrow`, `DayAfterTomorrow`) を含む集約。`PlaceID` や `DateTime` も属性として持つ。このレスポンス自体が集約ルート。
    *   `GetOtenkiASPResponse` (`internal/models/types.go`): `Element` エンティティのリスト (`Elements`) を含む集約。`Status` や `DateTime` も属性として持つ。このレスポンス自体が集約ルート。 (関連する `Raw*` 構造体も `internal/models/types.go` に定義)

*   **リポジトリ (Repositories)**: 集約の永続化や取得を担当するインターフェース。インフラストラクチャ層で実装される。
    *   `Client` 構造体 (`internal/api/api.go`): 外部 API (zutool API, Otenki ASP API) との通信を担当。以下のメソッドが集約を取得するリポジトリの役割を果たす。
        *   `GetPainStatus(areaCode string, setWeatherPoint *string) (models.GetPainStatusResponse, error)` (定義: `internal/api/api.go`)
        *   `GetWeatherPoint(keyword string) (models.GetWeatherPointResponse, error)` (定義: `internal/api/api.go`)
        *   `GetWeatherStatus(cityCode string) (models.GetWeatherStatusResponse, error)` (定義: `internal/api/api.go`)
        *   `GetOtenkiASP(cityCode string) (models.GetOtenkiASPResponse, error)` (定義: `internal/api/api.go`)

*   **アプリケーションサービス (Application Services)**: ユースケースを実現するための処理フローを定義する。ドメインオブジェクト（エンティティ、値オブジェクト、リポジトリ）を利用してタスクを実行する。
    *   `RunPainStatus` (`internal/commands/pain_status.go`): `pain_status` コマンドの実行ロジック。引数を解釈し、`Client.GetPainStatus` を呼び出し、結果を `Presenter` に渡す。
    *   `RunWeatherPoint` (`internal/commands/weather_point.go`): `weather_point` コマンドの実行ロジック。引数を解釈し、`Client.GetWeatherPoint` を呼び出し、結果を `Presenter` に渡す。
    *   `RunWeatherStatus` (`internal/commands/weather_status.go`): `weather_status` コマンドの実行ロジック。引数を解釈し、`Client.GetWeatherStatus` を呼び出し、結果を `Presenter` に渡す。
    *   `RunOtenkiAsp` (`internal/commands/otenki_asp.go`): `otenki_asp` コマンドの実行ロジック。引数を解釈し、`Client.GetOtenkiASP` を呼び出し、結果を `Presenter` に渡す。

*   **ドメインサービス (Domain Services)**: 特定のエンティティや値オブジェクトに属さないドメインロジック。このプロジェクトでは明確なドメインサービスは現時点では見当たらない。

*   **プレゼンター (Presenter)**: アプリケーションサービスから受け取ったデータをユーザーインターフェース（この場合は CLI）に適した形式で表示する。(`internal/presenter/`)
    *   `Presenter` インターフェース (`internal/presenter/presenter.go`)
    *   `JSONPresenter` (`internal/presenter/json.go`)
    *   `TablePresenter` (`internal/presenter/table.go`)
