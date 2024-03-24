package entities

import (
	"time"
)

// RepoList is a struct list of GitHub repositories.
type RepoList []RepoItem

func (v RepoList) Len() int { return len(v) }
func (v RepoList) Less(i, j int) bool {
	// Descending order
	return v[j].CreatedAt.Before(v[i].CreatedAt)
}

// RepoItem is a representation of a GitHub repository.
type RepoItem struct {
	ID              int64
	Name            string
	FullName        string
	Owner           string
	HTMLUrl         string
	Description     string
	LanguagesURL    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Size            int
	Language        string
	Languages       Languages
	LicenseName     string
	ForksCount      int
	OpenIssuesCount int
	WatchersCount   int
	AllowForking    bool
	HasIssues       bool
	HasProjects     bool
	HasDownloads    bool
	HasWiki         bool
	HasPages        bool
	HasDiscussions  bool
}

// Languages is a map of languages used in a repository.
type Languages map[string]int64

func (v Languages) Strings() []string {
	var langs []string
	for k := range v {
		langs = append(langs, k)
	}
	return langs
}

type Stats struct {
	AvgNumForksPerRepoByLanguage map[string]float32
	AvgNumOpenIssuesByLanguage   map[string]float32
	AvgSizeByLanguage            map[string]float32
	NumReposByLanguage           map[string]int
}
