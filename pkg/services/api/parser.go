package api

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	"github.com/jacobbrewer1/puppet-reporter/pkg/utils"
	"github.com/spf13/viper"
)

const (
	reportKeyHost          = "host"
	reportKeyPuppetVersion = "puppet_version"
	reportKeyEnvironment   = "environment"
	reportKeyExecutionTime = "time"
	reportKeyStatus        = "status"
)

var (
	// ErrMissingHost is returned when the host is missing from the report
	ErrMissingHost = errors.New("missing host in report")

	// ErrInvalidHost is returned when the host is invalid
	ErrInvalidHost = errors.New("invalid host in report")

	// ErrMissingPuppetVersion is returned when the puppet version is missing from the report
	ErrMissingPuppetVersion = errors.New("missing puppet version in report")

	// ErrMissingEnvironment is returned when the environment is missing from the report
	ErrMissingEnvironment = errors.New("missing environment in report")

	// ErrMissingExecutionTime is returned when the execution time is missing from the report
	ErrMissingExecutionTime = errors.New("missing execution time in report")

	// ErrMissingStatus is returned when the status is missing from the report
	ErrMissingStatus = errors.New("missing status in report")
)

func parsePuppetReport(content []byte) (*models.Report, error) {
	report := new(models.Report)

	report.Hash = utils.Sha256(content)

	vip := viper.New()
	vip.SetConfigType("yaml")
	err := vip.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	if err := parseHost(report, vip); err != nil {
		return nil, fmt.Errorf("parsing host: %w", err)
	}

	if err := parsePuppetVersion(report, vip); err != nil {
		return nil, fmt.Errorf("parsing puppet version: %w", err)
	}

	if err := parseEnvironment(report, vip); err != nil {
		return nil, fmt.Errorf("parsing environment: %w", err)
	}

	if err := parseExecutionTime(report, vip); err != nil {
		return nil, fmt.Errorf("parsing execution time: %w", err)
	}

	if err := parseStatus(report, vip); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}

	return report, nil
}

func parseHost(r *models.Report, vip *viper.Viper) error {
	host := vip.GetString(reportKeyHost)
	if host == "" {
		return ErrMissingHost
	}

	reg, _ := regexp.Compile("^([a-z0-9._-]+)$")
	if !reg.MatchString(host) {
		return ErrInvalidHost
	}

	r.Host = host

	return nil
}

func parsePuppetVersion(r *models.Report, vip *viper.Viper) error {
	version := vip.GetString(reportKeyPuppetVersion)
	if version == "" {
		return ErrMissingPuppetVersion
	}

	// Strip any quotes that might surround the version.
	version = strings.Replace(version, "'", "", -1)

	// Trim the version to 1 decimal place. (4.8.2 -> 4.8)
	elms := strings.Split(version, ".")
	if len(elms) > 2 {
		version = elms[0] + "." + elms[1]
	}

	// Convert the version to a float.
	v, err := strconv.ParseFloat(version, 64)
	if err != nil {
		return fmt.Errorf("failed to parse puppet_version '%s' as float", version)
	}

	r.PuppetVersion = v

	return nil
}

func parseEnvironment(r *models.Report, vip *viper.Viper) error {
	env := vip.GetString(reportKeyEnvironment)
	if env == "" {
		return ErrMissingEnvironment
	}

	reg, _ := regexp.Compile("^([A-Za-z0-9_]+)$")
	if !reg.MatchString(env) {
		return errors.New("the submitted 'environment' field failed our security check")
	}

	r.Environment = env

	return nil
}

func parseExecutionTime(r *models.Report, vip *viper.Viper) error {
	execTimeStr := vip.GetString(reportKeyExecutionTime)
	if execTimeStr == "" {
		return ErrMissingExecutionTime
	}

	// Strip any quotes that might surround the time.
	execTimeStr = strings.Replace(execTimeStr, "'", "", -1)

	// Parse "2024-02-17T02:00:09.572734022+00:00" as a time.Time
	execTime, err := time.Parse(time.RFC3339Nano, execTimeStr)
	if err != nil {
		return fmt.Errorf("failed to parse time '%s' as time.Time", execTimeStr)
	}

	r.ExecutedAt = execTime

	return nil
}

func parseStatus(r *models.Report, vip *viper.Viper) error {
	status := vip.GetString(reportKeyStatus)
	if status == "" {
		return ErrMissingStatus
	}

	r.State = status

	return nil
}
