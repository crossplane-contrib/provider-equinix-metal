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

	"github.com/packethost/packngo"
	"github.com/packethost/provider-packet/apis/server/v1alpha1"
	"github.com/packethost/provider-packet/pkg/clients"
)

// NewClient ... TODO
func NewClient(ctx context.Context, credentials []byte) (packngo.DeviceService, error) {
	client := clients.NewClient(ctx, credentials)

	return client.Devices, nil
}

// CreateFromDevice return packngo.DeviceCreateRequest created from Kubernetes
func CreateFromDevice(d *v1alpha1.Device) *packngo.DeviceCreateRequest {
	return &packngo.DeviceCreateRequest{
		Hostname:     d.Spec.ForProvider.Hostname,
		Plan:         d.Spec.ForProvider.Plan,
		Facility:     []string{d.Spec.ForProvider.Facility},
		OS:           d.Spec.ForProvider.OS,
		BillingCycle: d.Spec.ForProvider.BillingCycle,
		ProjectID:    d.Spec.ForProvider.ProjectID,
		UserData:     d.Spec.ForProvider.UserData,
		Tags:         d.Spec.ForProvider.Tags,
	}
}

// IsUpToDate returns true if the supplied Kubernetes resource does not differ from the
// supplied Packet resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func IsUpToDate(d *v1alpha1.Device, p *packngo.Device) bool {
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
	if !reflect.DeepEqual(d.Spec.ForProvider.Tags, p.Tags) {
		return false
	}

	return true
}

// NewUpdateDeviceRequest creates a request to update an instance suitable for
// use with the Packet API.
func NewUpdateDeviceRequest(d *v1alpha1.Device) *packngo.DeviceUpdateRequest {
	return &packngo.DeviceUpdateRequest{
		Hostname:      &d.Spec.ForProvider.Hostname,
		Locked:        &d.Spec.ForProvider.Locked,
		UserData:      &d.Spec.ForProvider.UserData,
		IPXEScriptURL: &d.Spec.ForProvider.IPXEScriptURL,
		AlwaysPXE:     &d.Spec.ForProvider.AlwaysPXE,
		Tags:          &d.Spec.ForProvider.Tags,
	}
}
