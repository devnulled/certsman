package certs

import (
	"testing"

	"github.com/devnulled/certsman/pkg/certsman"
	"github.com/stretchr/testify/assert"
)

func TestIssueCertificate(t *testing.T) {
	prefix := "myprefix-"
	hostname := "myhostname"

	expectedResult := prefix + hostname

	var strCertType = StringCertIssuer{StringPrefix: prefix, SleepyTimeSeconds: 0}

	req := certsman.CertificateRequest{Hostname: hostname, RequestID: "blah"}

	myCert, err := strCertType.IssueCertificate(req)

	assert.Nil(t, err, "An error shouldn't have occurred")
	assert.Equal(t, hostname, myCert.Hostname, "The expected hostname was not correct")
	assert.Equal(t, expectedResult, expectedResult, "The expecvted certificate body was not correct")
}

func BenchmarkIssueCertificate(b *testing.B) {

	prefix := "myprefix-"
	hostname := "myhostname"

	var strCertType = StringCertIssuer{StringPrefix: prefix, SleepyTimeSeconds: 0}

	req := certsman.CertificateRequest{Hostname: hostname, RequestID: "blah"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strCertType.IssueCertificate(req)
	}
}
