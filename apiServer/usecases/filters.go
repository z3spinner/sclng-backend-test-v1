package usecases

import "strconv"

// GetRepoListFilters is a struct to hold the parameters for the GetRepoList usecase
// use pointer values to allow null values
type GetRepoListFilters struct {
	Name          *string
	Language      *string
	License       *string
	AllowForking  *bool
	HasOpenIssues *bool
}

// NewGetRepoListFilteredFilters creates a new GetRepoListFilters struct with the provided values
func NewGetRepoListFilteredFilters(name, language, license, allowForkingCount, hasOpenIssues string) GetRepoListFilters {

	return GetRepoListFilters{
		Name:          toStr(name),
		Language:      toStr(language),
		License:       toStr(license),
		AllowForking:  toBool(allowForkingCount),
		HasOpenIssues: toBool(hasOpenIssues),

		// Note we could extend this with integer values.
		// To be useful integer values would need range filters.
	}
}

// toStr converts a string to a string pointer
func toStr(in string) *string {
	if in == "" {
		return nil
	}
	return &in
}

// toBool converts a string to a bool pointer
func toBool(in string) *bool {
	if in == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(in)
	if err != nil {
		return nil
	}
	return &parsed
}
