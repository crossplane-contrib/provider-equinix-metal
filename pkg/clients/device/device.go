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

	"github.com/hasheddan/stack-packet-demo/api/server/v1alpha1"
	"github.com/hasheddan/stack-packet-demo/pkg/clients"
	"github.com/packethost/packngo"
)

// NewClient ... TODO
func NewClient(ctx context.Context, credentials []byte) (packngo.DeviceService, error) {
	client := clients.NewClient(ctx, credentials)

	return client.Devices, nil
}

// CreateFromDevice ... TODO
func CreateFromDevice(d *v1alpha1.Device) *packngo.DeviceCreateRequest {
	return &packngo.DeviceCreateRequest{
		HostName:     d.Spec.Hostname,
		Plan:         d.Spec.Plan,
		Facility:     d.Spec.Facility,
		OS:           d.Spec.OS,
		BillingCycle: d.Spec.BillingCycle,
		ProjectID:    d.Spec.ProjectID,
		UserData:     d.Spec.UserData,
		Tags:         d.Spec.Tags,
	}
}

// NeedsUpdate ... TODO
func NeedsUpdate(d *v1alpha1.Device, p *packngo.Device) bool {
	return false
}
