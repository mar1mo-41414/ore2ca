package trust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mar1mo-41414/ore2ca/internal/config"
)

// installNSS tries to register the CA cert into Firefox NSS databases.
func installNSS(certPath string) (bool, error) {
	certutil, err := findNSSCertutil()
	if err != nil {
		return false, err
	}

	dbs := findNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Firefox / Chrome の NSS プロファイルが見つかりません")
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

func uninstallNSS() (bool, error) {
	certutil, err := findNSSCertutil()
	if err != nil {
		return false, err
	}
	dbs := findNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Firefox / Chrome の NSS プロファイルが見つかりません")
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

// findNSSCertutil returns the path to the NSS certutil tool.
// On Windows, PATH の certutil.exe は Windows 組み込みの別物なので
// Firefox インストールフォルダを直接検索する。
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
		return "", fmt.Errorf(
			"Firefox の certutil.exe が見つかりません。\n" +
				"Firefox がインストールされている場合、以下のパスを確認してください:\n" +
				"  C:\\Program Files\\Mozilla Firefox\\certutil.exe\n\n" +
				"Firefox が未インストールの場合は Firefox NSS 登録をスキップします。")
	}
	path, err := exec.LookPath("certutil")
	if err != nil {
		return "", certutilNotFoundError()
	}
	return path, nil
}

func findNSSDBs() []string {
	var dirs []string
	home, err := config.EffectiveUserHome()
	if err != nil {
		return nil
	}

	candidates := nssCandidates(home)
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

func certutilNotFoundError() error {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    brew install nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	case "linux":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    # Debian/Ubuntu:\n    sudo apt install libnss3-tools\n\n    # Fedora/RHEL:\n    sudo dnf install nss-tools\n\n    # Arch:\n    sudo pacman -S nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	}
	return fmt.Errorf("certutil not found")
}

func nssCandidates(home string) []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles", "*"),
			filepath.Join(home, "Library", "Application Support", "Chromium", "Default"),
		}
	case "linux":
		return []string{
			// Firefox
			filepath.Join(home, ".mozilla", "firefox", "*"),
			filepath.Join(home, ".firefox", "*"),
			filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox", "*"),
			// Chrome / Chromium (共有NSSデータベース)
			filepath.Join(home, ".pki", "nssdb"),
			filepath.Join(home, ".config", "chromium", "Default"),
			filepath.Join(home, ".config", "google-chrome", "Default"),
		}
	case "windows":
		appdata := os.Getenv("APPDATA")
		return []string{
			filepath.Join(appdata, "Mozilla", "Firefox", "Profiles", "*"),
		}
	}
	return nil
}
