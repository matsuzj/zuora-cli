# Phase 7: Usage + Meter

**依存**: Phase 2

## 実装チェックリスト

### Usage

- [ ] `zr usage post --file <path>` — `POST /v1/usage` (multipart/form-data で CSV ファイルアップロード。--file 必須。非同期: checkImportStatus レスポンスを返す)
- [ ] `zr usage create` — `POST /v1/object/usage`
- [ ] `zr usage get <id>` — `GET /v1/object/usage/{id}`
- [ ] `zr usage update <id>` — `PUT /v1/object/usage/{id}`
- [ ] `zr usage delete <id>` — `DELETE /v1/object/usage/{id}`

### Meter

- [ ] `zr meter run <meterId> <version>` — `POST /meters/run/{meterId}/{version}`
- [ ] `zr meter debug <meterId> <version>` — `POST /meters/debug/{meterId}/{version}`
- [ ] `zr meter status <meterId> <version>` — `GET /meters/{meterId}/{version}/runStatus`
- [ ] `zr meter summary <meterId> --run-type <type>` — `POST /meters/{meterId}/summary` (--run-type 必須、--body でグルーピング条件等を指定)
- [ ] `zr meter audit <meterId>` — `GET /meters/{meterId}/auditTrail/entries` (--export-type, --run-type, --from, --to 全て必須)
