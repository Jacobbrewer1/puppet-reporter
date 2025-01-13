package web

import (
	"strings"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

const (
	stylePrefix = "list-group-item"
	styleJoiner = "-"

	failureStyle   = "danger"
	changeStyle    = "info"
	unchangedStyle = "secondary"
)

func getReportStyle(rep *models.Report) string {
	switch rep.State {
	case models.ReportStateChanged:
		return strings.Join([]string{stylePrefix, changeStyle}, styleJoiner)
	case models.ReportStateFailed:
		return strings.Join([]string{stylePrefix, failureStyle}, styleJoiner)
	case models.ReportStateUnchanged:
		return strings.Join([]string{stylePrefix, unchangedStyle}, styleJoiner)
	default:
		return ""
	}
}
