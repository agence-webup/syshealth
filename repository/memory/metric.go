package memory

import (
	"errors"
	"webup/syshealth"
)

// GetMetricRepository returns a new in-memory metric repository
func GetMetricRepository() syshealth.MetricRepository {
	repo := metricRepository{
		metricsByServerID: map[string]syshealth.Data{},
	}
	return &repo
}

type metricRepository struct {
	metricsByServerID map[string]syshealth.Data
}

func (repo *metricRepository) Get(serverID string) (*syshealth.Data, error) {
	if data, ok := repo.metricsByServerID[serverID]; ok {
		return &data, nil
	}
	return nil, errors.New("unable to find data for server id")
}

func (repo *metricRepository) Store(serverID string, data syshealth.Data) error {
	repo.metricsByServerID[serverID] = data

	// log.Printf("metric received for server %v: %v\n", serverID, data)

	return nil
}
