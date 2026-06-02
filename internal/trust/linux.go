//go:build linux

package trust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ore2ca/ore2ca/internal/store"
)

func installPlatform(s *store.Store) (*Result, error) {
	r := &Result{}
	certPath := s.CACertPath()

	if err := installLinuxSystem(certPath); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}

	r.Firefox, r.FFErr = installNSS(certPath)
	return r, nil
}

func installLinuxSystem(certPath string) error {
	// Debian/Ubuntu
	if _, err := os.Stat("/usr/local/share/ca-certificates"); err == nil {
		dst := "/usr/local/share/ca-certificates/ore2ca.crt"
		data, err := os.ReadFile(certPath)
		if err != nil {
			return err
		}
		if err := writeFileRoot(dst, data); err != nil {
			return err
		}
		if out, err := exec.Command("sudo", "update-ca-certificates").CombinedOutput(); err != nil {
			return fmt.Errorf("update-ca-certificates: %w\n%s", err, out)
		}
		return nil
	}
	// RHEL/Fedora/CentOS
	if _, err := os.Stat("/etc/pki/ca-trust/source/anchors"); err == nil {
		dst := "/etc/pki/ca-trust/source/anchors/ore2ca.crt"
		data, err := os.ReadFile(certPath)
		if err != nil {
			return err
		}
		if err := writeFileRoot(dst, data); err != nil {
			return err
		}
		if out, err := exec.Command("sudo", "update-ca-trust", "extract").CombinedOutput(); err != nil {
			return fmt.Errorf("update-ca-trust: %w\n%s", err, out)
		}
		return nil
	}
	// Arch Linux
	if _, err := os.Stat("/etc/ca-certificates/trust-source/anchors"); err == nil {
		dst := "/etc/ca-certificates/trust-source/anchors/ore2ca.crt"
		data, err := os.ReadFile(certPath)
		if err != nil {
			return err
		}
		if err := writeFileRoot(dst, data); err != nil {
			return err
		}
		if out, err := exec.Command("sudo", "trust", "extract-compat").CombinedOutput(); err != nil {
			return fmt.Errorf("trust extract-compat: %w\n%s", err, out)
		}
		return nil
	}
	return fmt.Errorf("unsupported Linux distribution")
}

func writeFileRoot(dst string, data []byte) error {
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func uninstallPlatform(s *store.Store) error {
	paths := []string{
		"/usr/local/share/ca-certificates/ore2ca.crt",
		"/etc/pki/ca-trust/source/anchors/ore2ca.crt",
		"/etc/ca-certificates/trust-source/anchors/ore2ca.crt",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}
	// best-effort update
	exec.Command("sudo", "update-ca-certificates").Run()
	exec.Command("sudo", "update-ca-trust", "extract").Run()
	exec.Command("sudo", "trust", "extract-compat").Run()
	return nil
}
