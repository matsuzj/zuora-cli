# Read-only mode — the full allowlist

`--read-only`(または `ZR_READ_ONLY`)は書き込み API 呼び出しをリクエスト送信前に
ブロックする(終了コード 5)。env は **fail-closed**: 空でない値が既知の falsy
綴り(`false`/`0`/`no`/`off`)でなければ有効化される。`--read-only` フラグが
env に優先する。設計の経緯は
[docs/plans/archive/read-only-mode.md](plans/archive/read-only-mode.md)(歴史的文書)。

以下の allowlist は `internal/api/client.go` の実データから
`cmd/gen-readonly-doc` が生成し、`make lint` がドリフトを検出する
(README の destructive-list と同じ方式)。手で編集しないこと —
変更はゲート実装側(`client.go`)に行い、`go run ./cmd/gen-readonly-doc` の
出力でマーカー区間を更新する。

<!-- readonly-allowlist:begin -->
Under `--read-only` / `ZR_READ_ONLY` the API client allows:

- **GET / HEAD / OPTIONS** — always allowed.
- **POST** — allowed only for these read-only endpoints (exact match, after path normalization):
  - `v1/action/query`
  - `v1/action/querymore`
  - `commerce/charges/query`
  - `commerce/plans/query`
  - `commerce/plans/list`
  - `commerce/purchase-options/list`
  - `commerce/legacy/products/list`
  - `v1/orders/preview`
  - `v1/async/orders/preview`
  - `v1/subscriptions/preview`
- **POST** — allowed for these dynamic-path patterns (regex match):
  - `^v1/subscriptions/[^/]+/preview$`
  - `^meters/[^/]+/summary$`
  - `^commerce/products/[^/]+$`
- **PUT / DELETE / PATCH** — always blocked.
- **Data Query opt-in** (`--read-only-allow-data-query` / `ZR_READ_ONLY_ALLOW_DATA_QUERY`, default OFF — fails restrictive): additionally allows `POST query/jobs` (submit) and `DELETE ^query/jobs/[^/]+$` (cancel). Nothing else widens.
<!-- readonly-allowlist:end -->
