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
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/packethost/crossplane-provider-equinix-metal/apis/server/v1alpha2"
	"github.com/packethost/crossplane-provider-equinix-metal/apis/vlan/v1alpha1"
)

// ResolveReferences of this Assignment
func (mg *Assignment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.deviceId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.DeviceID,
		Reference:    mg.Spec.ForProvider.DeviceIDRef,
		Selector:     mg.Spec.ForProvider.DeviceIDSelector,
		To:           reference.To{Managed: &v1alpha2.Device{}, List: &v1alpha2.DeviceList{}},
		Extract:      v1alpha2.DeviceID(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.DeviceID = rsp.ResolvedValue
	mg.Spec.ForProvider.DeviceIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.virtualNetworkId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.VirtualNetworkID,
		Reference:    mg.Spec.ForProvider.VirtualNetworkIDRef,
		Selector:     mg.Spec.ForProvider.VirtualNetworkIDSelector,
		To:           reference.To{Managed: &v1alpha1.VirtualNetwork{}, List: &v1alpha1.VirtualNetworkList{}},
		Extract:      v1alpha1.VirtualNetworkID(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.VirtualNetworkID = rsp.ResolvedValue
	mg.Spec.ForProvider.VirtualNetworkIDRef = rsp.ResolvedReference

	return nil
}
