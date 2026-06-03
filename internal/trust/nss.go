package trust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mar1mo-41414/ore2ca/internal/config"
)

// --- browser detection ---

// isFirefoxInstalled returns true if Firefox is found on the system.
func isFirefoxInstalled() bool {
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/Applications/Firefox.app",
			"/Applications/Firefox Developer Edition.app",
			"/Applications/Firefox Nightly.app",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	case "linux":
		paths := []string{
			"/usr/bin/firefox",
			"/usr/bin/firefox-esr",
			"/usr/local/bin/firefox",
			"/snap/bin/firefox",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	case "windows":
		paths := []string{
			`C:\Program Files\Mozilla Firefox\firefox.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\firefox.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	}
	return false
}

// isChromeInstalled returns true if Chrome/Chromium/Brave/Edge is found.
// On macOS and Windows these browsers use the OS trust store, so this is
// only meaningful on Linux where they each maintain NSS databases.
func isChromeInstalled() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	paths := []string{
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/brave-browser",
		"/usr/bin/brave",
		"/usr/bin/microsoft-edge",
		"/usr/bin/microsoft-edge-stable",
		"/usr/local/bin/google-chrome",
		"/usr/local/bin/chromium",
		"/snap/bin/chromium",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// shouldDoFirefox returns true if Firefox registration should be attempted.
func shouldDoFirefox(opts Options) bool {
	if opts.explicit() {
		return opts.Firefox
	}
	return isFirefoxInstalled()
}

// shouldDoChrome returns true if Chrome NSS registration should be attempted.
// Only relevant on Linux — macOS/Windows Chrome uses the OS trust store.
func shouldDoChrome(opts Options) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if opts.explicit() {
		return opts.Chrome
	}
	return isChromeInstalled()
}

// --- NSS operations ---

func installNSSDBs(certPath string, dbs []string) (bool, error) {
	certutil, err := findNSSCertutil()
	if err != nil {
		return false, err
	}
	var lastErr error
	ok := 0
	for _, db := range dbs {
		cmd := exec.Command(certutil, "-A", "-d", db, "-t", "CT,,", "-n", "Ore2CA", "-i", certPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			lastErr = fmt.Errorf("certutil %s: %w\n%s", db, err, out)
		} else {
			ok++
		}
	}
	if ok == 0 {
		return false, lastErr
	}
	return true, nil
}

func uninstallNSSDBs(dbs []string) (bool, error) {
	certutil, err := findNSSCertutil()
	if err != nil {
		return false, err
	}
	var lastErr error
	ok := 0
	for _, db := range dbs {
		cmd := exec.Command(certutil, "-D", "-d", db, "-n", "Ore2CA")
		if out, err := cmd.CombinedOutput(); err != nil {
			lastErr = fmt.Errorf("certutil %s: %w\n%s", db, err, out)
		} else {
			ok++
		}
	}
	if ok == 0 {
		return false, lastErr
	}
	return true, nil
}

// installNSSForFirefox installs into Firefox-only NSS databases.
func installNSSForFirefox(certPath string) (bool, error) {
	dbs := findFirefoxNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Firefox の NSS プロファイルが見つかりません（Firefox を一度起動してプロファイルを作成してください）")
	}
	return installNSSDBs(certPath, dbs)
}

// installNSSForChrome installs into Chrome/Chromium NSS databases (Linux only).
func installNSSForChrome(certPath string) (bool, error) {
	dbs := findChromeNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Chrome/Chromium の NSS データベースが見つかりません（Chrome を一度起動してください）")
	}
	return installNSSDBs(certPath, dbs)
}

// uninstallNSSForFirefox removes from Firefox-only NSS databases.
func uninstallNSSForFirefox() (bool, error) {
	dbs := findFirefoxNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Firefox の NSS プロファイルが見つかりません")
	}
	return uninstallNSSDBs(dbs)
}

// uninstallNSSForChrome removes from Chrome/Chromium NSS databases (Linux only).
func uninstallNSSForChrome() (bool, error) {
	dbs := findChromeNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Chrome/Chromium の NSS データベースが見つかりません")
	}
	return uninstallNSSDBs(dbs)
}

// --- NSS database finders ---

func findFirefoxNSSDBs() []string {
	home, err := config.EffectiveUserHome()
	if err != nil {
		return nil
	}
	return collectNSSDBs(firefoxCandidates(home))
}

func findChromeNSSDBs() []string {
	home, err := config.EffectiveUserHome()
	if err != nil {
		return nil
	}
	return collectNSSDBs(chromeCandidates(home))
}

func collectNSSDBs(candidates []string) []string {
	var dirs []string
	for _, c := range candidates {
		matches, err := filepath.Glob(c)
		if err != nil {
			continue
		}
		for _, m := range matches {
			if _, err := os.Stat(filepath.Join(m, "cert9.db")); err == nil {
				dirs = append(dirs, "sql:"+m)
			} else if _, err := os.Stat(filepath.Join(m, "cert8.db")); err == nil {
				dirs = append(dirs, m)
			}
		}
	}
	return dirs
}

func firefoxCandidates(home string) []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles", "*"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".mozilla", "firefox", "*"),
			filepath.Join(home, ".firefox", "*"),
			filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox", "*"),
		}
	case "windows":
		appdata := os.Getenv("APPDATA")
		return []string{
			filepath.Join(appdata, "Mozilla", "Firefox", "Profiles", "*"),
		}
	}
	return nil
}

func chromeCandidates(home string) []string {
	// Chrome/Chromium use NSS only on Linux.
	if runtime.GOOS != "linux" {
		return nil
	}
	return []string{
		filepath.Join(home, ".pki", "nssdb"),
		filepath.Join(home, ".config", "google-chrome", "Default"),
		filepath.Join(home, ".config", "chromium", "Default"),
		filepath.Join(home, ".config", "BraveSoftware", "Brave-Browser", "Default"),
		filepath.Join(home, ".config", "microsoft-edge", "Default"),
	}
}

// --- certutil finder ---

func findNSSCertutil() (string, error) {
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\Mozilla Firefox\certutil.exe`,
			`C:\Program Files (x86)\Mozilla Firefox\certutil.exe`,
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
		return "", fmt.Errorf("Firefox の certutil.exe が見つかりません")
	}
	path, err := exec.LookPath("certutil")
	if err != nil {
		return "", certutilNotFoundError()
	}
	return path, nil
}

func certutilNotFoundError() error {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    brew install nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	case "linux":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    # Debian/Ubuntu:\n    sudo apt install libnss3-tools\n\n    # Fedora/RHEL:\n    sudo dnf install nss-tools\n\n    # Arch:\n    sudo pacman -S nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	}
	return fmt.Errorf("certutil not found")
}
