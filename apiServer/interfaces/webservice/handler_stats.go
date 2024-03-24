package webservice

import (
	"encoding/json"
	"net/http"

	"github.com/Scalingo/go-utils/logger"
)

// statsHandler returns a handler that responds with a JSON object containing the stats of the repositories
// - it accepts no query parameters
// it returns a JSON object containing the stats of the repositories
func (ws Webservice) statsHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// Check to see if the request is a GET request
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			stats, err := ws.uc.GetStats(
				r.Context(),
			)
			if err != nil {
				logger.Get(r.Context()).WithError(err).Error("Fail to get latest 100 repositories")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err = json.NewEncoder(w).Encode(
				Stats{
					AvgNumForksPerRepoByLanguage: stats.AvgNumForksPerRepoByLanguage,
					AvgNumOpenIssuesByLanguage:   stats.AvgNumOpenIssuesByLanguage,
					AvgSizeByLanguage:            stats.AvgSizeByLanguage,
					NumReposByLanguage:           stats.NumReposByLanguage,
				},
			)
			if err != nil {
				ws.log.WithError(err).Error("Fail to encode JSON")
			}
		},
	)
}
