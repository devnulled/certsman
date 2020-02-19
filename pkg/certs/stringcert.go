/*

The StringCertIssuer is a CertificateIssuer which provides a certificate based on
a simple genenrated string.

*/
package certs

import (
	"strings"
	"time"

	"github.com/devnulled/certsman/pkg/certsman"
	log "github.com/sirupsen/logrus"
)

// StringCertIssuer provides a simple certificate based on a configurable string
type StringCertIssuer struct {
	StringPrefix      string
	SleepEnabled      bool
	SleepyTimeSeconds time.Duration
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

	if i.SleepEnabled == true {
		// Would run as a go-routine if possible, but that might be cheating?
		sleepyTime(i.SleepyTimeSeconds*time.Second, req.Hostname)
	}

	return cert, nil
}

// SleepyTime is an artifical time to sleep
func sleepyTime(sleepTime time.Duration, hostname string) {
	log.Debug("String certificate sleeping for ", hostname)
	time.Sleep(sleepTime)
	log.Debug("String certificate sleeping complete for ", hostname)

}
