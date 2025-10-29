package utils

const ContextRedis = "ContextRedis"

// labels
const (
	AppName              = "app.kubernetes.io/name"
	AppInstance          = "app.kubernetes.io/instance"
	AppVersion           = "app.kubernetes.io/version"
	AppComponent         = "app.kubernetes.io/component"
	AppManagedBy         = "app.kubernetes.io/managed-by"
	AppManagedByOperator = "app.kubernetes.io/managed-by-operator"
	AppProcByOperator    = "app.kubernetes.io/processed-by-operator"
	AppTechnology        = "app.kubernetes.io/technology"
	AppPartOf            = "app.kubernetes.io/part-of"
	DeploymentSessionId  = "deployment.netcracker.com/sessionId"
)
