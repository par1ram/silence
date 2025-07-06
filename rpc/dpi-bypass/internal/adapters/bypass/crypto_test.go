package bypass

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	cert, err := generateSelfSignedCert()
	assert.NoError(t, err)
	assert.NotNil(t, cert)
	assert.NotEmpty(t, cert.Certificate)

	// Проверяем, что сертификат парсится как x509
	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "localhost", parsed.Subject.CommonName)
	assert.Contains(t, parsed.DNSNames, "localhost")
	assert.Contains(t, parsed.DNSNames, "127.0.0.1")
}
