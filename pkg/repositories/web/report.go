package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

var (
	// ErrNoReports is returned when no reports are found.
	ErrNoReports = errors.New("no reports found")
)

func (r *repository) GetReportsInPeriod(start, end time.Time) ([]*models.Report, error) {
	query := `
		SELECT
			r.id
		FROM
			report r
		LEFT OUTER JOIN report r2 ON r.host = r2.host AND r.executed_at < r2.executed_at
		WHERE r2.id IS NULL
		AND r.executed_at >= ? AND r.executed_at <= ?
		ORDER BY r.executed_at DESC
	`

	ids := make([]int, 0)
	if err := r.db.Select(&ids, query, start, end); err != nil {
		return nil, fmt.Errorf("failed to get report ids in period: %w", err)
	}

	if len(ids) == 0 {
		return nil, ErrNoReports
	}

	reports := make([]*models.Report, len(ids))
	for i, id := range ids {
		report, err := models.ReportById(r.db, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get report %d: %w", id, err)
		}
		reports[i] = report
	}

	return reports, nil

}

func (r *repository) GetLatestUniqueReportHosts(start, end time.Time) ([]*models.Report, error) {
	query := `
		SELECT
			r.id
		FROM
			report r
		LEFT OUTER JOIN report r2 ON r.host = r2.host AND r.executed_at < r2.executed_at
		WHERE r2.id IS NULL
		AND r.executed_at >= ? AND r.executed_at <= ?
		ORDER BY r.executed_at DESC
	`

	ids := make([]int, 0)
	if err := r.db.Select(&ids, query, start, end); err != nil {
		return nil, fmt.Errorf("failed to get report ids: %w", err)
	}

	if len(ids) == 0 {
		return nil, ErrNoReports
	}

	reports := make([]*models.Report, len(ids))
	for i, id := range ids {
		report, err := models.ReportById(r.db, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get report %d: %w", id, err)
		}
		reports[i] = report
	}

	return reports, nil
}
