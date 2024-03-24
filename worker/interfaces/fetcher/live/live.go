package fetcherLive

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const serviceName = "FetcherLive"

// FetcherLive is a service for getting repository information from the Github API
type FetcherLive struct {
	log logrus.FieldLogger
	cfg *config.Config

	// rateLimitCallback is called when rate limit headers are received from the API
	//  This allows external tracking of the rate limits imposed by the remote API
	rateLimitCallback func(remaining int, reset time.Time)
}

// SetRateLimitHeadersCallback sets the callback function for rate limit headers
func (f *FetcherLive) SetRateLimitHeadersCallback(callback func(remaining int, reset time.Time)) {
	f.rateLimitCallback = callback
}

// GetLatest100 returns the latest 100 public repositories from Github
func (f *FetcherLive) GetRepoList(ctx context.Context) (entities.RepoList, error) {

	url := "https://api.github.com/search/repositories?q=type:public&per_page=100&sort=created&order=desc"
	resBody, err := f.fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	var iRepoLost fetcher.RepoList
	err = json.Unmarshal(resBody, &iRepoLost)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing json")
	}

	// Convert from the interface type to entities type (I2E)
	eRepoList := fetcher.ConvertLatest100I2E(iRepoLost)

	return eRepoList, nil

}

// GetRepoLanguages returns the languages used in a repository
func (f *FetcherLive) GetRepoLanguages(ctx context.Context, url string) (entities.Languages, error) {

	resBody, err := f.fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	var iLangs fetcher.Languages
	err = json.Unmarshal(resBody, &iLangs)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing json")
	}

	// Convert from the interface type to entities type (I2E)
	eLangs := fetcher.ConvertLanguagesI2E(iLangs)

	return eLangs, nil

}

var _ fetcher.Service = (*FetcherLive)(nil)
