package dbRedis

import (
	"encoding/json"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/pkg/errors"
)

// Conversions
//  Conversion E2I represents the conversion of types from entities to interfaces layers.
//  Conversion I2E represents the conversion of types from interfaces to entities layers.

func ConvertRepoItemE2I(e entities.RepoItem) (RepoItem, error) {

	iLangs, err := ConvertLanguagesE2I(e.Languages)
	if err != nil {
		return RepoItem{}, err
	}

	return RepoItem{
		ID:              e.ID,
		Name:            e.Name,
		FullName:        e.FullName,
		Owner:           e.Owner,
		HTMLUrl:         e.HTMLUrl,
		Description:     e.Description,
		LanguagesURL:    e.LanguagesURL,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		Size:            e.Size,
		Language:        e.Language,
		Languages:       &iLangs,
		LicenseName:     e.LicenseName,
		ForksCount:      e.ForksCount,
		OpenIssuesCount: e.OpenIssuesCount,
		WatchersCount:   e.WatchersCount,
		AllowForking:    e.AllowForking,
		HasIssues:       e.HasIssues,
		HasProjects:     e.HasProjects,
		HasDownloads:    e.HasDownloads,
		HasWiki:         e.HasWiki,
		HasPages:        e.HasPages,
		HasDiscussions:  e.HasDiscussions,
	}, nil
}

func ConvertLanguagesE2I(e entities.Languages) (string, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return "nil", errors.Wrap(err, "Error marshaling JSON")
	}
	str := string(jsonData)
	return str, nil
}

func ConvertRepoListI2E(latest100 RepoList) entities.RepoList {
	out := make(entities.RepoList, len(latest100))
	for i := 0; i < len(out); i++ {
		out[i], _ = ConvertRepoItemI2E(latest100[i])
	}
	return out
}

func ConvertRepoItemI2E(i RepoItem) (entities.RepoItem, error) {

	eLangs, err := ConvertLanguagesI2E(i.Languages)
	if err != nil {
		return entities.RepoItem{}, err
	}

	return entities.RepoItem{
		ID:              i.ID,
		Name:            i.Name,
		FullName:        i.FullName,
		Owner:           i.Owner,
		HTMLUrl:         i.HTMLUrl,
		Description:     i.Description,
		LanguagesURL:    i.LanguagesURL,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		Size:            i.Size,
		Language:        i.Language,
		Languages:       eLangs,
		LicenseName:     i.LicenseName,
		ForksCount:      i.ForksCount,
		OpenIssuesCount: i.OpenIssuesCount,
		WatchersCount:   i.WatchersCount,
		AllowForking:    i.AllowForking,
		HasIssues:       i.HasIssues,
		HasProjects:     i.HasProjects,
		HasDownloads:    i.HasDownloads,
		HasWiki:         i.HasWiki,
		HasPages:        i.HasPages,
		HasDiscussions:  i.HasDiscussions,
	}, nil
}

func ConvertLanguagesI2E(i *string) (entities.Languages, error) {
	out := entities.Languages{}
	if i == nil {
		return out, nil
	}

	err := json.Unmarshal([]byte(*i), &out)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling languages JSON")
	}
	return out, nil
}
