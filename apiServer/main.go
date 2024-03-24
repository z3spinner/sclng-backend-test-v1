package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/config"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/interfaces/webservice"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/usecases/standard"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db/dbRedis"
	"github.com/sirupsen/logrus"
)

// with go modules enabled (GO111MODULE=on or outside GOPATH)

func main() {

	log := logger.Default().WithField("main", "apiServer")
	log.Info("Initializing app")
	cfg, err := config.New()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(1)
	}

	// Create a waitGroup to wait for all program services to gracefully exit
	var wg sync.WaitGroup

	log.Info("Startup")
	ctxServices, ctxServicesCancel := context.WithCancel(context.Background())
	if err := startup(ctxServices, &wg, log, cfg); err != nil {
		log.Errorf("exiting... failed to start services: %v", err)

		// Something went wrong when starting the services. Cancel the context and wait for the services to exit
		ctxServicesCancel()
		wg.Wait()

		os.Exit(1)
	}

	// All services have started.
	// Wait for them to stop before exiting (triggered by an interrupt signal from the OS).
	wg.Wait()

	// Server shutting down
	log.Info("Shutdown")

}

// startup starts the services required by the worker
// - db service
// - fetcher service
// It initialises the usecases layer and injects the services
func startup(ctx context.Context, wg *sync.WaitGroup, log logrus.FieldLogger, config *config.Config) error {
	var err error

	// Wait for an interrupt signal from the OS can cancel the context
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	go func(ctx context.Context, stop context.CancelFunc) {
		defer stop()
		<-ctx.Done()
	}(ctx, stop)

	// ***********************************************************
	// 1. Create the db service - using redis implementation
	// ***********************************************************
	db, err := dbRedis.NewDBServiceRedis(
		log, config.RedisHostPort, config.RedisPrefix,
	)
	if err != nil {
		return fmt.Errorf("error creating new redis service: %w", err)
	}
	if err = db.Start(ctx, wg); err != nil {
		return fmt.Errorf("error starting redis service: %w", err)
	}

	// ***********************************************************
	// 2. Create the usecases layer.
	//  - Inject the db service
	// ***********************************************************
	uc := standard.New(ctx, log, config, db)

	// ***********************************************************
	// 3. Create the webservice (interface layer).
	//  - Inject the usecases
	// ***********************************************************
	ws, err := webservice.New(log, config, uc)
	if err != nil {
		return fmt.Errorf("error creating new webservice: %w", err)
	}
	if err = ws.Start(ctx, stop, wg); err != nil {
		stop()
		return fmt.Errorf("error starting webservice: %w", err)
	}

	return nil
}
