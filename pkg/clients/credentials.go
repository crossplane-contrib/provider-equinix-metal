/*
Copyright 2019 The Crossplane Authors.

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

package clients

// Credentials is a common credential format used by various Equinix Metal Kubernetes
// providers
type Credentials struct {
	APIKey     string `json:"apiKey"`
	ProjectID  string `json:"projectID"`
	FacilityID string `json:"facilityID"`
}

// Using these constants causes Credential methods to return the credential
// configured values
const (
	CredentialAPIKey     = ""
	CredentialProjectID  = ""
	CredentialFacilityID = ""
)

// DefaultGetter provides setters for common Equinix Metal client properties
type DefaultGetter interface {
	GetProjectID(string) string
	GetFacilityID(string) string
}

// DefaultSetter provides setters for common Equinix Metal client properties
type DefaultSetter interface {
	SetProjectID(string)
	SetFacilityID(string)
}

// Defaulter provides getter and setters for common Equinix Metal client properties
type Defaulter interface {
	DefaultGetter
	DefaultSetter
}

// GetProjectID returns the supplied ProjectID or the ProjectID included with
// the Client credentials (if any)
func (c *Credentials) GetProjectID(projectID string) string {
	if projectID != "" {
		return projectID
	}
	return c.ProjectID
}

// GetFacilityID returns the supplied FacilityID or the FacilityID included with
// the Client credentials (if any)
func (c *Credentials) GetFacilityID(facilityID string) string {
	if facilityID != "" {
		return facilityID
	}
	return c.FacilityID
}

// GetAPIKey returns the supplied APIKey or the APIKey included with the
// Client credentials (if any)
func (c *Credentials) GetAPIKey(apiKey string) string {
	if apiKey != "" {
		return apiKey
	}
	return c.APIKey
}

// SetProjectID sets the default ProjectID for the client
func (c *Credentials) SetProjectID(projectID string) {
	c.ProjectID = projectID
}

// SetFacilityID sets the default FacilityID for the client
func (c *Credentials) SetFacilityID(facilityID string) {
	c.FacilityID = facilityID
}

// SetAPIKey sets the default APIKey for the client
func (c *Credentials) SetAPIKey(apiKey string) {
	c.APIKey = apiKey
}

// LateInitializeStringPtr returns `from` if `in` is empty and in other cases it
// returns `in`.
func LateInitializeStringPtr(in *string, from *string) *string {
	if in == nil {
		return from
	}
	return in
}

// LateInitializeString returns `from` if `in` is empty and in other cases it
// returns `in`.
func LateInitializeString(in string, from *string) string {
	if in == "" && from != nil {
		return *from
	}
	return in
}

// LateInitializeBoolPtr returns in if it's non-nil, otherwise returns from
func LateInitializeBoolPtr(in *bool, from *bool) *bool {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeIntPtr returns in if it's non-nil, otherwise returns from
func LateInitializeIntPtr(in *int, from *int) *int {
	if in != nil {
		return in
	}
	return from
}
