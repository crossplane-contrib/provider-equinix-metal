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

package device

import (
	"context"
	"fmt"
	"reflect"

	"github.com/packethost/packngo"
	"github.com/pkg/errors"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	apiresource "k8s.io/apimachinery/pkg/api/resource"

	"github.com/packethost/crossplane-provider-equinix-metal/apis/server/v1alpha2"
	"github.com/packethost/crossplane-provider-equinix-metal/pkg/clients"
)

const (
	errUnmarshalDate = "cannot unmarshal date"
)

// Client implements the Equinix Metal API methods needed to interact with
// Devices for the Equinix Metal Crossplane Provider
type Client interface {
	Get(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error)
	Create(*packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error)
	Delete(deviceID string, force bool) (*packngo.Response, error)
	Update(string, *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error)
}

// PortsClient implements the Equinix Metal API methods needed to interact with
// Ports for the Equinix Metal Crossplane Provider
type PortsClient interface {
	DeviceToNetworkType(string, string) (*packngo.Device, error)
	DeviceNetworkType(string) (string, error)
}

// build-time test that the interface is implemented
var _ Client = (&packngo.Client{}).Devices
var _ PortsClient = (&packngo.Client{}).DevicePorts

// ClientWithDefaults is an interface that provides Device services and
// provides default values for common properties
type ClientWithDefaults interface {
	Client
	PortsClient
	clients.DefaultGetter
}

// CredentialedClient is a credentialed client to Equinix Metal Device services
type CredentialedClient struct {
	Client
	PortsClient
	*clients.Credentials
}

var _ ClientWithDefaults = &CredentialedClient{}

// NewClient returns a Client implementing the Equinix Metal API methods needed to
// interact with Devices for the Equinix Metal Crossplane Provider
func NewClient(ctx context.Context, credentials []byte, projectID string) (ClientWithDefaults, error) {
	client, err := clients.NewClient(ctx, credentials)
	if err != nil {
		return nil, err
	}
	deviceClient := CredentialedClient{
		Client:      client.Client.Devices,
		PortsClient: client.Client.DevicePorts,
		Credentials: client.Credentials,
	}
	deviceClient.SetProjectID(projectID)
	return deviceClient, nil
}

// CreateFromDevice return packngo.DeviceCreateRequest created from Kubernetes
func CreateFromDevice(d *v1alpha2.Device, projectID string) *packngo.DeviceCreateRequest {
	return &packngo.DeviceCreateRequest{
		Hostname:     emptyIfNil(d.Spec.ForProvider.Hostname),
		Plan:         d.Spec.ForProvider.Plan,
		Facility:     []string{d.Spec.ForProvider.Facility},
		OS:           d.Spec.ForProvider.OS,
		BillingCycle: emptyIfNil(d.Spec.ForProvider.BillingCycle),
		ProjectID:    projectID,
		UserData:     emptyIfNil(d.Spec.ForProvider.UserData),
		Tags:         d.Spec.ForProvider.Tags,
	}
}

func emptyIfNil(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

// GetConnectionDetails extracts managed.ConnectionDetails out of
// packngo.Device.
func GetConnectionDetails(device *packngo.Device) managed.ConnectionDetails {
	// RootPassword is only in the device responses for 24h
	// TODO(displague) Handle devices without public IPv4
	if device.RootPassword == "" || device.GetNetworkInfo().PublicIPv4 == "" {
		return managed.ConnectionDetails{}
	}

	// TODO(displague) device.User is in the API but not included in packngo
	user := "root"
	port := "22" // ssh

	return managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(device.GetNetworkInfo().PublicIPv4),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(user),
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(device.RootPassword),
		runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(port),
	}
}

// GenerateObservation produces v1alpha2.DeviceObservation from packngo.Device
func GenerateObservation(device *packngo.Device) (v1alpha2.DeviceObservation, error) {
	// Update device status
	observation := v1alpha2.DeviceObservation{
		ID:     device.ID,
		Href:   device.Href,
		State:  device.State,
		Locked: device.Locked,
		IPv4:   device.GetNetworkInfo().PublicIPv4,
	}

	if device.Facility != nil {
		observation.Facility = device.Facility.Code
	}

	// TODO: investigate better way to do this
	observation.ProvisionPercentage = apiresource.MustParse(fmt.Sprintf("%.6f", device.ProvisionPer))

	if !observation.CreatedAt.IsZero() {
		if err := observation.CreatedAt.UnmarshalText([]byte(device.Created)); err != nil {
			return v1alpha2.DeviceObservation{}, errors.Wrap(err, errUnmarshalDate)
		}
	}
	if !observation.UpdatedAt.IsZero() {
		if err := observation.UpdatedAt.UnmarshalText([]byte(device.Updated)); err != nil {
			return v1alpha2.DeviceObservation{}, errors.Wrap(err, errUnmarshalDate)
		}
	}

	return observation, nil
}

