package api

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jacobbrewer1/goschema/usql"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	"github.com/jacobbrewer1/utils"
	"github.com/smallfish/simpleyaml"
)

type CompleteReport struct {
	Report    *models.Report
	Resources []*models.Resource
	Logs      []*models.LogMessage
}

func parsePuppetReport(content []byte) (*CompleteReport, error) {
	complete := new(CompleteReport)
	report := new(models.Report)

	report.Hash = utils.Sha256(content)

	yaml, err := simpleyaml.NewYaml(content)
	if err != nil {
		return nil, errors.New("failed to parse YAML")
	}

	if err := parseHost(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing host: %w", err)
	}

	if err := parsePuppetVersion(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing puppet version: %w", err)
	}

	if err := parseEnvironment(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing environment: %w", err)
	}

	if err := parseExecutionTime(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing execution time: %w", err)
	}

	if err := parseStatus(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}

	if err := parseRuntime(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing runtime: %w", err)
	}

	if err := parseResourceStates(report, yaml); err != nil {
		return nil, fmt.Errorf("parsing resource states: %w", err)
	}

	complete.Report = report

	if logs := parseLogs(yaml); len(logs) > 0 {
		complete.Logs = logs
	}

	resources, err := parseResources(yaml)
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

	return complete, nil
}

// parseHost reads the `host` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseHost(rep *models.Report, y *simpleyaml.Yaml) error {
	host, err := y.Get(reportKeyHost).String()
	if err != nil {
		return errors.New("failed to get 'host' from YAML")
	}
	reg, _ := regexp.Compile("^([a-z0-9._-]+)$")
	if !reg.MatchString(host) {
		return errors.New("the submitted 'host' field failed our security check")
	}
	rep.Host = host
	return nil
}

// parseEnvironment reads the `environment` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseEnvironment(rep *models.Report, y *simpleyaml.Yaml) error {
	envStr, err := y.Get(reportKeyEnvironment).String()
	if err != nil {
		return errors.New("failed to get 'environment' from YAML")
	}
	reg, _ := regexp.Compile("^([A-Za-z0-9_]+)$")
	if !reg.MatchString(envStr) {
		return errors.New("the submitted 'environment' field failed our security check")
	}
	rep.Environment = strings.ToUpper(envStr)
	return nil
}

// parseExecutionTime reads the `time` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseExecutionTime(rep *models.Report, y *simpleyaml.Yaml) error {
	// Get the time puppet executed
	at, err := y.Get(reportKeyTime).String()
	if err != nil {
		return errors.New("failed to get 'time' from YAML")
	}

	at = strings.Replace(at, "'", "", -1)

	execTime, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return errors.New("failed to parse 'time' from YAML")
	}

	rep.ExecutedAt = execTime

	return nil
}

// parseStatus reads the `status` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseStatus(rep *models.Report, y *simpleyaml.Yaml) error {
	s, err := y.Get(reportKeyStatus).String()
	if err != nil {
		return errors.New("failed to get 'status' from YAML")
	}

	rep.State = usql.NewEnum(strings.ToUpper(s))

	return nil
}

// parseRuntime reads the `metrics.time.values` parameters from the YAML
// and populates given report-structure with suitable values.
func parseRuntime(rep *models.Report, y *simpleyaml.Yaml) error {
	times, err := y.Get(reportKeyMetrics).Get(reportKeyTime).Get(reportKeyValues).Array()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile("Total ([0-9.]+)")

	runtime := ""
	for _, value := range times {
		match := r.FindStringSubmatch(fmt.Sprint(value))
		if len(match) == 2 {
			runtime = match[1]
		}
	}

	// Parse the runtime as a duration.
	d, err := time.ParseDuration(runtime + "s")
	if err != nil {
		return fmt.Errorf("failed to parse runtime '%s' as duration", runtime)
	}

	rep.Runtime = int(d.Seconds())

	return nil
}

