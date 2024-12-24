package api

const (
	reportKeyHost          = "host"
	reportKeyPuppetVersion = "puppet_version"
	reportKeyEnvironment   = "environment"
	reportKeyTime          = "time"
	reportKeyStatus        = "status"
	reportKeyMetrics       = "metrics"
	reportKeyValues        = "values"
	reportKeyLogs          = "logs"
	reportKeyResources     = "resource_statuses"

	resourceKeyResourceType = "resource_type"
	resourceKeyFile         = "file"
	resourceKeyLine         = "line"
	resourceKeyTitle        = "title"

	stateSkipped   = "skipped"
	stateFailed    = "failed"
	stateChanged   = "changed"
	stateUnchanged = "unchanged"

	unknownLineNum = -1
)
