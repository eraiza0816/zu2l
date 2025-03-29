## zutoolプロジェクトのGoへの移植計画

1.  **プロジェクトのセットアップ**
    *   Goのプロジェクトを作成します。
    *   必要な依存関係をGoのパッケージマネージャーでインストールします。
2.  **データモデルの移植**
    *   `zutool/models/*.py`に定義されているデータモデルを1つずつGoの構造体に変換します。
    *   `pydantic`のバリデーション機能をGoの標準機能またはサードパーティライブラリで代替します。
3.  **APIクライアントの移植**
    *   `zutool/api.py`に定義されているAPIクライアントをGoで実装します。
    *   `requests`ライブラリをGoの`net/http`パッケージで代替します。
    *   HTTPリクエストのタイムアウトやUser-Agentの設定を行います。
4.  **コマンドラインインターフェースの移植**
    *   `zutool/main.py`に定義されているコマンドラインインターフェースをGoで実装します。
    *   Cobraライブラリを使用して、コマンドラインインターフェースをGoで実装します。
    *   `rich`ライブラリをGoの標準機能またはサードパーティライブラリで代替します。
5.  **テストの移植**
    *   `tests/test_zutool.py`に定義されているテストコードをGoで実装します。
    *   `pytest`ライブラリをGoの標準機能またはサードパーティライブラリで代替します。
6.  **エラー処理の移植**
    *   Pythonのエラー処理をGoのエラー処理に置き換えます。
7.  **並行処理の移植**
    *   必要に応じて、Pythonの並行処理をGoのgoroutineで代替します。
8.  **ドキュメントの作成**
    *   Goのドキュメントツールを使用して、ドキュメントを作成します。

## DDDモデル

*   **エンティティ**:
    *   `_WeatherPoint` (zutool/models/get_weather_point.py): 天気予報の地点情報
    *   `_GetPainStatus` (zutool/models/get_pain_status.py): 痛み予報のステータス情報
    *   `_WeatherStatusByTime` (zutool/models/get_weather_status.py): 時間ごとの天気情報
*   **値オブジェクト**:
    *   `PressureLevelEnum` (zutool/models/enum_type.py): 気圧レベル
    *   `WeatherEnum` (zutool/models/enum_type.py): 天気
    *   `AreaEnum` (zutool/models/enum_type.py): 地域
*   **集約**:
    *   `GetWeatherPointResponse` (zutool/models/get_weather_point.py): 天気予報の地点情報の集約
    *   `GetPainStatusResponse` (zutool/models/get_pain_status.py): 痛み予報のステータス情報の集約
    *   `GetWeatherStatusResponse` (zutool/models/get_weather_status.py): 天気情報の集約
    *   `GetOtenkiASPResponse` (zutool/models/get_otenki_asp.py): お天気ASPの情報の集約
*   **リポジトリ**:
    *   `api.py`に定義されている`get_pain_status`, `get_weather_point`, `get_weather_status`, `get_otenki_asp`関数が、外部APIとの通信を行うリポジトリの役割を果たしています。
*   **アプリケーションサービス**:
    *   `main.py`に定義されている`func_pain_status`, `func_weather_point`, `func_weather_status`, `func_otenki_asp`関数が、コマンドラインからのリクエストを処理し、リポジトリを呼び出してデータを取得し、結果を表示するアプリケーションサービスの役割を果たしています。
