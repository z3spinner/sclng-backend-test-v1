package webservice

import (
	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
)

// Conversions
//  Conversion E2I represents the conversion of types from entities to interfaces layers.
//  Conversion I2E represents the conversion of types from interfaces to entities layers.

// convertRepoListE2I converts a RepoList from entities to RepoList from interfaces
func convertRepoListE2I(in entities.RepoList) []RepoItem {
	out := make([]RepoItem, len(in))
	for i, v := range in {
		out[i] = convertRepoItemE2I(v)
	}
	return out
}

func convertRepoItemE2I(in entities.RepoItem) RepoItem {
	return RepoItem{
		ID:              in.ID,
		Name:            in.Name,
		FullName:        in.FullName,
		Owner:           in.Owner,
		HTMLUrl:         in.HTMLUrl,
		Description:     in.Description,
		LanguagesURL:    in.LanguagesURL,
		CreatedAt:       in.CreatedAt,
		UpdatedAt:       in.UpdatedAt,
		Size:            in.Size,
		Language:        in.Language,
		Languages:       convertLanguagesE2I(in.Languages),
		LicenseName:     in.LicenseName,
		ForksCount:      in.ForksCount,
		OpenIssuesCount: in.OpenIssuesCount,
		WatchersCount:   in.WatchersCount,
		AllowForking:    in.AllowForking,
		HasIssues:       in.HasIssues,
		HasProjects:     in.HasProjects,
		HasDownloads:    in.HasDownloads,
		HasWiki:         in.HasWiki,
		HasPages:        in.HasPages,
		HasDiscussions:  in.HasDiscussions,
	}
}

func convertLanguagesE2I(in entities.Languages) Languages {
	out := make(Languages)
	for k, v := range in {
		out[k] = Language{Bytes: v}
	}
	return out
}
