package fetcher

import "time"

type FetcherError string

func (e FetcherError) Error() string { return string(e) }

// RepoList The types in this file are used to represent the data returned by the GitHub API.
type RepoList struct {
	TotalCount int    `json:"total_count"`
	Items      []Repo `json:"items"`
}

type Repo struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Owner           Owner     `json:"owner"`
	HTMLUrl         string    `json:"html_url"`
	Description     string    `json:"description"`
	LanguagesURL    string    `json:"languages_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Size            int       `json:"size"`
	Language        string    `json:"language"`
	License         License   `json:"license"`
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

type License struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Owner struct {
	Login string `json:"login"`
}

type Languages map[string]int64
