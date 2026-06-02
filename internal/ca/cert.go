package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/config"
	"github.com/mar1mo-41414/ore2ca/internal/store"
)

// Issue issues a certificate for domain with optional extra SANs.
func Issue(domain string, cfg *config.Config, s *store.Store, extraSANs ...string) (*store.CertMeta, error) {
	caCert, caKey, err := LoadCA(s)
	if err != nil {
		return nil, fmt.Errorf("load CA: %w", err)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	serial, err := nextSerial(s)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	notAfter := now.AddDate(0, 0, cfg.Certs.ValidDays)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{cfg.CA.Organization},
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	addSANs(template, domain)
	for _, san := range extraSANs {
		addSANs(template, san)
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.CertDir(domain), 0700); err != nil {
		return nil, err
	}

	if err := writeCert(s.CertPath(domain), certDER); err != nil {
		return nil, err
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	if err := writeKey(s.KeyPath(domain), keyDER); err != nil {
		return nil, err
	}

	if err := writeChain(s.ChainPath(domain), s.CACertPath()); err != nil {
		return nil, err
	}

	if err := writeFullchain(s.FullchainPath(domain), certDER, s.CACertPath()); err != nil {
		return nil, err
	}

	meta := &store.CertMeta{
		ID:        fmt.Sprintf("%04d", serial),
		Domain:    domain,
		SANs:      extraSANs,
		IssuedAt:  now,
		ExpiresAt: notAfter,
		Serial:    fmt.Sprintf("%d", serial),
	}
	return meta, s.SaveMeta(meta)
}

func addSANs(tmpl *x509.Certificate, domain string) {
	if ip := net.ParseIP(domain); ip != nil {
		tmpl.IPAddresses = appendIPUniq(tmpl.IPAddresses, ip)
		return
	}
	tmpl.DNSNames = appendStrUniq(tmpl.DNSNames, domain)
	// localhost は IP SANs も追加
	if domain == "localhost" {
		tmpl.IPAddresses = appendIPUniq(tmpl.IPAddresses, net.ParseIP("127.0.0.1"))
		tmpl.IPAddresses = appendIPUniq(tmpl.IPAddresses, net.ParseIP("::1"))
	}
	// ワイルドカードの場合はベースドメインも追加
	if strings.HasPrefix(domain, "*.") {
		base := strings.TrimPrefix(domain, "*.")
		tmpl.DNSNames = appendStrUniq(tmpl.DNSNames, base)
	}
}

func appendIPUniq(list []net.IP, ip net.IP) []net.IP {
	for _, existing := range list {
		if existing.Equal(ip) {
			return list
		}
	}
	return append(list, ip)
}

func appendStrUniq(list []string, s string) []string {
	for _, existing := range list {
		if existing == s {
			return list
		}
	}
	return append(list, s)
}

func nextSerial(s *store.Store) (int64, error) {
	serial, err := s.ReadSerial()
	if err != nil {
		return 0, err
	}
	return serial, s.WriteSerial(serial + 1)
}

func writeChain(chainPath, caCertPath string) error {
	data, err := os.ReadFile(caCertPath)
	if err != nil {
		return err
	}
	return os.WriteFile(chainPath, data, 0644)
}

func writeFullchain(fullchainPath string, certDER []byte, caCertPath string) error {
	f, err := os.OpenFile(fullchainPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	caData, err := os.ReadFile(caCertPath)
	if err != nil {
		return err
	}
	_, err = f.Write(caData)
	return err
}
