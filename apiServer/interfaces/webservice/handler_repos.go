package webservice

import (
	"encoding/json"
	"net/http"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/usecases"
)

// repoListHandler returns a http.Handler that handles the request to get the list of repositories
// it accepts the following query parameters:
// - name: string
// - language: string
// - license: string
// - allow_forking: string
// - has_open_issues: string
// it returns a JSON object containing the list of repositories
func (ws Webservice) reposHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// Check to see if the request is a GET request
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			// Get the filters from the query parameters
			filters := usecases.NewGetRepoListFilteredFilters(
				r.URL.Query().Get("name"),
				r.URL.Query().Get("language"),
				r.URL.Query().Get("license"),
				r.URL.Query().Get("allow_forking"),
				r.URL.Query().Get("has_open_issues"),
			)

			// Check the cache is valid
			ws.checkCacheValidity()

			// Check the local-memory cache first
			cacheKey := filters.CacheKey()
			iList, ok := ws.reposCache[cacheKey]

			// Cache miss
			if !ok {
				// Get the repository list from the usecases
				list, err := ws.uc.GetRepoListFiltered(
					r.Context(), filters,
				)
				if err != nil {
					logger.Get(r.Context()).WithError(err).Error("Fail to get latest 100 repositories")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Convert the list from the types used in the entities layer to those in the interfaces layer
				iList = convertRepoListE2I(list)

				// Store the data in the local-memory cache
				ws.reposMU.Lock()
				ws.reposCache[cacheKey] = iList
				ws.reposMU.Unlock()
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(
				RepoList{
					Items: iList,
				},
			)
			if err != nil {
				ws.log.WithError(err).Error("Fail to encode JSON")
			}
		},
	)
}
