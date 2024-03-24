package fetcherMock

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const serviceName = "FetcherMock"

// FetcherMock is a mock service for getting repository information from Github
type FetcherMock struct {
	log logrus.FieldLogger
	cfg *config.Config

	// dataDir is the location of the mock datafiles
	dataDir string

	// rateLimitCallback is called when rate limit headers are received from the API
	//  This allows external tracking of the rate limits imposed by the remote API (which in this case is mock)
	rateLimitCallback func(remaining int, reset time.Time)

	// mockRateLimiter is a fake limiter which simulates the rate limit headers received from the remote API.
	mockRateLimiter *mockRateLimiter

	// fileCache is a local cache for the mock data, so that we only load it once from disk
	fileCache      map[string][]byte
	fileCacheMutex sync.Mutex
}

// SetRateLimitHeadersCallback sets the rate limit callback function
func (f *FetcherMock) SetRateLimitHeadersCallback(callback func(remaining int, reset time.Time)) {
	f.rateLimitCallback = callback
}

// GetLatest100 gets mock latest 100 data
func (f *FetcherMock) GetRepoList(ctx context.Context) (
	entities.RepoList, error,
) {

	var iRepoList fetcher.RepoList

	// Load list.json from data folder
	listData, err := f.mockAPIRequest(ctx, f.dataDir+"list.json")
	if err != nil {
		return nil, errors.Wrapf(err, "error GetRepoList")
	}

	err = json.Unmarshal(listData, &iRepoList)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing json")
	}

	// Convert from the interface type to entities type (I2E)
	eRepoList := fetcher.ConvertLatest100I2E(iRepoList)

	return eRepoList, nil
}

// GetRepoLanguages gets mock repoLanguages data
func (f *FetcherMock) GetRepoLanguages(ctx context.Context, url string) (entities.Languages, error) {
	var iLangs fetcher.Languages

	// convert the url to a filename (replace "/" with "_")
	filename := strings.ReplaceAll(url, "/", "_") + ".json"

	// load the data from the file
	langData, err := f.mockAPIRequest(ctx, f.dataDir+filename)
	if err != nil {
		return nil, errors.Wrapf(err, "error GetRepoLanguages")
	}

	err = json.Unmarshal(langData, &iLangs)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing json")
	}

	// Convert from the interface type to entities type (I2E)
	eLangs := fetcher.ConvertLanguagesI2E(iLangs)

	return eLangs, nil

}

// mockAPIRequest loads a file from disk or cache and updates the simulated rate limiter
func (f *FetcherMock) mockAPIRequest(ctx context.Context, filename string) ([]byte, error) {

	// Check the mock rate limiter (simulated rate limits)
	// update the rate limit counts
	// See if this query should be rate limited
	err := f.mockRateLimiter.checkMockRateLimiting()
	if err != nil && !errors.Is(err, fetcher.ErrRateLimited) {
		return nil, errors.Wrap(err, "error checking mock rate limiter")
	}

	// Call the rate limit callback to update the local rate limit counts
	if f.rateLimitCallback != nil {
		f.rateLimitCallback(f.mockRateLimiter.get())
	}

	// if the request is rate limited, return with a short delay
	if errors.Is(err, fetcher.ErrRateLimited) {
		<-time.After(100 * time.Millisecond)
		return nil, fetcher.ErrRateLimited
	}

	// Get the file data from local cache or load it from disk
	var data []byte
	var ok bool
	f.fileCacheMutex.Lock()
	if data, ok = f.fileCache[filename]; !ok {
		data, err = os.ReadFile(filename)
		if err != nil {
			f.fileCacheMutex.Unlock()
			return nil, errors.Wrapf(err, "error reading file %s", filename)
		}
		f.fileCache[filename] = data
	}
	f.fileCacheMutex.Unlock()

	return f.simulatedAPICallWithTimeout(ctx, data)

}

func (f *FetcherMock) simulatedAPICall(ctx context.Context, data []byte) ([]byte, error) {
	sleepFor := time.Duration(rand.Float32()*f.cfg.MockFetcherAvgRequestSeconds*2000) * time.Millisecond
	select {
	case <-time.After(time.Until(time.Now().Add(sleepFor))):
		return data, nil
	case <-ctx.Done():
		return nil, fetcher.ErrRequestTimeout
	}
}

func (f *FetcherMock) simulatedAPICallWithTimeout(ctx context.Context, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(f.cfg.FetchTimeoutSeconds*1000)*time.Millisecond)
	defer cancel()
	return f.simulatedAPICall(ctx, data)
}

var _ fetcher.Service = (*FetcherMock)(nil)
