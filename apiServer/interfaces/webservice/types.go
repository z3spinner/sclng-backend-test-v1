package webservice

import "time"

// The types in this file are used to represent the data returned by our API.

// RepoList represents a list of repositories
type RepoList struct {
	Items []RepoItem `json:"repositories"`
}

// RepoItem represents a repository
type RepoItem struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Owner           string    `json:"owner"`
	HTMLUrl         string    `json:"html_url"`
	Description     string    `json:"description"`
	LanguagesURL    string    `json:"languages_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Size            int       `json:"size"`
	Language        string    `json:"language"`
	Languages       Languages `json:"languages"`
	LicenseName     string    `json:"license"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	WatchersCount   int       `json:"watchers_count"`
	AllowForking    bool      `json:"allow_forking"`
	HasIssues       bool      `json:"has_issues"`
	HasProjects     bool      `json:"has_projects"`
	HasDownloads    bool      `json:"has_downloads"`
	HasWiki         bool      `json:"has_wiki"`
	HasPages        bool      `json:"has_pages"`
	HasDiscussions  bool      `json:"has_discussions"`
}

// Languages is a map of languages used in a repository.
type Languages map[string]Language

type Language struct {
	Bytes int64 `json:"bytes"`
}

// Stats represents the statistics returned by the API
type Stats struct {
	AvgNumForksPerRepoByLanguage map[string]float32 `json:"avg_num_forks_per_repo_by_language"`
	AvgNumOpenIssuesByLanguage   map[string]float32 `json:"avg_num_open_issues_by_language"`
	AvgSizeByLanguage            map[string]float32 `json:"avg_size_by_language"`
	NumReposByLanguage           map[string]int     `json:"num_repos_by_language"`
}
