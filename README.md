# certsman
Some sample code that returns randomly but securely generated certificates for a given domain

## Requirements

One of the services I intend on rolling out very soon is an ACME-compliant certificate service. For now, we can treat it as a stub that simply returns the string foo-${domain} as a certificate when a HTTP GET request to /cert/{domain} is made. Implement a code base that:

* properly handles certificates for different domains
* allows certificates to live for 10 minutes before expiring
* sleeps for 10 seconds when a new certificate is generated 
* generates its own certificate and keeps its certificate up-to-date when it expires
* can generate multiple certificates at the same time