package ca

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/store"
)

// BuildCRL generates a CRL from all revoked certificates and writes it to
// ~/.ore2ca/ca/crl.pem.
//
// Note: requires the CA certificate to have SubjectKeyId set.
// If the CA was created with an older version of ore2ca, run `ore2ca init --force`
// to regenerate the CA certificate with SubjectKeyId.
func BuildCRL(s *store.Store) error {
	caCert, caKey, err := LoadCA(s)
	if err != nil {
		return fmt.Errorf("load CA: %w", err)
	}

	if len(caCert.SubjectKeyId) == 0 {
		return fmt.Errorf(
			"CA証明書に SubjectKeyId がありません。\n" +
				"`ore2ca init --force` でCAを再作成してからもう一度お試しください。",
		)
	}

	metas, err := s.ListCerts()
	if err != nil {
		return fmt.Errorf("証明書一覧の取得失敗: %w", err)
	}

	var entries []x509.RevocationListEntry
	for _, m := range metas {
		if !m.Revoked {
			continue
		}
		serial := new(big.Int)
		serial.SetString(m.Serial, 10)
		entries = append(entries, x509.RevocationListEntry{
			SerialNumber:   serial,
			RevocationTime: m.RevokedAt,
		})
	}

	now := time.Now()
	template := &x509.RevocationList{
		// CRL シリアル番号: Unix 時刻を使うことで単調増加を保証
		Number:                    big.NewInt(now.Unix()),
		ThisUpdate:                now,
		NextUpdate:                now.Add(7 * 24 * time.Hour), // 1 週間後
		RevokedCertificateEntries: entries,
	}

	crlDER, err := x509.CreateRevocationList(rand.Reader, template, caCert, caKey)
	if err != nil {
		return fmt.Errorf("CRL 生成失敗: %w", err)
	}

	f, err := os.OpenFile(s.CRLPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("CRL ファイル書き込み失敗: %w", err)
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: "X509 CRL", Bytes: crlDER})
}
