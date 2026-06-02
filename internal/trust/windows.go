//go:build windows

package trust

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

// powershellUTF8Prefix forces UTF-8 output in PowerShell to reduce garbled text.
const powershellUTF8Prefix = `$OutputEncoding = [System.Text.Encoding]::UTF8; [Console]::OutputEncoding = [System.Text.Encoding]::UTF8; `

func installPlatform(s *store.Store) (*Result, error) {
	r := &Result{}
	if err := installWindowsRoot(s.CACertPath()); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}
	// Modern Firefox (v110+) no longer ships certutil.exe.
	// Use Firefox Enterprise Policy instead.
	if err := installFirefoxPolicy(s.CACertPath()); err != nil {
		r.FFErr = err
	} else {
		r.Firefox = true
	}
	return r, nil
}

func uninstallPlatform(s *store.Store) (*UninstallResult, error) {
	r := &UninstallResult{}
	if err := uninstallWindowsRoot(s); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}
	if err := uninstallFirefoxPolicy(s.CACertPath()); err != nil {
		r.FFErr = err
	} else {
		r.Firefox = true
	}
	return r, nil
}

// --- OS trust store ---

func installWindowsRoot(certPath string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		powershellUTF8Prefix+fmt.Sprintf(
			`Import-Certificate -FilePath "%s" -CertStoreLocation Cert:\LocalMachine\Root`,
			certPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "0x80070005") ||
			strings.Contains(outStr, "UnauthorizedAccessException") ||
			strings.Contains(outStr, "AccessDenied") {
			return fmt.Errorf("管理者権限が必要です。\nPowerShell を「管理者として実行」して再度実行してください:\n\n    .\\ore2ca.exe trust")
		}
		return fmt.Errorf("Import-Certificate 失敗: %w\n%s", err, outStr)
	}
	return nil
}

func uninstallWindowsRoot(s *store.Store) error {
	// Read CN from cert to identify it in the store
	certPath := s.CACertPath()
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		powershellUTF8Prefix+fmt.Sprintf(
			`$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2 "%s"; `+
				`Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Thumbprint -eq $cert.Thumbprint } | Remove-Item`,
			certPath))
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

// --- Firefox Enterprise Policy ---

type firefoxPoliciesFile struct {
	Policies firefoxPolicies `json:"policies"`
}

type firefoxPolicies struct {
	Certificates firefoxCerts `json:"Certificates,omitempty"`
}

type firefoxCerts struct {
	Install []string `json:"Install,omitempty"`
}

func firefoxInstallDir() (string, error) {
	candidates := []string{
		`C:\Program Files\Mozilla Firefox`,
		`C:\Program Files (x86)\Mozilla Firefox`,
	}
	for _, d := range candidates {
		if _, err := os.Stat(filepath.Join(d, "firefox.exe")); err == nil {
			return d, nil
		}
	}
	return "", fmt.Errorf("Firefox のインストールフォルダが見つかりません")
}

func installFirefoxPolicy(certPath string) error {
	dir, err := firefoxInstallDir()
	if err != nil {
		return err
	}
	return writeFirefoxPolicy(dir, certPath, true)
}

func uninstallFirefoxPolicy(certPath string) error {
	dir, err := firefoxInstallDir()
	if err != nil {
		return err
	}
	return writeFirefoxPolicy(dir, certPath, false)
}

func writeFirefoxPolicy(firefoxDir, certPath string, add bool) error {
	distDir := filepath.Join(firefoxDir, "distribution")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("distribution ディレクトリ作成失敗: %w", err)
	}
	policiesPath := filepath.Join(distDir, "policies.json")

	pf := firefoxPoliciesFile{}
	if data, err := os.ReadFile(policiesPath); err == nil {
		_ = json.Unmarshal(data, &pf)
	}

	certs := pf.Policies.Certificates.Install
	if add {
		found := false
		for _, p := range certs {
			if strings.EqualFold(p, certPath) {
				found = true
				break
			}
		}
		if !found {
			certs = append(certs, certPath)
		}
	} else {
		filtered := certs[:0]
		for _, p := range certs {
			if !strings.EqualFold(p, certPath) {
				filtered = append(filtered, p)
			}
		}
		certs = filtered
	}
	pf.Policies.Certificates.Install = certs

	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(policiesPath, data, 0644); err != nil {
		if strings.Contains(err.Error(), "Access is denied") || strings.Contains(err.Error(), "access denied") {
			return fmt.Errorf("管理者権限が必要です。PowerShell を「管理者として実行」して再度実行してください")
		}
		return fmt.Errorf("policies.json 書き込み失敗: %w", err)
	}
	return nil
}
