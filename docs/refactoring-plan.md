# zuora-cli 全面リファクタリング計画

作成日: 2026-06-10
根拠: 42エージェントによるマルチエージェント監査(8次元 + ギャップ監査3領域)。主要指摘30件はすべて独立エージェントによる敵対的再計測を通過済み。本書の数値は検証で修正された後の実測値。

---

## 1. 現状サマリ

| 項目 | 実測値 |
|---|---|
| プロダクションコード | 15,406行(Go、380ファイル) |
| テストコード | 15,752行(178ファイル) |
| サブコマンドパッケージ | 129(`pkg/cmd/<resource>/<verb>/`、平均80.4行) |
| リーフコマンド数 | 144(cobra.Command リテラル172) |

蓄積した無駄の正体は大きく3種類:

1. **コピペ・ボイラープレート**: 詳細系コマンドの run() 本体が88ファイルにほぼ同一(ファイル間diffで77〜90%一致)、list系11コマンドの骨格が約60%一致、テスト側は `newTestRoot` ヘルパーが136ファイルに完全コピー(1,352行)。
2. **コピー間ドリフト(コピペが既に実害を出している箇所)**: `WithCheckSuccess` 漏れ10ファイル(成功扱いで exit 0 する既知バグクラス)、`--csv` を無視する28コマンド、delete系7コマンドで空200ボディの成否判定が3通りに分裂、order list系4ファイルが `nextPage` を黙って捨てる(結果セットの暗黙切り捨て)、エイリアスの予約名リストが本体登録と乖離(billrun/creditmemo/debitmemo がシャドー可能)。
3. **三重・四重管理**: ビルドゲートが Makefile / Taskfile.yml / ci.yml(+ .goreleaser.yml の ldflags)で手動同期、E2Eシェル9本が約40行のプリアンブルをコピペ(ログ切り捨て修正が9本中2本にしか届いていない)、docs/plans/ は全フェーズ「未着手」のまま凍結。

**削減見込み合計(検証済みの保守的見積り)**: プロダクション約 2,500〜3,500行、テスト約 3,500〜4,500行、シェル/ビルド約 400行、ドキュメント約 1,100行の整理。加えてドリフト由来のバグクラスを構造的に消す。

---

## 2. 全体方針

- **ランタイムヘルパー方式を採用、コード生成はしない**。コマンドの約9割は「メソッド+パス+フィールド」だけの純CRUDであり、宣言的な spec 構造体+共有ランナーで足りる。genuinely カスタムなコマンド(usage/post の multipart、contact/list の queryMore 自動追跡、account/get のネスト展開)は手書きのまま残す。
- **1リソース=1PR** で機械的に移行。既存の per-command テストをゴールデンテストとして「変更せずグリーン」を維持しながら本体だけ差し替え、その後テストを共有ハーネスへ移行。
- **挙動変更は意図的・明示的にのみ**行い、CHANGELOG に記録。ユーザー可視の変更(フラグ改名、エラー文言、出力精度)は非推奨エイリアスを1リリース以上維持。
- 各PRは `make ci` グリーン+カバレッジ床(73%)維持が条件。出力挙動・API挙動に触れるバッチは **E2Eスイート(live tenant)を手動実行**(AGENTS.md の規約どおり)。
- main はストリクト保護でPRが直列化するため、PRは小さく速く。`gh pr update-branch` での追従を前提にフェーズ内の並行PRは2〜3本まで。

### やらないこと(non-goals)

- コード生成基盤の導入(上記理由)。
- `config.Config`(15メソッド)のロールインターフェース分割 — 直接の消費者は12ファイルで、その大半が config/auth/alias コマンド本体。`auth.ConfigStore`(4メソッド)の消費者側宣言だけ行う。
- `contact/list` の共通ランナー化(ZOQL queryMore 自動追跡は正当な例外)。
- `interface{}` → `any` の一括置換(AGENTS.md の既存方針どおり、staticcheck 対象外)。
- Homebrew formula→cask 移行(AGENTS.md で明示禁止)。

---

## 3. フェーズ計画

依存関係: P0 → P1 →(P2 → P3)→ P4 → P5 → P6 → P7。P4 の大半と P6 は P3 と並行可能だが、同一ファイルを触る場合は P3 を優先。

```
P0 基盤整備 ─→ P1 顕在バグ修正 ─→ P2 共有ヘルパー基盤 ─→ P3 コマンド大移行
                                                              │
P4 internal層整理 ←──────────────────────────────────────────┘
P5 CLI表面の一貫性 → P6 可観測性 → P7 ドキュメント・仕上げ
```

---

### フェーズ0: リファクタリング基盤の整備(PR 3〜4本)

以降の全フェーズが依存する検証インフラを先に一本化する。

