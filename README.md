# OpenSSLの証明書呪文を二度と覚えたくない人へ

![version](https://img.shields.io/badge/version-v1.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![license](https://img.shields.io/badge/license-MIT-green)

ローカル開発で HTTPS が必要になるたびに、こんな呪文をググっていませんか？

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

**もう覚えなくていいです。**

---

## ore2ca（俺俺CA）

ローカル開発専用の認証局（CA）を作り、HTTPS 証明書の発行・信頼登録・管理をワンストップで行うツールです。

```bash
ore2ca init                    # CA を作る
ore2ca trust                   # OS / ブラウザに信頼登録
ore2ca issue localhost         # 証明書を発行する
```

以上。OpenSSL は不要です。

---

## 動作確認済み環境

| OS | Safari | Firefox | Chrome | Edge | Brave |
|---|---|---|---|---|---|
| macOS | ✅ | ✅ | ✅ | - | - |
| Linux | - | ✅ | ✅ | - | - |
| Windows | - | ✅ | ✅ | ✅ | - |
| iOS | ✅ | ✅ | ✅ | - | ✅ |

> **iOS** — `root.crt` をデバイスで開き、**設定 → 一般 → VPNとデバイス管理** でインストール後、  
> **設定 → 一般 → 情報 → 証明書信頼設定** で CA をオンにすることで全ブラウザが対応します。

---

## インストール

### バイナリをダウンロード（推奨）

[Releases ページ](https://github.com/mar1mo-41414/ore2ca/releases/latest) から OS / アーキテクチャに合ったアーカイブをダウンロードして展開し、`PATH` の通ったディレクトリに置くだけです。

| ファイル名 | 対象 |
|---|---|
| `ore2ca_vX.Y_darwin_arm64.tar.gz` | macOS Apple Silicon |
| `ore2ca_vX.Y_darwin_amd64.tar.gz` | macOS Intel |
| `ore2ca_vX.Y_linux_amd64.tar.gz` | Linux x86_64 |
| `ore2ca_vX.Y_linux_arm64.tar.gz` | Linux ARM64 |
| `ore2ca_vX.Y_windows_amd64.zip` | Windows x64 |

### go install

```bash
go install github.com/mar1mo-41414/ore2ca@latest
```

### 手元でビルド

```bash
# macOS / Linux
git clone https://github.com/mar1mo-41414/ore2ca
cd ore2ca
go build -o ore2ca .

# Windows (PowerShell ※ 管理者として実行)
git clone https://github.com/mar1mo-41414/ore2ca
cd ore2ca
go build -o ore2ca.exe .
.\ore2ca.exe init
```

---

## はじめかた（5分で完了）

### 1. CA を作成する

```bash
ore2ca init
```

`~/.ore2ca/ca/` にルートCA証明書と秘密鍵が生成されます。

CA名や組織名はオプションで指定できます：

```bash
ore2ca init --cn "My Local CA" --org "MyCompany" --country JP --years 10
```

| オプション | デフォルト値 | 説明 |
|---|---|---|
| `--cn` | `Ore2CA Local Root CA` | CA の Common Name |
| `--org` | `Ore2CA` | 組織名 |
| `--country` | `JP` | 国コード（2文字） |
| `--years` | `10` | CA 証明書の有効期間（年） |
| `--force` | - | 既存の CA を上書きする |

### 2. OS とブラウザに信頼登録する

```bash
ore2ca trust
```

> **macOS** — Firefox への登録には `certutil` が必要です：
> ```bash
> brew install nss && ore2ca trust
> ```

> **Linux** — Firefox / Chrome への登録には `certutil` が必要です：
> ```bash
> sudo apt install libnss3-tools   # Debian/Ubuntu
> sudo dnf install nss-tools       # Fedora/RHEL
> sudo pacman -S nss               # Arch
> sudo ore2ca trust
> ```

> **Windows** — `trust` / `untrust` は **管理者として実行した PowerShell** から実行してください。  
> Firefox への登録は Enterprise Policy 方式（`policies.json`）で自動処理されます。

### 3. 証明書を発行する

```bash
ore2ca issue localhost
ore2ca issue myapp.local
ore2ca issue jellyfin.home.arpa
ore2ca issue '*.home.arpa'                          # ワイルドカードも対応
ore2ca issue localhost --san 192.168.11.8           # LAN IPも1枚の証明書に追加
ore2ca issue myapp.local --san 10.0.0.5 --san 192.168.1.10  # 複数SAN指定可
```

発行された証明書は `~/.ore2ca/certs/<domain>/` に保存されます：

```
~/.ore2ca/certs/localhost/
├── cert.crt       サーバ証明書
├── cert.key       秘密鍵
├── chain.crt      中間証明書チェーン
└── fullchain.crt  フルチェーン（nginx 等で使用）
```

---

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `ore2ca init` | ローカル CA を作成 |
| `ore2ca import --cert <path>` | 別PCで作成したCA証明書をインポート |
| `ore2ca trust` | OS / ブラウザに CA を信頼登録 |
| `ore2ca untrust` | OS / ブラウザから CA の信頼登録を削除 |
| `ore2ca issue <domain>` | サーバ証明書を発行 |
| `ore2ca renew <id>` | 証明書を同じドメイン・SANで更新 |
| `ore2ca show <id>` | 証明書の詳細情報を表示 |
| `ore2ca export <id>` | 証明書を PKCS#12 (.pfx) 形式でエクスポート |
| `ore2ca list` | 発行済み証明書の一覧 |
| `ore2ca revoke <id>` | 証明書を失効（CRL を自動更新） |
| `ore2ca delete <id>` | 証明書を削除 |
| `ore2ca docker nginx` | nginx 向け Docker Compose 設定例を出力 |
| `ore2ca docker caddy` | Caddy 向け Docker Compose 設定例を出力 |
| `ore2ca docker traefik` | Traefik 向け Docker Compose 設定例を出力 |
| `ore2ca web` | ローカル管理 Web UI を起動（デフォルト: `http://localhost:8080`） |

---

## Docker での使い方

Caddy を使った最小構成の例：

```bash
ore2ca issue myapp.local
ore2ca docker caddy
```

出力された設定をそのまま `docker-compose.yml` と `Caddyfile` に貼り付けるだけです。

### LAN 内からもアクセスする場合

```bash
# localhost と LAN IP を1枚の証明書にまとめる
ore2ca issue localhost --san 192.168.11.8
```

Caddyfile に `default_sni` を設定することで、IP 直アクセス（SNI なし接続）にも対応できます：

```
{
    default_sni localhost
}

localhost, https://192.168.11.8 {
    tls /path/to/cert.crt /path/to/cert.key
    ...
}
```

---

## 対応プラットフォーム

| OS | システム信頼ストア | Firefox | Chrome |
|---|---|---|---|
| macOS | ✅ Keychain | ✅（要 `brew install nss`） | ✅ |
| Linux（Debian/Ubuntu） | ✅ | ✅（要 `apt install libnss3-tools`） | ✅ |
| Linux（RHEL/Fedora） | ✅ | ✅（要 `dnf install nss-tools`） | ✅ |
| Linux（Arch） | ✅ | ✅（要 `pacman -S nss`） | ✅ |
| Windows | ✅ PowerShell | ✅ Enterprise Policy 方式 | ✅ |

---

## 保存構造

```
~/.ore2ca/
├── ca/
│   ├── root.crt      ルートCA証明書
│   ├── root.key      秘密鍵（大切に）
│   ├── serial        シリアル番号管理
│   └── crl.pem       証明書失効リスト（revoke 時に自動更新）
├── certs/
│   ├── localhost/
│   ├── myapp.local/
│   └── ...
└── config.yaml
```

---

## よくある使い方

### localhost の HTTPS 開発環境

```bash
ore2ca init && ore2ca trust && ore2ca issue localhost
# あとはサーバに cert.crt と cert.key を渡すだけ
```

### ホームネットワークのサービス

```bash
ore2ca issue jellyfin.home.arpa
# /etc/hosts に 192.168.1.x jellyfin.home.arpa を追加
```

### ワイルドカード証明書

```bash
ore2ca issue '*.home.arpa'
# *.home.arpa と home.arpa の両方をカバー
```

### PKCS#12 (.pfx) にエクスポートする

Windows (IIS, mmc) や Java アプリへのインポートに使用できます。

```bash
ore2ca export <id>                        # <domain>.pfx を出力（パスワードなし）
ore2ca export <id> --out server.pfx       # 出力先を指定
ore2ca export <id> --password secretpass  # パスワードを設定
ore2ca export <id> --legacy               # Java 8 以前など旧ツール向け（3DES）
```

---

### 証明書の詳細を確認する

```bash
ore2ca list               # ID を確認
ore2ca show <id>          # SAN・有効期限・アルゴリズムなどを詳細表示
```

### 期限切れ前に証明書を更新する

```bash
ore2ca renew <id>                 # 同じドメイン・SANで再発行（有効期間はデフォルト値）
ore2ca renew <id> --days 90       # 有効期間を変えて再発行
```

既存の証明書ファイルが上書きされるため、サーバ側の設定変更は不要です。

---

### 開発終了時にクリーンアップする

信頼登録を取り消してシステムをクリーンな状態に戻せます。

```bash
ore2ca untrust   # OS・ブラウザの信頼ストアから CA を削除
```

CA証明書・発行済み証明書のファイルは残るので、再度 `ore2ca trust` で復元できます。

### テスト端末に CA を信頼登録する

開発機とは別のPC・スマートフォン等で俺俺証明書のサイトにアクセスしたい場合、
CA証明書（`~/.ore2ca/ca/root.crt`）をコピーしてインポートするだけです。

```bash
# テスト端末上で実行
ore2ca import --cert root.crt --trust
# ブラウザを再起動すれば完了
```

---

## 技術仕様

- **言語**: Go 1.22+
- **暗号**: ECDSA P-256（CA・サーバ証明書ともに）
- **依存**: OpenSSL 不使用。`crypto/x509` 標準ライブラリのみ。
- **証明書有効期間**: CA=10年、サーバ証明書=825日（デフォルト）
- **Windows Firefox 登録**: Enterprise Policy（`policies.json`）方式

---

## ライセンス

MIT
