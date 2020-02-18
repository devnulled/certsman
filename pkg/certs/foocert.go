package certs

import (
	"github.com/devnulled/certsman/pkg/certsman"
)

// StringCertIssuer provides a simple certificate based on a configurable string
type StringCertIssuer struct {
	StringPrefix string
}

func (i StringCertIssuer) IssueCertificate(req certsman.CertificateRequest) (certsman.Certificate, error) {

	certBody := i.StringPrefix + req.Hostname

	cert := certsman.Certificate{
		Hostname:        req.Hostname,
		CertificateBody: certBody,
	}
	return cert, nil
}
