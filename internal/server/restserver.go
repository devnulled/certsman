package server

import (
	"github.com/devnulled/certsman/pkg/certgen"
	log "github.com/sirupsen/logrus"
)

func RunServer() {
	log.Info("hi there")
	certRequest := certgen.CertificateRequest{
		Hostname: "yellow.com",
	}

	log.Info(certRequest)

	/*
		var certGenerator tokencert.TokenCertIssuer

		thisCert, certErr := certGenerator.TokenCertIssuer.IssueCertificate(certRequest)

		if certErr == nil {
			log.Info(thisCert)
		} else {
			log.Error(certErr)
		}
	*/
}
