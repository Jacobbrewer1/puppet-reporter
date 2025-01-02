package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) UploadReport(l *slog.Logger, r *http.Request, body0 *api.UploadReportJSONBody) (*api.ReportDetails, error) {
	bts, err := body0.File.Bytes()
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusBadRequest, err, "error reading file")
	}

	rep, err := parsePuppetReport(bts)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusBadRequest, err, "error parsing report")
	}

	existingRep, err := s.r.GetReportByHash(rep.Report.Hash)
	if err != nil && !errors.Is(err, repo.ErrReportNotFound) {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "error getting report")
	} else if existingRep != nil {
		return nil, uhttp.NewHTTPError(http.StatusConflict, fmt.Errorf("report with hash %s already exists", rep.Report.Hash), "report already exists")
	}

	if err := s.r.SaveReport(rep.Report); err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "error saving report")
	}

	wg := new(sync.WaitGroup)

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range rep.Resources {
			rep.Resources[i].ReportId = rep.Report.Id
		}
	}()
	go func() {
		defer wg.Done()
		for i := range rep.Logs {
			rep.Logs[i].ReportId = rep.Report.Id
		}
	}()

	wg.Wait()

	if err := s.r.SaveResources(rep.Resources); err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "error saving resources")
	}

	if err := s.r.SaveLogs(rep.Logs); err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "error saving logs")
	}

	go updateMetrics(rep)

	respReport := s.modelAsApiReport(rep.Report)
	respLogs := make([]api.LogMessage, len(rep.Logs))
	for i, log := range rep.Logs {
		respLogs[i] = *s.modelAsApiLogMessage(log)
	}

	respResources := make([]api.Resource, len(rep.Resources))
	for i, resource := range rep.Resources {
		respResources[i] = *s.modelAsApiResource(resource)
	}

	respReportDetails := &api.ReportDetails{
		Logs:      respLogs,
		Report:    *respReport,
		Resources: respResources,
	}

	return respReportDetails, nil
}

func updateMetrics(rep *CompleteReport) {
	totalReports.WithLabelValues(strings.ToLower(rep.Report.State), strings.ToLower(rep.Report.Environment)).Inc()
}
