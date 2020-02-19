package certs

import (
	"crypto/rand"
	"math/big"

	"github.com/devnulled/certsman/pkg/certsman"
)

// TokenCertIssuer provides a simple token based certificate generated from a securely generated string
type TokenCertIssuer struct {
	KeyLength int
}

// IssueCertificate Returns a generated token based certificate
func (t TokenCertIssuer) IssueCertificate(req certsman.CertificateRequest) (certsman.Certificate, error) {

	certStr, err := cryptoGenerator(t.KeyLength)

	if err == nil {
		cert := certsman.Certificate{
			Hostname:        req.Hostname,
			CertificateBody: certStr,
		}

		return cert, nil
	}

	return certsman.Certificate{}, err

}

// cryptoGenerator returns a securely generated random ASCII string.
// It reads random numbers from crypto/rand and searches for printable characters.
// It will return an error if the system's secure random number generator fails to
// function correctly, in which case the caller must not continue.
func cryptoGenerator(length int) (string, error) {
	result := ""
	for {
		if len(result) >= length {
			return result, nil
		}
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}
		n := num.Int64()
		// Make sure that the number/byte/letter is inside
		// the range of printable ASCII characters (excluding space and DEL)
		if n > 32 && n < 127 {
			result += string(n)
		}
	}
}
