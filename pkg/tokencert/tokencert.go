// Provides a simple token based certificate generated from a securely generated string
package tokencert

import (
	"crypto/rand"
	"math/big"

	"github.com/devnulled/certsman/pkg/certgen"
)

// TokenCertIssuer user-defined type
type TokenCertIssuer struct{}

// Returns a generated certificate
func (t TokenCertIssuer) IssueCertificate(req certgen.CertificateRequest) (certgen.Certificate, error) {

	certStr, err := cryptoGenerator(1028)

	if err == nil {
		cert := certgen.Certificate{
			Hostname:        req.Hostname,
			CertificateBody: certStr,
		}
		return cert, nil
	} else {
		return certgen.Certificate{}, err
	}

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
