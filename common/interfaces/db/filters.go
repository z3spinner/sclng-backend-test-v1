package db

// GetRepoListFilters is a struct to hold the parameters for the GetRepoList db method
// use pointer values to allow null values
type GetRepoListFilters struct {
	Name          *string
	Language      *string
	License       *string
	AllowForking  *bool
	HasOpenIssues *bool
}
