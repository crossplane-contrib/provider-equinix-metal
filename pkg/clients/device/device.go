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

	"github.com/hasheddan/stack-packet-demo/api/server/v1alpha1"
	"github.com/hasheddan/stack-packet-demo/pkg/clients"
	"github.com/packethost/packngo"
)

// NewClient ... TODO
func NewClient(ctx context.Context, credentials []byte) (packngo.DeviceService, error) {
	client := clients.NewClient(ctx, credentials)

	return client.Devices, nil
}

// CreateFromDevice return packngo.DeviceCreateRequest created from Kubernetes
func CreateFromDevice(d *v1alpha1.Device) *packngo.DeviceCreateRequest {
	return &packngo.DeviceCreateRequest{
		Hostname:     d.Spec.Hostname,
		Plan:         d.Spec.Plan,
		Facility:     []string{d.Spec.Facility},
		OS:           d.Spec.OS,
		BillingCycle: d.Spec.BillingCycle,
		ProjectID:    d.Spec.ProjectID,
		UserData:     d.Spec.UserData,
		Tags:         d.Spec.Tags,
	}
}

// NeedsUpdate returns true if the supplied Kubernetes resource differs from the
// supplied Packet resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func NeedsUpdate(d *v1alpha1.Device, p *packngo.Device) bool {
	if d.Spec.Hostname != p.Hostname {
		return true
	}
	if d.Spec.Locked != &p.Locked {
		return true
	}
	if d.Spec.UserData != p.UserData {
		return true
	}
	if d.Spec.IPXEScriptURL != p.IPXEScriptURL {
		return true
	}
	if d.Spec.AlwaysPXE != p.AlwaysPXE {
		return true
	}
	if !reflect.DeepEqual(d.Spec.Tags, p.Tags) {
		return true
	}

	return false
}

// NewUpdateDeviceRequest creates a request to update an instance suitable for
// use with the Packet API.
func NewUpdateDeviceRequest(d *v1alpha1.Device) *packngo.DeviceUpdateRequest {
	return &packngo.DeviceUpdateRequest{
		Hostname:      &d.Spec.Hostname,
		Locked:        d.Spec.Locked,
		UserData:      &d.Spec.UserData,
		IPXEScriptURL: &d.Spec.IPXEScriptURL,
		AlwaysPXE:     &d.Spec.AlwaysPXE,
		Tags:          &d.Spec.Tags,
	}
}
