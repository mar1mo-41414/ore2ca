# ore2ca（俺俺CA）実装報告書

作成日: 2026-06-02

---

## 概要

ローカル開発向け認証局（CA）管理ツール。OpenSSL コマンドへの依存をなくし、証明書のライフサイクル全体をワンストップで管理することを目的として開発した。

---

## 使用言語・技術スタック

| 項目 | 内容 |
|---|---|
| 言語 | Go 1.22 |
| CLIフレームワーク | [cobra](https://github.com/spf13/cobra) v1.8.1 |
| 設定ファイル | [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) |
| 暗号ライブラリ | Go 標準ライブラリ `crypto/x509`、`crypto/ecdsa` |
| 外部コマンド依存 | `certutil`（Firefox NSS登録時のみ）、`security`（macOS Keychain）、OS証明書更新コマンド（Linux） |
| OpenSSL依存 | **なし** |

暗号処理は Go 標準ライブラリのみで完結させた。OpenSSL の外部プロセス呼び出しは一切行っていない。

---

## プロジェクト構成

```
ore2ca/
├── main.go                    エントリーポイント
├── cmd/                       CLIコマンド定義
│   ├── root.go                ルートコマンド
│   ├── init.go                ore2ca init
│   ├── trust.go               ore2ca trust
│   ├── issue.go               ore2ca issue
│   ├── list.go                ore2ca list
│   ├── revoke.go              ore2ca revoke
│   ├── delete.go              ore2ca delete
│   └── docker.go              ore2ca docker
├── internal/                  内部パッケージ（外部非公開）
│   ├── ca/
│   │   ├── ca.go              CA生成ロジック
│   │   └── cert.go            証明書発行ロジック
│   ├── store/
│   │   └── store.go           ファイルシステム管理・メタデータ管理
│   ├── trust/
│   │   ├── trust.go           信頼登録インターフェース
│   │   ├── darwin.go          macOS Keychain 実装
│   │   ├── linux.go           Linux システム証明書ストア実装
│   │   ├── windows.go         Windows 実装
│   │   └── nss.go             Firefox NSS 登録実装（クロスプラットフォーム）
│   └── config/
│       └── config.go          設定ファイル（config.yaml）読み書き
└── pkg/ore2ca/
    └── api.go                 将来のGUI・Web UI向けパブリックAPI
```

---

## 実装済み機能

### CA管理

- **`ore2ca init`**: ECDSA P-256 によるルートCA生成
  - Common Name / 組織名 / 国コード / 有効年数をオプションで指定可能
  - `--force` で既存CAの上書き可能
  - 証明書・秘密鍵を `~/.ore2ca/ca/` に保存

### 信頼登録

- **`ore2ca trust`**: OS信頼ストアへの登録
  - macOS: `security add-trusted-cert` でシステムキーチェーンに登録
  - Linux: Debian/Ubuntu / RHEL系 / Arch の各ディストリビューションに対応
  - Windows: `x509.SystemCertPool` 経由で登録
  - Firefox NSS: `certutil` を使いプロファイルを自動検出して登録
  - `certutil` 未インストール時はOS別インストール手順をガイド表示

### 証明書発行

- **`ore2ca issue <domain>`**: ECDSA P-256 サーバ証明書の発行
  - SAN（Subject Alternative Name）自動付与
  - `localhost` → DNS SAN + `127.0.0.1` / `::1` IP SAN を自動追加
  - ワイルドカード（`*.home.arpa`）→ ベースドメイン（`home.arpa`）も自動追加
  - IP アドレス直接指定対応
  - 以下のファイルを生成: `cert.crt`、`cert.key`、`chain.crt`、`fullchain.crt`
  - メタデータ（発行日・有効期限・シリアル番号）を `meta.json` に記録

### 証明書管理

- **`ore2ca list`**: 発行済み証明書の一覧表示（期限切れ・失効状態を表示）
- **`ore2ca revoke <id>`**: 証明書の失効マーク（ファイルは削除しない）
- **`ore2ca delete <id>`**: 証明書ファイルの完全削除（確認プロンプトあり、`-y` でスキップ）

### Docker支援

- **`ore2ca docker nginx`**: nginx 向け Docker Compose + nginx.conf 設定例出力
- **`ore2ca docker caddy`**: Caddy 向け Docker Compose + Caddyfile 設定例出力
- **`ore2ca docker traefik`**: Traefik v3 向け Docker Compose + 動的設定例出力
- 出力パスはユーザー環境の実際の証明書パスを埋め込み済み

### ライブラリAPI

- **`pkg/ore2ca`**: 将来のGUI・Web UI向けに全機能をパブリックAPIとして公開

---

## 動作確認済み内容

- macOS + Firefox での end-to-end テスト
  - `ore2ca init` → `ore2ca trust` → `ore2ca issue localhost`
  - Docker + Caddy で HTTPS サーバを起動
  - Firefox でアドレスバー鍵マーク表示・証明書警告なしを確認
  - ECDSA P-256、TLS 1.3、HTTP/2 で通信成立

---

## できなかった部分・既知の制限

### CRL（証明書失効リスト）の未実装

`ore2ca revoke` は内部メタデータに失効フラグを立てるのみで、RFC 5280 準拠の CRL ファイル（`.crl`）を生成・配信する機能は未実装。クライアントが CRL を取得して検証する仕組みは持っていない。

### OCSP（Online Certificate Status Protocol）の未実装

証明書のリアルタイム失効確認プロトコル。将来の ACMEサーバ機能と合わせて実装予定。

### Windows 信頼登録の不完全さ

現在の Windows 実装は `x509.SystemCertPool` を使用しているが、これはプロセス内のみ有効で OS 全体へのインストールには不十分。`certmgr.msc` への登録は PowerShell (`Import-Certificate`) または Windows API (`CertOpenSystemStore`) を使う必要があり、未対応。

### 証明書の再発行（renew）コマンドの未実装

既存証明書を同じドメインで更新する `ore2ca renew <id>` コマンドが未実装。現状は `ore2ca delete` → `ore2ca issue` で代替。

### `ore2ca docker` の出力形式

現状は設定例を標準出力に表示するのみで、ファイルへの直接書き出し（`--output`）や既存 `docker-compose.yml` へのマージは未対応。

---

## 将来構想（未実装）

- `ore2ca web`: ローカル管理 Web UI の起動
- ACMEサーバ機能（Let's Encrypt 互換 API）
- Kubernetes Secret 出力（`ore2ca k8s secret`）
- Docker Compose の自動生成・既存ファイルへのマージ
- Headscale / Tailscale MagicDNS との連携
- CRL / OCSP の実装
- 証明書の自動更新（`ore2ca renew`）

---

## 所感

OpenSSL コマンドの呪文を調べる手間をなくすという目的は達成できた。特に SAN の自動付与と Firefox NSS への自動登録は、手作業では間違えやすい部分であり、ツール化の効果が大きい箇所だった。

Go 標準ライブラリの `crypto/x509` は証明書生成に必要な機能が揃っており、OpenSSL への依存を完全に排除できた点は設計方針通りに実現できた。

Windows 対応については実機テストができておらず、信頼登録部分に課題が残る。
