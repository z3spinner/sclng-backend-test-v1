package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db/dbRedis"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	fetcherLive "github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher/live"
	fetcherMock "github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher/mock"
	"github.com/Scalingo/sclng-backend-test-v1/worker/usecases/standard"
	"github.com/sirupsen/logrus"
)

// with go modules enabled (GO111MODULE=on or outside GOPATH)

func main() {

	log := logger.Default().WithField("main", "worker")
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
	if err = startup(ctxServices, &wg, log, cfg); err != nil {
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
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func(ctx context.Context, stop context.CancelFunc) {
		defer stop()
		<-ctx.Done()
	}(ctx, stop)

	// ***********************************************************
	// 1. Create the db service - using redis implementation
	// ***********************************************************
	dbService, err := dbRedis.NewDBServiceRedis(
		log, config.RedisHostPort, config.RedisPrefix,
	)
	if err != nil {
		return fmt.Errorf("error creating new redis service: %w", err)
	}
	if err = dbService.Start(ctx, wg); err != nil {
		return fmt.Errorf("error starting redis service: %w", err)
	}

	// ***********************************************************
	// 2. Create the fetcher service (Configured in ENV)
	//  - live: fetches data from the live Github API
	//  - mock: fetches data from a local mock data files
	// ***********************************************************
	var fetcherService fetcher.Service
	switch strings.ToLower(config.UseFetcher) {
	case "live":
		// Create the Github fetcher service
		log.WithField("fetcher", "live").Info("configuring fetcher service")
		fetcherService, err = fetcherLive.New(log, config)
		if err != nil {
			return fmt.Errorf("error creating new github fetcher service: %w", err)
		}
	case "mock":
		// Create a mock fetcher service
		log.WithField("fetcher", "mock").Info("configuring fetcher service")
		fetcherService, err = fetcherMock.New(log, "worker/interfaces/fetcher/mock/data", config)
		if err != nil {
			return fmt.Errorf("error creating new mock fetcher service: %w", err)
		}
	default:
		return fmt.Errorf("unknown fetcher service: %s", config.UseFetcher)
	}

	// ***********************************************************
	// 3. Create the usecases layer.
	//  - Inject the db and fetcher services
	// ***********************************************************
	usecases := standard.New(ctx, log, config, dbService, fetcherService)
	err = usecases.RunWorker(ctx, wg)
	if err != nil {
		return fmt.Errorf("error running worker: %w", err)
	}

	return nil
}
