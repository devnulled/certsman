package memstorage

import (
	"github.com/bluele/gcache"
	"github.com/devnulled/certsman/pkg/certgen"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

// InMemStorage in a type of persistence which is a machine local memory cache
type InMemStorage struct {
	Cache gcache.Cache
}

// CreateCertificate creates a cached certificate record in memory
func (i InMemStorage) CreateCertificate(req certgen.CertificateRequest, cert certgen.Certificate) (bool, error) {
	i.Cache.Set(req.Hostname, cert)
	return true, nil
}

// RetrieveCertificate retrives a cached certificate record from memory
func (i InMemStorage) RetrieveCertificate(req certgen.CertificateRequest) (certgen.Certificate, error) {

	var cert certgen.Certificate
	cachedCert, err := i.Cache.Get(req.Hostname)

	if err != nil {
		return certgen.Certificate{}, err
	}

	mapstructure.Decode(cachedCert, &cert)

	log.Info(cert)

	return cert, nil
}

// UpdateCertificate updates a cached certificate record from memory but is currently not implemented
func (i InMemStorage) UpdateCertificate(req certgen.CertificateRequest, prevCert certgen.Certificate, currentCert certgen.Certificate) (certgen.Certificate, error) {
	return certgen.Certificate{}, nil
}

// DeleteCertificate removes a cached certificate record from memory but is currently not implemented
func (i InMemStorage) DeleteCertificate(req certgen.CertificateRequest) (bool, error) {
	return true, nil
}
