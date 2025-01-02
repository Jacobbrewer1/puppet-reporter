package api

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

func (r *repository) SaveResources(resources []*models.Resource) error {
	//return models.InsertManyResources(r.db, resources...)
	return errors.New("test")
}

func (r *repository) GetResourcesByReportID(reportID int) ([]*models.Resource, error) {
	sqlStr := `SELECT id FROM resource WHERE report_id = ?`

	ids := make([]int, 0)
	if err := r.db.Select(&ids, sqlStr, reportID); err != nil {
		return nil, fmt.Errorf("get resource ids by report id: %w", err)
	}

	resources := make([]*models.Resource, 0)
	for _, id := range ids {
		resource, err := models.ResourceById(r.db, id)
		if err != nil {
			return nil, fmt.Errorf("get resource by id: %w", err)
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
