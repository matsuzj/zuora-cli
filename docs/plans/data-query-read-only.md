# Data Query を Read-Only から除外し、専用 `zr data-query` コマンドを追加

Status: shipped (PR #411, 2026-06-29)

> **本文は計画当時のまま**。実装との差分は直下の「実装メモ」が正。

## 実装メモ(2026-07-08 追記 — 計画からの逸脱)

- **パッケージ構成は計画の逆が出荷された**: 本計画(Codex 6 パスレビュー済み)はフラット単一パッケージ `pkg/cmd/dataquery` を第一候補としたが、実装レビュー(PR #411 の P1)で `scripts/gen-destructive-list.sh` がコマンド名をディレクトリ構造から導出する制約と衝突することが判明し、代替案だったサブパッケージ構成 `pkg/cmd/data-query/{submit,get,list,cancel,run}` + 共有 `dqutil/` が採用された。
- `run --timeout` は #506 で `--wait-timeout` に改名(global `--timeout` との二重定義回避)。
- 本文中のカバレッジ床 73% は当時の値。現行は total 83.0% / per-package 60.0(Makefile が正)。
- Open items の live 検証は 2026-06-29 に apac-sandbox で実施済み — 結果は `pkg/cmd/data-query/dqutil/dqutil.go` の日付入り live-verified コメントが正。

## Context

Zuora **Data Query**（非同期 Trino SQL：`POST /query/jobs` 投入 → `GET /query/jobs/{id}` ポーリング → 完了後に S3 の `dataFile` をダウンロード）は **読み取り専用**の機能だが、CLI の read-only ガードは HTTP メソッドベースで `POST`/`DELETE` をデフォルトでブロックする（`internal/api/client.go:476-498`）。そのため Data Query は `--read-only` / `ZR_READ_ONLY` 下で投入も取消もできない。

さらに調査の結果、**Data Query 用のコマンドは CLI に存在しない**（`grep "query/jobs"` = 0 件）。既存の `zr query`（`pkg/cmd/query/query.go`）は **ZOQL**（`POST /v1/action/query`）で別物であり、これは既に allowlist 済み（`client.go:437-438`）。

**ゴール（ユーザー確認済み）:**
1. 専用コマンド `zr data-query` を新設（submit / get / list / cancel / run+download）。
2. read-only モードで Data Query の **submit（`POST /query/jobs`）と cancel（`DELETE /query/jobs/{id}`）を許可するかを「判断」できるようにする**。GET 系（status/list）は元々許可。
   - **既定はブロック（オプトイン）**: `ZR_READ_ONLY` 単体では Data Query も従来どおりブロック（read-only の安全網を維持）。
   - 明示的なトグル **`--read-only-allow-data-query` フラグ / `ZR_READ_ONLY_ALLOW_DATA_QUERY` 環境変数**（フラグが env に優先、`--read-only`/`ZR_READ_ONLY` と同じ流儀）を立てたときだけ、read-only でも Data Query の POST/DELETE を通す。
   - このトグルは **Data Query のみ**を広げる。通常の write（`POST /v1/accounts` 等）は立てても引き続きブロック。

---

## レビュー反映（Codex 第二意見 + 一次コード確認）

Codex（`codex exec --sandbox read-only`、Zuora docs を web 確認）と当方のコード確認で以下を反映済み:
- **P1** import cycle 回避 → フラット単一パッケージ `pkg/cmd/dataquery`。
- **P1** submit ボディに必須 `output.target:"S3"` を追加。
- **P1** S3 ダウンロードを厳密化（token 無し client・https/host 検証・リダイレクト拒否・`DisableCompression`・ctx・非2xx拒否・アトミック rename）。
- **P1** `run` の timeout は既存 `order job-status` と同名 `--interval`/`--timeout` に統一（global `--timeout` との二重定義回避、同一 ctx が submit+poll+download を束ねる）。
- **P2** check-success は既定のまま（`WithoutCheckSuccess` 不要）。multi-entity は明示 defer。download/near-miss テストを追加。
- **P3** `cmdtest` で新 env を中和、`list` に `--status`/`--page-size`。env fail-safe 方向・read-only ゲート形は妥当と確認。

**Codex 第2パス（更新版プランへの再レビュー）で追加反映:**
- **P1** S3 ダウンロードのハング防止 → Transport は `http.DefaultTransport` を Clone（既定の各種 timeout を維持）+ `--download-timeout`（既定 10m）で必ず有限締切。`io.Copy`/`Close`/`Rename` の各エラーを確認。
- **P2** ダウンロードのテスト容易性 → `downloadDataFile` に `*http.Client` 注入シーム（テストは `httptest.NewTLSServer().Client()`。`http://` の `NewServer` は https 必須で弾かれるため）。
- **P2** SQL 入力は **位置引数 `<SQL>` と `--file` の排他**（`MaximumNArgs(1)`+validator）。
- **P2** `list` は手書きをやめ **`listcmd.New`+`Spec{ItemsKey:"data"}`**（`AGENTS.md` の宣言的ランナー指針）。
- **P2** read-only ブロック時の **Data Query 専用ヒントを必須化**（`ReadOnlyError.Hint`）。
- **P2** DELETE near-miss の矛盾を訂正（`purge` は matcher 上「許可」側＝ブロック near-miss から除外。UUID narrow は live 後の任意ハードニング）。
- **P3** `run --output -`（stdout ストリーム）対応、timeout のフェーズ区別（poll/download）、キュー制限ヒント、exit code 規約を明記。
- **確認:** フラット package は `auth`/`config` の先例どおり妥当。check-success 既定 OK。read-only ゲートは fail-closed で妥当。POST submit の Idempotency-Key はプロセス内リトライの重複投入防止に有用（手動再実行は新キー＝別ジョブ）。

**Codex 第3パス:** CONVERGED（実質的指摘なし）。**Codex 第4パス（観点を変えたフレッシュなアドバーサリアル、非決定性対策）で追加反映（P1 なし・P2×5）:**
- globalflags は `SetReadOnly`/`SetReadOnlyAllowDataQuery` を**無条件設定**にして再 Apply で状態が残らないよう冪等化（+回帰テスト）。
- `list` の status クエリ名は **`queryStatus`**（`--status` フラグ→`Query:"queryStatus"`）、`pageSize`、`data` は配列（web 確認）。
- 重複ジョブ対策に **`--idempotency-key`** を任意追加し、曖昧失敗時にキーを surface。
- `failed` の理由は `errorMessage`/`message`/`error` フォールバック + `--json` で `data` 生出力。
- `--output -` と整形フラグ併用を拒否（stdout 混在防止）。
- **再確認（問題なし）:** read-only ゲートは checkHost/auth/Idempotency-Key 付与より前に走り、ブロックされた書き込みは送信されない（オフホスト絶対 URL も checkHost で停止）。run の timeout 合成（global ctx を local `--timeout` で narrow）は正しく、`defer cancel()` を poll/download それぞれに付ける。

**Codex 第5パス（P1 なし・P2×2、最新追加分の整合確認）:**
- `--output -` の拒否は **ユーザーが明示した整形フラグのみ**で判定（`default_output=json` の暗黙 `--json` 注入で誤拒否しないよう `Changed()` 基準）。
- `TestRootGlobalFlags`（`root_test.go:63`）のハードコード flag 一覧に `read-only-allow-data-query` を追加。
- **確認:** `WithHeader("Idempotency-Key",…)` は `Header.Set` で自動キーを上書き（重複しない）。無条件 setter で sticky 状態解消。`extractPath` は matcher と一致。`ReadOnlyError` テストは substring なので Hint 追加で壊れない。

**Codex 第6パス（P1 なし・P2×2、第5パス案の自己訂正）:**
- `--output -` は「拒否」方式が誤り（pflag は programmatic `Set` も `Changed=true` にし、`PersistentPreRunE→RunE` 順で区別不能）→ **`--output -` では stdout に job メタデータを一切レンダリングしない**（生バイトのみ）方式へ変更。整形フラグは stdout に無効。
- 再 Apply 回帰は **`readOnly=true, allow=true`→`readOnly=true, allow=false` で `POST /query/jobs` が再ブロック**を検証（「両 false」では無意味）。`SetReadOnly` true→false 単体リセットも別途。

---

## Part 1 — read-only ガードに「Data Query 許可の判断」を追加

read-only 判定は `Do`（`client.go:232-234`）が `isReadOnlyAllowed(method, path)`（`client.go:476-498`）で行う。`extractPath` で小文字化＋先頭スラッシュ除去されるため、判定キーは `query/jobs` / `query/jobs/{id}`（`v1/` 前置きなし。`commerce/...` と同形式）。

**方針:** `isReadOnlyAllowed` の常時許可セット（GET・既存 POST allowlist）には **Data Query を入れない**。代わりに Data Query 専用エンドポイントを識別する関数と、オプトインのトグルを追加し、`Do` で条件分岐する。

1. **Data Query エンドポイント識別**（新規・純関数。`isReadOnlyAllowed` は変更しない）:
   ```go
   // isDataQueryWrite は read-only で「明示許可した場合のみ」通す Data Query の
   // 書き込み系（投入 POST / 取消 DELETE）を識別する。テナントデータは変更しない。
   func isDataQueryWrite(method, path string) bool {
       p := extractPath(path)
       switch strings.ToUpper(method) {
       case http.MethodPost:
           return p == "query/jobs"
       case http.MethodDelete:
           return dataQueryJobPattern.MatchString(p) // ^query/jobs/[^/]+$
       }
       return false
   }
   var dataQueryJobPattern = regexp.MustCompile(`^query/jobs/[^/]+$`)
   ```
2. **オプトイン状態を Client に持たせる**:
   - フィールド `readOnlyAllowDataQuery bool`（`client.go:26-41` の `Client` に追加）。
   - 設定メソッド `func (c *Client) SetReadOnlyAllowDataQuery(v bool) { c.readOnlyAllowDataQuery = v }`（`SetReadOnly` の隣）。
3. **`Do` のガードを条件分岐**（`client.go:232-234`）:
   ```go
   if c.readOnly && !isReadOnlyAllowed(method, path) {
       if !(c.readOnlyAllowDataQuery && isDataQueryWrite(method, path)) {
           return nil, &ReadOnlyError{Method: method, Path: path}
       }
   }
   ```

**設計上の注意（fail-closed 維持）:** 既定（トグル off）では Data Query も従来どおりブロック。トグル on でも広がるのは `POST query/jobs` と `DELETE query/jobs/{id}` の2つだけ（near-miss は引き続きブロック）。PUT/PATCH や他の write は影響を受けない。

## Part 1b — オプトインのトグル配線（`pkg/cmd/globalflags/globalflags.go`）

既存の `--read-only` / `ZR_READ_ONLY`（`globalflags.go:33` 登録、`114-117` precedence、`187-189` で `client.SetReadOnly`、`196-212` の `EnvReadOnly`）と**同形**で実装:

1. **フラグ登録**（`Register`、`globalflags.go:33` 付近）:
   ```go
   cmd.PersistentFlags().Bool("read-only-allow-data-query", false,
       "In read-only mode, also allow Data Query submit/cancel (POST /query/jobs, DELETE /query/jobs/{id})")
   ```
2. **環境変数パーサ** `EnvReadOnlyAllowDataQuery()`（`EnvReadOnly` を踏襲、ただし **fail-safe の向きは逆**）:
   - truthy（`1/true/yes/on/...`）→ true、それ以外（未認識値含む）→ **false**。
   - 理由: これは安全網を**緩める**オプトイン。未認識値で誤って許可しないよう、保守側 = false に倒す（`EnvReadOnly` は逆に未認識→true に倒している点と対照的）。
3. **precedence + 配線**（`Apply`、`globalflags.go:114-117` / `187-189` を踏襲）:
   ```go
   allowDQ, _ := cmd.Flags().GetBool("read-only-allow-data-query")
   if !cmd.Flags().Changed("read-only-allow-data-query") {
       allowDQ = EnvReadOnlyAllowDataQuery()
   }
   // 両方とも「無条件」に bool を設定する（Codex 第4パス P2）。現状 SetReadOnly は
   // `if readOnly { ... }` と条件付きで、同一 factory に Apply が複数回かかると前回の
   // 状態が残りうる（sticky-wrapper edge）。false も明示設定して冪等にする:
   client.SetReadOnly(readOnly)
   client.SetReadOnlyAllowDataQuery(allowDQ)
   ```
   回帰テスト: 同一 factory に `readOnly=true, allow=true` を Apply→次に `readOnly=true, allow=false` を Apply し、2回目で `POST /query/jobs` が再びブロックされることを検証（read-only off では DQ write は元々通るため、必ず read-only=true のまま allow を false に落として確認する）。
4. **テストハーネス**（`pkg/cmdtest/cmdtest.go:34-42`）: 既存の `ZR_READ_ONLY` 中和に加えて `ZR_READ_ONLY_ALLOW_DATA_QUERY` も中和し、開発者の環境変数がテストへ漏れないようにする。

**UX（必須／Codex P2）:** read-only かつトグル off で Data Query がブロックされたとき、汎用の「Remove --read-only ...」だけでは正しい解（`--read-only-allow-data-query`）に辿り着けない。`ReadOnlyError`（`internal/api/errors.go:162-176`）に `Hint` フィールドを足し、`Do` で `c.readOnly && isDataQueryWrite(method,path) && !c.readOnlyAllowDataQuery` のとき **Data Query 専用ヒント**（`--read-only-allow-data-query` / `ZR_READ_ONLY_ALLOW_DATA_QUERY=1` を案内）を載せる。これは任意ではなくコア UX として実装する。

---

## Part 2 — 新コマンド `zr data-query`

**レイアウト（import cycle 回避＝Codex P1）:** ヘルパー（投入ボディ生成・`data` unwrap・terminal-status・S3 ダウンロード）を submit/run/get 等で共有するため、**フラット単一パッケージ** `pkg/cmd/dataquery`（親 + 全サブコマンドを `package dataquery` の複数ファイルで）にする。`billrun` の「サブパッケージ別」レイアウトは親が子を import するため、子が共有ヘルパーを親から import すると循環する。フラットなら循環せず共有も容易で、**既存の `pkg/cmd/auth`（`login.go`/`logout.go`/`status.go`/`token.go` がすべて `package auth`）や `config` と同じ先例**に沿う（Codex 第2パスで確認）。代替は `pkg/cmd/dataquery/dqutil` 子パッケージ。

**ファイル構成（すべて `package dataquery`）:**
- `dataquery.go` — 親 `NewCmdDataQuery(f)`（`Use: "data-query <command>"`）。`pkg/cmd/root/root.go:95` 付近に `cmd.AddCommand(dataquerycmd.NewCmdDataQuery(f))` を追加。
- `submit.go` / `get.go` / `list.go` / `cancel.go` / `run.go`
- `helpers.go` — 共有: (a) フラグ→投入ボディ生成、(b) `unwrapData(raw)`（`{"data":{...}}` を降りる）、(c) `isTerminalStatus`、(d) `downloadDataFile`。

**投入ボディの形（Codex P1：`output.target` 必須）:**
```json
{ "query": "<SQL>", "outputFormat": "JSON", "compression": "NONE",
  "output": { "target": "S3" } }
```
`output` / `output.target` は **必須**（Zuora 仕様）。任意で `columnSeparator`（DSV）、`sourceData`（LIVE/WAREHOUSE）、`readDeleted`、`useIndexJoin`。**正確な JSON キー名は実装前に API リファレンスで確定**し、submit テストで body 形を厳密検証する。

**サブコマンド設計:**

| コマンド | メソッド/パス | 実装 | 備考 |
|---|---|---|---|
| `submit "<SQL>"` | `POST /query/jobs` | `cmdutil.RunDetail`（`Action{Method:"POST", Path:"/query/jobs", Body, Fields: unwrapData→id/queryStatus...}`） | 非同期。job id を即返す。**check-success は既定のまま**（Data Query 応答に `success` キーが無く誤検知しない＝`checksuccess_test.go:58` で実証済み。`WithoutCheckSuccess` は付けない／Codex P2） |
| `get <job-id>` (alias `status`) | `GET /query/jobs/{id}` | `cmdutil.RunDetail` | id/queryStatus/outputRows/processingTime/`dataFile` を表示。GET は元々 read-only OK |
| `list` | `GET /query/jobs` | **`listcmd.New`+`listcmd.Spec`**（`ItemsKey:"data"`／`data` は配列。手書きにしない＝`AGENTS.md:40,82`） | `--status`→`Flag{Query:"queryStatus", Enum:[accepted/in_progress/completed/failed/cancelled]}`（**クエリ名は `status` でなく `queryStatus`**／Codex 第4パス web 確認）、`--page-size`→`Flag{Query:"pageSize", Int:true, OmitZero:true}`。`nextPage`/hint は listcmd 標準 |
| `cancel <job-id>` | `DELETE /query/jobs/{id}` | `cmdutil.RequireConfirm` + `client.Delete` + `cmdutil.RenderDeleteResult` | `cmdutil.AddConfirmFlag(cmd,&confirm,"cancellation")`。`billrun cancel` を踏襲。id は `url.PathEscape` |
| `run "<SQL>"` | submit + poll + (任意) download | **手書きランナー**（`detail.go:48-50` が polling を変種として明記。`order job-status` を踏襲） | 下記 |

**`run` の挙動（`pkg/cmd/order/job-status/job_status.go` を踏襲）:**
1. flags: `--interval`（既定 5s）/ `--timeout`（既定 0=無制限）。`order job-status` と**同名・同既定**にして一貫性を持たせる（独自 `--poll-interval` 等は作らない／Codex P1 の timeout 二重化を回避）。`--timeout>0` のとき `ctx, cancel = context.WithTimeout(cmd.Context(), timeout)`。
2. `client.SetContext(ctx)` で submit/poll/（後述）download まで**同一 ctx** が deadline と Ctrl-C を担保（`main.go:32-35` の signal context + globalの `zr --timeout` とも合成）。
3. `POST /query/jobs` 投入 → `unwrapData` で job id 取得。
4. `queryStatus` が終端（`completed`/`failed`/`cancelled`）まで `GET /query/jobs/{id}` をポーリング。待機は **`cmdutil.SleepContext(ctx, interval)`**（生 `time.Sleep` 禁止）。ポーリングの deadline 超過は `order job-status` 形でラップし、**キュー制限のヒントを付ける（Codex P3）**: 例「gave up waiting for job <id> after <t> (last status: accepted; job may still be queued by tenant concurrency limits — use `get <id>` or raise `--timeout`)」。
5. `completed`: 行数・処理時間・`dataFile` URL をサマリ（stderr）。`--output <file>` 指定時は `downloadDataFile` で取得。`--output -` は **stdout へ生バイトのみ**・サマリ/進捗は stderr のみ。**この時は job メタデータを stdout にレンダリングしない**（`output.Render*` を呼ばず、`--json`/`--jq`/`--template`/`--csv` は stdout に作用しない＝ヘルプに明記）。これで生データとメタデータの混在は構造的に起きず、拒否ロジック自体が不要。
  - **なぜ「拒否」にしないか（Codex 第5/6パス）:** `globalflags.Apply` は非 TTY + `default_output=json` のとき `cmd.Flags().Set("json","true")`（`globalflags.go:75`）を呼び、**pflag は programmatic Set も `Changed=true` にする**。`PersistentPreRunE`→`RunE` の順で走るため `run` の RunE 時点では `Changed("json")` が既に true で、暗黙注入とユーザー明示を区別できない。したがって「`--output -` と整形フラグ併用を拒否」する方式は誤検知する。代わりに上記の「`--output -` では stdout=生バイトのみ・メタデータ非レンダリング」で回避する。ダウンロード中の deadline は「download timed out」系に**フェーズを区別**して表現（ポーリングの文言と混同しない／Codex P3）。
6. `failed`/`cancelled`: 非ゼロ終了。失敗理由は **`errorMessage` 単一キーに依存せず** `errorMessage`/`message`/`error` のいずれか + job id/status を surface し、`--json` では `data` 生オブジェクトを出す（実フィールド名が docs 未確定のため／Codex 第4パス P2）。**exit code 規約**: deadline=exit 1（`context.DeadlineExceeded`）、Ctrl-C=exit 130（`context.Canceled`）、read-only ブロック=exit 5。

**`downloadDataFile(ctx, url, dstPath, httpClient)`（S3 取得、Codex P1/P2：セキュリティ・バイト保全・タイムアウト・テスト容易性）:**
- `dataFile` は **Zuora 以外のホスト（S3 署名付き URL）**。`api.Client` は使わない（`checkHost`（`client.go:206-228`）がオフホストを拒否し、Bearer トークンを外部へ送ってはならない）。
- URL を `url.Parse` し、**`https` 必須・`Host` 非空・userinfo 無し**を検証。
- **本番クライアントは `Authorization` 無し**。Transport は **`http.DefaultTransport.(*http.Transport).Clone()`** をベースに `DisableCompression = true`（DialContext/TLSHandshakeTimeout/proxy 等の既定を失わない。署名 URL の gzip 自動展開でバイトが壊れるのを防ぐ。特に `--compression GZIP/ZIP` で圧縮ファイルを保つ）。`CheckRedirect` は **リダイレクト拒否**（署名クエリの外部ホスト漏洩防止）。
- **ハング防止（Codex P1）:** `--timeout 0` だとポーリングに deadline が無くダウンロードも無制限になりうる。ダウンロードは `context.WithTimeout(ctx, downloadTimeout)`（`--download-timeout`、既定 10m）で必ず有限の締切を持たせる。
- **テスト容易性（Codex P2）:** `downloadDataFile` は `*http.Client` を**引数で注入可能**にする（本番は上記ハード化済みクライアント、テストは `httptest.NewTLSServer().Client()` を渡して自己署名証明書を信頼）。`https` 必須チェックは `https://127.0.0.1` を通すので TLS テストサーバで検証できる。
- `http.NewRequestWithContext(ctx, GET, ...)` → 非 2xx は拒否 → `io.Copy` で temp へストリーム → `query.go:104-128` の **temp + `os.Rename` アトミック**。`io.Copy`/`Close`/`Rename` の各エラーを確認してから既存ファイルを置換し、失敗時は temp 削除＋既存ファイル保持。

**フラグ（投入系、submit と run で共有）:**
- SQL 入力: **位置引数 `<SQL>` か `--file <path>` の排他（どちらか一方必須／Codex P2）**。`Args: cobra.MaximumNArgs(1)` + RunE 冒頭で「両方指定／両方無し」をエラーにする validator（submit と run の両方に適用）。
- `--output-format`（`CSV`/`JSON`/`TSV`/`DSV`、既定 `JSON`=JSON Lines）→ Data Query の `outputFormat`。**CLI 表示用の `--json`/`--csv` グローバルフラグ（ジョブメタデータの表示形式）とは別物**（命名衝突回避で `--output-format`）。enum は `cmdutil.EnumCompletion`。
- `--compression`（`NONE`/`GZIP`/`ZIP`、既定 `NONE`）、`--column-separator`（DSV 用）、`--source`（`LIVE`/`WAREHOUSE`、既定 `LIVE`）、`--read-deleted`、`--use-index-join`。
- **`--idempotency-key`（任意／Codex 第4パス P2）:** `POST /query/jobs` は `Idempotency-Key` を受け付ける（`api.WithHeader`）。クライアントはプロセス内リトライ（401/success-envelope/429）では同一キーを再利用し重複投入を防ぐが、**ジョブ生成後に非リトライの transport error が返ると手動再実行で別キー＝ジョブ二重作成**になりうる。明示キーを渡せるようにし、曖昧な POST 失敗時は使用キーを stderr に surface する。
- `run` のみ: `--output <file>`（`-` で stdout ストリーム）、`--interval`、`--timeout`、`--download-timeout`（既定 10m）。

**再利用するもの:** `cmdutil.RunDetail` / `cmdutil.RenderDeleteResult` / `cmdutil.RequireConfirm` / `cmdutil.AddConfirmFlag` / `cmdutil.SleepContext` / `cmdutil.GetString`・`GetDecimal` / `cmdutil.EnumCompletion`、`output.FromCmd`・`Render`・`RenderDetail`・`Column`・`DetailField`、`query.go` のアトミック export パターン、`order job-status` のポーリング骨格。

**スコープ外（明示的に defer／Codex P2）:** Multi-entity / Multi-Org（`Zuora-Entity-Ids` / `Zuora-Org-Ids`）。CLI にグローバルな entity 選択機構は無く（per-request `WithHeader` のみ、`pagination.go:31`）、当面は `zr api --header` 経由で対応可能。将来 `--entity-id` 追加を別タスクとする旨を README に明記。

---

## Part 3 — テスト

1. **read-only ガード単体**（`internal/api/client_test.go`、既存 `TestClient_ReadOnly_*` を踏襲。`SetReadOnly`/`SetReadOnlyAllowDataQuery` を組み合わせる）:
   - `readOnly=true, allowDQ=false`（既定）：`POST /query/jobs` も `DELETE /query/jobs/job-123` も**ブロック**。
   - `readOnly=true, allowDQ=true`：`POST /query/jobs` と `DELETE /query/jobs/job-123` が**許可**。
   - `readOnly=true, allowDQ=true` でも **near-miss はブロック**：`POST /query/jobs/abc`（id 付き POST）、`DELETE /query/jobs`（id 無し collection）、`PUT /query/jobs/job-123`、複数セグメント `DELETE /query/jobs/abc/def`、末尾スラッシュ `DELETE /query/jobs/abc/`、そして **`POST /v1/accounts`（通常 write）** は依然ブロック（トグルが Data Query 以外を広げないことの確認）。
   - **DELETE matcher の幅（Codex P2、訂正）:** `^query/jobs/[^/]+$` は**単一セグメントの任意値**を許可するため、`DELETE /query/jobs/purge` のような英数字 1 語は（仮に存在すれば）**許可されてしまう**＝これは near-miss(ブロック)ではない。charset では action 語と id を区別できないので、purge を「ブロックされる near-miss」には**含めない**。許容する理由: Zuora の Data Query DELETE は `/query/jobs/{id}` のみで、コマンドは実 job-id（`url.PathEscape`）しか送らない＋トグルは明示オプトイン。**ライブ検証で job-id が正準 UUID と判明したら matcher を UUID 形に狭める**（任意ハードニング、open items 参照）。
   - `readOnly=false`：トグルに関わらず全許可。
2. **globalflags**（`pkg/cmd/globalflags/readonly_test.go` を踏襲）:
   - `EnvReadOnlyAllowDataQuery` の truthy/falsy + **未認識値→false**（保守側）。
   - フラグ > env の precedence、`Apply` が `SetReadOnlyAllowDataQuery` を正しく呼ぶこと（`root_test.go` の `TestRootReadOnly*` 形式の統合テスト：`--read-only --read-only-allow-data-query` で data-query submit が通り、`account create` はブロックされる）。
   - **既存テストの更新（Codex 第5パス P2）:** `TestRootGlobalFlags`（`root_test.go:63`）は persistent flag 一覧をハードコードしているので `read-only-allow-data-query` を追加。
   - **再 Apply 冪等性（Codex 第4/6パス）:** 意味のある回帰は **`readOnly=true, allow=true` → 再 Apply `readOnly=true, allow=false` の後で `POST /query/jobs` が再びブロックされる**こと（read-only off だと DQ write はそもそも通るので「両 false」では証明にならない）。加えて `SetReadOnly` の true→false リセット単体テストも別に持つ。
3. **コマンド層**（`cmdtest.Run`/`cmdtest.OK`、`pkg/cmd/query/query_test.go` を踏襲。ハーネスは `ZR_READ_ONLY` と `ZR_READ_ONLY_ALLOW_DATA_QUERY` を無効化）:
   - submit：POST ボディ形状（`query`/`outputFormat` 等）と `data.id` 表示。
   - get：`data` を降りて status/dataFile を表示。
   - cancel：`--confirm` 無しでエラー、有りで `DELETE` 発行。
   - run：httptest で「`accepted`→`in_progress`→`completed`」を返すモック + `dataFile` を別ハンドラで配信し、`--output` でファイル生成を検証。`--timeout` 到達で「gave up waiting」エラーも検証。
   - **download のセキュリティ/正しさ（Codex P2）:** `downloadDataFile` の注入可能 `*http.Client` シームに **`httptest.NewTLSServer().Client()`** を渡して（`https` 必須チェックを満たす）、dataFile リクエストに **`Authorization` ヘッダが無い**こと、**クロスホストのリダイレクトを辿らない**こと、ctx キャンセルを尊重すること、**gzip バイトが保全**される（`DisableCompression`）こと、非 2xx で rename されず既存ファイルが無傷なこと、`--output -` で stdout に生バイトが出ること、を明示アサート。
4. CI: `make ci`（`staticcheck`/`govulncheck`/coverage floor 73%）をローカルで通す。

---

## Part 4 — ドキュメント

- `README.md` に `zr data-query` セクション（submit/get/list/cancel/run）と、**read-only 時は既定でブロック・`--read-only-allow-data-query`（または `ZR_READ_ONLY_ALLOW_DATA_QUERY=1`）でオプトイン許可**である旨を明記。read-only の説明（`--read-only`/`ZR_READ_ONLY`）にトグルを追記。
- 必要なら `AGENTS.md` の read-only 説明に「Data Query の POST/DELETE は read-only では既定ブロック、専用トグルでのみ許可」を追記。
- E2E は方針どおりバッシュ手動ゲート（`project_e2e_tooling_decision` 参照）。任意で `tests/e2e-dataquery.sh` を追加可能だが必須ではない。

---

## Verification（実行時）

```bash
make build
# 既定: read-only では Data Query もブロック（安全網維持）
ZR_READ_ONLY=1 ./bin/zr data-query submit "SELECT accountnumber FROM account"   # → exit 5 / blocked
# オプトインで Data Query のみ許可（核心）
ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=1 \
  ./bin/zr data-query submit "SELECT accountnumber, balance FROM account WHERE balance > 100"
./bin/zr --read-only --read-only-allow-data-query \
  data-query run "SELECT accountnumber FROM account" --output /tmp/dq.json
ZR_READ_ONLY=1 ./bin/zr data-query get <job-id>          # GET は元々 OK（トグル不要）
ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=1 ./bin/zr data-query cancel <job-id> --confirm
# トグルを立てても通常 write はブロックされ続ける（退行防止の要）
ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=1 ./bin/zr account create --body '{}'   # → exit 5 / blocked
make ci   # staticcheck + govulncheck + coverage floor
```
ユニット/コマンドテスト：`go test ./internal/api/... ./pkg/cmd/globalflags/... ./pkg/cmd/dataquery/...`

---

## Open items / 実装前にライブ検証すべき点

Codex が Zuora の API リファレンスで一次確認した結果と、なお実機確認が望ましい点:
- **確認済み（Codex + docs）:** パスは `/query/jobs` と `/query/jobs/{job-id}`（**`/v1/` 前置きなし**）。応答は `{"data": ...}` 封筒で、completed 時 `dataFile` は `data` 配下。statuses は `accepted/in_progress/completed/failed/cancelled`。submit ボディは `output.target` 必須。→ Part 1 の allowlist キー `query/jobs` と一致（`extractPath` 出力＝小文字・先頭スラッシュ無し）。
- **check-success は対応不要（解決済み）:** 既定の success-flag 検査でよい。Data Query 応答に top-level `success` が無く、`TestCheckSuccess_NoSuccessField_PassesThrough`（`checksuccess_test.go:58`）が通過を実証。`WithoutCheckSuccess` は使わない。

実装前に live tenant で確認（このリポジトリの慣例＝実機 probe で確定）:
1. **投入ボディの正確なキー名**：`sourceData` / `columnSeparator` / `readDeleted` / `useIndexJoin` の綴りと `output` の必須サブフィールドを API リファレンスで確定し、submit テストで body 形を厳密検証。
2. **`data` 配下のフィールド名**：`queryStatus`（vs `status`）、`outputRows`、`processingTime`、`dataFile`、失敗時 `errorMessage` の実名を確定。
3. **cancel の応答形状**：204 か JSON か（`RenderDeleteResult` は両対応済み）。
4. **S3 `dataFile` の実挙動**：Content-Encoding / リダイレクト有無 / オブジェクトヘッダ（Codex も live 未確認）。`downloadDataFile` の `DisableCompression`・リダイレクト拒否で安全側に倒してあるが、実取得で 1 回検証。
5. `--output-format` 既定値（`JSON`=JSONL を提案。実運用に合わせ調整可）。
