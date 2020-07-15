/*
Copyright 2020 The Crossplane Authors.

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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualNetworkSpec defines the desired state of VirtualNetwork
type VirtualNetworkSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  VirtualNetworkParameters `json:"forProvider"`
}

// VirtualNetworkStatus defines the observed state of VirtualNetwork
type VirtualNetworkStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     VirtualNetworkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualNetwork is a managed resource that represents a Packet VirtualNetwork
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="HOSTNAME",type="string",JSONPath=".spec.forProvider.hostname"
// +kubebuilder:printcolumn:name="FACILITY",type="string",JSONPath=".status.atProvider.facility"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,packet}
type VirtualNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualNetworkSpec   `json:"spec,omitempty"`
	Status VirtualNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualNetworkList contains a list of VirtualNetworks
type VirtualNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualNetwork `json:"items"`
}

// VirtualNetworkParameters define the desired state of a Packet Virtual Network.
// https://www.packet.com/developers/api/vlans/#create-an-virtual-network
//
// Reference values are used for optional parameters to determine if
// LateInitialization should update the parameter after creation.
type VirtualNetworkParameters struct {
	// +immutable
	// +required
	Facility string `json:"facility"`

	// +optional
	Description *string `json:"description,omitempty"`
}

// VirtualNetworkObservation is used to reflect in the Kubernetes API, the observed
// state of the VirtualNetwork resource from the Packet API.
type VirtualNetworkObservation struct {
	ID           string       `json:"id"`
	Href         string       `json:"href,omitempty"`
	VXLAN        int          `json:"vxlan,omitempty"`
	FacilityCode string       `json:"facility_code,omitempty"`
	CreatedAt    *metav1.Time `json:"createdAt,omitempty"`
}
