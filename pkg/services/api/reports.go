package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
	"github.com/jacobbrewer1/uhttp"
	"github.com/jacobbrewer1/utils"
)

func (s *service) GetReports(l *slog.Logger, r *http.Request, params api.GetReportsParams) (*api.ReportResponse, error) {
	paginationDetails, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusBadRequest, err, "failed to get pagination details")
	}

	filts, err := s.getReportsFilters(&params)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusBadRequest, err, "failed to parse filters")
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
			return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "error getting reports")
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

	return resp, nil
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
		filters.State = utils.Ptr(string(*params.State))
	}

	if params.From != nil {
		filters.From = params.From
	}

	if params.To != nil {
		filters.To = params.To
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
		Status:         api.ReportStatus(report.State),
		TotalChanged:   int64(report.Changed),
		TotalFailed:    int64(report.Failed),
		TotalResources: int64(report.Total),
		TotalSkipped:   int64(report.Skipped),
	}
}

func (s *service) GetReport(l *slog.Logger, r *http.Request, hash string) (*api.ReportDetails, error) {
	report, err := s.reportDetailsByHash(hash)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrReportNotFound):
			return nil, uhttp.NewHTTPError(http.StatusNotFound, err, "report not found", fmt.Sprintf("hash: %s", hash))
		default:
			return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "failed to get report", fmt.Sprintf("hash: %s", hash))
		}
	}

	return report, nil
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
