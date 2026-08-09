# PMO Agent

![CI](https://github.com/ymd38/pmo-agent/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

プロジェクト統制基盤（PMO プラットフォーム）。プロジェクトコードの発行から工数・コスト管理、経営レポートまでを一気通貫で管理するモノレポです。

- 仕様の正本: [`docs/SPEC.md`](docs/SPEC.md)
- デザインシステム: [`DESIGN.md`](DESIGN.md)
- 開発ガイド: [`CLAUDE.md`](CLAUDE.md) / [`api/CLAUDE.md`](api/CLAUDE.md) / [`apps/pmo-dashboard/CLAUDE.md`](apps/pmo-dashboard/CLAUDE.md)

## English Overview

**PMO Agent** is a project governance platform for PMO (Project Management Office) operations — issuing project codes, tracking work hours and true project costs (external spend + internal labor), and generating executive reports, all in one place.

**Stack**: Go 1.25 (Gin / dig / GORM, clean architecture), Nuxt 4 + Tailwind CSS v4, MySQL 8.0, golang-migrate, n8n, Docker Compose, GitHub Actions.

**Highlights**:

- Role-based access control stored in the database (`roles` / `functions` / `role_functions`) — new roles require data changes, not code changes
- JWT (HS256) auth in httpOnly cookies with refresh-token rotation; invitation/reset via single-use hashed tokens
- Field-level response control by role (e.g. unit rates are visible to PMO admins only)
- Immutable project codes issued per program (`INV-2026-0001-001`), enforced at the usecase layer
- Automated daily / weekly / executive reporting via n8n workflows, with AI-generated PMO commentary (OpenAI via LangChain nodes)

**Quick start** (Docker Desktop only — no host Go/Node/MySQL needed):

```bash
make env && make up && make migrate-up && make seed-link
# open the printed set-password link, then visit http://localhost:3000
```

---

## アーキテクチャ

```
[ タスク管理ツール ] → [ Google Sheets ]
                           ↓
                    [ n8n (:5678) ] 日次収集・AI分析
                           ↓ mySql ノードで直接書き込み
[ MySQL 8.0 ]  ←→  [ Go API (:8080) ]  ←→  [ pmo-dashboard (:3000) ]
```

| コンポーネント | 技術 | 役割 |
|---|---|---|
| `api/` | Go 1.25 / Gin / dig / GORM | RESTful API。クリーンアーキテクチャ（handler → usecase → repository） |
| `apps/pmo-dashboard/` | Nuxt 4 / TypeScript strict / Tailwind v4 | PMO管理画面（経営層・PMO管理者・PM向け） |
| `apps/worktrack/` | Nuxt 4（未実装） | 工数入力UI（全メンバー向け） |
| `db/migrations/` | golang-migrate | 連番スキーマ・seedマイグレーション |
| `n8n/workflows/` | n8n | 進捗収集・日次/週次/エグゼクティブレポート自動生成ワークフロー |

設計の特徴:

- **RBAC を DB で管理** — `roles` / `functions` / `role_functions` テーブル。新ロール・新権限はデータ追加のみで対応（開放閉鎖原則）
- **認証**: パスワード（bcrypt）+ JWT HS256 を httpOnly Cookie に格納。リフレッシュトークンはDB保管・ローテーション。招待・リセットは単回利用のハッシュ化トークン
- **ロール別フィールド制御**: 単価（`grade_rates`）は `pmo_admin` のみ参照可。コスト集計APIはロールで返却フィールドを変える
- **プロジェクトコードの不変性**: プログラム単位でプレフィックス発行（`INV-2026-0001`）、承認時に枝番発行（`-001`）。発行後の変更は usecase 層で拒否

---

## n8n ワークフロー（レポート自動生成）

[`n8n/workflows/`](n8n/workflows/) に、進捗収集とレポート生成を自動化する2つのワークフロー定義（JSON）を同梱しています。

| ワークフロー | スケジュール | 処理内容 |
|---|---|---|
| `daily_workflow.json` | 毎日10時 | 進行中プロジェクトの進捗データを Google Sheets から収集 → MySQL に保存 → 直近7日を分析して日次レポート生成 → AI（OpenAI）が PMO 視点のコメントを付与 |
| `weekly_workflow.json` | 毎週月曜9時 | 日次レポートを集約して週次サマリーを生成 → 全プロジェクト横断のエグゼクティブレポートを作成 → AI が週次・経営向けコメントを生成 |

n8n は `make up` で一緒に起動し、**http://localhost:5678** で開けます。

利用手順:

1. http://localhost:5678 を開く（初回はオーナーアカウントの作成を求められます）
2. ワークフロー JSON をインポートする。`n8n/workflows/` はコンテナ内の `/home/node/workflows` にマウントされているので、そのパスから読み込めます
3. クレデンシャルを設定する: **MySQL** / **Google Sheets** / **OpenAI**
4. インポートしたワークフローをオンにする

> **MySQL クレデンシャルのホスト名**は `localhost` ではなく **`mysql`**（ポート `3306`）を指定してください。n8n は compose ネットワーク内から接続するため、ホスト公開ポート（`MYSQL_PORT`）ではなくコンテナ内ポートを使います。
>
> 認証情報とワークフローは `n8n_data` ボリュームに保存されるため、`make down` では消えません。`make reset` や `docker compose down -v` を実行すると**消えます**。

### Google Sheets フォーマット

各プロジェクトのスプレッドシートはシート名「**プロジェクト進捗**」で、以下の列構成にします。

| 列 | 列名 | 内容 |
|---|---|---|
| A | フェーズ | 要件定義、設計、実装、テスト など |
| B | タスク | フェーズ内のタスク名 |
| C | ステータス | 未着手 / 進行中 / 完了 |
| D | 進捗 | 進捗率（0〜100） |
| E | 期限 | タスクの期限日 |

> **注意（情報取得元）**: 進捗データの取得元は環境に依存します。同梱ワークフローはタスク管理ツールのデータを Google Sheets 経由で受け取る構成のため、利用する環境に合わせて取得ノード（Google Sheets 部分）を各自のタスク管理ツール（Backlog / Jira / GitHub Issues 等）の API ノードに差し替えてください。取得後の整形・保存・AI分析のフローはそのまま流用できます。

---

## 前提

**Docker Desktop だけあれば動きます。** Go・Node・MySQL・golang-migrate のホストへのインストールは不要です。
すべての操作は `make` 経由で行います（`docker` / `go` / `npm` を直接叩く必要はありません）。

利用可能なコマンドの一覧は次で確認できます:

```bash
make help
```

---

## クイックスタート

```bash
# 1. 環境変数ファイルを用意（初回のみ）
make env

# 2. 全サービスを起動（初回はイメージビルドで数分かかります）
make up

# 3. データベースにマイグレーションを適用
make migrate-up

# 4. 初期管理者のパスワード設定リンクを発行
make seed-link
#   → 出力された http://localhost:3000/set-password?token=... を
#     ブラウザで開き、パスワードを設定する
```

設定が終わったら **http://localhost:3000** を開いてログインします。

> 初期管理者は `admin@example.com`（パスワード未設定）。
> 別アカウントのリンクが欲しい場合: `make seed-link email=<メールアドレス>`

---

## 動作確認チェックリスト

ログイン後、以下の流れで主要機能を確認できます。

| # | 操作 | 期待結果 |
|---|---|---|
| 1 | `/`（未ログイン） | 公開LPが表示される（ログインリンクあり） |
| 2 | `/login` でログイン | `/home` ダッシュボードへ遷移 |
| 3 | ナビ「プログラム」→「新規プログラム」 | 種別(例 `INV`)＋会計年度＋名称を入力 |
| 4 | プログラム作成 | コードが自動採番される（`INV-2026-0001`）。連続作成で連番が増える |
| 5 | プログラム詳細 →「新規プロジェクト」 | プロジェクトを起案（status=planning・コード未発行） |
| 6 | プロジェクトの「コード発行」 | 枝番が採番され（`INV-2026-0001-001`）status=進行中になる |
| 7 | ナビ「メンバー管理」→「新規メンバー」 | 作成すると招待リンクが表示される |
| 8 | ナビ「属性マスタ」 | カテゴリ選択 → 値の追加・無効化ができる |
| 9 | ログアウト | `/login` に戻る |

### API だけ叩いて確認する場合

```bash
# ヘルスチェック
curl http://localhost:8080/api/health        # {"status":"ok"}

# ログイン（Cookie 取得）
curl -c /tmp/c.txt -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"<設定したパスワード>"}'

# 自分の権限を確認
curl -b /tmp/c.txt http://localhost:8080/api/auth/me
```

---

## よく使う make コマンド

| コマンド | 説明 |
|---|---|
| `make up` | 全サービス起動（mysql / api / pmo-dashboard / n8n） |
| `make down` | 全サービス停止 |
| `make restart` | 再起動 |
| `make ps` | 起動中コンテナ確認 |
| `make logs` | 全ログ追従（`make api-logs` / `make web-logs` で個別） |
| `make migrate-up` / `make migrate-down` | マイグレーション適用 / 1件ロールバック |
| `make migrate-create name=add_xxx` | 新規マイグレーション作成 |
| `make seed-link [email=...]` | パスワード設定リンク発行 |
| `make db-cli` | MySQL CLI に接続 |
| `make reset` | **DBを全削除して作り直す**（クリーンな状態に戻したいとき） |
| `make test` | バックエンド(`-race`)＋フロントのテスト実行 |
| `make rebuild` | 依存更新時にイメージを再ビルドして起動 |

---

## ポート一覧

ホスト側の公開ポートはすべて環境変数で変更できます（コンテナ内のポートは固定）。

| サービス | 既定 URL / ポート | 変更する環境変数 | コンテナ内 |
|---|---|---|---|
| pmo-dashboard | http://localhost:3000 | `PMO_DASHBOARD_PORT` | `3000` |
| api | http://localhost:8080 | `API_PORT` | `api:${API_PORT}` |
| mysql | `localhost:3306` | `MYSQL_PORT` | `mysql:3306` |
| n8n | http://localhost:5678 | `N8N_PORT` | `n8n:${N8N_PORT}` |

`API_PORT` はホスト公開ポートと API がコンテナ内で listen するポートの両方を指します（`config.go` も同じ変数を読みます）。ホスト側とコンテナ側で番号が食い違わないよう、1つの変数で揃えています。`N8N_PORT` も同じ扱いです。

> `N8N_PORT` に **5679 は指定できません**。n8n 内部の Task Broker が 5679 を使うため、衝突して `n8n Task Broker's port 5679 is already in use` で起動に失敗します。

> ホストで別の MySQL が `3306` を使っている場合は、`MYSQL_PORT=3307` のように変更してください。

URL のような複合値は、設定した「部品」から `docker-compose.yml` が組み立てます。**ポートやホスト名を変えても、URL 側を手で直す必要はありません。**

| 組み立てられる値 | 既定の組み立て | 用途 |
|---|---|---|
| `NUXT_PUBLIC_API_BASE` | `http://${APP_HOST}:${API_PORT}` | ブラウザ → API |
| `APP_BASE_URL` | `http://${APP_HOST}:${PMO_DASHBOARD_PORT}` | パスワード設定リンクの生成元 |
| `DB_PASSWORD` / `DB_NAME` | `MYSQL_ROOT_PASSWORD` / `MYSQL_DATABASE` を流用 | API → MySQL |

`APP_HOST`（既定 `localhost`）は、ブラウザからアクセスするホスト名です。LAN 内の別マシンから開くときに IP へ変えると、API の URL も同時に追随します。

リバースプロキシ配下など、ホスト名とポートの組み合わせで表せない場合のみ、`APP_BASE_URL` などを直接指定して組み立てを上書きしてください（`.env.example` の「上書き用」セクション）。SSR 用の `NUXT_API_BASE_SERVER` は compose ネットワーク内の通信でコンテナ内ポート固定のため、変更不要です。

worktrack（工数入力UI）と Storybook は未実装のため、現在 compose では無効化しています。

---

## トラブルシュート

- **API が想定と違う DB に繋ぎにいく（`connection refused` / `Access denied`）** — 起動ログの `DB_HOST=... DB_PORT=...` 行と `infra: DB接続先 ...` 行を見比べてください。前者は設定値、後者は MySQL 自身の応答です。接続先のホストとポートは `mysql:3306` 固定です（compose ネットワーク内の構成そのものなので設定項目ではありません）。**`MYSQL_PORT` はホストへの公開ポートで、API の接続先には使えません。** 認証情報と DB 名の上書きは `PMO_DB_USER` / `PMO_DB_PASSWORD` / `PMO_DB_NAME` という **`PMO_` 付きの名前**で行います。`DB_USER` のような一般的な名前を使わないのは、他プロジェクト用にシェルへ export された値を拾ってしまう事故を防ぐためです（実例: `DB_PORT=5433` が漏れ込み、MySQL に PostgreSQL のポートで接続しにいった）。
- **`required variable ... is missing a value` で止まる** — `MYSQL_ROOT_PASSWORD` / `MYSQL_DATABASE` / `JWT_SECRET` は必須です。空のまま起動すると空パスワードの MySQL や空の署名鍵で立ち上がってしまうため、コンテナを作る前に compose が止めます。エラーは1件ずつ出るので、表示された変数を設定して再実行してください。なお **この検証は `make down` / `make ps` / `make logs` でも走ります** — 変数を export していないターミナルからは停止操作もできない点に注意してください（起動したターミナルから実行するか、同じ変数を export してください）。
- **`make up` でポート競合エラー** — 既定の 3000 / 8080 / 3306 が他プロセスで使われています。競合プロセスを停止するか、`PMO_DASHBOARD_PORT` / `API_PORT` / `MYSQL_PORT` を空いているポートに変更してください（例: `PMO_DASHBOARD_PORT=3002 make up`）。
- **ログイン直後の画面で一瞬エラーが出る** — API コンテナの起動直後はSSRの初回取得が間に合わないことがあります。リロードで解消します。
- **マイグレーションが `dirty` で失敗する** — `make migrate-force version=<直前の成功番号>` で解除してから `make migrate-up`。
- **データを完全に作り直したい** — `make reset`（全データ削除 → 再マイグレーション。その後 `make seed-link`）。
- **依存（go.mod / package.json）を変更した** — `make rebuild` でイメージを作り直します。

---

## 構成

```
pmo-agent/
├── api/                   # Go API（Gin + dig + GORM, JWT=HS256）
├── apps/
│   ├── pmo-dashboard/     # PMO管理画面（Nuxt 4 + Tailwind v4）
│   └── worktrack/         # 工数入力UI（未実装）
├── db/migrations/         # golang-migrate 連番マイグレーション
├── n8n/workflows/         # 進捗収集・レポート自動生成ワークフロー
├── docs/SPEC.md           # 仕様の正本
├── docker-compose.yml
└── Makefile               # すべての開発タスク
```

## License

[MIT](LICENSE)