**P0-1. ビルドゲートの一本化**
- 現状: Makefile(63行)と Taskfile.yml(107行)が12ターゲットを100%重複定義。ci.yml は make を一切呼ばず全ステップをインライン再実装(カバレッジゲートの awk が3重コピー、gofmt チェック3重)。実害ドリフト: CI のビルドは ldflags なし(ci.yml:63)、staticcheck はCIが @latest / ローカルは PATH 任せでバージョン乖離。.goreleaser.yml:21-25 が4つ目の ldflags コピー。
- 作業: (1) ci.yml の各ステップを `make modverify` / `make fmtcheck` / `make lint` / `make vuln` / `make cover` / `make build` への委譲に変更(ステップは分割維持してCIアノテーションを保つ。staticcheck インストールステップは残す)。(2) Taskfile.yml は `cmds: [make <target>]` の1行委譲に縮約(go-task ユーザー向けに薄く残す)。(3) staticcheck のバージョンを固定(tools.go か Makefile 内で `@v…` 指定)し、ローカル/CI を一致させる。
- 削減: 委譲方式で約85〜90行(Taskfile 削除なら約125行)。以後ゲート変更は1ファイル編集。
- リスク: 低。PR2本に分割(ci.yml委譲→グリーン確認→Taskfile縮約)。
- **実装メモ(2026-06-11, PR #64)**: staticcheck の固定は tools.go / Makefile `@v…` ではなく **go.mod の `tool` ディレクティブ**(v0.7.0)で実装。`go tool staticcheck` 化により CI のインストールステップ自体が不要になったため「インストールステップは残す」は撤回(dependabot が tool 依存を bump、ローカル/CI が構造的に同一バージョン)。govulncheck の `./...` は tool 依存を走査しない(バイナリには乗らないため許容、Makefile に注記)。**Part 2(Taskfile縮約)への注意**: `build` タスクの sources/generates フィンガープリントと `e2e` タスクの `{{.CLI_ARGS}}` パススルーは bare な `cmds: [make <target>]` に畳むと黙って消える — 薄いラッパー内で維持するか、挙動差を明示して受け入れること。**→ Part 2 実装済み(2026-06-11)**: 全タスクを `make <target>` 1行委譲に縮約(107→72行)、`build` の fingerprint は維持、`e2e` は Makefile に追加した `ARGS` 変数で `{{.CLI_ARGS}}` をパススルー(`make e2e ARGS="local"` 単体でも使える)。`task e2e` が毎回再ビルドになる点のみ挙動差(make build は .PHONY)— 正確性優先で受容。**P0-1 完了**。

**P0-2. E2E共通ライブラリの抽出**
- 現状: 9スイート計2,752行のうち、36行のプリアンブルが8/9本に逐語一致。`_drain_log` 修正(ログ末尾切り捨て対策)は e2e-order.sh:29-32 と e2e-subscription-write.sh:26-29 の2本にしか届いておらず、**残り7本でログ切り捨てバグが現役**。レートプランID `4c6059a8…` が5本にハードコード(env override があるのは4本)、`run_retry` は1本のみ。
- 作業: `tests/lib/e2e-common.sh` を新設し、色/pass/fail/skip/カウンタ/ログ設定(`_drain_log` 込み)/run()/run_retry()/expect_ok()/expect_fail()/require_auth()/`RATE_PLAN_ID="${ZR_E2E_RATE_PLAN_ID:-…}"` を集約。各スイートは `source` + 各自のチェックだけに。e2e-subscription-write.sh に欠けている expect_fail も移植。
- 移行順: オフラインで検証できる e2e-local.sh → 読み取り中心の e2e-zoql-omnichannel.sh → 残り。1スイート1コミット、移行前後のログを diff。
- 削減: 約295〜300行。修正が9箇所→1箇所になる。
- 注意: e2e-local.sh の `zr` 関数再定義と XDG_CONFIG_HOME 隔離はスイート固有のままライブラリに入れない。

**P0-3. テストの決定性修正(即日級)**
- `TestEnvReadOnly_Unset`(pkg/cmd/root/readonly_test.go:43)が環境変数 `ZR_READ_ONLY` を unset せずに「未設定」を assert しており、ドキュメントが推奨する安全設定を入れている開発機で `go test ./...` が**現に落ちる**。`t.Setenv` + unset で修正。

**フェーズ0完了条件**: `make ci` グリーン、`make e2e` 9スイートのログが移行前と同等、任意の開発機で `go test ./...` が決定的に通る。

---

### フェーズ1: 顕在バグの解消(PR 5〜6本、構造変更なしの小粒修正)

リファクタリング以前に「今ユーザーが踏む」確定バグを潰す。すべて監査でファイル:行まで特定済み。

**P1-1. `WithCheckSuccess` 欠落 10ファイルへの追加**
- 対象(検証済みの全リスト): account/get:42, account/summary:39, account/payment-methods:38, account/payment-methods-default:39, account/payment-methods-cascading:39, account/list:64, contact/get:39, contact/snapshot:41, subscription/versions:39, subscription/metrics:53。
- Zuora は論理失敗を HTTP 200 + `{"success":false}` で返すため、これらは**空の詳細を表示して exit 0 する**(AGENTS.md が「再発バグクラス」と明記するそのもの)。書き込み系は78/78サイト準拠済みで、GETだけがドリフトした。
- success フィールドを持たないレスポンスにはチェックが no-op(client.go:313-326 が *bool で判定)なので、10ファイル全部に付けて安全。
- 各ファイルに「success:false で非ゼロ exit」のテストを追加し、AGENTS.md の「テストが噛むことを証明する」手順(本修正を revert してテストが落ちることを確認)を実施。E2Eで旧挙動を encode していないか確認。

**P1-2. order list系 4ファイルの `nextPage` 復元**
- list-by-subscription / list-by-subscription-owner / list-by-invoice-owner / list-pending の4ファイルは `--page/--page-size` を受け取るのにボディ構造体に `NextPage` がなく(grep -c = 0)、**結果が切り詰められても何のヒントも出ない**。各ファイル約5行のスタブ修正(本格統合はP3)。

**P1-3. エイリアス/エントリポイント周りの修正(このフェーズ最大の塊)**
- cmd/zr/main.go(110行、**テスト0件**)に root.go の二重管理が2つあり、両方とも実ドリフト済み:
  - `builtinCommands` マップ29件 vs root.go 登録31件。billrun/creditmemo/debitmemo が漏れており、`zr alias set creditmemo "account list"` で**ビルトインがシャドーされることを実機再現済み**。
  - 値を取るグローバルフラグのリスト(main.go:69-70 の5綴り)が root.go の8 persistent flag 定義の手動コピー。
  - `strings.Fields` での分割により、**`zr query "<ZOQL>"` をラップするエイリアスが壊れている**(クォート破壊、実機再現済み。gh は shlex を使用)。
  - `store.Load()` エラーの黙殺(壊れた aliases.yml で全エイリアスが無言で無効化)。alias ファイルの解決パスも `config.Dir()` と `cfg.ConfigDir()` の2系統。
  - `alias set` に予約名ガードがない(シャドーするエイリアスと、作れるのに効かない死にエイリアスの両方を黙って作成)。
- 作業順序(これ自体が安全な移行の手本): (1) 展開ロジックを純関数 `ExpandAlias(rootCmd, args, store)` として抽出し、現挙動を固定する characterization テストを先に書く(クォート/ビルトイン/--env 値スキップ/--env=形式/-e/破損YAML)。(2) rootCmd を展開**前**に構築し、予約名セットとフラグ arity を `rootCmd.Commands()` / `PersistentFlags().VisitAll`(`Value.Type() != "bool"`)から導出 — 2つの手動リストを削除。(3) `github.com/google/shlex` を導入して分割を置換(gh と同じ依存)。(4) Load エラーを stderr 警告にして通常ディスパッチへフォールスルー。(5) `alias set` に予約名拒否+自己参照拒否を追加(書き込み側ガード。`alias delete` は予約名でも動くまま残し、既存の汚染エントリを除去可能に)。(6) e2e-local.sh に「設定したエイリアスを実行する」ケースを追加(現状、展開を通すE2Eはゼロ)。
- リスク: NewCmdRoot は構築時に副作用なし(PersistentPreRunE は実行時のみ)なので展開前構築は安全。テスト先行で挙動を固定してから1コミット1修正。

**P1-4. Ctrl-C 即応化(2件)**
- `zr order job-status --watch` のポーリングが生 `time.Sleep(5s)`(job_status.go:84)+ NotifyContext のため、**Ctrl-C が最大5秒殺され、2発目も無効**(実測5.001s)。`cmdutil.SleepContext`(internal/api/client.go:135-147 の sleepWithContext を昇格)を新設して置換、`--interval`/`--timeout` フラグを追加(デフォルト5s/無制限で現挙動維持)。50msでキャンセル→500ms以内に return するテストを追加(time.Sleep 実装では必ず落ちる)。
- `zr auth login` / `zr auth token` が context.Background() ラッパー(oauth.go の Refresh/Token)経由のため、**ハングした OAuth リクエスト中は30秒間中断不能**。login.go:114 → `ForceRefreshContext(cmd.Context(), …)`(single-flight ロックも正しく取るようになる)、token.go:35 → `TokenContext(cmd.Context(), …)` に変更。ラッパー関数自体の削除はP4で。

**フェーズ1完了条件**: 上記すべてに「修正前は落ちる」テストが付き、E2E(認証系+order系+local)グリーン。ここまでで一度リリースを切ることを推奨(以後の大規模変更と分離するため)。

---

### フェーズ2: 共有ヘルパー基盤の構築(PR 4〜5本)

P3 の大量移行が乗る土台。**この段階では既存コマンドを書き換えない**(ヘルパー+ヘルパー自身のテストのみ)。例外は P2-1 の機械的一括変更。

**P2-1. `WithCheckSuccess` のデフォルト反転**
- 現状はオプトイン(125箇所/122ファイルが手で付与、AGENTS.md がレビュー規約で防御)。これを `newRequestConfig` でデフォルト true に反転し、`WithoutCheckSuccess()` を追加。生 `zr api` パスのみ非変更系メソッドでオプトアウト(pkg/cmd/api/api.go:97-101 の既存の特例をそのまま表現)。125箇所の呼び出し引数を一括削除。
- これで P1-1 のバグクラスが**構造的に再発不能**になり、AGENTS.md の該当規約とレビュー負担が消える。インラインのエンベロープ検査(client.go:313-326)は `successEnvelopeError(body)` として response.go/errors.go 側へ抽出。
- リスク: success フィールドのない 2xx ボディには no-op(checksuccess_test.go:56-79 が既に固定)なので list/object-query 系は無影響。E2E全スイートで確認。P3 の前に行うことで、ランナーがこのオプションを一切持ち回らずに済む。

**P2-2. `pkg/output` の入口統一**
- `output.RenderJSON(ios, rawJSON, opts)` を新設: 正準順 JQ > JSON > Template > (CSV方針) > PrintJSON。現状28ファイルが6行の手書き分岐をコピペし、うち全てが `--csv` を黙殺、28ファイルが `--json`/`--template` の優先順を RenderDetail と逆に実装している(--json+--template 併用は root で拒否済みのため実害は限定的だが、統一はここで宣言)。Render / RenderDetail の重複する先頭3分岐(formatter.go:43-51 / 68-76)も RenderJSON 呼び出しに畳む。
- `output.RenderSuccess(ios, opts, humanMsg)` を新設: delete系が10箇所コピペしている `{"success": true}` 合成+分岐+stderrメッセージを1関数に。
- `cmdutil.RenderDeleteResult`(または P3 ランナーのオプション)で 204 / 空200 / 非JSON 200 の方針を一本化。**決定事項**: 空200ボディは成功扱いを推奨(WithCheckSuccess が論理失敗を上流で弾くため)。現状は contact/fulfillment×2/omnichannel が成功・order/usage/account がエラーの3方針に分裂しており、どちらに寄せても挙動変更 — 独立コミット+3レスポンス形状のテスト付きで。
- json.go 内部の重複(prettyJSON 抽出、emptyBody ガード4箇所、sanitizeCell/sanitizeCSVCell 統合)もここで実施(約27行、CWE-1236 等の load-bearing コメントは維持)。

**P2-3. `cmdutil` の小物ヘルパー**
- `AddBodyFlag(cmd, &v, required)` / `AddConfirmFlag(cmd, &v)`(--body 定義54ファイル・ヘルプ2変種、--confirm ヘルプ8変種を正準化)、`RequireFlag` 相当は P5 の MarkFlagRequired 移行で吸収するため**ここでは作らない**。
- `GetBool` / `GetInt` / `GetMoney`(固定2桁)を追加。account/get:87-113 と account/summary:103 の私製 getNumber/getBool/getInt(逐語一致の重複あり)を削除し、`.(string)`/`Sprintf("%v")` バイパス10箇所を置換。**決定事項**: 金額表示は現行の `%.2f`("50.00")を GetMoney で維持(E2E・ユーザースクリプトの互換のため。GetDecimal への正規化は採らない)。
- `output.Column.Field` の削除: 85箇所で代入されるが**一度も読まれない** write-only フィールド。gh 風自動抽出を匂わせる誤誘導なので構造体から削除(84箇所は sed 可能、query.go:180 のみ手修正)。

**P2-4. `pkg/cmdtest` テストハーネス**
- `cmdtest.Run(t, parent, newCmd, handler, args…) (stdout, stderr, err)`: newTestRoot 構築(persistent flags 全部入り)+ httptest.NewServer + NewTestFactory + 実行 + ストリーム回収を1呼び出しに。現状 `newTestRoot` が136ファイルに完全コピー(1,352行)、実行ブロックが426箇所。
- エンベロープビルダー `cmdtest.OK(method, path, body)` / `cmdtest.Reasons(code, msg)` / `cmdtest.Status(code, body)`: `"success": true` 手書き75ファイル、`"reasons"` 手書き45ファイル、同形の SuccessFalse テスト44本を1行化。OK() は expected method/path を取り、既存のリクエスト assert を保てる署名にする。
- 代表レスポンスのゴールデンフィクスチャ `pkg/cmdtest/fixtures/*.json`: E2E(実テナント)から採取・サニタイズして登録。AGENTS.md の fixture masking 警告への構造的対策(同一ファイル内に同じフィクスチャを2回ペーストしている例: invoice/list/list_test.go:33-45 と 65-77)。

**フェーズ2完了条件**: 新ヘルパー全部に単体テスト、`make ci` グリーン、既存コマンドは P2-1 以外無変更。

---### フェーズ3: コマンド大移行(PR 20〜25本、本丸)

129 verb パッケージ(10,378行)の機械部分を宣言化する。**期待削減はプロダクション約 2,000〜2,800行+テスト約 2,500〜3,500行**。

**P3-1. 詳細系ランナー `cmdutil.RunDetail`**
- 設計: `Action{Method, Path string; Body io.Reader; ReqOpts []api.RequestOption; Fields func(raw map[string]interface{}) []output.DetailField; Success func(raw) string}` + `RunDetail(cmd, f, act)`。HttpClient 取得 → リクエスト → `FromCmd` → `parsing response: %w` ラップ付き Unmarshal → RenderDetail → ErrOut 成功メッセージ、を一手に引き受ける。
- 検証時の補正を反映した設計上の注意: SuccessMsg は単純な string では足りない(refund.go:83-85 のような「id があるときだけ2値補間」があるため func フィールドにする)。各コマンドは NewCmdX(cobra 配線・フラグ・ドキュメント)を保持し、runX が Action 組み立てに縮む(80行 → 50〜60行/ファイル)。
- 対象: RenderDetail を使う88ファイル(うち55は末尾まで同一テンプレート)。**正味削減 約1,200〜1,800行**。
- 手書きのまま残す例外: usage/post(multipart)、account/get・account/summary(ネスト展開とカスタムヘルパー — GetMoney 移行のみ)、meter系の特殊レスポンス。
- 移行手順: ランナー+単体テストを先に独立PRで → 1リソース=1PR で機械移行。**既存 _test.go は変更せずグリーン維持**(`parsing response: %w` の文言まで含めて互換)。最初のリソース移行後と全移行後に E2E。

**P3-2. list系ランナー `pkg/cmdutil/listcmd`**
- 設計: `Spec{Use, Short, Long; Args; Path func(args []string, flags map[string]string) string; Flags []QueryFlag; ItemsKey string; Columns []ColumnSpec; NextPageHint …}`。検証時の補正を反映: (a) `QueryFlag` は型・デフォルト値を持てるようにする(account/list の IntVar `--page-size` デフォルト20、repeatable StringArray + `WithQuerySlice`)。(b) `Path` は positional args と flags の両方を受ける(subscription/list は `--account` フラグからパスを構築)。(c) `ColumnSpec` に `Money bool`(%.2f 維持)を持たせる — 5コマンドが %.2f で金額描画しており、GetDecimal 化は出力変更("100.00"→"100")になるため**現行表示を既定で維持**。
- ランナーは reqOpts 組み立て → client.Get → ItemsKey でデコード → セル抽出 → output.Render → **正準の nextPage ヒント**(cobra.CommandPath() + 非デフォルトのフラグ値 + positional args から次コマンドを再構築)。現状4通りに分裂しているページネーション挙動(コピペ可能な次コマンド提示/曖昧なヒント/無言切り捨て/自動全件)が最良形に統一される。
- 対象: 標準テーブルlist 11コマンド(1,276行)→ **正味削減 約600〜750行**。order list-by-* 4兄弟(各101行、89〜90%一致)はここで spec 4個に畳む。commitment/plan/product-legacy の3listは Columns を定義してテーブル+CSV対応に昇格(レスポンス形状は既存テストのモックとE2Eで確認してから)。
- 移行順(監査の推奨どおり): creditmemo+debitmemo(最もクリーンな79%一致ペア)→ invoice/payment/order → order 4兄弟 → account/subscription(cursor/URL再構築フックが必要なため最後)。contact/list は対象外。
- 各リソースPRでテストも cmdtest + spec ベースに移行し、**シナリオはユニオンを取る**(現状、クローンごとに success=false / paging / --json のテストが歯抜けで異なるサブセットしかない。ランナーのテストスイートで全行動×全コマンドを1回で担保)。

**P3-3. 出力分岐テールの掃討**
- P3-1/2 に乗らない28ファイル(charge/*, plan/*, ramp/*, subscription/changelog*, commitment/*, invoice/files 等の JSON-only コマンド)の手書き分岐を `output.RenderJSON` 呼び出しに置換(**正味約250〜280行**)。query.go:164-172,183-185 の冗長プレディスパッチ(直後の Render が同じ判定を再実行)も削除。
- **決定事項**: JSON-only コマンドでの `--csv` は「明示エラー」を推奨(黙殺は現状バグ。passthrough より診断的)。挙動変更として CHANGELOG に記載。

**P3-4. テストハーネス移行の完遂**
- 残りのコマンドテストを cmdtest へ(136ファイルの newTestRoot 削除、426実行ブロックの1行化、44 SuccessFalse のビルダー化)。**正味削減 約2,500〜3,500行(pkg/cmd テストツリーの25〜30%)**。
- 注意: ディスパッチ/バリデーションの共通挙動はランナー側で一度だけテストし、per-command テストは「リクエスト形状(method/path/body)+フィールドマッピング+成功メッセージ」に絞る。ただし**各コマンド最低1本は実レスポンス形状のフィクスチャによる end-to-end 風テストを残す**(過剰中央化で per-command 回帰を見逃さないため)。
- `go test -v` のテスト名リストを移行前後で比較し、ケースの黙失を防ぐ。リソース移行ごとにカバレッジ床を確認(共有ランナーに集中するため通常は上がる)。

**フェーズ3完了条件**: 全リソース移行済み、E2E 9スイートグリーン、カバレッジ床維持、新コマンド追加の作法(spec を書くだけ)を AGENTS.md に追記。

---

### フェーズ4: internal 層とテストスイートの整理(PR 6〜8本、P3と一部並行可)

**P4-1. デッドコード削除(検証済みリスト、計約100行)**
- 完全死(テスト含め呼び出しゼロ): `api.WithUserAgent`、`factory.NewTestFactoryReadOnly`、`iostreams.IsStderrTerminal`、`iostreams.ColorEnabled`。
  - **注意**: ColorEnabled は本リポジトリ唯一の NO_COLOR 処理で、README.md:276 が NO_COLOR 対応を明記している。実際には色出力が存在せず(tablewriter 無着色+サニタイザが ESC を除去)、間接依存の fatih/color も自前で NO_COLOR を尊重するため、**削除+README.md:276 の行も削除**で整合させる(配線する価値が生じたら再実装)。
- 自パッケージのテストだけが延命させているもの: `Client.Patch`(テスト2箇所を `Do(http.MethodPatch,…)` に書き換えて削除)、`Response.JSON`/`Response.String`(専用の29行テストファイル response_json_test.go ごと削除)、`TokenSource.ForceRefresh`/`Refresh`/`Token` の context なしラッパー3本(P1-4 で呼び出し側は移行済み)、`alias.Store.Len`(assert.Len(s.List(), N) に書き換え)。
- `internal/config.Dir()` は P1-3 で唯一の呼び出しが消えるため削除。

**P4-2. internal/api の整理**
- With*/Set* の二重構成面を解消: **オプション(b)を採用** — テスト専用の WithVerbose/WithReadOnly/WithContext を削除し、テストを本番と同じ Set* 経路に書き換え(17行削減+本番経路がテストで直接叩かれるようになる)。WithHTTPClient はテスト注入シームとして明示的に残す(doc コメントでその旨を宣言)。
- Idempotency-Key 述語の単一ソース化: client.go:261 を `carriesIdempotencyKey(method)` 呼び出しに(SafeToRetry のユーザー向け約束が2コピーの一致に依存している現状を解消。重複支払い級のドリフトハザード)。PUT除外の根拠コメントは carriesIdempotencyKey に1本化し他は参照に。
- トランスポートエラーの形状統一(低優先): APIError に `Err error` + `Unwrap()` を持たせ、retry.go:96 と :161 の2経路を同形に。exit code は不変。
- テスト統合: 11ファイル → 本番ファイル対応の約6ファイルへ。**逐語移動 → 重複削除の2コミット**(移動コミットで `go test -v` リスト一致を確認)。検証で確定した注意点: hostcheck の same-host 絶対URLテストは唯一のカバレッジなので統合先で保持、cancel-during-backoff 2本は別ブランチ(5xx backoff と 429 Retry-After)なので**両方残す**。正味削減約130〜180行。
- 実測7.74秒を1テストで消費する `TestClient_ServerError_ExitCode`(パッケージ計7.83秒)を削除または newNoSleepClient 化(パッケージテストが0.5秒未満に)。

**P4-3. internal/config / internal/auth の整理**
- **MockConfig の委譲化**: 93行の手書き再実装が fileConfig とバリデーション乖離(YYYY-MM-DD 検査スキップ、AddEnvironment 検査スキップ、RemoveEnvironment が未知envで成功、mutex なし)。fileConfig へ委譲し Save だけ spy フック化。公開テストAPI(NewMockConfig / .Envs / SaveError / SaveCallCount)は不変なので142テストファイルの import 変更不要。
- **configtest パッケージへの移動**(委譲化の後、純リネームPRで): mock が internal/config のカバレッジを17.8ポイント圧迫している(65.4%→83.2%)。同様に MockCredentialStore → `auth.StaticCredentialStore` へ改名(**本番の login.go:110 が "Mock" 型に依存している**誤誘導の解消)し、テスト用エイリアスを残す。factorytest / iostreamstest も同パターンで移動し、`deadcode -test ./...` が空になる状態にして **Makefile の ci ターゲットに deadcode ゲートを追加**。
- **`default_output` の決定**: set/get/list/validate/persist まで完備で**どこからも読まれない**設定キー。**配線を推奨**(root.go の PersistentPreRunE で、フォーマットフラグ未指定時に `default_output=json` を --json の既定として扱う。TTY ガード付き)。配線しない判断なら36行+テスト群を削除し、`config set default_output` には有効キーを列挙するエラーを返す。どちらにせよ「設定できるのに効かない」現状は解消。
- ZR_CLIENT_ID/ZR_CLIENT_SECRET の判定一本化: 4ファイル12箇所に再導出され、login だけ per-variable フォールバックという**別ルール**になっている。`auth.EnvCredentials()` に集約し、both-or-nothing に統一(login の挙動変更は CHANGELOG へ)。
- TokenSource の表面縮小: 5公開メソッド → TokenContext + ForceRefreshContext(+便宜 Token)に。ロック取得ブロックの2重コピーを `lockEnv` ヘルパーに。`auth.ConfigStore`(4メソッド)を internal/auth 側で宣言し、factory の TokenSource 配線3重複を `factory.tokenSource()` に集約。
- auth テスト統合: 7ファイル576行 → 約3ファイルへ(テーブル駆動)。検証で見つかった**本物の欠落テスト**を追加: oauth.go:142-145 の 200文字超ボディ切り詰めブランチ(現状、名前だけ「Truncates」のテストが26バイトしか送っていない)、実 keyringStore の Set/Get/Delete(現状0%)。

**P4-4. カバレッジゲートの実質化**
- 73%は集計のみで、配下に床割れ12パッケージ(factory 39.5%、iostreams 50.0%、internal/config 65.4%、internal/auth 71.2% など)を隠している。**per-package 床をラチェット方式で導入**: 初期値は現状最低値の直下(38%)で導入即グリーン → factory/iostreams に的を絞ったテストを足してから段階的に引き上げ。登録専用の親パッケージ(0%)は除外リストへ。internal/build はちょうど60.0%なので >= 比較で。
- P3 で浮いたテスト行数の一部をここ(全コマンドが共有するインフラ)へ振り向ける。

**フェーズ4完了条件**: deadcode -test が空、per-package ゲート稼働、internal/api と internal/auth のテストが「1本番ファイル=1テストファイル」対応。

---

### フェーズ5: CLI表面の一貫性(PR 5〜7本、ユーザー可視 — リリースノート必須)

P3 完了後に行う(spec ベースのコマンドに対する変更は1箇所で済むため)。すべて非推奨エイリアス付きで段階移行。

**P5-1. account系フラグ語彙の正準化**
- 現状3綴り5意味/8コマンド: `--account` が「アカウントキー(path param)」の意味で3コマンド、「accountNumber(query param)」の意味で2コマンド — **同じ綴りで違うID種を受け、間違えるとエラーではなく空結果**。
- 正準化: エンドポイントが受けるものに合わせ `--account-key`(payment/subscription/invoice list)/`--account-number`(commitment系)へ改名、creditmemo/debitmemo/contact の `--account-id`/`--account-number` は現状維持。旧名は `MarkDeprecated` + `MarkHidden` で1リリース以上維持。`cmdutil.AddAccountKeyFlag` 系ヘルパーで再発防止。subscription/list:123 のページネーションヒント内の補間も更新。E2E の `--account` 使用箇所を同PRで更新。

**P5-2. 必須フラグの cobra 化**
- 手書き `is required` ガード65箇所(文言17種)vs `MarkFlagRequired` 5箇所。ヘルプにも補完にも必須が表示されない(`zr order create --help` の --body に必須表示なし、を実機確認済み)。
- `AddBodyFlag(required=true)` で50箇所、単独フラグ13箇所は MarkFlagRequired、compound は `MarkFlagsOneRequired`/`MarkFlagsRequiredTogether`(cobra v1.10.2 で利用可)。**--confirm の19箇所は対象外**(RequireConfirm は同意ゲートであり必須入力ではない)。
- エラー文言が cobra 標準(`required flag(s) "body" not set`)に変わる挙動変更 — テスト assert 約50箇所とE2Eの stderr grep を同コミットで機械更新。約190行の削減。

**P5-3. ヘルプ/補完の整備**
- `Example:` フィールド移行: 現状 **0/172 コマンドが Example: を使わず**、132ファイルが Long 内に Examples ブロックを埋め込み。機械的に移設(約400行の移動)し、AGENTS.md に規約追記+ ci に `grep -rln 'Examples:' pkg/cmd --include='*.go' --exclude='*_test.go'` の空チェックを追加。
- 動的補完: completion コマンドを出荷しながら ValidArgsFunction / RegisterFlagCompletionFunc が**ゼロ**。`cmdutil.EnumCompletion`(--policy×3、--periods-type×2、--status×3、--export-type/--run-type)、`EnvNamesCompletion`(--env、config から取得、失敗時 NoFileComp)、config get/set のキー補完、alias delete のエイリアス名補完 — 計約12登録。API を叩く補完はスコープ外。
- プレースホルダ統一: `<meterId>`→`<meter-id>`(5ファイル)、usage系の `<id>`→`<usage-id>`(3ファイル)。subscription の `<subscription-number>` 3箇所は**改名せず** Long に「このエンドポイントは number のみ受け付ける」と1文追記(API制約の文書化)。
- charge/plan の `get` を positional 引数化(`zr charge get <charge-key>`): CLI 中この2つだけが `--key` 必須フラグ方式。--key は deprecated hidden で1リリース維持。
- query の `--csv` ローカル再定義(root の persistent flag をシャドー)を削除し継承値を読む。`--page-size` の string/int 混在を IntVar に正規化し `cmdutil.AddPagingFlags` に集約(--cursor/--limit/--paginate は文書化された例外として維持)。

**フェーズ5完了条件**: `zr <何か> --help` の文法・必須表示・Examples 位置が全コマンド一様、E2E グリーン、CHANGELOG にユーザー可視変更を列挙、リリースを1本切る。

---

### フェーズ6: 可観測性とライフサイクル(PR 3〜4本)

**P6-1. --verbose のリトライ可視化**
- 現状、verbose はリクエスト1行とレスポンス1行のみで、**その間の最大3リトライ・1〜4秒のバックオフ・最大180秒の Retry-After 待ち・401トークン交換がすべて不可視**(retry.go に verbose 参照ゼロ)。README は「--verbose でリクエストを確認してから本番を触れ」と案内しており、約束と実態が乖離。
- `Client.vlogf` を追加し、doWithRetry の各判断点に `* retrying (attempt 2/4) after 1.4s backoff` / `* HTTP 429, honoring Retry-After: 30s` / `* HTTP 401, refreshing token and resending` 等の `*` プレフィクス行(gh 流儀)を約8行。既存の sleep シームでテスト可能。リフレッシュ後トークンの `Bearer ***` マスクを assert するテストを必ず付ける。

**P6-2. auth の可観測化**
- internal/auth は出力ゼロ(キャッシュヒット/ネットワークリフレッシュ/強制リフレッシュ/資格情報ソースが区別不能)。OAuth POST は自前 http.Client のため client.go の verbose フックを**完全に素通り**。TokenSource に nil チェック付き `Logf` を追加し、factory→root の verbose 配線から注入。シークレット・トークン値は絶対に出さない(漏えい否定テスト付き)。

**P6-3. ボディログ(ゲート付き)**
- gh の `GH_DEBUG=api` に倣い、第2レベル(`ZR_DEBUG=api` または `--verbose --verbose`)でリクエスト/レスポンスボディを出力。4KB キャップ、multipart は Content-Type 判定でスキップ、stderr のみ。請求データ=PII なので既定の --verbose には載せない。

**P6-4. ZR_ENV の実装**
- docs/plans が約束し README の表が「—」と認める未実装env var。`envOverrideConfig` の既存検証パスに約6行で乗る(フラグ > env var の優先順は ZR_READ_ONLY と同型)。**削除でなく実装を推奨**(credentials と read-only は既に env var パリティがあり、environment だけが欠けている。gh GH_HOST 相当)。README の2表を更新。
- (任意)context.Canceled を exit code 130 にマップ(main.go)。E2E の exit 期待を grep してから。

---

### フェーズ7: ドキュメントと再発防止(PR 2〜3本)

**P7-1. docs/plans の精算**
- 進捗表は11フェーズ全部「未着手」、チェックボックス197個未チェックのまま、**129コマンドが出荷済み** — しかも AGENTS.md:66 と「セッション引継ぎガイド」が新規エージェントをこの偽情報に誘導している(本リポジトリの主開発者はAIエージェントなので実害が大きい)。
- アーキテクチャ節を `docs/architecture.md` に抽出(AGENTS.md:66 の参照先を変更)、phase-*.md は `docs/plans/archive/` へ「歴史的文書」バナー付きで移動。**例外**: phase-05-pending.md はテナント制約による未消化E2E項目の生きたTODOなので、docs/e2e-test-skips.md に統合して保全。docs/zuora-api-reference.md(526行、3月から未更新)は「まだ正か」の検証を1回かける。本計画書(本ファイル)が進行管理を引き継ぐ。
- 約1,071行の誤誘導ドキュメントの整理。

**P7-2. README / AGENTS.md の重複解消**
- 実在の矛盾を修正: AGENTS.md:8 の test コマンドにカバレッジフラグが欠落(Makefile/README とは不一致)。AGENTS.md を貢献者ワークフローの単一情報源にし、README の Development 節は「前提条件+ `make ci` を実行、詳細は AGENTS.md」に縮約(約18〜22行)。AGENTS.md:27 の「9 suites」リテラルを glob 表現に。README.md:114 の破壊的コマンド19個の手書きリストは `grep -rl RequireConfirm pkg/cmd` で生成するスクリプトに置換。
- P3〜P5 で確立した新規約を AGENTS.md に追記: 「新コマンドは Action/Spec を書く(手書き run() は例外コマンドのみ)」「Examples は Example: フィールドへ」「options 構造体スタイルが正準」「フラグ語彙表(--account-key/--account-number/--account-id の使い分け)」。

---

## 4. 削減見込みサマリ(検証済み数値ベース)

| フェーズ | プロダクション | テスト | その他 |
|---|---|---|---|
| P0 | — | — | ビルド約85〜125行、E2E約295行 |
| P1 | 修正のみ(±0前後) | +characterizationテスト | — |
| P2 | 約150〜200(dispatch内部重複、Column.Field、私製ヘルパー) | ハーネス新設(+) | — |
| P3 | 約2,000〜2,800(detail 1,200〜1,800 + list 600〜750 + tail 250〜280 + delete 120〜140) | 約2,500〜3,500 | — |
| P4 | 約250〜300(デッドコード約100 + mock委譲60 + default_output 36 + API表面) | 約400〜500(統合+envelope) | — |
| P5 | 約190(必須ガード)+ヘルパー化 | assert更新(±0) | — |
| P7 | — | — | docs約1,100行整理 |
| **計** | **約2,600〜3,500行(プロダクションの17〜23%)** | **約3,000〜4,000行** | **約1,600行** |

---

## 5. 主要な意思決定ポイント(着手前に確定すべきもの)

| # | 決定事項 | 推奨 |
|---|---|---|
| 1 | JSON-only コマンドでの `--csv` | 明示エラー(黙殺は現状バグ) |
| 2 | delete の空200ボディ | 成功扱い(WithCheckSuccess が上流で論理失敗を弾く) |
| 3 | 金額表示 | `%.2f` 維持(GetMoney / ColumnSpec.Money)。出力互換優先 |
| 4 | `default_output` | 配線する(TTYガード付き)。削除でも可、現状維持のみ不可 |
| 5 | Taskfile.yml | make への1行委譲で存続 |
| 6 | ZR_ENV | 実装する(ドキュメント削除ではなく) |
| 7 | 出力フォーマット優先順 | JQ > JSON > Template に統一(--json+--template 併用は root で拒否済み) |

---

## 6. 検証プロトコル(全フェーズ共通)

1. PRごと: `make ci`(= gofmt / vet / staticcheck / govulncheck / race test / カバレッジ床 / build)。
2. 挙動を意図的に変えるコミットは独立させ、「修正前は落ちるテスト」を同梱(AGENTS.md の prove-the-test-bites)。
3. 出力・API・auth に触れるバッチ: `make e2e`(live sandbox、手動ゲート)。P3 は最初のリソース移行後と完了後の最低2回+リソースごとに該当スイート。
4. テストファイルの統合・移動は「逐語移動 → 重複削除」の2コミットに分け、`go test -v` のテスト名リスト一致を移動コミットで確認。
5. リリースは P1 後・P3 後・P5 後の3回を目安に。各リリース前にリリースコミット上で E2E(AGENTS.md の規約)。
