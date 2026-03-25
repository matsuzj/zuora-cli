# Phase 5: Commerce — 未確認項目

Phase 5 の実装は完了したが、以下の項目はテナント制約により E2E テストで未確認。
Commerce API が有効なテナントで手動検証が必要。

## 未確認項目

### Commerce API エンドポイント（テナント未対応で SKIP）

- [ ] `zr product create` — `POST /commerce/products` の実 API 呼び出し
- [ ] `zr product update` — `PUT /commerce/products` の実 API 呼び出し
- [ ] `zr product get <key>` — `GET /commerce/products/{product_key}` の実 API 呼び出し
- [ ] `zr product list-legacy` — `POST /commerce/legacy/products/list` の実 API 呼び出し
- [ ] `zr plan create` — `POST /commerce/plans` の実 API 呼び出し
- [ ] `zr plan update` — `PUT /commerce/plans` の実 API 呼び出し
- [ ] `zr plan get --key` — `POST /commerce/plans/query` の実 API 呼び出し
- [ ] `zr plan list` — `POST /commerce/plans/list` の実 API 呼び出し
- [ ] `zr plan purchase-options --plan` — `POST /commerce/purchase-options/list` の実 API 呼び出し
- [ ] `zr charge create` — `POST /commerce/charges` の実 API 呼び出し
- [ ] `zr charge update` — `PUT /commerce/charges` の実 API 呼び出し
- [ ] `zr charge get --key` — `POST /commerce/charges/query` の実 API 呼び出し
- [ ] `zr charge update-tiers` — `PUT /commerce/tiers` の実 API 呼び出し

### 確認済み項目（参考）

- [x] 全14コマンドのバリデーションテスト（必須フラグ/引数の欠落） — E2E PASS
- [x] `cobra.NoArgs` による余分な位置引数の拒否 — E2E PASS
- [x] `zr rateplan get` — v1 API `/v1/rateplans/{id}` の実 API 呼び出し — E2E PASS
- [x] `--json` / `--jq` 出力フォーマット — E2E PASS
- [x] エラーハンドリング（存在しないリソースの取得） — E2E PASS
- [x] 全14コマンドのユニットテスト（HTTP メソッド・パス・ヘッダー検証） — go test PASS

## 検証方法

Commerce API が有効なテナントに接続して以下を実行:

```bash
# 1. 環境切り替え
zr auth login  # Commerce API 有効テナントに接続

# 2. E2E テスト再実行
bash tests/e2e-commerce.sh

# 3. Step 6 の product list-legacy / plan list が PASS になることを確認
```
