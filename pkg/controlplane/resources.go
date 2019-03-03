/*
Copyright 2018 The v3io Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
