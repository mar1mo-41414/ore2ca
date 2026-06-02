//go:build darwin

package trust

import (
	"fmt"
	"os/exec"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

func installPlatform(s *store.Store) (*Result, error) {
	r := &Result{}
	certPath := s.CACertPath()

	// macOS Keychain
	cmd := exec.Command("sudo", "security", "add-trusted-cert",
		"-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain",
		certPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.OSErr = fmt.Errorf("keychain: %w\n%s", err, out)
	} else {
		r.OS = true
	}

	// Firefox NSS
	r.Firefox, r.FFErr = installNSS(certPath)
	return r, nil
}

func uninstallPlatform(s *store.Store) error {
	certPath := s.CACertPath()
	cmd := exec.Command("sudo", "security", "remove-trusted-cert", "-d", certPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain remove: %w\n%s", err, out)
	}
	return nil
}
