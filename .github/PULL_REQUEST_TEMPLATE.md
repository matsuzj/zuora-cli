<!-- この骨格は AGENTS.md「Autonomous pipeline & PR contract」の PR 本文契約
     (#528)。branch protection は必須レビュー 0 のため、PR 本文がマージ判断の
     唯一の信頼面 — 各節を省略しない(該当なしなら「なし」と書く)。 -->

## Summary

<!-- 1〜3 行: 何が変わるか。詳細な背景は Closes 先の issue に。 -->

## Deviations & deliberately-not-done

<!-- プラン/issue からの逸脱と、意図的にやらなかったこと。
     先送り(deferral)は open issue か LIVE-UNVERIFIED マーカーへの参照が必須。
     PR 散文だけの先送りはマージと同時に蒸発する(#521)。 -->

## Review gates

- make ci: <!-- green / 該当ゲートの結果 -->
- Codex review: <!-- パスごとの指摘数の推移(found→fixed)、2 連続クリーンで収束 -->
- live E2E: <!-- API/auth/出力挙動に触れる場合は N/M スイート green を必須で記載。
                対象外ならその理由(docs のみ、等)を明記 -->
- bite 証明: <!-- 新テスト/ゲートが実際に咬む証拠(revert で赤、等)。該当時のみ -->

<!-- Closes は 1 文 1 issue で書く(改行区切り)。
     「Closes #A, #B」のカンマ連結は最初の 1 件しか自動クローズしない
     (#444 → #438 で実際に不発)。 -->
Closes #
