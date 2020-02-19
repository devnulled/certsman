package certsman

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// CertificateIssuer provides a contract for various types of certificates to be generated/issued from
type CertificateIssuer interface {
	IssueCertificate(req CertificateRequest) (Certificate, error)
}

// CertificateRequest provides a contract for a request issued from a client to create/retrive a certificate for a given hostname
type CertificateRequest struct {
	// Generated on each request for tracing/debugging purposes
	RequestID string
	// The hostname being requested for a certificate
	Hostname string
}

// CertificateResponse provides a contract to respond to a request for a Certificate
type CertificateResponse struct {
	// The original request ID
	RequestID string

	// Whether or not the request was successful
	IsSuccess bool

	// Status code being returned to the client
	StatusCode int

	// A fatal error that happened during the request
	Error error

	// The hostname the certificate was requested for
	CertificateHostname string

	// The actual certificate being returned
	Certificate Certificate

	// Normally I'd create a single enum for these, but don't want to mess with the golang
	// tooling to do that at the moment since enums aren't supported as a native data type
	WasCreated bool
	WasCached  bool
}

// Certificate provides a contract for certificates to be created, stored, and returned to clients
type Certificate struct {
	// Hostname for the certificate
	Hostname string
	// The body that represents the certificate
	CertificateBody string

	// TODO: Should actually be a datetime
	// How long until this certificate expires
	Expiration time.Duration
}

// CerfificateService provides a contract for a particular certificate implementation, and it's backing persistance implementation
type CerfificateService struct {
	// The particular certificate Issuer
	Issuer      CertificateIssuer
	Persistence CertificatePersistenceProvider
}

// GetOrCreateCertificate interacts with a CertificateService to either retrieve a stored certificate and respond with it, or generate a new one, store it, and respond with it
func (svc CerfificateService) GetOrCreateCertificate(req CertificateRequest) CertificateResponse {
	storedCert, retErr := svc.Persistence.RetrieveCertificate(req)

	if retErr != nil {
		log.Debug("No cert found in persistance for ", req.Hostname)
		// The certificate must not exist.  Create a new one and store.
		newCert, createErr := svc.Issuer.IssueCertificate(req)

		if createErr != nil {
			// Something bad happened.  Lets bail.
			log.Error("Unable to create cert for ", req.Hostname)
			resp := marshallErrResponse(req, createErr)
			return resp
		}

		// Created new cert.  Lets store it and then respond.
		_, storeErr := svc.Persistence.CreateCertificate(req, newCert)

		if storeErr != nil {
			// Something bad happened.  Lets bail.
			log.Error("Unable to store cert for ", req.Hostname)
			resp := marshallErrResponse(req, storeErr)
			return resp
		}

		// Persistence was also successful! Lets return the cert now that it's been stored
		resp := marshallCertificateResponse(req, newCert, true, false)
		return resp
	}

	// Found the cert in persistence, lets return it
	log.Debug("Cert found in persistance for ", req.Hostname)
	resp := marshallCertificateResponse(req, storedCert, true, false)
	return resp
}

// marshallCertificateResponse marshalls various data into a successful certificate response
func marshallCertificateResponse(req CertificateRequest, cert Certificate, wasCreated bool, wasCached bool) CertificateResponse {
	resp := CertificateResponse{
		StatusCode:          200,
		IsSuccess:           true,
		Error:               nil,
		CertificateHostname: req.Hostname,
		Certificate:         cert,
		WasCreated:          wasCreated,
		WasCached:           wasCached,
	}
	return resp
}

// marshallErrResponse marshalls an error into an CertificateResponse
func marshallErrResponse(req CertificateRequest, err error) CertificateResponse {
	resp := CertificateResponse{
		StatusCode:          500,
		IsSuccess:           false,
		Error:               err,
		CertificateHostname: req.Hostname,
		Certificate:         Certificate{},
		WasCreated:          false,
		WasCached:           false,
	}
	return resp
}
