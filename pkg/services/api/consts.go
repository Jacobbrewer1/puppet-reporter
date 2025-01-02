package api

const (
	reportKeyHost           = "host"
	reportKeyPuppetVersion  = "puppet_version"
	reportKeyEnvironment    = "environment"
	reportKeyTime           = "time"
	reportKeyStatus         = "status"
	reportKeyMetrics        = "metrics"
	reportKeyValues         = "values"
	reportKeyLogs           = "logs"
	reportKeyResources      = "resources"
	reportKeyResourceStates = "resource_statuses"

	resourceKeyResourceType = "RESOURCE_TYPE"
	resourceKeyFile         = "FILE"
	resourceKeyLine         = "LINE"
	resourceKeyTitle        = "TITLE"

	stateSkipped = "SKIPPED"
	stateFailed  = "FAILED"
	stateChanged = "CHANGED"

	unknownLineNum = -1
)
