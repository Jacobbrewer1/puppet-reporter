package api

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

func (r *repository) SaveLogs(logs []*models.LogMessage) error {
	//return models.InsertManyLogMessages(r.db, logs...)
	return errors.New("not implemented")
}

func (r *repository) GetLogsByReportID(reportID int) ([]*models.LogMessage, error) {
	sqlStr := `SELECT id FROM log_message WHERE report_id = ?`

	ids := make([]int, 0)
	if err := r.db.Select(&ids, sqlStr, reportID); err != nil {
		return nil, fmt.Errorf("get log ids by report id: %w", err)
	}

	logs := make([]*models.LogMessage, 0)
	for _, id := range ids {
		log, err := models.LogMessageById(r.db, id)
		if err != nil {
			return nil, fmt.Errorf("get log by id: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
