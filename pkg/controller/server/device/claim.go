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

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	computev1alpha1 "github.com/crossplane/crossplane/apis/compute/v1alpha1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/packethost/crossplane-provider-packet/apis/server/v1alpha2"
)

// SetupDeviceClaimScheduling adds a controller that reconciles Device claims
// that include a class selector but omit their class and resource references by
// picking a random matching DeviceClass, if any.
func SetupDeviceClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(computev1alpha1.MachineInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.MachineInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.DeviceClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupDeviceClaimDefaulting adds a controller that reconciles Device claims
// that omit their resource ref, class ref, and class selector by choosing a
// default DeviceClass if one exists
func SetupDeviceClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(computev1alpha1.MachineInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.MachineInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.DeviceClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupDeviceClaimBinding adds a controller that reconciles Device claims with Device, dynamically provisioning them if needed.
func SetupDeviceClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(computev1alpha1.MachineInstanceGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
		resource.ClassKind(v1alpha2.DeviceClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.DeviceGroupVersionKind),
		claimbinding.WithBinder(claimbinding.NewAPIBinder(mgr.GetClient(), mgr.GetScheme())),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureDevice),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames),
		),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.DeviceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.DeviceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.DeviceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.Device{}}, &resource.EnqueueRequestForClaim{}).
		For(&computev1alpha1.MachineInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureDevice configures the supplied resource (presumed
// to be a Device) using the supplied resource claim (presumed
// to be a RedisCluster) and resource class.
func ConfigureDevice(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	_, cmok := cm.(*computev1alpha1.MachineInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), computev1alpha1.MachineInstanceGroupVersionKind)
	}

	rc, csok := cs.(*v1alpha2.DeviceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.DeviceClassGroupVersionKind)
	}

	c, mgok := mg.(*v1alpha2.Device)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.DeviceGroupVersionKind)
	}

	spec := &v1alpha2.DeviceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rc.SpecTemplate.ForProvider,
	}

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rc.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rc.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rc.SpecTemplate.ReclaimPolicy

	c.Spec = *spec

	return nil
}
