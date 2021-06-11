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

package vlan

import (
	"context"

	"github.com/packethost/packngo"
	"github.com/pkg/errors"

	"github.com/packethost/crossplane-provider-equinix-metal/apis/vlan/v1alpha1"
	"github.com/packethost/crossplane-provider-equinix-metal/pkg/clients"
)

const (
	errUnmarshalDate = "cannot unmarshal date"
)

// Client implements the Equinix Metal API methods needed to interact with VirtualNetworks for
// the Equinix Metal Crossplane Provider
type Client interface {
	List(projectID string, listOpt *packngo.ListOptions) (*packngo.VirtualNetworkListResponse, *packngo.Response, error)
	Create(*packngo.VirtualNetworkCreateRequest) (*packngo.VirtualNetwork, *packngo.Response, error)
	Get(vlanID string, getOpt *packngo.GetOptions) (*packngo.VirtualNetwork, *packngo.Response, error)
	Delete(virtualNetworkID string) (*packngo.Response, error)
}

// build-time test that the interface is implemented
var _ Client = (&packngo.Client{}).ProjectVirtualNetworks

// ClientWithDefaults is an interface that provides VirtualNetwork services and
// provides default values for common properties
type ClientWithDefaults interface {
	Client
	clients.DefaultGetter
}

// CredentialedClient is a credentialed client to Equinix Metal VirtualNetwork services
type CredentialedClient struct {
	Client
	*clients.Credentials
}

var _ ClientWithDefaults = &CredentialedClient{}

// NewClient returns a Client implementing the Equinix Metal API methods needed to
// interact with VirtualNetworks for the Equinix Metal Crossplane Provider
func NewClient(ctx context.Context, config *clients.Credentials) (ClientWithDefaults, error) {
	client, err := clients.NewClient(ctx, config)
	if err != nil {
		return nil, err
	}
	vlanClient := CredentialedClient{
		Client:      client.Client.ProjectVirtualNetworks,
		Credentials: client.Credentials,
	}
	vlanClient.SetProjectID(config.ProjectID)
	return vlanClient, nil
}

// CreateFromVirtualNetwork return packngo.VirtualNetworkCreateRequest created from Kubernetes
func CreateFromVirtualNetwork(d *v1alpha1.VirtualNetwork, projectID string) *packngo.VirtualNetworkCreateRequest {
	return &packngo.VirtualNetworkCreateRequest{
		Facility:    d.Spec.ForProvider.Facility,
		Metro:       d.Spec.ForProvider.Metro,
		Description: emptyIfNil(d.Spec.ForProvider.Description),
		ProjectID:   projectID,
	}
}

func emptyIfNil(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

// GenerateObservation produces v1alpha1.VirtualNetworkObservation from packngo.VirtualNetwork
func GenerateObservation(vlan *packngo.VirtualNetwork) (v1alpha1.VirtualNetworkObservation, error) {
	// Update vlan status
	observation := v1alpha1.VirtualNetworkObservation{
		ID:           vlan.ID,
		Href:         vlan.Href,
		VXLAN:        vlan.VXLAN,
		FacilityCode: vlan.FacilityCode,
	}

	if !observation.CreatedAt.IsZero() {
		if err := observation.CreatedAt.UnmarshalText([]byte(vlan.CreatedAt)); err != nil {
			return v1alpha1.VirtualNetworkObservation{}, errors.Wrap(err, errUnmarshalDate)
		}
	}

	return observation, nil
}

// LateInitialize fills the empty fields in *v1alpha2.VirtualNetworkParameters with the
// values seen in packngo.VirtualNetwork
func LateInitialize(in *v1alpha1.VirtualNetworkParameters, vlan *packngo.VirtualNetwork) {
	if vlan == nil {
		return
	}

	in.Description = clients.LateInitializeStringPtr(in.Description, &vlan.Description)
}

// IsUpToDate returns true if the supplied Kubernetes resource does not differ
// from the supplied Equinix Metal resource. It considers only fields that can be
// modified in place without deleting and recreating the instance, which are
// immutable.
func IsUpToDate(d *v1alpha1.VirtualNetwork, p *packngo.VirtualNetwork) bool {
	if !nilOrEqualStr(&d.Spec.ForProvider.Facility, p.FacilityCode) {
		return false
	}
	if !nilOrEqualStr(d.Spec.ForProvider.Description, p.Description) {
		return false
	}

	return true
}

// nilOrEqualStr is true if a (aPtr) is non-nil and equal to b
func nilOrEqualStr(aPtr *string, b string) bool {
	return (aPtr == nil || *aPtr == b)
}
