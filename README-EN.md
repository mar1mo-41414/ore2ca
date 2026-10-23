# Stop Memorizing OpenSSL Certificate Spells

[日本語 README →](README.md)

![version](https://img.shields.io/badge/version-v1.4-blue)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)
![license](https://img.shields.io/badge/license-MIT-green)

Are you Googling this incantation every time you need local HTTPS?

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

**You don't have to anymore.**

---

## ore2ca

A local CA (Certificate Authority) manager for development.
Create, trust, and issue HTTPS certificates — all in one tool.

```bash
ore2ca init                    # Create a CA
ore2ca trust                   # Register CA with OS / browsers
ore2ca issue localhost         # Issue a certificate
```

That's it. No OpenSSL required.

---

## Verified Platforms

All major browsers on macOS, Linux, Windows, iOS, and Android confirmed working.

| OS | Safari | Firefox | Chrome | Edge | Brave |
|---|---|---|---|---|---|
| macOS (Apple Silicon / Intel) | ✅ | ✅ | ✅ | - | - |
| Linux x86_64 | - | ✅ | ✅ | - | - |
| Linux ARM64 (Raspberry Pi) | - | ✅ | - | - | - |
| Windows (x64 / ARM) | - | ✅ | ✅ | ✅ | - |
| iOS | ✅ | ✅ | ✅ | - | ✅ |
| Android | - | ✅ | ✅ | - | - |
| ChromeOS Flex | - | - | ✅ | - | - |

> macOS verified on both Apple Silicon (macOS 27) and Intel (macOS 12 Monterey).

> **iOS** — Open `root.crt` on the device, install via **Settings → General → VPN & Device Management**,
> then enable the CA under **Settings → General → About → Certificate Trust Settings**.

> **Android** — Register the CA in the system store first. Firefox requires an extra step; see [below](#android).

> **ChromeOS Flex** — Use ore2ca inside the Linux environment (Crostini) and register the CA manually in Chrome. See [below](#chromeos).

---

## Installation

### Windows (winget)

```powershell
winget install mar1mo-41414.ore2ca
```

### Download Binary

Download the archive for your OS and architecture from the [Releases page](https://github.com/mar1mo-41414/ore2ca/releases/latest),
extract it, and place the binary in a directory on your `PATH`.

| File | Target |
|---|---|
| `ore2ca_vX.Y_darwin_arm64.tar.gz` | macOS Apple Silicon |
| `ore2ca_vX.Y_darwin_amd64.tar.gz` | macOS Intel |
| `ore2ca_vX.Y_linux_amd64.tar.gz` | Linux x86_64 |
| `ore2ca_vX.Y_linux_arm64.tar.gz` | Linux ARM64 |
| `ore2ca_vX.Y_windows_amd64.zip` | Windows x64 |
| `ore2ca_vX.Y_windows_arm64.zip` | Windows ARM64 |

### go install

```bash
go install github.com/mar1mo-41414/ore2ca@latest
```

### Build from Source

```bash
# macOS / Linux
git clone https://github.com/mar1mo-41414/ore2ca
cd ore2ca
go build -o ore2ca .

# Windows (PowerShell — run as Administrator)
git clone https://github.com/mar1mo-41414/ore2ca
cd ore2ca
go build -o ore2ca.exe .
.\ore2ca.exe init
```

---

## Getting Started (5 minutes)

### 1. Create a CA

```bash
ore2ca init
```

A root CA certificate and private key are saved to `~/.ore2ca/ca/`.

You can customize the CA identity:

```bash
ore2ca init --cn "My Local CA" --org "MyCompany" --country US --years 10
```

| Flag | Default | Description |
|---|---|---|
| `--cn` | `Ore2CA Local Root CA` | CA Common Name |
| `--org` | `Ore2CA` | Organization |
| `--country` | `JP` | Country code (2 letters) |
| `--years` | `10` | CA certificate validity (years) |
| `--cert-days` | `825` | Default validity for server certificates (days) |
| `--force` | - | Overwrite existing CA |

### 2. Register the CA with OS and Browsers

```bash
ore2ca trust
```

ore2ca auto-detects installed browsers and registers only the ones found.
If Firefox is not installed, its step is silently skipped.

To target a specific browser:

```bash
ore2ca trust --firefox   # OS + Firefox only
ore2ca trust --chrome    # OS + Chrome only (Linux)
```

> **macOS** — Firefox registration requires `certutil`:
> ```bash
> brew install nss && ore2ca trust
> ```

> **Linux** — Firefox/Chrome registration requires `certutil`:
> ```bash
> sudo apt install libnss3-tools   # Debian/Ubuntu
> sudo dnf install nss-tools       # Fedora/RHEL
> sudo pacman -S nss               # Arch
> sudo ore2ca trust
> ```

> **Windows** — Run `trust` / `untrust` from **PowerShell opened as Administrator**.
> Firefox registration is handled automatically via Enterprise Policy (`policies.json`).

### 3. Issue a Certificate

```bash
ore2ca issue localhost
ore2ca issue myapp.local
ore2ca issue jellyfin.home.arpa
ore2ca issue '*.home.arpa'                          # Wildcard supported
ore2ca issue localhost --san 192.168.1.1            # Add a LAN IP as SAN
ore2ca issue myapp.local --san 10.0.0.5 --san 192.168.1.10
```

Certificates are saved to `~/.ore2ca/certs/<domain>/`:

```
~/.ore2ca/certs/localhost/
├── cert.crt       Server certificate
├── cert.key       Private key
├── chain.crt      Certificate chain
└── fullchain.crt  Full chain (for nginx etc.)
```

---

## Command Reference

| Command | Description |
|---|---|
| `ore2ca init` | Create a local CA |
| `ore2ca import --cert <path>` | Import a CA certificate created on another machine |
| `ore2ca trust` | Register CA with OS / browsers |
| `ore2ca untrust` | Remove CA from OS / browsers |
| `ore2ca issue <domain>` | Issue a server certificate |
| `ore2ca renew <id>` | Renew a certificate with the same domain/SANs |
| `ore2ca show <id>` | Show certificate details |
| `ore2ca export <id>` | Export certificate as PKCS#12 (.pfx) |
| `ore2ca list` | List issued certificates |
| `ore2ca revoke <id>` | Revoke a certificate (updates CRL automatically) |
| `ore2ca delete <id>` | Delete a certificate |
| `ore2ca docker nginx` | Output a Docker Compose example for nginx |
| `ore2ca docker caddy` | Output a Docker Compose example for Caddy |
| `ore2ca docker traefik` | Output a Docker Compose example for Traefik |
| `ore2ca web` | Launch the local management Web UI (default: `http://localhost:8080`) |
| `ore2ca serve` | Serve the CA certificate to mobile devices via QR code |
| `ore2ca config show` | Display current configuration |
| `ore2ca config set cert-days <n>` | Change the default certificate validity |
| `ore2ca uninstall` | Remove CA trust (equivalent to `untrust`) |
| `ore2ca uninstall --purge` | Remove trust + delete `~/.ore2ca` entirely |
| `ore2ca completion` | Generate shell completion script |

---

## Common Workflows

### Local HTTPS Development

```bash
ore2ca init && ore2ca trust && ore2ca issue localhost
# Pass cert.crt and cert.key to your server
```

### Home Network Service

```bash
ore2ca issue jellyfin.home.arpa
# Add: 192.168.1.x jellyfin.home.arpa  to /etc/hosts
```

### Wildcard Certificate

```bash
ore2ca issue '*.home.arpa'
# Covers both *.home.arpa and home.arpa
```

### PKCS#12 (.pfx) Export

For Windows IIS, MMC, or Java applications:

```bash
ore2ca export <id>                        # Output <domain>.pfx (no password)
ore2ca export <id> --out server.pfx       # Specify output path
ore2ca export <id> --password secretpass  # Set a password
ore2ca export <id> --legacy               # For Java 8 or older tools (3DES)
```

---

### Change Default Certificate Validity

The default validity for `ore2ca issue` (without `--days`) can be changed:

```bash
ore2ca config show              # View current settings
ore2ca config set cert-days 365 # Change to 365 days for future issuance
```

You can also set it at CA creation time:

```bash
ore2ca init --cert-days 365
```

---

### Distribute CA to Mobile Devices (ore2ca serve)

Use `ore2ca serve` to distribute your CA certificate to smartphones via QR code.

```bash
# Run on your development machine
ore2ca serve
```

A QR code is printed in the terminal. Scan it with your phone's camera and open the URL.

- **iOS** — Tap "Install Configuration Profile (.mobileconfig)".
  After installing the profile, you must still enable the CA under
  **Settings → General → About → Certificate Trust Settings** — this step cannot be automated.
- **Android** — Tap "Download CA Certificate (.crt)" and import it via the system certificate manager.

```bash
ore2ca serve --port 8888   # Use a different port
ore2ca serve --once        # Shut down automatically after one download
```

> ⚠ `ore2ca serve` exposes your CA certificate on the LAN. Do not run it on untrusted networks.

---

### Inspect and Renew Certificates

```bash
ore2ca list               # Find the ID
ore2ca show <id>          # Show SAN, expiry, algorithm etc.
ore2ca renew <id>         # Renew with same domain/SANs (uses default validity)
ore2ca renew <id> --days 90   # Renew with a different validity period
```

---

### Clean Up

Remove the CA trust registration:

```bash
ore2ca untrust              # Remove CA from OS / browsers
# or
ore2ca uninstall            # Same as untrust
ore2ca uninstall --purge    # Remove trust + delete ~/.ore2ca entirely (-y to skip prompt)
```

Without `--purge`, certificate files remain, so you can restore with `ore2ca trust` later.

---

### Registering CA on Another Device

To access your self-signed HTTPS site from a different PC or smartphone,
copy `root.crt` (`~/.ore2ca/ca/root.crt`) to that device and run:

```bash
# On the target device
ore2ca import --cert root.crt --trust
# Restart the browser — done
```

---

### Android Trust Registration <a id="android"></a>

**Chrome (system CA is enough)**

1. Transfer `root.crt` to the Android device
2. Go to **Settings → Security → Encryption & credentials → Install a certificate → CA certificate**
3. Select `root.crt` and install
4. Restart Chrome → verify the padlock icon

> Menu names vary by Android version and manufacturer.

**Firefox for Android (extra step required)**

Android Firefox does not read the system certificate store by default.
After registering the CA in the system, follow these steps to load it in Firefox:

1. Enter the following URL in Firefox's address bar:
   ```
   chrome://geckoview/content/config.xhtml
   ```
2. Search for **`security.enterprise_roots.enabled`**
3. Toggle the value from **`false` to `true`**
4. Restart Firefox → verify the padlock icon

> `about:config` does not work in recent Firefox for Android.
> `chrome://geckoview/content/config.xhtml` is the correct settings page.

---

### ChromeOS Flex Trust Registration <a id="chromeos"></a>

Chrome on ChromeOS uses its own certificate management rather than the OS certificate store.

**1. Prepare the CA certificate in Linux (Crostini)**

```bash
# Run inside the Linux environment
ore2ca init
ore2ca trust          # Register with Linux system / Firefox
cp ~/.ore2ca/ca/root.crt ~/   # Copy to a location accessible from ChromeOS
```

**2. Register the CA in ChromeOS Chrome**

1. Open `chrome://settings/certificates` in Chrome
2. Select the **Authorities** tab
3. Click **Import**
4. In the file picker, open the **Linux files** tab and select `root.crt`
5. Check "Trust this certificate for identifying websites" and click OK
6. Restart Chrome

**3. Verify**

Start an HTTPS server (e.g., Caddy) in the Linux environment and navigate to
`https://<Linux environment IP>` to confirm the padlock icon appears.

> Find the Linux environment IP with `ip addr` (usually in the `100.115.92.x` range).

---

## Shell Completion

```bash
# Bash (current session)
source <(ore2ca completion bash)

# Bash (persistent) — Linux:
ore2ca completion bash > /etc/bash_completion.d/ore2ca
# macOS:
ore2ca completion bash > $(brew --prefix)/etc/bash_completion.d/ore2ca

# Zsh (current session)
source <(ore2ca completion zsh)

# Fish
ore2ca completion fish | source

# PowerShell
ore2ca completion powershell >> $PROFILE
```

---

## OCSP Responder (separate tool)

`ore2ca revoke` updates `~/.ore2ca/ca/crl.pem`, but if browsers or HTTP clients
need to verify revocation via OCSP, use **ore2ca-ocsp**:

```bash
# Install
go install github.com/mar1mo-41414/ore2ca-ocsp@latest

# Start (auto-reads ~/.ore2ca)
ore2ca-ocsp
```

ore2ca-ocsp is a separate binary that reflects `ore2ca revoke` results **immediately without restart**.

> See the [ore2ca-ocsp repository](https://github.com/mar1mo-41414/ore2ca-ocsp) for details.

---

## Technical Specifications

- **Language**: Go 1.25+
- **Cryptography**: ECDSA P-256 (CA and server certificates)
- **Dependencies**: No OpenSSL. Uses only Go's `crypto/x509` standard library.
- **Certificate validity**: CA = 10 years, server certificates = 825 days (default, configurable)
- **Windows Firefox registration**: Enterprise Policy (`policies.json`) method

---

## License

MIT
