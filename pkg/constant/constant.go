package constant

// default user
const DefaultUserID = "admin"
const DefaultUserPassword = "password"
const DefaultUserEmail = "hello@instill-ai.com"
const DefaultUserCustomerID = ""
const DefaultUserCompanyName = "Instill AI"
const DefaultUserDisplayName = "Instill"
const DefaultUserRole = "hobbyist"
const DefaultUserNewsletterSubscription = true
const DefaultJwtExpiration = 86400 * 7
const DefaultJwtIssuer = "http://localhost:8080"
const DefaultJwtAudience = "http://localhost:8080"

const PresetOrgID = "preset"
const PresetOrgUID = "63196cec-1c95-49e8-9bf6-9f9497a15f72"
const PresetOrgDisplayName = "Preset"

// HeaderUserIDKey is the header key for the User ID
const HeaderUserIDKey = "Instill-User-Id"

// HeaderUserUIDKey is the header key for the User UID
const HeaderUserUIDKey = "Instill-User-Uid"
const HeaderVisitorUIDKey = "Instill-Visitor-Uid"

const HeaderAuthType = "Instill-Auth-Type"

const DefaultTokenType = "Bearer"
const AccessTokenKeyFormat = "access_token:%s:owner_permalink"
const HeaderAuthorization = "Authorization"

const MaxPayloadSize = 1024 * 1024 * 32

// Filter enum
const (
	Start              string = "start"
	Stop               string = "stop"
	OwnerName          string = "owner_name"
	PipelineID         string = "pipeline_id"
	PipelineUID        string = "pipeline_uid"
	PipelineReleaseID  string = "pipeline_release_id"
	PipelineReleaseUID string = "pipeline_release_uid"
	TriggerMode        string = "trigger_mode"
	Status             string = "status"
	Email              string = "email"
	UserID             string = "id"
)

// Metric data enum
const (
	PipelineOwnerUID           string = "owner_uid"
	PipelineTriggerMeasurement string = "pipeline.trigger"
	PipelineMode               string = "pipeline_mode"
	PipelineTriggerID          string = "pipeline_trigger_id"
	TriggerTime                string = "trigger_time"
	ExecuteTime                string = "execute_time"
	ComputeTimeDuration        string = "compute_time_duration"
	Completed                  string = "completed"
	Errored                    string = "errored"
)
