package pkg

//ERRORS
const (
	ErrResourceNotFoundMessage      = 404
	ErrResourceNotAccessibleMessage = 403
	ErrResourceNotGoodMessage       = 400
	ErrResourceInternal             = 500
)

//Paths
const (
	SQLPath        = "apps/sql"
	ConnectorsPath = "apps/connectors"

	GroupsPath          = "groups"
	UsersPath           = "users"
	ServiceAccountsPath = "service-accounts"

	AclsPath   = "kafka/acls"
	TopicsPath = "kafka/topics"
	QuotasPath = "kafka/quotas"

	SchemasPath       = "schemas"
	AlertSettingsPath = "alert-settings"
	PoliciesPath      = "policies"

	ConnectionsFilePath        = "connections"
	ConnectionsAPIPath         = "v1/connection/connections"
	ConnectionTemplatesAPIPath = "v1/connection/connection-templates"
)
