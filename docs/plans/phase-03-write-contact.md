# Phase 3: Account + Subscription (Write) + Contact

**依存**: Phase 2

## 実装チェックリスト

### Account — Write 操作

- [ ] `zr account create` — `POST /v1/accounts`
- [ ] `zr account update <key>` — `PUT /v1/accounts/{account-key}`
- [ ] `zr account delete <key>` — `DELETE /v1/accounts/{account-key}` (async, --confirm 必須)
- [ ] `zr account set-cascading <key>` — `PUT /v1/accounts/{account-key}/payment-methods/cascading`

### Subscription — Write 操作

- [ ] `zr subscription create` — `POST /v1/subscriptions`
- [ ] `zr subscription update <key>` — `PUT /v1/subscriptions/{subscription-key}`
- [ ] `zr subscription cancel <key>` — `PUT /v1/subscriptions/{subscription-key}/cancel`
- [ ] `zr subscription suspend <key>` — `PUT /v1/subscriptions/{subscription-key}/suspend`
- [ ] `zr subscription resume <key>` — `PUT /v1/subscriptions/{subscription-key}/resume`
- [ ] `zr subscription renew <key>` — `PUT /v1/subscriptions/{subscription-key}/renew` (--body 必須、更新条件 JSON)
- [ ] `zr subscription delete <key>` — `PUT /v1/subscriptions/{subscription-key}/delete`
- [ ] `zr subscription preview` — `POST /v1/subscriptions/preview`
- [ ] `zr subscription preview-change <key>` — `POST /v1/subscriptions/{subscription-key}/preview`
- [ ] `zr subscription update-custom-fields <num> <ver>` — `PUT /v1/subscriptions/{subscriptionNumber}/versions/{version}/customFields` (--body 必須。注意: version 指定必須。version なしの API は現行 v1 に存在しない)

### Contact

- [ ] `zr contact list --account-id <accountId>` — `POST /v1/action/query` で ZOQL `SELECT Id, FirstName, LastName, Email FROM Contact WHERE AccountId = '{accountId}'` を実行 (v1 に list API がないため ZOQL 代替。account number ではなく account ID を要求。account number からの解決が必要な場合は先に `zr account get` で ID を取得)
- [ ] `zr contact get <id>` — `GET /v1/contacts/{contactId}`
- [ ] `zr contact create` — `POST /v1/contacts`
- [ ] `zr contact update <id>` — `PUT /v1/contacts/{contactId}`
- [ ] `zr contact delete <id>` — `DELETE /v1/contacts/{contactId}`
- [ ] `zr contact transfer <id>` — `PUT /v1/contacts/{contactId}/transfer`
- [ ] `zr contact scrub <id>` — `PUT /v1/contacts/{contactId}/scrub`
- [ ] `zr contact snapshot <snapshot-id>` — `GET /v1/contact-snapshots/{contact-snapshot-id}` (注意: contact ID ではなく snapshot ID を指定)

### Sign Up

- [ ] `zr signup` — `POST /v1/sign-up` (アカウント+支払い+サブスク一括作成)

## コマンド詳細仕様

### account コマンド (Write)

| サブコマンド | API | フラグ | テーブル出力カラム |
|---|---|---|---|
| `create` | `POST /v1/accounts` | `--body` | 作成結果 (ID, Number) |
| `update <key>` | `PUT /v1/accounts/{key}` | `--body` | 更新結果 |
| `delete <key>` | `DELETE /v1/accounts/{key}` | `--confirm` | 削除結果 (async) |
| `set-cascading <key>` | `PUT /v1/accounts/{key}/payment-methods/cascading` | `--body` | 更新結果 |

### subscription コマンド (Write)

| サブコマンド | API | フラグ | テーブル出力カラム |
|---|---|---|---|
| `create` | `POST /v1/subscriptions` | `--body` | 作成結果 |
| `update <key>` | `PUT /v1/subscriptions/{key}` | `--body` | 更新結果 |
| `cancel <key>` | `PUT /v1/subscriptions/{key}/cancel` | `--policy` (必須), `--effective-date` (SpecificDate 時必須)。または `--body` で全フィールド指定 | 結果 |
| `suspend <key>` | `PUT /v1/subscriptions/{key}/suspend` | `--policy` (必須), `--suspend-date` (SpecificDate 時), `--periods`+`--periods-type` (FixedPeriodsFromToday 時)。または `--body` | 結果 |
| `resume <key>` | `PUT /v1/subscriptions/{key}/resume` | `--policy` (必須), `--resume-date` (SpecificDate 時), `--periods`+`--periods-type` (FixedPeriodsFromSuspendDate 時)。または `--body` | 結果 |
| `renew <key>` | `PUT /v1/subscriptions/{key}/renew` | `--body` (必須) | 結果 |
| `delete <key>` | `PUT /v1/subscriptions/{key}/delete` | `--confirm` | 結果 |
| `preview` | `POST /v1/subscriptions/preview` | `--body` | プレビュー結果 |
| `preview-change <key>` | `POST /v1/subscriptions/{key}/preview` | `--body` | プレビュー結果 |
