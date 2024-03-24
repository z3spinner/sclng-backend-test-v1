package fetcherLive

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/pkg/errors"
)

// Helper function to fetch a URL from Github
// Uses appropriate headers and returns the body as a byte slice
// Inspects the rate limit headers and calls the rateLimit callback
func (f *FetcherLive) fetch(ctx context.Context, url string) ([]byte, error) {

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(f.cfg.FetchTimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "client could not create request")
	}
	req.WithContext(ctx)
	req.Header = map[string][]string{
		"Accept": {"application/vnd.github.json"},
	}

	// Make the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "client error making request")
	}

	// Call the rateLimit callback to update the local rate limit counts
	err = f.updateRateLimiter(res.Header)
	if err != nil {
		return nil, err
	}

	// Check status code
	// Rate limited requests return a 403 or 429 status code
	// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#exceeding-the-rate-limit
	if res.StatusCode == 403 || res.StatusCode == 429 {
		return nil, fetcher.ErrRateLimited
	}

	if res.StatusCode != 200 {
		return nil, errors.Errorf("status %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "client could not read response body")
	}

	return resBody, nil
}
