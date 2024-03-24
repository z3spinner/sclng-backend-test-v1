package db

import (
	"context"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
)

var ErrNotFound = DBError("not found")

type Service interface {
	SetRepoList(ctx context.Context, list entities.RepoList) error
	SetRepoItemLanguages(ctx context.Context, repoID int64, langs entities.Languages) error

	GetRepoList(ctx context.Context, filters GetRepoListFilters) (entities.RepoList, error)
	GetRepoItem(ctx context.Context, repoID int64) (entities.RepoItem, error)

	GetAvgNumForksPerRepoByLanguage(ctx context.Context) (map[string]float32, error)
	GetAvgNumOpenIssuesByLanguage(ctx context.Context) (map[string]float32, error)
	GetAvgSizeByLanguage(ctx context.Context) (map[string]float32, error)
	GetNumReposByLanguage(ctx context.Context) (map[string]int, error)
}
