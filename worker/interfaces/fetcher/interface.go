package fetcher

import (
	"context"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
)

var ErrRateLimited = FetcherError("rate limited")
var ErrRequestTimeout = FetcherError("request timeout")

type Service interface {
	SetRateLimitHeadersCallback(callback func(remaining int, reset time.Time))
	GetRepoList(ctx context.Context) (entities.RepoList, error)
	GetRepoLanguages(ctx context.Context, url string) (entities.Languages, error)
}
