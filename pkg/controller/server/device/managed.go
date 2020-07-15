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

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha2 "github.com/packethost/crossplane-provider-packet/apis/server/v1alpha2"
	packetv1alpha2 "github.com/packethost/crossplane-provider-packet/apis/v1alpha2"
	packetclient "github.com/packethost/crossplane-provider-packet/pkg/clients"
	devicesclient "github.com/packethost/crossplane-provider-packet/pkg/clients/device"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Error strings.
const (
	errManagedUpdateFailed = "cannot update Device custom resource"
	errProviderSecretNil   = "cannot find Secret reference on Provider"
	errGetProvider         = "cannot get Provider"
	errGetProviderSecret   = "cannot get Provider Secret"
	errGenObservation      = "cannot generate observation"
	errNewClient           = "cannot create new Device client"
	errNotDevice           = "managed resource is not a Device"
	errGetDevice           = "cannot get Device"
	errCreateDevice        = "cannot create Device"
	errUpdateDevice        = "cannot modify Device"
	errDeleteDevice        = "cannot delete Device"
)

// SetupDevice adds a controller that reconciles Devices
func SetupDevice(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha2.DeviceGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha2.DeviceGroupVersionKind),
		managed.WithExternalConnecter(&connecter{kube: mgr.GetClient()}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha2.Device{}).
		Complete(r)
}

type connecter struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, projectID string) (devicesclient.ClientWithDefaults, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	g, ok := mg.(*v1alpha2.Device)
	if !ok {
		return nil, errors.New(errNotDevice)
	}

	p := &packetv1alpha2.Provider{}
	n := meta.NamespacedNameOf(g.Spec.ProviderReference)
	if err := c.kube.Get(ctx, n, p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	newClientFn := devicesclient.NewClient
	if c.newClientFn != nil {
		newClientFn = c.newClientFn
	}
	client, err := newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.ProjectID)

	return &external{kube: c.kube, client: client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube   client.Client
	client devicesclient.ClientWithDefaults
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	d, ok := mg.(*v1alpha2.Device)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDevice)
	}

	// Observe device
	device, _, err := e.client.Get(meta.GetExternalName(d), nil)
	if packetclient.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDevice)
	}

	current := d.Spec.ForProvider.DeepCopy()
	devicesclient.LateInitialize(&d.Spec.ForProvider, device)
	if !cmp.Equal(current, &d.Spec.ForProvider) {
		if err := e.kube.Update(ctx, d); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	d.Status.AtProvider, err = devicesclient.GenerateObservation(device)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	// Set Device status and bindable
	switch d.Status.AtProvider.State {
	case v1alpha2.StateActive:
		d.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(d)
	case v1alpha2.StateProvisioning:
		d.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha2.StateQueued,
		v1alpha2.StateDeprovisioning,
		v1alpha2.StateFailed,
		v1alpha2.StateInactive,
		v1alpha2.StatePoweringOff,
		v1alpha2.StateReinstalling:
		d.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	upToDate, _ := devicesclient.IsUpToDate(d, device)

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: devicesclient.GetConnectionDetails(device),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	d, ok := mg.(*v1alpha2.Device)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDevice)
	}

	d.Status.SetConditions(runtimev1alpha1.Creating())

	create := devicesclient.CreateFromDevice(d, e.client.GetProjectID(packetclient.CredentialProjectID))
	device, _, err := e.client.Create(create)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDevice)
	}

	d.Status.AtProvider.ID = device.ID
	meta.SetExternalName(d, device.ID)
	if err := e.kube.Update(ctx, d); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errManagedUpdateFailed)
	}

	return managed.ExternalCreation{ConnectionDetails: devicesclient.GetConnectionDetails(device)}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	d, ok := mg.(*v1alpha2.Device)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDevice)
	}

	// NOTE(hasheddan): we must get the device again to see what type of update
	// we need to make
	device, _, err := e.client.Get(meta.GetExternalName(d), nil)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetDevice)
	}

	// NOTE(hasheddan): if the update is for the network type we return early
	// and do any updates on subsequent reconciles
	if _, n := devicesclient.IsUpToDate(d, device); n && d.Spec.ForProvider.NetworkType != nil {
		_, err := e.client.DeviceToNetworkType(meta.GetExternalName(d), *d.Spec.ForProvider.NetworkType)
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDevice)
	}
	_, _, err = e.client.Update(meta.GetExternalName(d), devicesclient.NewUpdateDeviceRequest(d))

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDevice)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	d, ok := mg.(*v1alpha2.Device)
	if !ok {
		return errors.New(errNotDevice)
	}
	d.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(meta.GetExternalName(d))
	return errors.Wrap(resource.Ignore(packetclient.IsNotFound, err), errDeleteDevice)
}
