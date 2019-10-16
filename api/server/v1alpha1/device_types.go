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

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/packethost/packngo"
	corev1 "k8s.io/api/core/v1"
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

// DeviceSpec defines the desired state of Device
type DeviceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	DeviceParameters             `json:",inline"`
}

// DeviceStatus defines the observed state of Device
type DeviceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	ID           string            `json:"id"`
	Href         string            `json:"href,omitempty"`
	Hostname     string            `json:"hostname,omitempty"`
	State        string            `json:"state,omitempty"`
	ProvisionPer resource.Quantity `json:"provisionPer,omitempty"`
	IPv4         string            `json:"ipv4,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Device is the Schema for the devices API
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
}

// SetBindingPhase of this Device.
func (c *Device) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	c.Status.SetBindingPhase(p)
}

// GetBindingPhase of this Device.
func (c *Device) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return c.Status.GetBindingPhase()
}

// SetConditions of this Device.
func (c *Device) SetConditions(cd ...runtimev1alpha1.Condition) {
	c.Status.SetConditions(cd...)
}

// SetClaimReference of this Device.
func (c *Device) SetClaimReference(r *corev1.ObjectReference) {
	c.Spec.ClaimReference = r
}

// GetClaimReference of this Device.
func (c *Device) GetClaimReference() *corev1.ObjectReference {
	return c.Spec.ClaimReference
}

// SetNonPortableClassReference of this Device.
func (c *Device) SetNonPortableClassReference(r *corev1.ObjectReference) {
	c.Spec.NonPortableClassReference = r
}

// GetNonPortableClassReference of this Device.
func (c *Device) GetNonPortableClassReference() *corev1.ObjectReference {
	return c.Spec.NonPortableClassReference
}

// SetWriteConnectionSecretToReference of this Device.
func (c *Device) SetWriteConnectionSecretToReference(r corev1.LocalObjectReference) {
	c.Spec.WriteConnectionSecretToReference = r
}

// GetWriteConnectionSecretToReference of this Device.
func (c *Device) GetWriteConnectionSecretToReference() corev1.LocalObjectReference {
	return c.Spec.WriteConnectionSecretToReference
}

// GetReclaimPolicy of this Device.
func (c *Device) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return c.Spec.ReclaimPolicy
}

// SetReclaimPolicy of this Device.
func (c *Device) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	c.Spec.ReclaimPolicy = p
}

// +kubebuilder:object:root=true

// DeviceList contains a list of Devices
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

// DeviceParameters ...
type DeviceParameters struct {
	Hostname              string                           `json:"hostname"`
	Plan                  string                           `json:"plan"`
	Facility              string                           `json:"facility"`
	OS                    string                           `json:"operatingSystem"`
	BillingCycle          string                           `json:"billingCycle"`
	ProjectID             string                           `json:"projectID"`
	UserData              string                           `json:"userdata,omitempty"`
	Tags                  []string                         `json:"tags,omitempty"`
	Locked                *bool                            `json:"locked,omitemtpy"`
	IPXEScriptURL         string                           `json:"ipxe_script_url,omitempty"`
	PublicIPv4SubnetSize  int                              `json:"public_ipv4_subnet_size,omitempty"`
	AlwaysPXE             bool                             `json:"always_pxe,omitempty"`
	HardwareReservationID string                           `json:"hardware_reservation_id,omitempty"`
	CustomData            string                           `json:"customdata,omitempty"`
	UserSSHKeys           []string                         `json:"user_ssh_keys,omitempty"`
	ProjectSSHKeys        []string                         `json:"project_ssh_keys,omitempty"`
	Features              map[string]string                `json:"features,omitempty"`
	IPAddresses           []packngo.IPAddressCreateRequest `json:"ip_addresses,omitempty"`
}

// DeviceClassSpecTemplate ...
type DeviceClassSpecTemplate struct {
	runtimev1alpha1.NonPortableClassSpecTemplate `json:",inline"`
	DeviceParameters                             `json:",inline"`
}

// +kubebuilder:object:root=true

// DeviceClass ...
type DeviceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// DeviceSpec.
	SpecTemplate DeviceParameters `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// DeviceClassList ...
type DeviceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceClass `json:"items"`
}
