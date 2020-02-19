# certsman
Some sample golang code that returns randomly generated strings as certificates for a given domain

## Requirements

One of the services I intend on rolling out very soon is an ACME-compliant certificate service. For now, we can treat it as a stub that simply returns the string foo-${domain} as a certificate when a HTTP GET request to /cert/{domain} is made. Implement a code base that:

* properly handles certificates for different domains
* allows certificates to live for 10 minutes before expiring
* sleeps for 10 seconds when a new certificate is generated 
* generates its own certificate and keeps its certificate up-to-date when it expires
* can generate multiple certificates at the same time

## Building

Built with go 1.13.x

To do a full build, run tests, then build mac and linux binaries, simply run `make`

## Development

To build, test, and run a Mac binary locally, simply run `make build-test`

The binary will launch a server at http://localhost:8080

If you want to run a simple load test against this server, open a new terminal and run `make basic-load-test`

You can alter the logging level, and disable things like the arbitrary sleep in `internal/server/restserver.go`

## Design Notes

The way that the application makes sure to keep the cert for itself is pretty hacky and far from perfect.  I just ran out of time to fully work through the problem.

I think the best way to tackle this problem would be to add a function to the cache whenever it expires an item, to send the key name to a channel.  Whatever function is waiting on the other end of this channel would check to see if the cert being expired matched the cert for the server itself.  If so, it would automatically go create one again.

## API

### GET /cert/{domain}

Returns a simple string based certificate given the domain name you pass in.  If one has already been created and isn't expired,
returns the one which was already created.  Otherwise, generates a new one.

### GET /certtest/

A convenience method to return a certificate which has been generated/retrieved for a randomly generated domain name. This
is just a testing URL for convenience as it's hard to use load testing tools with generated URI's.

