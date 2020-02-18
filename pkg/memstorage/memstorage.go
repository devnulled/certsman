package memstorage

import (
	"github.com/bluele/gcache"
	"github.com/devnulled/certsman/pkg/certgen"
	log "github.com/sirupsen/logrus"
)

type InMemStorage struct {
	Cache gcache.Cache
}

func (i InMemStorage) CreateCertificate(req certgen.CertificateRequest, cert certgen.Certificate) (bool, error) {
	i.Cache.Set(req.Hostname, cert)
	return true, nil
}

func (i InMemStorage) RetrieveCertificate(req certgen.CertificateRequest) (certgen.Certificate, error) {
	cert, err := i.Cache.GetIFPresent(req.Hostname)

	if err != nil {
		return certgen.Certificate{}, err
	}

	log.Info(cert)

	return certgen.Certificate{}, nil
}

func (i InMemStorage) UpdateCertificate(req certgen.CertificateRequest, prevCert certgen.Certificate, currentCert certgen.Certificate) (certgen.Certificate, error) {
	return certgen.Certificate{}, nil
}

func (i InMemStorage) DeleteCertificate(req certgen.CertificateRequest) (bool, error) {
	return true, nil
}
