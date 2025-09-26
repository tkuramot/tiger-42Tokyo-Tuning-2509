# 42Tokyo Tuning the backend Contest 2025

## ディレクトリ構造
```
.
├── .da          # SSL証明書などの保管場所
├── benchmarker  # ベンチマーカー
├── document     # 各種ドキュメント
└── webapp       # バックエンド、フロントエンド、Nginx、MySQL、E2Eテストの実装
```

## ドキュメント

競技開始後、[はじめに](./md/start/01_Scenario.md)を読み[最初にやること（VM の環境構築）](./md/setup/01_Start.md)を完了してください。

- [はじめに](./md/start/01_Scenario.md)
- [最初にやること（VM の環境構築）](./md/setup/01_Start.md)
- アプリケーションについて
  - [概要](./md/app/01_Service.md)
  - [環境](./md/app/02_Environment.md)
- 環境構築について
  - [【再掲】最初にやること（VM の環境構築）](./md/setup/01_Start.md)
  - [ローカル環境での開発](./md/setup/02_Local.md)
  - [Go の開発環境について](./md/setup/03_Go.md)
- サービスと競技について
  - [サービス概要](./md/app/01_Service.md)
  - [競技概要](./md/rules/01_Contest.md)
  - [レギュレーション](./md/rules/02_Regulation.md)
  - [採点について](./md/rules/03_Scoring.md)
  - [スクリプトの紹介](./md/app/03_Scripts.md)
- [FAQ](./md/01_FAQ.md)
- [API 設計書](./api-specs/openapi_defn.yaml)
