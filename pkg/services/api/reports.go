package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) GetReports(w http.ResponseWriter, r *http.Request, params *api.GetReportsParams) {
	l := logging.LoggerFromRequest(r)

	paginationDetails, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		l.Error("Failed to get pagination details", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusBadRequest, "failed to get pagination details", err)
		return
	}

	filts, err := s.getReportsFilters(params)
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
			l.Error("Error getting reports", slog.String(logging.KeyError, err.Error()))
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

func (s *service) getReportsFilters(params *api.GetReportsParams) (*repo.GetReportsFilters, error) {
	filters := new(repo.GetReportsFilters)
	if params == nil {
		return filters, nil
	}

	if params.Environment != nil {
		filters.Environment = params.Environment
	}

	if params.Host != nil {
		filters.Host = params.Host
	}

	if params.State != nil {
		filters.State = params.State
	}

	return filters, nil
}

func (s *service) modelAsApiReport(report *models.Report) *api.Report {
	return &api.Report{
		Environment:    report.Environment,
		ExecutedAt:     report.ExecutedAt,
		Hash:           report.Hash,
		Host:           report.Host,
		Id:             int64(report.Id),
		PuppetVersion:  float32(report.PuppetVersion),
		RuntimeSeconds: int64(report.Runtime),
		Status:         report.State,
		TotalChanged:   int64(report.Changed),
		TotalFailed:    int64(report.Failed),
		TotalResources: int64(report.Total),
		TotalSkipped:   int64(report.Skipped),
	}
}

func (s *service) GetReport(w http.ResponseWriter, r *http.Request, hash string) {
	l := logging.LoggerFromRequest(r)

	report, err := s.reportDetailsByHash(hash)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrReportNotFound):
			l.Error("Report not found", slog.String(logging.KeyError, err.Error()))
			uhttp.SendErrorMessageWithStatus(w, http.StatusNotFound, "report not found", err)
			return
		default:
			l.Error("Failed to get report details", slog.String(logging.KeyError, err.Error()))
			uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "failed to get report details", err)
			return
		}
	}

	if err := uhttp.EncodeJSON(w, http.StatusOK, report); err != nil {
		l.Error("Failed to encode response", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "failed to encode response", err)
	}
}

func (s *service) modelAsApiLogMessage(log *models.LogMessage) *api.LogMessage {
	return &api.LogMessage{
		Message: log.Message,
	}
}

func (s *service) modelAsApiResource(resource *models.Resource) *api.Resource {
	return &api.Resource{
		File:   resource.File,
		Line:   int64(resource.Line),
		Name:   resource.Name,
		Status: api.Status(resource.Status),
		Type:   resource.Type,
	}
}

func (s *service) reportDetailsByHash(hash string) (*api.ReportDetails, error) {
	report, err := s.r.GetReportByHash(hash)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrReportNotFound):
			return nil, fmt.Errorf("report not found: %w", err)
		default:
			return nil, fmt.Errorf("failed to get report: %w", err)
		}
	}

	reportResp := s.modelAsApiReport(report)

	resources, err := s.r.GetResourcesByReportID(report.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	resourceResp := make([]api.Resource, len(resources))
	for i, resource := range resources {
		resourceResp[i] = *s.modelAsApiResource(resource)
	}

	logs, err := s.r.GetLogsByReportID(report.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	logResp := make([]api.LogMessage, len(logs))
	for i, log := range logs {
		logResp[i] = *s.modelAsApiLogMessage(log)
	}

	resp := &api.ReportDetails{
		Logs:      logResp,
		Report:    *reportResp,
		Resources: resourceResp,
	}

	return resp, nil
}
