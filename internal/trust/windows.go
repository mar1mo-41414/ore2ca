//go:build windows

package trust

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

// powershellUTF8Prefix forces UTF-8 output in PowerShell to avoid garbled text.
const powershellUTF8Prefix = `$OutputEncoding = [System.Text.Encoding]::UTF8; [Console]::OutputEncoding = [System.Text.Encoding]::UTF8; `

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
		powershellUTF8Prefix+fmt.Sprintf(
			`Import-Certificate -FilePath "%s" -CertStoreLocation Cert:\LocalMachine\Root`,
			certPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "0x80070005") || strings.Contains(outStr, "UnauthorizedAccessException") || strings.Contains(outStr, "AccessDenied") {
			return fmt.Errorf("管理者権限が必要です。\nPowerShell を「管理者として実行」して再度実行してください:\n\n    .\\ore2ca.exe trust")
		}
		return fmt.Errorf("Import-Certificate 失敗: %w\n%s", err, outStr)
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
		powershellUTF8Prefix+fmt.Sprintf(
			`Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -like "*%s*" } | Remove-Item`,
			cn))
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "0x80070005") || strings.Contains(outStr, "UnauthorizedAccessException") {
			return fmt.Errorf("管理者権限が必要です。PowerShell を「管理者として実行」して再度実行してください")
		}
		return fmt.Errorf("Remove-Item 失敗: %w\n%s", err, outStr)
	}
	return nil
}
