package api

import "github.com/jacobbrewer1/puppet-reporter/pkg/models"

func (r *repository) SaveResources(resources []*models.Resource) error {
	return models.InsertManyResources(r.db, resources...)
}
