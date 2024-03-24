package fetcher

import "github.com/Scalingo/sclng-backend-test-v1/common/entities"

// Conversions
//  Conversion E2I represents the conversion of types from entities to interfaces layers.
//  Conversion I2E represents the conversion of types from interfaces to entities layers.

func ConvertLatest100I2E(latest100 RepoList) entities.RepoList {
	out := make(entities.RepoList, len(latest100.Items))
	for i := 0; i < len(out); i++ {
		out[i] = ConvertLatest100ItemI2E(latest100.Items[i])
	}
	return out
}

func ConvertLatest100ItemI2E(i Repo) entities.RepoItem {
	return entities.RepoItem{
		ID:              i.ID,
		Name:            i.Name,
		FullName:        i.FullName,
		Owner:           i.Owner.Login,
		HTMLUrl:         i.HTMLUrl,
		Description:     i.Description,
		LanguagesURL:    i.LanguagesURL,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		Size:            i.Size,
		Language:        i.Language,
		Languages:       nil,
		LicenseName:     i.License.Name,
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
	}
}

func ConvertLanguagesI2E(i Languages) entities.Languages {
	out := entities.Languages{}
	for lang, count := range i {
		out[lang] = count
	}
	return out
}
