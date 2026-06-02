package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/config"
)

type CertMeta struct {
	ID        string    `json:"id"`
	Domain    string    `json:"domain"`
	SANs      []string  `json:"sans,omitempty"` // extra SANs passed to ore2ca issue --san
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Serial    string    `json:"serial"`
	Revoked   bool      `json:"revoked"`
	RevokedAt time.Time `json:"revoked_at,omitempty"`
}

type Store struct {
	baseDir string
}

func New() (*Store, error) {
	dir, err := config.HomeDir()
	if err != nil {
		return nil, err
	}
	s := &Store{baseDir: dir}
	return s, s.init()
}

func (s *Store) init() error {
	dirs := []string{
		s.baseDir,
		s.caDir(),
		s.certsDir(),
		s.revokedDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) BaseDir() string    { return s.baseDir }
func (s *Store) caDir() string      { return filepath.Join(s.baseDir, "ca") }
func (s *Store) certsDir() string   { return filepath.Join(s.baseDir, "certs") }
func (s *Store) revokedDir() string { return filepath.Join(s.baseDir, "revoked") }

func (s *Store) CACertPath() string { return filepath.Join(s.caDir(), "root.crt") }
func (s *Store) CAKeyPath() string  { return filepath.Join(s.caDir(), "root.key") }
func (s *Store) SerialPath() string { return filepath.Join(s.caDir(), "serial") }
func (s *Store) CRLPath() string    { return filepath.Join(s.caDir(), "crl.pem") }

func (s *Store) CAExists() bool {
	_, err := os.Stat(s.CACertPath())
	return err == nil
}

func (s *Store) CertDir(domain string) string {
	return filepath.Join(s.certsDir(), sanitizeDomain(domain))
}

func (s *Store) CertPath(domain string) string {
	return filepath.Join(s.CertDir(domain), "cert.crt")
}

func (s *Store) KeyPath(domain string) string {
	return filepath.Join(s.CertDir(domain), "cert.key")
}

func (s *Store) ChainPath(domain string) string {
	return filepath.Join(s.CertDir(domain), "chain.crt")
}

func (s *Store) FullchainPath(domain string) string {
	return filepath.Join(s.CertDir(domain), "fullchain.crt")
}

func (s *Store) MetaPath(domain string) string {
	return filepath.Join(s.CertDir(domain), "meta.json")
}

func (s *Store) SaveMeta(meta *CertMeta) error {
	if err := os.MkdirAll(s.CertDir(meta.Domain), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.MetaPath(meta.Domain), data, 0600)
}

func (s *Store) LoadMeta(domain string) (*CertMeta, error) {
	data, err := os.ReadFile(s.MetaPath(domain))
	if err != nil {
		return nil, err
	}
	var meta CertMeta
	return &meta, json.Unmarshal(data, &meta)
}

func (s *Store) ListCerts() ([]*CertMeta, error) {
	entries, err := os.ReadDir(s.certsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var metas []*CertMeta
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(s.certsDir(), e.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta CertMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		metas = append(metas, &meta)
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].IssuedAt.Before(metas[j].IssuedAt)
	})
	return metas, nil
}

func (s *Store) FindByID(id string) (*CertMeta, error) {
	metas, err := s.ListCerts()
	if err != nil {
		return nil, err
	}
	for _, m := range metas {
		if m.ID == id || strings.HasPrefix(m.ID, id) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("certificate not found: %s", id)
}

func (s *Store) DeleteCert(domain string) error {
	return os.RemoveAll(s.CertDir(domain))
}

func (s *Store) ReadSerial() (int64, error) {
	data, err := os.ReadFile(s.SerialPath())
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	var serial int64
	_, err = fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &serial)
	return serial, err
}

func (s *Store) WriteSerial(serial int64) error {
	return os.WriteFile(s.SerialPath(), []byte(fmt.Sprintf("%d\n", serial)), 0600)
}

func sanitizeDomain(domain string) string {
	return strings.ReplaceAll(domain, "*", "_wildcard_")
}
