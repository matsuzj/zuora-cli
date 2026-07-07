# Architecture

現行実装の構造リファレンス(2026-06-13 時点、フェーズ0〜6 リファクタリング後)。
進行管理は [docs/refactoring-plan.md](refactoring-plan.md)(各フェーズ末尾の実装メモが履歴)。

## レイアウト

```
cmd/zr/            main: エイリアス展開(resolveAliasArgs)→ root.NewCmdRoot + ExecuteContext → exit code 変換
pkg/cmd/<group>/<action>/   1コマンド=1パッケージ(gh CLI 流儀)
pkg/cmd/data-query/  submit/get/list/cancel/run + 共有 dqutil/(グループ内共有ヘルパの先例 — gen-destructive-list.sh のディレクトリ導出制約により、フラット単一パッケージ案は PR #411 で不採用)
pkg/cmd/root/      ルートコマンド組み立て
pkg/cmd/factory/   DI コンテナ(IOStreams / Config / HttpClient / AuthToken、遅延初期化)
pkg/cmd/globalflags/  グローバルフラグの定義(Register)と適用(Apply)の単一ソース
pkg/cmd/alias/     alias の永続化(aliases.yml を所有・読み書き)
pkg/cmdutil/       共有ランナーとヘルパー(下記)
pkg/cmdutil/listcmd/  テーブル list コマンドの宣言的ランナー(Spec)
pkg/output/        出力パイプライン(JSON/jq/template/CSV/table、サニタイズ)
pkg/cmdtest/       コマンドテストハーネス(Run / ハンドラ OK・Reasons・Status・ObjectCRUDFailure・Route / リクエスト契約マッチャ Expect)
internal/api/      HTTP クライアント(リトライ・読み取り専用ゲート・verbose)
internal/auth/     OAuth TokenSource(キャッシュ・単一飛行ロック・資格情報ストア)
internal/config/   config.yml / environments.yml / tokens.yml
```

## コマンドの書き方(正準)

新しいコマンドは**宣言的ランナー**で書く。手書き `runE` は例外コマンド
(account get/summary の nested 特殊、order job-status の watch ループ等)のみ。

- **detail 系**(1リクエスト→詳細表示): `cmdutil.RunDetail(cmd, f, cmdutil.Action{...})`
  — Method/Path/Body/ReqOpts/Fields/SuccessMsg を宣言。空 2xx は RenderSuccess、
  delete 系は `cmdutil.RenderDeleteResult`。
- **list 系**(テーブル一覧): `listcmd.New(f, listcmd.Spec{...})`
  — Flags(Required/Int/OmitZero/Repeatable/DeprecatedName/Enum)、Path、ItemsKey、
  Columns(Money は %.2f)、NextPage(正準ページネーションヒントを自動生成)。
- **JSON-only 系**: `output.RenderJSONOnly`(裸の --csv は明示エラー)。
- オプションは options 構造体(`opts := &xxxOptions{Factory: f}`)に集約。
- 例示は cobra の `Example:` フィールド(`make lint` が Long 内 Examples: を拒否)。
- フラグ語彙: `--account-key`(ID/番号両用のパスparam)/ `--account-number`
  (accountNumber クエリ)/ `--account-id`(accountId クエリ)。必須フラグは
  cobra(`MarkFlagRequired` / `AddBodyFlag(required)`)、条件付き必須のみ手書き。

## HTTP クライアント(internal/api)

- 認証ヘッダは factory の TokenSource から。401 は単一飛行の強制リフレッシュ後に再送。
- リトライ: 冪等メソッドの transport/5xx を最大3回(指数バックオフ+ジッタ)、
  429 は Retry-After(上限60s)。POST/PATCH は Idempotency-Key を自動付与
  (**PUT は Zuora が Idempotency-Key を拒否するため付与しない**)。
- 成功フラグ検査はデフォルト有効(HTTP 200 + success:false は非ゼロ終了)。
  `WithoutCheckSuccess` は zr api の GET/HEAD のみ。
- 読み取り専用ゲート(--read-only / ZR_READ_ONLY、fail-closed)はリクエスト送信前に判定。
- verbose: `-v` で診断行(`* ` プレフィクス、リトライ判断点含む)、`-vv` / `ZR_DEBUG=api`
  でボディ(4KB上限、multipart スキップ、PII のためレベル2限定)。
- リダイレクトはオフホスト/HTTPS→HTTP 降格を拒否(資格情報保護)。

## 認証(internal/auth)

- OAuth client_credentials。トークンは tokens.yml にキャッシュ、環境ごとの
  単一飛行ロックでリフレッシュを直列化。
- 資格情報: `ZR_CLIENT_ID`/`ZR_CLIENT_SECRET` は **both-or-nothing**(完全ペアが
  keyring に勝つ。片側だけは無視)→ OS keyring → login プロンプト。
- TokenSource.Logf(nilガード)が verbose 時の観測行を出す。秘密値は絶対に出さない。

## 出力(pkg/output)

優先順位: `--jq` > `--json` > `--template` > `--csv` > table/detail 既定。
`--json --template` の組のみ拒否(README の優先順位契約)。非TTY +
`default_output=json` + フォーマットフラグ無しのとき JSON に切替。
セルは制御文字/BiDi/改行をサニタイズ、CSV は formula injection 対策付き。

## 環境選択

`--env` フラグ > `ZR_ENV` > config の active_environment。未知名は即エラー
(exact-match、黙殺フォールバックなし)。

## テスト

- ユニット: pkg/cmdtest ハーネス(実 globalflags.Apply を通す。ZR_READ_ONLY /
  ZR_DEBUG / ZR_ENV を中和)。fixture は実レスポンス形から(fixture-masking 防止、
  AGENTS.md 参照)。
- E2E: tests/e2e-*.sh(全スイート — 数は `ls tests/e2e-*.sh` が正、live sandbox、
  手動ゲート)。スキップは期待エラーコード限定。
  [docs/e2e-test-skips.md](e2e-test-skips.md) が正。
- ゲート: `make ci`(CI と同一)+ per-package カバレッジフロア + deadcode +
  Examples: 残存検査。
