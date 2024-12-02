package api

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	"github.com/jacobbrewer1/puppet-reporter/pkg/utils"
	"github.com/spf13/viper"
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

	// ErrMissingRuntime is returned when the runtime is missing from the report
	ErrMissingRuntime = errors.New("missing runtime in report")

	// ErrInvalidRuntime is returned when the runtime is invalid
	ErrInvalidRuntime = errors.New("invalid runtime in report")
)

type CompleteReport struct {
	Report    *models.Report
	Resources []*models.Resource
	Logs      []string
}

func parsePuppetReport(content []byte) (*models.Report, error) {
	complete := new(CompleteReport)
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

	if err := parseRuntime(report, vip); err != nil {
		return nil, fmt.Errorf("parsing runtime: %w", err)
	}

	if err := parseResourceStates(report, vip); err != nil {
		return nil, fmt.Errorf("parsing resources: %w", err)
	}

	complete.Report = report

	if logs := parseLogs(vip); len(logs) > 0 {
		complete.Logs = logs
	}

	resources, err := parseResources(vip)
	if err != nil {
		return nil, fmt.Errorf("parsing resources: %w", err)
	}

	sort.Slice(resources, func(i, j int) bool {
		if resources[i].File != resources[j].File {
			return resources[i].File < resources[j].File
		}

		if resources[i].Line != resources[j].Line {
			return resources[i].Line < resources[j].Line
		}

		if resources[i].Type != resources[j].Type {
			return resources[i].Type < resources[j].Type
		}

		if resources[i].Name != resources[j].Name {
			return resources[i].Name < resources[j].Name
		}

		// Resources are equal.
		return false
	})

	complete.Resources = resources

	return report, nil
}

func parseResources(vip *viper.Viper) ([]*models.Resource, error) {
	gotResources := vip.GetStringMap(reportKeyResources)
	resources := make([]*models.Resource, 0)

	for _, r := range gotResources {
		m := make(map[string]string)
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				st := key.MapIndex(key)

				k, val := key.Interface(), st.Interface()
				m[k.(string)] = fmt.Sprintf("%v", val)
			}
		}

		skip, err := strconv.ParseBool(m[stateSkipped])
		if err != nil {
			return nil, fmt.Errorf("failed to parse skipped '%s' as bool", m[stateSkipped])
		}
		change, err := strconv.ParseBool(m[stateChanged])
		if err != nil {
			return nil, fmt.Errorf("failed to parse changed '%s' as bool", m[stateChanged])
		}
		fail, err := strconv.ParseBool(m[stateFailed])
		if err != nil {
			return nil, fmt.Errorf("failed to parse failed '%s' as bool", m[stateFailed])
		}

		res := parseResource(m)
		switch {
		case skip:
			res.Status = models.ResourceStatusSkipped
		case change:
			res.Status = models.ResourceStatusChanged
		case fail:
			res.Status = models.ResourceStatusFailed
		default:
			res.Status = models.ResourceStatusUnchanged
		}

		resources = append(resources, res)
	}

	return resources, nil
}

func parseResource(m map[string]string) *models.Resource {
	res := &models.Resource{
		Name: m[resourceKeyTitle],
		Type: m[resourceKeyResourceType],
		File: m[resourceKeyFile],
	}

	if line, err := strconv.Atoi(m[resourceKeyLine]); err == nil {
		res.Line = line
	} else {
		slog.Warn(fmt.Sprintf("failed to parse line '%s' as int", m[resourceKeyLine]))
		res.Line = unknownLineNum
	}

	return res
}

func parseLogs(vip *viper.Viper) []string {
	return vip.GetStringSlice(reportKeyLogs)
}

func parseResourceStates(r *models.Report, vip *viper.Viper) error {
	gotResources := vip.GetStringSlice(reportKeyResourceStatus)

	totalReg, _ := regexp.Compile("Total ([0-9]+)")
	failedReg, _ := regexp.Compile("Failed ([0-9]+)")
	skippedReg, _ := regexp.Compile("Skipped ([0-9]+)")
	changedReg, _ := regexp.Compile("Changed ([0-9]+)")

	totalStr := ""
	failedStr := ""
	skippedStr := ""
	changedStr := ""

	for _, r := range gotResources {
		mt := totalReg.FindStringSubmatch(r)
		if len(mt) == 2 {
			totalStr = mt[1]
		}

		mf := failedReg.FindStringSubmatch(r)
		if len(mf) == 2 {
			failedStr = mf[1]
		}

		ms := skippedReg.FindStringSubmatch(r)
		if len(ms) == 2 {
			skippedStr = ms[1]
		}

		mc := changedReg.FindStringSubmatch(r)
		if len(mc) == 2 {
			changedStr = mc[1]
		}
	}

	if totalStr == "" || failedStr == "" || skippedStr == "" || changedStr == "" {
		return errors.New("failed to parse resource metrics")
	}

	total, err := strconv.Atoi(totalStr)
	if err != nil {
		return fmt.Errorf("failed to parse total '%s' as int", totalStr)
	}

	failed, err := strconv.Atoi(failedStr)
	if err != nil {
		return fmt.Errorf("failed to parse failed '%s' as int", failedStr)
	}

	skipped, err := strconv.Atoi(skippedStr)
	if err != nil {
		return fmt.Errorf("failed to parse skipped '%s' as int", skippedStr)
	}

	changed, err := strconv.Atoi(changedStr)
	if err != nil {
		return fmt.Errorf("failed to parse changed '%s' as int", changedStr)
	}

	r.Total = total
	r.Failed = failed
	r.Skipped = skipped
	r.Changed = changed

	return nil
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

func parseRuntime(r *models.Report, vip *viper.Viper) error {
	times := vip.GetStringSlice(reportKeyRuntimes)
	if len(times) == 0 {
		return ErrMissingRuntime
	}

	reg, _ := regexp.Compile("Total ([0-9.]+)")

	runtime := ""
	for _, t := range times {
		match := reg.FindStringSubmatch(t)
		if len(match) == 2 {
			runtime = match[1]
		}
	}

	if runtime == "" {
		return ErrInvalidRuntime
	}

	dur, err := time.ParseDuration(runtime + "s")
	if err != nil {
		return fmt.Errorf("failed to parse runtime '%s' as time.Duration", runtime)
	}

	r.Runtime = int(dur.Seconds())

	return nil
}
