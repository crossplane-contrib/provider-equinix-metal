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

package v1alpha2

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/packethost/packngo"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// StateActive indicates device is in active state
	StateActive = "active"

	// StateProvisioning indicates device is in provisioning state
	StateProvisioning = "provisioning"

	// StateQueued indicates device is in queued state
	StateQueued = "queued"
)

// TODO: make optional parameters pointers and add +optional

// DeviceSpec defines the desired state of Device
type DeviceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  DeviceParameters `json:"forProvider"`
}

// DeviceStatus defines the observed state of Device
type DeviceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     DeviceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Device is a managed resource that represents a Packet Device
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="HOSTNAME",type="string",JSONPath=".status.atProvider.hostname"
// +kubebuilder:printcolumn:name="IPV4",type="string",JSONPath=".status.atProvider.ipv4"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,packet}
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeviceList contains a list of Devices
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

// DeviceParameters define the desired state of a Packet device.
// https://www.packet.com/developers/api/#devices
type DeviceParameters struct {
	Hostname              string                           `json:"hostname,omitempty"`
	Plan                  string                           `json:"plan"`
	Facility              string                           `json:"facility"`
	OS                    string                           `json:"operatingSystem"`
	Description           string                           `json:"description,omitempty"`
	BillingCycle          string                           `json:"billingCycle,omitempty"`
	UserData              string                           `json:"userdata,omitempty"`
	Tags                  []string                         `json:"tags,omitempty"`
	Locked                bool                             `json:"locked,omitempty"`
	IPXEScriptURL         string                           `json:"ipxeScriptUrl,omitempty"`
	PublicIPv4SubnetSize  int                              `json:"publicIPv4SubnetSize,omitempty"`
	AlwaysPXE             bool                             `json:"alwaysPXE,omitempty"`
	HardwareReservationID string                           `json:"hardwareReservationID,omitempty"`
	CustomData            string                           `json:"customData,omitempty"`
	UserSSHKeys           []string                         `json:"userSSHKeys,omitempty"`
	ProjectSSHKeys        []string                         `json:"projectSSHKeys,omitempty"`
	Features              map[string]string                `json:"features,omitempty"`
	IPAddresses           []packngo.IPAddressCreateRequest `json:"ipAddresses,omitempty"`
}

// DeviceObservation is used to reflect in the Kubernetes API, the observed
// state of the Device resource from the Packet API.
type DeviceObservation struct {
	ID                  string            `json:"id"`
	Href                string            `json:"href,omitempty"`
	Hostname            string            `json:"hostname,omitempty"`
	Description         string            `json:"description,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	State               string            `json:"state,omitempty"`
	ProvisionPercentage resource.Quantity `json:"provisionPercentage,omitempty"`
	IPv4                string            `json:"ipv4,omitempty"`
	Locked              bool              `json:"locked"`
	BillingCycle        string            `json:"billingCycle,omitempty"`
	NetworkType         string            `json:"networkType,omitempty"`
	CreatedAt           metav1.Time       `json:"createdAt,omitempty"`
	UpdatedAt           metav1.Time       `json:"updatedAt,omitempty"`

	// IQN string is omitted
	// ImageURL *string is omitted
	// Facility map is omitted (see FacilityCode, FacilityID)
	// HardwareReservation map is omitted
	// IPAddresses []map is omitted
	// NetworkPorts []map is omitted
	// OperatingSystem map is omitted
	// Plan map is omitted
	// Project map is omitted
	// ShortID string is omitted
	// SSHKeys []map is omitted
	// Volumes []map is omitted

	// TODO(displague) should user+pass yield a secret?
	// User string is omitted
	// RootPassword string is omitted (available for 24 hours)
}

// DeviceClassSpecTemplate is a template for the spec of a dynamically provisioned Device.
type DeviceClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       DeviceParameters `json:"forProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DeviceClass is a resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:resource:scope=Cluster
type DeviceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// DeviceSpec.
	SpecTemplate DeviceClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// DeviceClassList contains a list of device resource classes.
type DeviceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceClass `json:"items"`
}
