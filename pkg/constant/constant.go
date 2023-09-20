package constant

// default user
const DefaultUserID = "admin"
const DefaultUserPassword = "password"
const DefaultUserEmail = "hello@instill.tech"
const DefaultUserCustomerId = ""
const DefaultUserOrgName = "Instill AI"
const DefaultUserFirstName = "Instill"
const DefaultUserLastName = "AI"
const DefaultUserRole = "hobbyist"
const DefaultUserNewsletterSubscription = true
const DefaultJwtExpiration = 86400 * 7
const DefaultJwtIssuer = "http://localhost:8080"
const DefaultJwtAudience = "http://localhost:8080"

// HeaderUserIDKey is the header key for the User ID
const HeaderUserIDKey = "user-id"

// HeaderUserUIDKey is the header key for the User UID
const HeaderUserUIDKey = "jwt-sub"

// MaxApiKeyNum is the maximum number of API keys
const MaxApiKeyNum = 10

// Filter enum
const (
	Start              string = "start"
	Stop               string = "stop"
	PipelineID         string = "pipeline_id"
	PipelineUID        string = "pipeline_uid"
	PipelineReleaseID  string = "pipeline_release_id"
	PipelineReleaseUID string = "pipeline_release_uid"
	ConnectorID        string = "connector_id"
	ConnectorUID       string = "connector_uid"
	TriggerMode        string = "trigger_mode"
	Status             string = "status"
)

// Metric data enum
const (
	OwnerUID                    string = "owner_uid"
	PipelineTriggerMeasurement  string = "pipeline.trigger"
	ConnectorExecuteMeasurement string = "connector.execute"
	PipelineMode                string = "pipeline_mode"
	PipelineTriggerID           string = "pipeline_trigger_id"
	ConnectorExecuteID          string = "connector_execute_id"
	ConnectorDefinitionUID      string = "connector_definition_uid"
	TriggerTime                 string = "trigger_time"
	Executetime                 string = "execute_time"
	ComputeTimeDuration         string = "compute_time_duration"
	Completed                   string = "completed"
	Errored                     string = "errored"
)
