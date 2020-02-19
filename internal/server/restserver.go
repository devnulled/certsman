package server

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/devnulled/certsman/pkg/certs"
	"github.com/devnulled/certsman/pkg/certsman"
	"github.com/devnulled/certsman/pkg/storage"

	"github.com/bluele/gcache"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Default hostname.  Should be refactored to be config driven or resolve on its own
const MyHostname = "localhost"

// Default key length for the TokenCerts that get generated
const DefaultTokenCertKeyLength = 1024

// Default expiration for certificates
const DefaultCertDurationMinutes = 10

// Maximum amount of certificates that get stored using the in-memory persistence
const InMemoryCertStorageLimit = 20000

// How long the server waits for connections to drain before doing a graceful shutdown
const DefaultServerGracefulTimeoutSeconds = 15

// Default address for the HTTP server to listen on
const DefaultServerAddress = "0.0.0.0:8080"

// Default name for the cert to self-generate
const DefaultCertServerName = "secure.certsman-server.com"

// Default amount of time to sleep as specified by the requirements
const DefaultArtificalSleepSeconds = 10

// Default string to use for the string cert issuer
const DefaultStringCertPrefix = "foo-"

// How long to wait before shutting the server down to let connections drain
var wait = time.Second * DefaultServerGracefulTimeoutSeconds

// Memory cache store
var memoryCache gcache.Cache

// InMemory persistence
var inMemPersist storage.InMemStorage

// The cert service which generates string based certificates
var stringCertIssuer certs.StringCertIssuer

// The cert service that that is compromised of the previous two impls
var stringCertService certsman.CerfificateService

// RunServer starts and runs the server
func RunServer() {

	log.SetLevel(log.InfoLevel)
	log.Info("certsman starting...")

	// Create a channel to communicate the expired cache items
	// expireChannel := make(chan string)

	memoryCache = gcache.New(InMemoryCertStorageLimit).
		ARC().
		EvictedFunc(func(key, value interface{}) {
			// The cache seems to be lazy expired
			// IE, only gets expunged when accessed OR
			// if the collection capacity is reached during
			// an addition.
			//
			// Trying to use this to expire the cache only
			// creates a race condition
			log.WithFields(log.Fields{
				"Hostname": fmt.Sprint(key),
			}).Debug("Certificate expired from cache")

			// expireChannel <- fmt.Sprint(key)
		}).
		//Expiration(time.Second * 20).
		Expiration(time.Minute * DefaultCertDurationMinutes).
		Build()

	// Process expired cache keys and renew our own cert if it gets evicted
	// go cacheExpirationHandler(expireChannel)

	inMemPersist = storage.InMemStorage{Cache: memoryCache}

	stringCertIssuer = certs.StringCertIssuer{
		StringPrefix:      DefaultStringCertPrefix,
		SleepEnabled:      true,
		SleepyTimeSeconds: DefaultArtificalSleepSeconds}

	stringCertService = certsman.CerfificateService{
		Issuer:      stringCertIssuer,
		Persistence: inMemPersist,
	}

	log.Info("Generating initial server cert")
	// In theory, this request should block as the server needs its own cert to startup successfully
	selfCertIssuer()

	log.Info("Starting timed task to refresh server cert in background")
	go selfCertIssueTimer()

	log.Info("Starting up certsman server at ", DefaultServerAddress)

	// Start our HTTP router/handler
	r := mux.NewRouter()
	r.HandleFunc("/cert/{hostname}", certificateGetHandler).Methods("GET")
	r.HandleFunc("/certtest/", certTestGetHandler).Methods("GET")

	srv := &http.Server{
		Addr: DefaultServerAddress,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
	log.Info("certsman stopping...")
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("certsman out!")
	os.Exit(0)
}

func cacheExpirationHandler(hostnames chan string) {
	hostname := <-hostnames

	if hostname == DefaultCertServerName {
		log.Info("Server cert expired from cache. Refreshing.")
		go selfCertIssuer()
	}
}

// certTestGetHandler is a convenience method for load testing
func certTestGetHandler(w http.ResponseWriter, r *http.Request) {
	// only use 2 chars so that some of the lookups are cached
	mux.Vars(r)["hostname"] = randomString(4)
	certificateGetHandler(w, r)
}

// randomString is used by certTestGetHandler to generate random hostnames
func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// certificateGetHandler is the main HTTP handler for certificate requests
func certificateGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	initHostname := vars["hostname"]

	if len(initHostname) <= 1 {
		http.Error(w, "Invalid hostname provided", http.StatusNotAcceptable)
		return
	}

	hostname := strings.ToLower(initHostname)

	reqID := requestIDGenerator()

	req := certsman.CertificateRequest{
		RequestID: reqID,
		Hostname:  hostname,
	}

	log.WithFields(log.Fields{
		"RequestID": reqID,
		"Hostname":  hostname,
	}).Trace("Certificate request recieved")

	resp := stringCertService.GetOrCreateCertificate(req)

	if !resp.IsSuccess {
		http.Error(w, resp.Error.Error(), resp.StatusCode)
		return
	}

	respBody := []byte(resp.Certificate.CertificateBody)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
	return
}

// requestIDGenerator generates RequestIds.  With more time, would use something like Zipkin.
func requestIDGenerator() string {
	u := uuid.NewV4()

	return u.String()
}

// selfCertIssuer is a convenience method for generating the certificate for the server itself
func selfCertIssuer() {
	// TODO: This is kind of hacky, maybe would refactor later to not use requests perhaps?
	reqID := requestIDGenerator()

	req := certsman.CertificateRequest{
		RequestID: reqID,
		Hostname:  DefaultCertServerName,
	}

	log.WithFields(log.Fields{
		"RequestID": reqID,
		"Hostname":  DefaultCertServerName,
	}).Debug("Updating self-cert for server")
	stringCertService.GetOrCreateCertificate(req)
}

// selfCertIssueTimer runs as a go routine every 5 seconds to refresh the servers own certificate.
// I think this is a better solution than having it wakeup every 10 minutes because of drift.
// Because it's own cert is cached for 10 minutes before its regenerated, the overhead of this call
// every 5 seconds is very minimal.  I think this tradeoff is much better than going up to a whole 10
// minutes without a certificate.
// There is opportunity to come-up with a much more complex way to do this that would ensure that
// a current certificate is always there, but I think it's beyond the scope of this exercise.
func selfCertIssueTimer() {
	tick := time.Tick(5 * time.Second)

	for range tick {
		log.Trace("Cert issuer timer is now refreshing the server certificate")
		selfCertIssuer()
	}
}
