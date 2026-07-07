# 監査 ground truth — 検証済み非問題(再フラグ禁止台帳)

監査・バグハント・fixture 忠実性チェックを行うエージェントは、**着手前にこの
台帳を読み、ここにある項目を再フラグしないこと**。各項目は live プローブまたは
公式リファレンスで確定済みで、検証手段と日付を伴う。新たに「非問題」を確定
させた監査は、同じ形式(何を・どう検証したか・日付)でここに追記する。

背景: 2026-06-30 の 27 エージェント triage は FALSE_POSITIVE バケット 6 件中
4 件を誤分類していた(記憶から Zuora フィールド名を推測)。応答形状の問題を
確定できるのは live プローブだけであり、その結果がセッション間で失われると
同じ誤検知が再生産される — この台帳はそれを防ぐ。

## Live プローブ確定済みの応答形状(2026-06-30, apac-sandbox)

- `GET /v1/subscriptions/{key}` — **top-level `name` は存在しない**。識別子は
  `subscriptionNumber`(#438 で修正済み)。
- `GET /v1/creditmemos` — `unappliedAmount` を持ち、**`balance` は無い**。
  debitmemos は逆に `balance` を持つ(#418)。
- creditmemo の status enum は **`Canceled`**(米綴り・l 1 つ)(#422)。
- `/v1/accounts/{key}/payment-methods` の `creditcard[]` — マスク番号は
  **`cardNumber`**(`"************1111"` 形式)。`creditCardMaskNumber` は
  ZOQL 側のフィールドで REST 応答には無い(#421)。
- `/v1/async-jobs/{jobId}` — 根は `{status, errors, result, success}`。
  jobId/orderNumber/accountNumber は根に**無く** `result` 配下にネスト
  (commit 系 = `{orderNumber, accountNumber, subscriptions[]}` / preview 系 =
  `{invoices, creditMemos}`)。`result` は object なので GetString 厳禁
  (#419/#460 — preview-async の非破壊 dry-run で live 確認)。
- contact は `address2` + `zipCode` を持つ(#427、live 確認)。
  ※ contact-**snapshot** 側の `postalCode` 綴りは live 未検証の仮定なので
  この台帳には載せない — `make pending-live` のマーカーが管理する。

## 確定 FALSE POSITIVE(再フラグ禁止)

- **order activate/create/cancel が top-level `"status"` を読むのは正しい** —
  Zuora の Activate-Order 応答は `status`(Draft/Pending/Completed/Scheduled)
  を根に持つ(developer.zuora.com で確認、2026-06-28)。
- **account payment-methods の項目が `isDefault` を持つのは正しい** —
  #56 で live 確認済み。
- **invoice items の `chargeAmount` は JSON number(float)** — float64 の
  まま扱うのが正しい(#415 を FP クローズ、live 確認)。
- **`/query/jobs`(Data Query jobs list)に `nextPage` カーソルは無い** —
  pageSize=1・40+ ジョブでプローブしても現れない。カーソル・ページネーションは
  実装不能であり「More results ヒントが出ない」のは仕様(#437 FP)。
- **`DELETE /v1/async/orders/{key}` は正しいメソッド** — 偽 id への DELETE は
  HTTP 400 order-not-found(=メソッド受理)、PUT は 405。「PUT が正」と
  主張していた検証メモの方が誤りだった(#417 FP、#452 で api-reference を訂正)。

## プロセス上の教訓(監査ワークフローの設計者向け)

- **triage の FALSE_POSITIVE バケットを信用しない**。応答形状の主張は live
  プローブでのみ確定する(上記 4/6 誤分類の実績)。
- **live プローブは doc の主張に勝つ** — プロジェクト自身の検証メモ含む
  (#417 の教訓: メモに従って「修正」していたら動くコマンドを壊していた)。
- 非破壊プローブの実績ある手口: 偽 id への PUT/DELETE でルート存在確認
  (リソース別の 404 コード: payment cancel 50000040 / unapply 53840040 /
  transfer 53830040)、preview-async を dry-run として使う、pageSize=1 で
  ページネーション有無を見る。**書き込みを伴うプローブは
  `scripts/require-sandbox.sh` を先に通すこと。**
- **加算的なフィールド追加(キー置換ではない)は live 未検証でも安全に出荷
  できる** — 未知キーは空欄レンダーになるだけ(#433 の判断根拠)。
- 監査は必ず `origin/main` の worktree に対して行う(stale WIP checkout への
  監査が修正済みバグ 2 件を「confirmed」と再報告した実績)。実装前は
  `scripts/preflight-issue.sh` を使う。

## 未検証の仮定はここに書かない

「まだ確定していない」ものはこの台帳ではなく `LIVE-UNVERIFIED` マーカー
(`make pending-live` が一覧、docs/e2e-test-skips.md に生成台帳)で管理する。
この台帳は**確定した非問題**専用 — 二重管理しない。
