package trust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// installNSS tries to register the CA cert into Firefox NSS databases.
func installNSS(certPath string) (bool, error) {
	certutil, err := exec.LookPath("certutil")
	if err != nil {
		return false, certutilNotFoundError()
	}

	dbs := findNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("no Firefox NSS profile found")
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

func findNSSDBs() []string {
	var dirs []string
	home, err := os.UserHomeDir()
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

func uninstallNSS() (bool, error) {
	certutil, err := exec.LookPath("certutil")
	if err != nil {
		return false, certutilNotFoundError()
	}
	dbs := findNSSDBs()
	if len(dbs) == 0 {
		return false, fmt.Errorf("Firefox NSS プロファイルが見つかりません")
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

func certutilNotFoundError() error {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    brew install nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	case "linux":
		return fmt.Errorf("certutil が見つかりません。以下を実行してインストールしてください:\n\n    # Debian/Ubuntu:\n    sudo apt install libnss3-tools\n\n    # Fedora/RHEL:\n    sudo dnf install nss-tools\n\n    # Arch:\n    sudo pacman -S nss\n\nインストール後、再度 ore2ca trust を実行してください。")
	case "windows":
		return fmt.Errorf("certutil が見つかりません。Firefox NSS への登録には certutil が必要です。")
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
			filepath.Join(home, ".mozilla", "firefox", "*"),
			filepath.Join(home, ".firefox", "*"),
			filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox", "*"),
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
