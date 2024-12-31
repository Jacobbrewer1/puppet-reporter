package api

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
	"github.com/jacobbrewer1/uhttp"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *service) UploadReport(w http.ResponseWriter, r *http.Request) {
	l := logging.LoggerFromRequest(r)

	if r.Body == http.NoBody {
		l.Debug("No body in request")
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "No body in request")
		return
	}

	bdy, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("Error reading body", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error reading body", err)
		return
	}

	fileBody := &api.UploadReportMultipartBody{
		File: new(openapi_types.File),
	}
	fileBody.File.InitFromBytes(bdy, defaultFileName)

	bts, err := fileBody.File.Bytes()
	if err != nil {
		l.Error("Error reading file", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error reading file", err)
		return
	}

	rep, err := parsePuppetReport(bts)
	if err != nil {
		l.Error("Error parsing puppet report", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error parsing puppet report", fmt.Errorf("error parsing puppet report: %w", err))
		return
	}

	existingRep, err := s.r.GetReportByHash(rep.Report.Hash)
	if err != nil && !errors.Is(err, repo.ErrReportNotFound) {
		l.Error("Error getting report by hash", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error getting report by hash", err)
		return
	} else if existingRep != nil {
		l.Debug("Report already exists", slog.String("hash", rep.Report.Hash))
		uhttp.SendMessageWithStatus(w, http.StatusConflict, "Report already exists")
		return
	}

	if err := s.r.SaveReport(rep.Report); err != nil {
		l.Error("Error saving report", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error saving report", err)
		return
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
		l.Error("Error saving resources", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error saving resources", err)
		return
	}

	if err := s.r.SaveLogs(rep.Logs); err != nil {
		l.Error("Error saving logs", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error saving logs", err)
		return
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

	uhttp.MustEncode(w, http.StatusCreated, respReportDetails)
}

func updateMetrics(rep *CompleteReport) {
	totalReports.WithLabelValues(strings.ToLower(rep.Report.State), strings.ToLower(rep.Report.Environment)).Inc()
}