// parseResources looks for the counts of resources which have been
// failed, changed, skipped, etc, and updates the given report-structure
// with those values.
func parseResources(y *simpleyaml.Yaml) ([]*models.Resource, error) {
	rs, err := y.Get(reportKeyResourceStates).Map()
	if err != nil {
		return nil, errors.New("failed to get 'resource_statuses' from YAML")
	}

	resources := make([]*models.Resource, 0)

	for _, v2 := range rs {
		m := make(map[string]string)
		v := reflect.ValueOf(v2)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)

				// Store the key/val in the map.
				k, v := key.Interface(), strct.Interface()
				m[strings.ToUpper(k.(string))] = fmt.Sprint(v)
			}
		}

		res := parseResource(m)

		const trueStr = "true"

		switch {
		case m[stateSkipped] == trueStr:
			res.Status = models.ResourceStatusSkipped
		case m[stateChanged] == trueStr:
			res.Status = models.ResourceStatusChanged
		case m[stateFailed] == trueStr:
			res.Status = models.ResourceStatusFailed
		default:
			res.Status = models.ResourceStatusUnchanged
		}

		resources = append(resources, res)
	}

	return resources, nil
}

// parseLogs updates the given report with any logged messages.
func parseLogs(y *simpleyaml.Yaml) []*models.LogMessage {
	logs, err := y.Get(reportKeyLogs).Array()
	if err != nil {
		slog.Error("failed to get 'logs' from YAML", slog.String(logging.KeyError, err.Error()))
		return nil
	}

	logged := make([]*models.LogMessage, 0)

	const (
		keyMessage = "message"
		keySource  = "source"
	)

	for _, v2 := range logs {
		// create a map
		m := make(map[string]string)
		v := reflect.ValueOf(v2)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)

				// Store the key/val in the map.
				key, val := key.Interface(), strct.Interface()
				m[key.(string)] = fmt.Sprint(val)
			}
		}
		if len(m[keyMessage]) > 0 {
			logged = append(logged, &models.LogMessage{
				Message: m[keySource] + " : " + m[keyMessage],
			})
		}
	}

	return logged
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

// parseResourceStates updates the given report with details of any resource
// which was failed, changed, or skipped.
func parseResourceStates(rep *models.Report, y *simpleyaml.Yaml) error {
	gotResources, err := y.Get(reportKeyMetrics).Get(reportKeyResources).Get(reportKeyValues).Array()
	if err != nil {
		return fmt.Errorf("failed to get 'metrics.resources.values' from YAML: %w", err)
	}

	totalReg, _ := regexp.Compile("Total ([0-9]+)")
	failedReg, _ := regexp.Compile("Failed ([0-9]+)")
	skippedReg, _ := regexp.Compile("Skipped ([0-9]+)")
	changedReg, _ := regexp.Compile("Changed ([0-9]+)")

	totalStr := ""
	failedStr := ""
	skippedStr := ""
	changedStr := ""

	for _, r := range gotResources {
		mt := totalReg.FindStringSubmatch(fmt.Sprint(r))
		if len(mt) == 2 {
			totalStr = mt[1]
		}

		mf := failedReg.FindStringSubmatch(fmt.Sprint(r))
		if len(mf) == 2 {
			failedStr = mf[1]
		}

		ms := skippedReg.FindStringSubmatch(fmt.Sprint(r))
		if len(ms) == 2 {
			skippedStr = ms[1]
		}

		mc := changedReg.FindStringSubmatch(fmt.Sprint(r))
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

	rep.Total = total
	rep.Failed = failed
	rep.Skipped = skipped
	rep.Changed = changed

	return nil
}

func parsePuppetVersion(rep *models.Report, y *simpleyaml.Yaml) error {
	version, err := y.Get(reportKeyPuppetVersion).String()
	if err != nil {
		return errors.New("failed to get 'puppet_version' from YAML")
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

	rep.PuppetVersion = v

	return nil
}
