//go:build windows

package trust

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"golang.org/x/sys/windows/registry"
)

func installPlatform(s *store.Store) (*Result, error) {
	r := &Result{}
	certPath := s.CACertPath()

	if err := installWindowsRoot(certPath); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}

	r.Firefox, r.FFErr = installNSS(certPath)
	return r, nil
}

func installWindowsRoot(certPath string) error {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("invalid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	store, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	store.AddCert(cert)
	return nil
}

func uninstallPlatform(s *store.Store) error {
	return fmt.Errorf("Windows uninstall: please remove 'Ore2CA Local Root CA' from certmgr.msc manually")
}
