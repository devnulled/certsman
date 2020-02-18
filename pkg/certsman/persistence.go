package certsman

// CertificatePersistenceProvider provides a simple contract to use for anything that persists Certificates to memory, databases, cache, disk, etc.
type CertificatePersistenceProvider interface {
	CreateCertificate(req CertificateRequest, cert Certificate) (bool, error)
	RetrieveCertificate(req CertificateRequest) (Certificate, error)
	UpdateCertificate(req CertificateRequest, prevCert Certificate, currentCert Certificate) (Certificate, error)
	DeleteCertificate(req CertificateRequest) (bool, error)
}
