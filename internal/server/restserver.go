package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
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

var inMemPersist = storage.InMemStorage{Cache: inMemCache}

// The cert service which generates string based certificates
var stringCertIssuer = certs.StringCertIssuer{StringPrefix: "foo-"}

// The cert service that that is compromised of the previous two impls
var stringCertService = certsman.CerfificateService{Issuer: stringCertIssuer, Persistence: inMemPersist}

// RunServer starts and runs the server
func RunServer() {
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	log.Info("Starting up certsman server at ", DefaultServerAddress)

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
}

func certificateGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	hostname := vars["hostname"]

	if len(hostname) <= 3 {
		http.Error(w, "Invalid hostname provided", http.StatusNotAcceptable)
	}

	reqID := requestIDGenerator()

	req := certsman.CertificateRequest{
		RequestID: reqID,
		Hostname:  hostname,
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

	log.Info("Request: ", req)
	log.Info("Response: ", resp)

}

func requestIDGenerator() string {
	u := uuid.NewV4()

	return u.String()
}
