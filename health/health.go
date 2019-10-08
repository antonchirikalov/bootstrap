package health

import "fmt"

type HealthCheckFunc func() (bool, error)

type HealthCheck struct {
	Name  string
	Check HealthCheckFunc
}

type HealthCheckCollection struct {
	healthChecks []*HealthCheck
}

func NewHealthCheckCollection() *HealthCheckCollection {
	return &HealthCheckCollection{
		healthChecks: []*HealthCheck{},
	}
}

func (hcc *HealthCheckCollection) AddHealthCheck(name string, check HealthCheckFunc) {
	hc := HealthCheck{
		Name:  name,
		Check: check,
	}
	hcc.healthChecks = append(hcc.healthChecks, &hc)
}

func (hcc *HealthCheckCollection) IsHealthy() (bool, error) {
	for _, hc := range hcc.healthChecks {
		ok, err := hc.Check()
		if !ok {
			return false, fmt.Errorf("%v: %v", hc.Name, err)
		}
	}

	return true, nil
}
