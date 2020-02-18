package certgen

import (
	"time"
)

// CertificateIssuer provides a contract for various types of certificates to be generated/issued from
type CertificateIssuer interface {
	IssueCertificate(req CertificateRequest) (Certificate, error)
}

// CertificateRequest provides a contract for a request issued from a client to create/retrive a certificate for a given hostname
type CertificateRequest struct {
	Hostname string
}

// CertificateResponse provides a contract to respond to a request for a Certificate
type CertificateResponse struct {
	StatusCode          int
	Error               error
	CertificateHostname string
	Certificate         Certificate

	// Normally I'd create a single enum for these, but don't want to mess with the golang
	// tooling to do that at the moment since enums aren't supported as a native data type
	WasCreated bool
	WasCached  bool
}

// Certificate provides a contract for certificates to be created, stored, and returned to clients
type Certificate struct {
	Hostname        string
	CertificateBody string
	Expiration      time.Duration
}

// GetOrCreateCertificate interacts with a persistence provider to either retrieve a stored certificate and respond with it, or generate a new one, store it, and respond with it
func GetOrCreateCertificate(req CertificateRequest, ci CertificateIssuer, cpr CertificatePersistenceProvider) CertificateResponse {
	storedCert, retErr := cpr.RetrieveCertificate(req)

	if retErr == nil {
		// Certificate must not be cached.  Create a new one and store.
		newCert, createErr := ci.IssueCertificate(req)

		if createErr != nil {
			// Something bad happened.  Lets bail.
			resp := marshallErrResponse(req, createErr)
			return sleepyResponder(resp)
		}

		// Created new cert.  Lets store it and then respond.
		_, storeErr := cpr.CreateCertificate(req, newCert)

		if storeErr != nil {
			// Something bad happened.  Lets bail.
			resp := marshallErrResponse(req, storeErr)
			return sleepyResponder(resp)
		}

		// Persistence was also successful! Lets return the cert now that it's been stored
		resp := marshallCertificateResponse(req, newCert, true, false)
		return sleepyResponder(resp)
	}

	// Found the cert in persistence, lets return it
	resp := marshallCertificateResponse(req, storedCert, true, false)
	return sleepyResponder(resp)
}

// marshallCertificateResponse marshalls various data into a successful certificate response
func marshallCertificateResponse(req CertificateRequest, cert Certificate, wasCreated bool, wasCached bool) CertificateResponse {
	resp := CertificateResponse{
		StatusCode:          200,
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
		Error:               err,
		CertificateHostname: req.Hostname,
		Certificate:         Certificate{},
		WasCreated:          false,
		WasCached:           false,
	}

	return resp
}

// sleepyResponder sleeps for a bit, then responds
func sleepyResponder(resp CertificateResponse) CertificateResponse {

	//TODO: Make this configurable somewhere else
	time.Sleep(1 * time.Second)

	return resp
}
