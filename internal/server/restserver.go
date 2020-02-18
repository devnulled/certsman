package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bluele/gcache"
	"github.com/devnulled/certsman/pkg/certgen"
	"github.com/devnulled/certsman/pkg/memstorage"
	"github.com/devnulled/certsman/pkg/tokencert"
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
const InMemoryCertStorageLimit = 2000

// How long the server waits for connections to drain before doing a graceful shutdown
const DefaultServerGracefulTimeoutSeconds = 15

const DefaultServerAddress = "0.0.0.0:8080"

// How long to wait before shutting the server down to let connections drain
var wait = time.Second * DefaultServerGracefulTimeoutSeconds

//TODO: Refactor the below to not use values
// The shared in-memory persistance cache
var inMemCache = gcache.New(InMemoryCertStorageLimit).
	ARC().
	Expiration(time.Minute * DefaultCertDurationMinutes).
	Build()

var inMemPersist = memstorage.InMemStorage{Cache: inMemCache}

// The service which generates the certificates
var tokenCertIssuer = tokencert.TokenCertIssuer{KeyLength: DefaultTokenCertKeyLength}

// The cert service that that is compromised of the previous two impls
var tokenCertService = certgen.CerfificateService{Issuer: tokenCertIssuer, Persistence: inMemPersist}

// RunServer starts and runs the server
func RunServer() {
	log.Info("Starting up certsman server at ", DefaultServerAddress)
	// Build our cache collection for the in-memory persistence
	/*
		gc := gcache.New(InMemoryCertStorageLimit).
			ARC().
			Expiration(time.Minute * DefaultCertDurationMinutes).
			Build()
	*/

	// Start our HTTP router/handler
	r := mux.NewRouter()
	r.HandleFunc("/cert/{hostname}", certificateGetHandler).Methods("GET")

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

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("Shutting down certsman server gracefully")
	os.Exit(0)

	/*
		certRequest := certgen.CertificateRequest{
			Hostname: "yellow.com",
		}

		tokenCertGenerator := tokencert.TokenCertIssuer{KeyLength: DefaultTokenCertKeyLength}

		thisCert, certErr := tokenCertGenerator.IssueCertificate(certRequest)

		if certErr == nil {
			log.Info(thisCert)
			gc.Set(certRequest.Hostname, thisCert)
		} else {
			log.Error(certErr)
		}
	*/
}

func certificateGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	hostname := vars["hostname"]
	reqID := requestIDGenerator()

	req := certgen.CertificateRequest{
		RequestID: reqID,
		Hostname:  hostname,
	}

	resp := tokenCertService.GetOrCreateCertificate(req)

	if !resp.IsSuccess {
		http.Error(w, resp.Error.Error(), resp.StatusCode)
		return
	}

	log.Info(req)

}

func requestIDGenerator() string {
	u := uuid.NewV4()

	return u.String()
}
