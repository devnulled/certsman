package server

import (
	"time"

	"github.com/bluele/gcache"
	"github.com/devnulled/certsman/pkg/certgen"
	"github.com/devnulled/certsman/pkg/tokencert"
	log "github.com/sirupsen/logrus"
)

func RunServer() {

	gc := gcache.New(2000).
		ARC().
		Expiration(time.Minute * 10).
		Build()

	certRequest := certgen.CertificateRequest{
		Hostname: "yellow.com",
	}

	certGenerator := tokencert.TokenCertIssuer{KeyLength: 1024}

	thisCert, certErr := certGenerator.IssueCertificate(certRequest)

	if certErr == nil {
		log.Info(thisCert)
		gc.Set(certRequest.Hostname, thisCert)
	} else {
		log.Error(certErr)
	}

}
