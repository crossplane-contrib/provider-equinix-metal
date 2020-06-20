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
	"strings"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	computev1alpha1 "github.com/crossplaneio/crossplane/apis/compute/v1alpha1"
	"github.com/packethost/provider-packet/apis/server/v1alpha1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// A ClaimSchedulingController reconciles MachineInstance
// claims that include a class selector but omit their class and resource
// references by picking a random matching DeviceClass, if
// any.
type ClaimSchedulingController struct{}

// SetupWithManager sets up the
// ClaimSchedulingController using the supplied manager.
func (c *ClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		computev1alpha1.MachineInstanceKind,
		v1alpha1.DeviceKind,
		v1alpha1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.MachineInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
			resource.ClassKind(v1alpha1.DeviceClassGroupVersionKind),
		))
}

// A ClaimDefaultingController reconciles MachineInstance
// claims that omit their resource ref, class ref, and class selector by
// choosing a default DeviceClass if one exists.
type ClaimDefaultingController struct{}

// SetupWithManager sets up the
// ClaimDefaultingController using the supplied manager.
func (c *ClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		computev1alpha1.MachineInstanceKind,
		v1alpha1.DeviceKind,
		v1alpha1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha1.MachineInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
			resource.ClassKind(v1alpha1.DeviceClassGroupVersionKind),
		))
}

// A ClaimController reconciles MachineInstance claims with
// Devices, dynamically provisioning them if needed.
type ClaimController struct{}

// SetupWithManager adds a controller that reconciles MachineInstance resource claims.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		computev1alpha1.MachineInstanceKind,
		v1alpha1.DeviceKind,
		v1alpha1.Group))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha1.DeviceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha1.DeviceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha1.DeviceGroupVersionKind), mgr.GetScheme()),
	))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(computev1alpha1.MachineInstanceGroupVersionKind),
		resource.ClassKind(v1alpha1.DeviceClassGroupVersionKind),
		resource.ManagedKind(v1alpha1.DeviceGroupVersionKind),
		resource.WithManagedBinder(resource.NewAPIManagedStatusBinder(mgr.GetClient(), mgr.GetScheme())),
		resource.WithManagedFinalizer(resource.NewAPIManagedStatusUnbinder(mgr.GetClient())),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureDevice),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha1.Device{}}, &resource.EnqueueRequestForClaim{}).
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

	rc, csok := cs.(*v1alpha1.DeviceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha1.DeviceClassGroupVersionKind)
	}

	c, mgok := mg.(*v1alpha1.Device)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha1.DeviceGroupVersionKind)
	}

	spec := &v1alpha1.DeviceSpec{
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
