package standard

import (
	"context"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/pkg/errors"
)

// wait blocks until the given time or the context is cancelled
func (s *Standard) wait(ctx context.Context, until time.Time) {
	select {
	case <-time.After(time.Until(until)):
		// Blocked
	case <-ctx.Done():
		// Context cancelled
	}
}

func (s *Standard) getRateLimitSleepUntilTime() time.Time {
	// The time we should sleep after the resetTime has passed
	sleepOverTime := time.Duration(s.cfg.SleepoverDurationSeconds) * time.Second
	return s.ratelimits.GetResetTime().Add(sleepOverTime)
}

// waitForRateLimiter waits until the rate limiter is reset
func (s *Standard) waitForRateLimiter(ctx context.Context) {

	sleepUntil := s.getRateLimitSleepUntilTime()

	s.log.WithField("workerID", ctx.Value("workerID")).
		Warnf("rate limited, sleep to %s", sleepUntil.Format("2006-01-02T15:04:05"))

	s.wait(ctx, sleepUntil)
}

// retryOrWait retries the job until it succeeds or waits for the rate limiter
func (s *Standard) retryOrWait(ctx context.Context, job func(ctx context.Context) error) error {

	// First check our local copy of the rate limits
	if s.ratelimits.GetRemainingCount() == 0 {
		s.wait(ctx, s.getRateLimitSleepUntilTime())
	}

	retries := 3
	for {
		// Now lets try the job
		err := job(ctx)
		if err != nil {
			if errors.Is(err, fetcher.ErrRateLimited) {
				s.waitForRateLimiter(ctx)
			} else if errors.Is(err, fetcher.ErrRequestTimeout) && retries > 0 {
				// Try again after a short wait
				s.wait(ctx, time.Now().Add(2*time.Second))
				retries--
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}
