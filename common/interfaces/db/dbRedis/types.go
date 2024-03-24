package dbRedis

import (
	"fmt"
	"time"
)

type RedisError string

func (e RedisError) Error() string { return string(e) }

// RepoList The types in this file are used to represent the data returned by the GitHub API.
type RepoList []RepoItem

type repoKey string

type RepoItem struct {
	ID              int64     `redis:"id" json:"id"`
	Name            string    `redis:"name" json:"name"`
	FullName        string    `redis:"full_name" json:"full_name"`
	Owner           string    `redis:"owner" json:"owner"`
	HTMLUrl         string    `redis:"html_url" json:"html_url"`
	Description     string    `redis:"description" json:"description"`
	LanguagesURL    string    `redis:"languages_url" json:"languages_url"`
	CreatedAt       time.Time `redis:"created_at" json:"created_at"`
	UpdatedAt       time.Time `redis:"updated_at" json:"updated_at"`
	Size            int       `redis:"size" json:"size"`
	Language        string    `redis:"language" json:"language"`
	Languages       *string   `redis:"languages,omitempty" json:"languages,omitempty"`
	LicenseName     string    `redis:"license" json:"license"`
	ForksCount      int       `redis:"forks_count" json:"forks_count"`
	OpenIssuesCount int       `redis:"open_issues_count" json:"open_issues_count"`
	WatchersCount   int       `redis:"watchers_count" json:"watchers_count"`
	AllowForking    bool      `redis:"allow_forking" json:"allow_forking"`
	HasIssues       bool      `redis:"has_issues" json:"has_issues"`
	HasProjects     bool      `redis:"has_projects" json:"has_projects"`
	HasDownloads    bool      `redis:"has_downloads" json:"has_downloads"`
	HasWiki         bool      `redis:"has_wiki" json:"has_wiki"`
	HasPages        bool      `redis:"has_pages" json:"has_pages"`
	HasDiscussions  bool      `redis:"has_discussions" json:"has_discussions"`
}

func getRepoKey(id int64) repoKey {
	return repoKey(fmt.Sprintf("repo:%d", id))
}

func (r RepoItem) getKey() repoKey {
	return getRepoKey(r.ID)
}
