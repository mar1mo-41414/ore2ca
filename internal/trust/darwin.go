//go:build darwin

package trust

import (
	"fmt"
	"os/exec"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

func installPlatform(s *store.Store, opts Options) (*Result, error) {
	r := &Result{}
	certPath := s.CACertPath()

	// macOS Keychain — always
	cmd := exec.Command("sudo", "security", "add-trusted-cert",
		"-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain",
		certPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.OSErr = fmt.Errorf("keychain: %w\n%s", err, out)
	} else {
		r.OS = true
	}

	// Firefox (NSS) — auto-detect or explicit --firefox
	if shouldDoFirefox(opts) {
		ok, err := installNSSForFirefox(certPath)
		r.Firefox = BrowserResult{Registered: ok, Err: err}
	} else {
		r.Firefox = BrowserResult{Skipped: true}
	}

	// Chrome on macOS uses Keychain → already covered by OS trust above
	r.Chrome = BrowserResult{Skipped: true}

	return r, nil
}

func uninstallPlatform(s *store.Store, opts Options) (*UninstallResult, error) {
	r := &UninstallResult{}

	cmd := exec.Command("sudo", "security", "remove-trusted-cert", "-d", s.CACertPath())
	if out, err := cmd.CombinedOutput(); err != nil {
		r.OSErr = fmt.Errorf("keychain remove: %w\n%s", err, out)
	} else {
		r.OS = true
	}

	if shouldDoFirefox(opts) {
		ok, err := uninstallNSSForFirefox()
		r.Firefox = BrowserResult{Registered: ok, Err: err}
	} else {
		r.Firefox = BrowserResult{Skipped: true}
	}

	r.Chrome = BrowserResult{Skipped: true}

	return r, nil
}
