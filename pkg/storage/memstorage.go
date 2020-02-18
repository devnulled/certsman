package storage

import (
	"github.com/bluele/gcache"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/devnulled/certsman/pkg/certsman"
)

// InMemStorage in a type of persistence which is a machine local memory cache
type InMemStorage struct {
	Cache gcache.Cache
}

// CreateCertificate creates a cached certificate record in memory
func (i InMemStorage) CreateCertificate(req certsman.CertificateRequest, cert certsman.Certificate) (bool, error) {
	i.Cache.Set(req.Hostname, cert)
	log.Debug("Storing certificate for ", req.Hostname)
	return true, nil
}

// RetrieveCertificate retrives a cached certificate record from memory
func (i InMemStorage) RetrieveCertificate(req certsman.CertificateRequest) (certsman.Certificate, error) {

	var cert certsman.Certificate
	cachedCert, err := i.Cache.Get(req.Hostname)

	if err != nil {
		log.Debug("Unable to find cached certificate for ", req.Hostname)
		return certsman.Certificate{}, err
	}

	mapstructure.Decode(cachedCert, &cert)

	log.Debug("Found cached certificate for ", req.Hostname)
	return cert, nil
}

// UpdateCertificate updates a cached certificate record from memory but is currently not implemented
func (i InMemStorage) UpdateCertificate(req certsman.CertificateRequest, prevCert certsman.Certificate, currentCert certsman.Certificate) (certsman.Certificate, error) {
	return certsman.Certificate{}, nil
}

// DeleteCertificate removes a cached certificate record from memory but is currently not implemented
func (i InMemStorage) DeleteCertificate(req certsman.CertificateRequest) (bool, error) {
	return true, nil
}
