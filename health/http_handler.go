package health

import (
	"net/http"
)

// GetServiceHealth ranges over a set of health checks, validates that each is ok
// and responds accordingly
// This is a very minor variant of what we have in the go-svc-bootstrap, modified
// only to support the mux style definitions that we're using now.
func GetServiceHealth(healthChecks *HealthCheckCollection, serviceName string) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Server", serviceName)

		healthy, err := healthChecks.IsHealthy()
		if healthy {
			res.WriteHeader(http.StatusOK)
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}

	}
}
