package api

import "github.com/jacobbrewer1/puppet-reporter/pkg/models"

func (r *repository) SaveLogs(logs []*models.LogMessage) error {
	return models.InsertManyLogMessages(r.db, logs...)
}
