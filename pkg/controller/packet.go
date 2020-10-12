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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/packethost/crossplane-provider-equinix-metal/pkg/controller/ports/assignment"
	"github.com/packethost/crossplane-provider-equinix-metal/pkg/controller/server/device"
	"github.com/packethost/crossplane-provider-equinix-metal/pkg/controller/vlan/virtualnetwork"
)

// Setup creates all Equinix Metal controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		assignment.SetupAssignment,
		device.SetupDevice,
		virtualnetwork.SetupVirtualNetwork,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}

	controllers := []interface {
		SetupWithManager(ctrl.Manager) error
	}{}
	for _, c := range controllers {
		if err := c.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
