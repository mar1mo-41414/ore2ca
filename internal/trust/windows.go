//go:build windows

package trust

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

func installPlatform(s *store.Store) (*Result, error) {
	r := &Result{}
	if err := installWindowsRoot(s.CACertPath()); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}
	r.Firefox, r.FFErr = installNSS(s.CACertPath())
	return r, nil
}

func installWindowsRoot(certPath string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(`Import-Certificate -FilePath "%s" -CertStoreLocation Cert:\LocalMachine\Root`, certPath))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Import-Certificate: %w\n%s", err, out)
	}
	return nil
}

func uninstallPlatform(s *store.Store) error {
	data, err := os.ReadFile(s.CACertPath())
	if err != nil {
		return err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("CA証明書のPEM解析に失敗しました")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	cn := cert.Subject.CommonName
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(`Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -like "*%s*" } | Remove-Item`, cn))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Remove-Item: %w\n%s", err, out)
	}
	return nil
}
