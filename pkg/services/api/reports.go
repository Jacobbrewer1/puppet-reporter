package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
	"github.com/jacobbrewer1/puppet-reporter/pkg/utils"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) GetReports(w http.ResponseWriter, r *http.Request) {
	l := logging.LoggerFromRequest(r)

	paginationDetails, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		l.Error("Failed to get pagination details", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusBadRequest, "failed to get pagination details", err)
		return
	}

	filts, err := s.getReportsFilters(r)
	if err != nil {
		l.Error("Failed to parse filters", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusBadRequest, "failed to parse filters", err)
		return
	}

	reports, err := s.r.GetReports(paginationDetails, filts)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reports = &pagefilter.PaginatedResponse[models.Report]{
				Items: make([]*models.Report, 0),
				Total: 0,
			}
		default:
			slog.Error("Error getting reports", slog.String(logging.KeyError, err.Error()))
			uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "error getting reports", err)
			return
		}
	}

	respArray := make([]api.Report, len(reports.Items))
	for i, report := range reports.Items {
		respArray[i] = *s.modelAsApiReport(report)
	}

	resp := &api.ReportResponse{
		Reports: respArray,
		Total:   reports.Total,
	}

	if err := uhttp.EncodeJSON(w, http.StatusOK, resp); err != nil {
		l.Error("Failed to encode response", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "failed to encode response", err)
	}
}

func (s *service) getReportsFilters(r *http.Request) (*repo.GetReportsFilters, error) {
	filters := new(repo.GetReportsFilters)

	return filters, nil
}

func (s *service) modelAsApiReport(report *models.Report) *api.Report {
	return &api.Report{
		Environment:    &report.Environment,
		ExecutedAt:     &report.ExecutedAt,
		Hash:           &report.Hash,
		Host:           &report.Host,
		PuppetVersion:  utils.Ptr(float32(report.PuppetVersion)),
		RuntimeSeconds: utils.Ptr(int64(report.Runtime)),
		Status:         &report.State,
		TotalChanged:   utils.Ptr(int64(report.Changed)),
		TotalFailed:    utils.Ptr(int64(report.Failed)),
		TotalResources: utils.Ptr(int64(report.Total)),
		TotalSkipped:   utils.Ptr(int64(report.Skipped)),
	}
}
