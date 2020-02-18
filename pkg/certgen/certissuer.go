package certgen

import (
	"time"
)

// Provides a contract for various types of certificates to be generated/issued from
type CertificateIssuer interface {
	IssueCertificate(req CertificateRequest) (Certificate, error)
}

// Provides a contract for a request issued from a client to create/retrive a certificate for a given hostname
type CertificateRequest struct {
	Hostname string
}

// Provides a contract to respond to a request for a Certificate
type CertificateResponse struct {
	Certificate Certificate
	// Normally I'd create a single enum for these, but don't want to mess with the golang
	// tooling to do that at the moment since they aren't supported as a native data type
	WasCreated bool
	WasCached  bool
}

// Provides a contract for certificates to be returned to clients
type Certificate struct {
	Hostname        string
	CertificateBody string
	Expiration      time.Duration
}

// Interacts with a persistence provider to either retrieve a stored certificate and respond with it, or generate a new one, store it, and respond with it
func GetOrCreateCertificate(req CertificateRequest, ci CertificateIssuer, cpr CertificatePersistenceProvider) CertificateResponse {
	/*
		storedCert, retErr := cpr.RetrieveCertificate(req)
		if storedCert != nil {
			return storedCert
		} else if retErr != nil {
			return CertificateResponse{}
		} else {
			return CertificateResponse{}
		}
	*/

	return CertificateResponse{}
}