// LateInitialize fills the empty fields in *v1alpha2.DeviceParameters with the
// values seen in packngo.Device
func LateInitialize(in *v1alpha2.DeviceParameters, device *packngo.Device) {
	if device == nil {
		return
	}

	if device.OS != nil {
		in.OS = clients.LateInitializeString(in.OS, &device.OS.Slug)
	}

	if device.Plan != nil {
		in.Plan = clients.LateInitializeString(in.Plan, &device.Plan.Slug)
	}

	networkType := device.GetNetworkType()
	in.NetworkType = clients.LateInitializeStringPtr(in.NetworkType, &networkType)

	in.Hostname = clients.LateInitializeStringPtr(in.Hostname, &device.Hostname)
	in.BillingCycle = clients.LateInitializeStringPtr(in.BillingCycle, &device.BillingCycle)
	in.IPXEScriptURL = clients.LateInitializeStringPtr(in.IPXEScriptURL, &device.IPXEScriptURL)
	in.UserData = clients.LateInitializeStringPtr(in.UserData, &device.UserData)
	in.AlwaysPXE = clients.LateInitializeBoolPtr(in.AlwaysPXE, &device.AlwaysPXE)
	in.Locked = clients.LateInitializeBoolPtr(in.Locked, &device.Locked)

	for _, n := range device.Network {
		if n.Public && n.AddressFamily == 4 {
			in.PublicIPv4SubnetSize = clients.LateInitializeIntPtr(in.PublicIPv4SubnetSize, &n.CIDR)
		}
	}

	// Facility is required with a supported "any" value
	// AtProvider.Facility will reflect the Equinix Metal selected

	// TODO(displague) CustomData is string on input and a map when fetched
	// What's the format? Should it always be a map in k8s?
	// in.CustomData = device.CustomData

	// TODO(displague) Description is not yet supported
	// in.Description = device.Description

	if in.Tags == nil {
		in.Tags = device.Tags
	}
}

// IsUpToDate returns true if the supplied Kubernetes resource does not differ
// from the supplied Equinix Metal resource. It considers only fields that can be
// modified in place without deleting and recreating the instance, which are
// immutable.
func IsUpToDate(d *v1alpha2.Device, p *packngo.Device) (upToDate bool, networkTypeUpToDate bool) {
	networkType := p.GetNetworkType()
	networkIsUpToDate := nilOrEqualStr(d.Spec.ForProvider.NetworkType, networkType)

	if !nilOrEqualStr(d.Spec.ForProvider.Hostname, p.Hostname) {
		return false, networkIsUpToDate
	}
	if !nilOrEqualStr(d.Spec.ForProvider.UserData, p.UserData) {
		return false, networkIsUpToDate
	}
	if !nilOrEqualStr(d.Spec.ForProvider.IPXEScriptURL, p.IPXEScriptURL) {
		return false, networkIsUpToDate
	}

	if !nilOrEqualBool(d.Spec.ForProvider.Locked, p.Locked) {
		return false, networkIsUpToDate
	}

	if !nilOrEqualBool(d.Spec.ForProvider.AlwaysPXE, p.AlwaysPXE) {
		return false, networkIsUpToDate
	}

	// TODO(displague) CustomData is string vs map[string]interface{}
	/* TODO(displague) missing: https://github.com/packethost/packngo/pull/182
	if d.Spec.ForProvider.Description != p.Description {
		return false
	}
	*/

	if !reflect.DeepEqual(d.Spec.ForProvider.Tags, p.Tags) {
		return false, networkIsUpToDate
	}

	return true, networkIsUpToDate
}

// nilOrEqualStr is true if a (aPtr) is non-nil and equal to b
func nilOrEqualStr(aPtr *string, b string) bool {
	return (aPtr == nil || *aPtr == b)
}

// nilOrEqualBool is true if a (aPtr) is non-nil and equal to b
func nilOrEqualBool(aPtr *bool, b bool) bool {
	return (aPtr == nil || *aPtr == b)
}

// NewUpdateDeviceRequest creates a request to update an instance suitable for
// use with the Equinix Metal API.
func NewUpdateDeviceRequest(d *v1alpha2.Device) *packngo.DeviceUpdateRequest {
	return &packngo.DeviceUpdateRequest{
		Hostname:      d.Spec.ForProvider.Hostname,
		Locked:        d.Spec.ForProvider.Locked,
		UserData:      d.Spec.ForProvider.UserData,
		IPXEScriptURL: d.Spec.ForProvider.IPXEScriptURL,
		AlwaysPXE:     d.Spec.ForProvider.AlwaysPXE,
		Tags:          &d.Spec.ForProvider.Tags,
		Description:   d.Spec.ForProvider.Description,
		CustomData:    d.Spec.ForProvider.CustomData,
	}
}
