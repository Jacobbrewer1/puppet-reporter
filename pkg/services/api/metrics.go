package api

import (
	"github.com/jacobbrewer1/puppet-reporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var appNameSuffix = utils.PackageName(&service{})

var (
	// totalReports is a counter for the total number of reports processed
	totalReports = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "total_reports",
			Namespace: utils.AppName(appNameSuffix),
			Help:      "Total number of reports processed",
		},
		[]string{"state", "environment"},
	)
)
