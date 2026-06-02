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
		return false, fmt.Errorf("certutil not found (install libnss3-tools or nss-tools)")
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
