package webservice

import (
	"encoding/json"
	"net/http"
)

// pongHandler returns a handler that responds with a JSON object containing the string "pong"
func (ws Webservice) pongHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// Check to see if the request is a GET request
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
			if err != nil {
				ws.log.WithError(err).Error("Fail to encode JSON")
			}
		},
	)
}
