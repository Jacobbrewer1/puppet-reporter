package api

const (
	reportKeyHost           = "host"
	reportKeyPuppetVersion  = "puppet_version"
	reportKeyEnvironment    = "environment"
	reportKeyExecutionTime  = "time"
	reportKeyStatus         = "status"
	reportKeyMetrics        = "metrics"
	reportKeyRuntimes       = reportKeyMetrics + ".time.values"
	reportKeyResourceStatus = reportKeyMetrics + ".resources.values"
	reportKeyLogs           = "logs"
	reportKeyResources      = "resource_statuses"

	resourceKeyResourceType = "resource_type"
	resourceKeyFile         = "file"
	resourceKeyLine         = "line"
	resourceKeyTitle        = "title"

	stateSkipped = "skipped"
	stateFailed  = "failed"
	stateChanged = "changed"

	unknownLineNum = -1
)
