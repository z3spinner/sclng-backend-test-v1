package webservice

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/apiServer/config"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/usecases"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

const serviceName = "Webservice"
const gracefulShutdownTimeout = 5 * time.Second

type Webservice struct {
	log        logrus.FieldLogger
	serverPort int
	uc         usecases.Usecases
}

// New creates a new webservice
func New(log logrus.FieldLogger, config *config.Config, uc usecases.Usecases) (*Webservice, error) {

	if log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.APIServerPort == 0 {
		return nil, fmt.Errorf("port is required")
	}

	return &Webservice{
		log: log.WithFields(
			logrus.Fields{
				"service": serviceName,
				"port":    config.APIServerPort,
			},
		),
		serverPort: config.APIServerPort,
		uc:         uc,
	}, nil
}

// Start starts the webservice in a goroutine
// Returns an error if the service fails to start
// Graceful shutdown when an interrupt signal is received from the OS
func (ws Webservice) Start(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) error {

	if ctx == nil {
		return fmt.Errorf("parent context is required")
	}

	if wg == nil {
		return fmt.Errorf("wait group is required")
	}

	// Increment the wait group counter
	wg.Add(1)

	// Declare an error variable to store the error from the go routine
	var routineErr error

	// Start the service in a goroutine
	go func() {
		defer wg.Done()

		// New multiplexer and register handler functions
		mux := http.NewServeMux()
		mux.Handle("/ping", ws.pongHandler())
		mux.Handle("/repos", ws.reposHandler())
		mux.Handle("/stats", ws.statsHandler())

		// Use negroni to create a middleware stack (because included in go.mod of this exercise)
		n := negroni.Classic()
		n.UseHandler(mux)

		// Create the http server (this approach enables graceful shutdown).
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", ws.serverPort),
			Handler: n,
		}

		//stop := make(chan error, 1)
		//defer close(stop)

		// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
		go func() {
			ws.log.Printf("apiServer listening on %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				ws.log.Errorf("error starting server: %v", err)
				//stop <- err
				stop()
				return
			}
		}()

		// Listen for the context to be cancelled. This blocks until the signal is received.
		select {
		case <-ctx.Done():
			ws.log.Info("shutting down server")
			//case routineErr = <-stop:
			//	ws.log.Infof("stopping with error %v", routineErr)
		}

		// Ask the http server to shut down, stop accepting new requests and wait for existing requests to finish.
		// The context is used to set a deadline for the shutdown process.
		ws.log.Info("graceful shutdown started")
		ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctxGrace); err != nil {
			ws.log.Errorf("%s: forced to shutdown: %v", serviceName, err)
			return
		}
		ws.log.Info("graceful shutdown complete")

	}()

	return routineErr
}
