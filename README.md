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

I tried to use the cache expiration to communicate via a channel to pick-up on when the cache entry for the
hostname of the server expired. This would allow the service to generate a new cert automatically when it was reaped
from the cache.  

However, the cache is lazy reaped, so it gets expunged when either the cache is full, or
upon access when the key is accessed but is expired.  Using a reaper method to automatically update the cert creates a race condition so it really isn't a solution at all. You can look through `internal/server/restserver.go` to see some of my leftovers from trying this out.

For now, the best way I can think of to keep it updated is just to access its own cert frequently so that it keeps it
up to date.  The overhead of trying to load the servers cert from cache is very low.

I think running a timer every 10 minutes is going to have drift and get off pretty easily.

The only other thing I can think of at the moment would be to introduce my own background routine on a timer that
reaps the expired cache on it's own, and then uses that process to announce an event when the servers cert has
been reaped, so that it can be created again.

After some more testing, I now realize that dupli

## API

### GET /cert/{domain}

Returns a simple string based certificate given the domain name you pass in.  If one has already been created and isn't expired,
returns the one which was already created.  Otherwise, generates a new one.

### GET /certtest/

A convenience method to return a certificate which has been generated/retrieved for a randomly generated domain name. This
is just a testing URL for convenience as it's hard to use load testing tools with generated URI's.

