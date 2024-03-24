package standard

import (
	"context"

	"github.com/Scalingo/sclng-backend-test-v1/apiServer/config"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/usecases"
	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/sirupsen/logrus"
)

// Standard is the standard implementation of the worker usecases
type Standard struct {
	ctx context.Context
	log logrus.FieldLogger
	cfg *config.Config
	db  db.Service

	// NOTE: We could include a local memory cache here to reduce the traffic and load on the database.

}

func (s Standard) GetRepoListFiltered(ctx context.Context, filters usecases.GetRepoListFilters) (
	entities.RepoList, error,
) {

	list, err := s.db.GetRepoList(
		ctx, db.GetRepoListFilters{
			Name:          filters.Name,
			Language:      filters.Language,
			License:       filters.License,
			AllowForking:  filters.AllowForking,
			HasOpenIssues: filters.HasOpenIssues,
		},
	)
	if err != nil {
		return entities.RepoList{}, err
	}

	return list, nil
}

func (s Standard) GetStats(ctx context.Context) (entities.Stats, error) {
	var err error
	out := entities.Stats{}

	if out.AvgNumForksPerRepoByLanguage, err = s.db.GetAvgNumForksPerRepoByLanguage(ctx); err != nil {
		return entities.Stats{}, err
	}

	if out.NumReposByLanguage, err = s.db.GetNumReposByLanguage(ctx); err != nil {
		return entities.Stats{}, err
	}

	if out.AvgNumOpenIssuesByLanguage, err = s.db.GetAvgNumOpenIssuesByLanguage(ctx); err != nil {
		return entities.Stats{}, err
	}

	if out.AvgSizeByLanguage, err = s.db.GetAvgSizeByLanguage(ctx); err != nil {
		return entities.Stats{}, err
	}

	return out, nil
}

func New(
	ctx context.Context, log logrus.FieldLogger, cfg *config.Config, db db.Service,
) *Standard {
	uc := Standard{
		ctx: ctx,
		log: log,
		cfg: cfg,
		db:  db,
	}
	return &uc
}

var _ usecases.Usecases = (*Standard)(nil)
