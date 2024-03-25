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
	cfg        *config.Config
	serverPort int
	uc         usecases.Usecases

	// We Include a local memory cache here to reduce the traffic and load on the database.
	// Response times drop significantly when the data is cached.
	//  - No cache: ~ < 10ms (approx)
	//  - With cache:  < 300µs (approx)
	statsCache *Stats
	reposMU    *sync.Mutex
	reposCache map[string][]RepoItem

	// We have a naive cache invalidation strategy here. If the timestamp is older than a certain age, we invalidate the cache.
	cacheTimeStamp time.Time
}

// New creates a new webservice
func New(log logrus.FieldLogger, cfg *config.Config, uc usecases.Usecases) (*Webservice, error) {

	if log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.APIServerPort == 0 {
		return nil, fmt.Errorf("port is required")
	}

	return &Webservice{
		log: log.WithFields(
			logrus.Fields{
				"service": serviceName,
				"port":    cfg.APIServerPort,
			},
		),
		cfg:        cfg,
		serverPort: cfg.APIServerPort,
		uc:         uc,
		reposMU:    &sync.Mutex{},
		reposCache: make(map[string][]RepoItem),
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

func (ws *Webservice) checkCacheValidity() {

	// The cache is old. Invalidate it. (This covers the case where the cacheTimeStamp isZero also)
	if time.Since(ws.cacheTimeStamp).Seconds() > float64(ws.cfg.RequestMemCacheMaxAgeSeconds) {
		ws.reposMU.Lock()
		ws.reposCache = make(map[string][]RepoItem)
		ws.statsCache = nil
		ws.reposMU.Unlock()
		ws.cacheTimeStamp = time.Now()
	}
}
