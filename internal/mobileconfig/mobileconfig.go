// Package mobileconfig generates iOS configuration profiles (.mobileconfig)
// for distributing CA certificates to Apple devices.
package mobileconfig

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"text/template"
)

const profileTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadCertificateFileName</key>
			<string>root.crt</string>
			<key>PayloadContent</key>
			<data>{{.CertBase64}}</data>
			<key>PayloadDescription</key>
			<string>Adds {{.CommonName}} as a trusted CA certificate</string>
			<key>PayloadDisplayName</key>
			<string>{{.CommonName}}</string>
			<key>PayloadIdentifier</key>
			<string>com.ore2ca.cert.{{.ContentUUID}}</string>
			<key>PayloadType</key>
			<string>com.apple.security.root</string>
			<key>PayloadUUID</key>
			<string>{{.ContentUUID}}</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
		</dict>
	</array>
	<key>PayloadDescription</key>
	<string>Install {{.CommonName}} as a trusted root CA certificate</string>
	<key>PayloadDisplayName</key>
	<string>{{.CommonName}}</string>
	<key>PayloadIdentifier</key>
	<string>com.ore2ca.profile.{{.TopUUID}}</string>
	<key>PayloadRemovalDisallowed</key>
	<false/>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>{{.TopUUID}}</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>
`

type profileData struct {
	CommonName string
	CertBase64 string
	ContentUUID string
	TopUUID     string
}

// newUUID generates an RFC4122 v4 UUID using crypto/rand.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// Generate creates a .mobileconfig plist for the given CA certificate.
// commonName is the CA's CommonName; certDER is the raw DER-encoded certificate bytes.
func Generate(commonName string, certDER []byte) ([]byte, error) {
	contentUUID, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("UUID生成失敗: %w", err)
	}
	topUUID, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("UUID生成失敗: %w", err)
	}

	data := profileData{
		CommonName:  commonName,
		CertBase64:  base64.StdEncoding.EncodeToString(certDER),
		ContentUUID: contentUUID,
		TopUUID:     topUUID,
	}

	t, err := template.New("mobileconfig").Parse(profileTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
