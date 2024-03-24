package usecases

import (
	"context"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
)

type Usecases interface {
	GetRepoListFiltered(ctx context.Context, filters GetRepoListFilters) (entities.RepoList, error)
	GetStats(ctx context.Context) (entities.Stats, error)
}
