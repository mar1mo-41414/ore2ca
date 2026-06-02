# OpenSSLの証明書呪文を二度と覚えたくない人へ

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
ore2ca trust                   # OS / Firefox に信頼登録
ore2ca issue localhost         # 証明書を発行する
```

以上。OpenSSL は不要です。

---

## インストール

```bash
go install github.com/mar1mo-41414/ore2ca@latest
```

または手元でビルド：

```bash
# macOS / Linux
git clone https://github.com/mar1mo-41414/ore2ca
cd ore2ca
go build -o ore2ca .

# Windows (PowerShell)
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

### 2. OS と Firefox に信頼登録する

```bash
ore2ca trust
```

macOS の場合、Firefox への登録には `certutil` が必要です：

```bash
brew install nss
ore2ca trust
```

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
| `ore2ca trust` | OS / Firefox に CA を信頼登録 |
| `ore2ca untrust` | OS / Firefox から CA の信頼登録を削除 |
| `ore2ca issue <domain>` | サーバ証明書を発行 |
| `ore2ca list` | 発行済み証明書の一覧 |
| `ore2ca revoke <id>` | 証明書を失効 |
| `ore2ca delete <id>` | 証明書を削除 |
| `ore2ca docker nginx` | nginx 向け Docker Compose 設定例を出力 |
| `ore2ca docker caddy` | Caddy 向け Docker Compose 設定例を出力 |
| `ore2ca docker traefik` | Traefik 向け Docker Compose 設定例を出力 |

---

## Docker での使い方

Caddy を使った最小構成の例：

```bash
ore2ca issue myapp.local
ore2ca docker caddy
```

出力された設定をそのまま `docker-compose.yml` と `Caddyfile` に貼り付けるだけです。

---

## 対応プラットフォーム

| OS | システム信頼ストア | Firefox NSS |
|---|---|---|
| macOS | ✓ Keychain | ✓（要 `brew install nss`） |
| Linux（Debian/Ubuntu） | ✓ | ✓（要 `apt install libnss3-tools`） |
| Linux（RHEL/Fedora） | ✓ | ✓（要 `dnf install nss-tools`） |
| Linux（Arch） | ✓ | ✓（要 `pacman -S nss`） |
| Windows | ✓ | ✓ |

---

## 保存構造

```
~/.ore2ca/
├── ca/
│   ├── root.crt      ルートCA証明書
│   ├── root.key      秘密鍵（大切に）
│   └── serial        シリアル番号管理
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

### 開発終了時にクリーンアップする

信頼登録を取り消してシステムをクリーンな状態に戻せます。

```bash
ore2ca untrust   # OS・Firefox の信頼ストアから CA を削除
```

CA証明書・発行済み証明書のファイルは残るので、再度 `ore2ca trust` で復元できます。

---

### テスト端末に CA を信頼登録する

開発機とは別のPC・スマートフォン等で俺俺証明書のサイトにアクセスしたい場合、  
CA証明書（`~/.ore2ca/ca/root.crt`）をコピーしてインポートするだけです。

```bash
# テスト端末上で実行
ore2ca import --cert root.crt --trust
# ブラウザを再起動すれば完了
```

`--trust` を省いて後から `ore2ca trust` を個別に実行することもできます。

---

## 技術仕様

- **言語**: Go 1.22+
- **暗号**: ECDSA P-256（CA・サーバ証明書ともに）
- **依存**: OpenSSL 不使用。`crypto/x509` 標準ライブラリのみ。
- **証明書有効期間**: CA=10年、サーバ証明書=825日（デフォルト）

---

## ライセンス

MIT
