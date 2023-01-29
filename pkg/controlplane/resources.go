/*
Copyright 2019 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package v3ioc

type SessionAttributes struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Plane         string `json:"plane,omitempty"`
	InterfaceType string `json:"interfaceType,omitempty"`
}

type UserAttributes struct {
	AssignedPolicies       []string `json:"assigned_policies,omitempty"`
	AuthenticationScheme   string   `json:"authentication_scheme,omitempty"`
	CreatedAt              string   `json:"created_at,omitempty"`
	DataAccessMode         string   `json:"data_access_mode,omitempty"`
	Department             string   `json:"department,omitempty"`
	Description            string   `json:"description,omitempty"`
	Email                  string   `json:"email,omitempty"`
	Enabled                bool     `json:"enabled,omitempty"`
	FirstName              string   `json:"first_name,omitempty"`
	JobTitle               string   `json:"job_title,omitempty"`
	LastName               string   `json:"last_name,omitempty"`
	PasswordChangedAt      string   `json:"password_changed_at,omitempty"`
	PhoneNumber            string   `json:"phone_number,omitempty"`
	SendPasswordOnCreation bool     `json:"send_password_on_creation,omitempty"`
	UID                    int      `json:"uid,omitempty"`
	UpdatedAt              string   `json:"updated_at,omitempty"`
	Username               string   `json:"username,omitempty"`
	Password               string   `json:"password,omitempty"`
}

type ContainerAttributes struct {
	Name string `json:"name,omitempty"`
}

type ClusterInfoAttributes struct {
	Endpoints ClusterInfoEndpoints `json:"endpoints,omitempty"`
}

type ClusterInfoEndpoints struct {
	AppServices []ClusterInfoEndpointAddress `json:"app_services,omitempty"`
}

type ClusterInfoEndpointAddress struct {
	ServiceID   string   `json:"service_id,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Urls        []string `json:"urls,omitempty"`
	APIUrls     []string `json:"api_urls,omitempty"`
}

type EventAttributes struct {
	SystemEvent    bool           `json:"system_event,omitempty"`
	Source         string         `json:"source,omitempty"`
	Kind           string         `json:"kind,omitempty"`
	Description    string         `json:"description,omitempty"`
	Severity       Severity       `json:"severity,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Visibility     Visibility     `json:"visibility,omitempty"`
	Classification Classification `json:"classification,omitempty"`

	TimestampUint64  uint64 `json:"timestamp_uint64,omitempty"`
	TimestampIso8601 string `json:"timestamp_iso8601,omitempty"`

	ParametersUint64 []ParameterUint64 `json:"parameters_uint64,omitempty"`
	ParametersText   []ParameterText   `json:"parameters_text,omitempty"`

	InvokingUserID string `json:"invoking_user_id,omitempty"`
	AuditTenant    string `json:"audit_tenant,omitempty"`
}

type AffectedResource struct {
	ResourceType string `json:"resource_type,omitempty"`
	IDStr        string `json:"id_str,omitempty"`
	ID           uint64 `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
}

type ParameterText struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type ParameterUint64 struct {
	Name  string `json:"name,omitempty"`
	Value uint64 `json:"value,omitempty"`
}

type Severity string

const (
	UnknownSeverity  Severity = "unknown"
	DebugSeverity    Severity = "debug"
	InfoSeverity     Severity = "info"
	WarningSeverity  Severity = "warning"
	MajorSeverity    Severity = "major"
	CriticalSeverity Severity = "critical"
)

type Visibility string

const (
	UnknownVisibility      Visibility = "unknown"
	InternalVisibility     Visibility = "internal"
	ExternalVisibility     Visibility = "external"
	CustomerOnlyVisibility Visibility = "customerOnly"
)

type Classification string

const (
	UnknownClassification Classification = "unknown"
	HwClassification      Classification = "hw"
	UaClassification      Classification = "ua"
	BgClassification      Classification = "bg"
	SwClassification      Classification = "sw"
	SLAClassification     Classification = "sla"
	CapClassification     Classification = "cap"
	SecClassification     Classification = "sec"
	AuditClassification   Classification = "audit"
	SystemClassification  Classification = "system"
)

// AccessKeyAttributes holds info about a access key
type AccessKeyAttributes struct {
	TTL           int      `json:"ttl,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
	ExpiresAt     int      `json:"expires_at,omitempty"`
	GroupIds      []string `json:"group_ids,omitempty"`
	UID           int      `json:"uid,omitempty"`
	GIDs          []int    `json:"gids,omitempty"`
	TenantID      string   `json:"tenant_id,omitempty"`
	Kind          string   `json:"kind,omitempty"`
	Plane         Plane    `json:"plane,omitempty"`
	InterfaceKind string   `json:"interface_kind,omitempty"`
	Label         string   `json:"label,omitempty"`
}

type ClusterConfigurationReloadAttributes struct {
	JobID string `json:"job_id,omitempty"`
}

type Plane string

const (
	ControlPlane Plane = "control"
	DataPlane    Plane = "data"
)

type Kind string

const (
	SessionKind   Kind = "session"
	AccessKeyKind Kind = "accessKey"
)

type JobState string

const (
	JobStateCreated      JobState = "created"
	JobStateDispatched   JobState = "dispatched"
	JobStateInProgress   JobState = "in_progress"
	JobStateCompleted    JobState = "completed"
	JobStateCanceled     JobState = "canceled"
	JobStateRepublishing JobState = "republishing"
	JobStateFailed       JobState = "failed"
)

type JobWebHook struct {
	Method   string `json:"method,omitempty"`
	Target   string `json:"target,omitempty"`
	Resource string `json:"resource,omitempty"`
	Status   int    `json:"status,omitempty"`
	Payload  string `json:"payload,omitempty"`
}

type JobAttributes struct {
	Kind                   string       `json:"kind,omitempty"`
	Params                 string       `json:"params,omitempty"`
	UIFields               string       `json:"ui_fields,omitempty"`
	MaxTotalExecutionTime  int          `json:"max_total_execution_time,omitempty"`
	MaxWorkerExecutionTime int          `json:"max_worker_execution_time,omitempty"`
	Delay                  float64      `json:"delay,omitempty"`
	State                  JobState     `json:"state,omitempty"`
	Result                 string       `json:"result,omitempty"`
	CreatedAt              string       `json:"created_at,omitempty"`
	OnSuccess              []JobWebHook `json:"on_success,omitempty"`
	OnFailure              []JobWebHook `json:"on_failure,omitempty"`
	LinkedResources        string       `json:"linked_resources,omitempty"`
	UpdatedAt              string       `json:"updated_at,omitempty"`
	Handler                string       `json:"handler,omitempty"`
}

type AppServicesManifest struct {
	Type       string                        `json:"type,omitempty"`
	ID         int                           `json:"id,omitempty"`
	Attributes AppServicesManifestAttributes `json:"attributes,omitempty"`
}

type AppServicesManifestAttributes struct {
	AppServices            interface{} `json:"app_services,omitempty"`
	ClusterName            string      `json:"cluster_name,omitempty"`
	LastModificationJob    string      `json:"last_modification_job,omitempty"`
	RunningModificationJob string      `json:"running_modification_job,omitempty"`
	State                  string      `json:"state,omitempty"`
	TenantID               string      `json:"tenant_id,omitempty"`
	TenantName             string      `json:"tenant_name,omitempty"`
}
