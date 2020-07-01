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

	"github.com/packethost/packngo"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/packethost/crossplane-provider-packet/apis/server/v1alpha1"
	packetv1alpha1 "github.com/packethost/crossplane-provider-packet/apis/v1alpha1"
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
	errNewClient           = "cannot create new Device client"
	errNotDevice           = "managed resource is not a Device"
	errGetDevice           = "cannot get Device"
	errCreateDevice        = "cannot create Device"
	errUpdateDevice        = "cannot modify Device"
	errDeleteDevice        = "cannot delete Device"
)

// SetupDevice adds a controller that reconciles Devices
func SetupDevice(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DeviceGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DeviceGroupVersionKind),
		managed.WithExternalConnecter(&connecter{kube: mgr.GetClient()}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Device{}).
		Complete(r)
}

type connecter struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte) (packngo.DeviceService, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	g, ok := mg.(*v1alpha1.Device)
	if !ok {
		return nil, errors.New(errNotDevice)
	}

	p := &packetv1alpha1.Provider{}
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
	client, err := newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{kube: c.kube, client: client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube   client.Client
	client packngo.DeviceService
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	d, ok := mg.(*v1alpha1.Device)
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

	// Update device status
	d.Status.AtProvider.ID = device.ID
	d.Status.AtProvider.Hostname = device.Hostname
	d.Status.AtProvider.Href = device.Href
	d.Status.AtProvider.State = device.State

	for _, n := range device.Network {
		if n.Public && n.AddressFamily == 4 {
			d.Status.AtProvider.IPv4 = n.Address
		}
	}
	// TODO: investigate better way to do this
	d.Status.AtProvider.ProvisionPer = apiresource.MustParse(fmt.Sprintf("%.6f", device.ProvisionPer))

	// Set Device status and bindable
	// TODO: identify deleting state
	switch d.Status.AtProvider.State {
	case v1alpha1.StateActive:
		d.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(d)
	case v1alpha1.StateProvisioning:
		d.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.StateQueued:
		d.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  devicesclient.IsUpToDate(d, device),
		ConnectionDetails: managed.ConnectionDetails{},
	}

	// TODO: propagate secret info

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	d, ok := mg.(*v1alpha1.Device)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDevice)
	}

	d.Status.SetConditions(runtimev1alpha1.Creating())

	create := devicesclient.CreateFromDevice(d)
	device, _, err := e.client.Create(create)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDevice)
	}

	d.Status.AtProvider.ID = device.ID
	meta.SetExternalName(d, device.ID)
	if err := e.kube.Update(ctx, d); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errManagedUpdateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	d, ok := mg.(*v1alpha1.Device)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDevice)
	}

	_, _, err := e.client.Update(meta.GetExternalName(d), devicesclient.NewUpdateDeviceRequest(d))

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDevice)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	d, ok := mg.(*v1alpha1.Device)
	if !ok {
		return errors.New(errNotDevice)
	}
	d.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(meta.GetExternalName(d))
	return errors.Wrap(resource.Ignore(packetclient.IsNotFound, err), errDeleteDevice)
}
