package fetcherLive

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// New creates a new live fetcher service
func New(log logrus.FieldLogger, cfg *config.Config) (*FetcherLive, error) {

	log = log.WithFields(
		logrus.Fields{
			"service": serviceName,
		},
	)

	return &FetcherLive{
		log: log,
		cfg: cfg,
	}, nil
}

// updateRateLimiter updates the local tracking of the server rate limits. This allows to avoid clobbering the
// external API
func (f *FetcherLive) updateRateLimiter(headers http.Header) (err error) {
	// We get the X-RateLimit headers from the response and pass them to the RateLimit Callback
	if f.rateLimitCallback != nil {

		remainingStr := headers.Get("X-RateLimit-Remaining")
		resetStr := headers.Get("X-RateLimit-Reset")
		if remainingStr == "" || resetStr == "" {
			return nil
		}

		remainingInt64, err := strconv.ParseInt(remainingStr, 10, 64)
		if err != nil {
			return errors.Wrap(err, "error reading X-RateLimit-Remaining header")
		}

		resetSec, err := strconv.ParseInt(resetStr, 10, 64)
		if err != nil {
			return errors.Wrap(err, "error reading X-RateLimit-Reset")
		}
		resetTime := time.Unix(resetSec, 0)

		f.rateLimitCallback(int(remainingInt64), resetTime)
	}

	return nil
}
