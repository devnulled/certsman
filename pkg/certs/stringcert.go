package certs

import (
	"strings"

	"github.com/devnulled/certsman/pkg/certsman"
	log "github.com/sirupsen/logrus"
)

// StringCertIssuer provides a simple certificate based on a configurable string
type StringCertIssuer struct {
	StringPrefix string
}

// IssueCertificate returns a string based certificate
func (i StringCertIssuer) IssueCertificate(req certsman.CertificateRequest) (certsman.Certificate, error) {
	var certBuilder strings.Builder
	certBuilder.WriteString(i.StringPrefix)
	certBuilder.WriteString(req.Hostname)

	cert := certsman.Certificate{
		Hostname:        req.Hostname,
		CertificateBody: certBuilder.String(),
	}
	log.Info("String certificate issued for ", req.Hostname)
	return cert, nil
}
