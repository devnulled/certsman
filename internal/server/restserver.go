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

// InMemory persistance
var inMemPersist storage.InMemStorage

// The cert service which generates string based certificates
var stringCertIssuer certs.StringCertIssuer

// The cert service that that is compromised of the previous two impls
var stringCertService certsman.CerfificateService

// RunServer starts and runs the server
func RunServer() {

	log.SetLevel(log.InfoLevel)
	log.Info("certsman starting...")

	memoryCache = gcache.New(InMemoryCertStorageLimit).
		ARC().
		EvictedFunc(func(key, value interface{}) {
			//TODO: Send this back to a channel to reprocess the servers cert
			log.Debug("Expired from cache: ", fmt.Sprint(key))
		}).
		Expiration(time.Minute * DefaultCertDurationMinutes).
		Build()

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

}

func certTestGetHandler(w http.ResponseWriter, r *http.Request) {
	// only use 2 chars so that some of the lookups are cached
	mux.Vars(r)["hostname"] = randomString(4)
	certificateGetHandler(w, r)
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func certificateGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	hostname := vars["hostname"]

	if len(hostname) <= 1 {
		http.Error(w, "Invalid hostname provided", http.StatusNotAcceptable)
		return
	}

	reqID := requestIDGenerator()

	req := certsman.CertificateRequest{
		RequestID: reqID,
		Hostname:  strings.ToLower(hostname),
	}

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

func requestIDGenerator() string {
	u := uuid.NewV4()

	return u.String()
}

func selfCertIssuer() {
	// TODO: This is kind of hacky, maybe would refactor later to not use requests perhaps?
	reqID := requestIDGenerator()

	req := certsman.CertificateRequest{
		RequestID: reqID,
		Hostname:  DefaultCertServerName,
	}

	log.Debug("Generating/updating self-cert for server ", DefaultCertServerName)

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
		log.Debug("Cert issuer timer is now refreshing the server certificate")
		selfCertIssuer()
	}
}
