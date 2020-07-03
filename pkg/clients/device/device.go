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
	"reflect"

	"github.com/packethost/crossplane-provider-packet/apis/server/v1alpha2"
	"github.com/packethost/crossplane-provider-packet/pkg/clients"
	"github.com/packethost/packngo"
)

// Client implements the Packet API methods needed to interact with Devices for
// the Packet Crossplane Provider
type Client interface {
	Get(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error)
	Create(*packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error)
	Delete(deviceID string) (*packngo.Response, error)
	Update(string, *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error)
}

// build-time test that the interface is implemented
var _ Client = (&packngo.Client{}).Devices

type ClientWithDefaults interface {
	Client
	clients.DefaultGetter
}

type DeviceClient struct {
	Client
	*clients.Credentials
}

var _ ClientWithDefaults = &DeviceClient{}

// NewClient returns a Client implementing the Packet API methods needed to
// interact with Devices for the Packet Crossplane Provider
func NewClient(ctx context.Context, credentials []byte, projectID string) (ClientWithDefaults, error) {
	client, err := clients.NewClient(ctx, credentials)
	if err != nil {
		return nil, err
	}
	deviceClient := DeviceClient{
		Client:      client.Client.Devices,
		Credentials: client.Credentials,
	}
	deviceClient.SetProjectID(projectID)
	return deviceClient, nil
}

// CreateFromDevice return packngo.DeviceCreateRequest created from Kubernetes
func CreateFromDevice(d *v1alpha2.Device, projectID string) *packngo.DeviceCreateRequest {
	return &packngo.DeviceCreateRequest{
		Hostname:     d.Spec.ForProvider.Hostname,
		Plan:         d.Spec.ForProvider.Plan,
		Facility:     []string{d.Spec.ForProvider.Facility},
		OS:           d.Spec.ForProvider.OS,
		BillingCycle: d.Spec.ForProvider.BillingCycle,
		ProjectID:    projectID,
		UserData:     d.Spec.ForProvider.UserData,
		Tags:         d.Spec.ForProvider.Tags,
	}
}

// IsUpToDate returns true if the supplied Kubernetes resource does not differ from the
// supplied Packet resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func IsUpToDate(d *v1alpha2.Device, p *packngo.Device) bool {
	if d.Spec.ForProvider.Hostname != p.Hostname {
		return false
	}
	if d.Spec.ForProvider.Locked != p.Locked {
		return false
	}
	if d.Spec.ForProvider.UserData != p.UserData {
		return false
	}
	if d.Spec.ForProvider.IPXEScriptURL != p.IPXEScriptURL {
		return false
	}
	if d.Spec.ForProvider.AlwaysPXE != p.AlwaysPXE {
		return false
	}

	// TODO(displague) CustomData is string vs map[string]interface{}
	/* TODO(displague) missing: https://github.com/packethost/packngo/pull/182
	if d.Spec.ForProvider.Description != p.Description {
		return false
	}
	*/

	if !reflect.DeepEqual(d.Spec.ForProvider.Tags, p.Tags) {
		return false
	}

	return true
}

// NewUpdateDeviceRequest creates a request to update an instance suitable for
// use with the Packet API.
func NewUpdateDeviceRequest(d *v1alpha2.Device) *packngo.DeviceUpdateRequest {
	return &packngo.DeviceUpdateRequest{
		Hostname:      &d.Spec.ForProvider.Hostname,
		Locked:        &d.Spec.ForProvider.Locked,
		UserData:      &d.Spec.ForProvider.UserData,
		IPXEScriptURL: &d.Spec.ForProvider.IPXEScriptURL,
		AlwaysPXE:     &d.Spec.ForProvider.AlwaysPXE,
		Tags:          &d.Spec.ForProvider.Tags,
		Description:   &d.Spec.ForProvider.Description,
		CustomData:    &d.Spec.ForProvider.CustomData,
	}
}
