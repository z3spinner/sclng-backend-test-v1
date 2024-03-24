package standard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/entities"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/Scalingo/sclng-backend-test-v1/worker/usecases"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Standard is the standard implementation of the worker usecases
type Standard struct {
	ctx        context.Context
	log        logrus.FieldLogger
	cfg        *config.Config
	db         db.Service
	fetch      fetcher.Service
	ratelimits entities.RateLimits
}

// RunWorker runs the worker in a goroutine and manages the lifecycle
func (s *Standard) RunWorker(ctx context.Context, wg *sync.WaitGroup) error {

	if ctx == nil {
		return fmt.Errorf("parent context is required")
	}

	if wg == nil {
		return fmt.Errorf("wait group is required")
	}

	wg.Add(1)

	var routineErr error

	// stop channel is used for the worker to signal that it has stopped
	stop := make(chan bool)
	defer close(stop)

	s.log.Info("RunWorker started")

	go func() {
		for {
			routineErr = s.doWork(ctx)
			if routineErr != nil {
				if errors.Is(routineErr, fetcher.ErrRequestTimeout) {
					s.log.Warn("Request timeout, logging and continuing")
				} else {
					s.log.WithError(routineErr).Error("Worker failed")
					stop <- true
					return
				}
			}
			<-time.After(5 * time.Second)
		}
	}()

	// Wait for the context to be cancelled or the stop signal
	select {
	case <-ctx.Done():
		break
	case <-stop:
		break
	}

	s.log.Info("RunWorker stopped")

	wg.Done()

	return routineErr
}

// A callback function that is called when the rate limit headers are received after a Github api fetch
func (s *Standard) onFetcherRateLimitHeaders(remaining int, reset time.Time) {
	s.ratelimits.SetRateLimits(remaining, reset)
}

func New(
	ctx context.Context, log logrus.FieldLogger, cfg *config.Config, db db.Service, fetch fetcher.Service,
) *Standard {
	uc := Standard{
		ctx:        ctx,
		log:        log,
		cfg:        cfg,
		db:         db,
		fetch:      fetch,
		ratelimits: entities.NewRateLimits(),
	}
	fetch.SetRateLimitHeadersCallback(uc.onFetcherRateLimitHeaders)
	return &uc
}

var _ usecases.Usecases = (*Standard)(nil)
