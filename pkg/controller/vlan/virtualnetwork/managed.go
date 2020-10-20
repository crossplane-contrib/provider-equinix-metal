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

package virtualnetwork

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packetv1beta1 "github.com/packethost/crossplane-provider-equinix-metal/apis/v1beta1"
	"github.com/packethost/crossplane-provider-equinix-metal/apis/vlan/v1alpha1"
	packetclient "github.com/packethost/crossplane-provider-equinix-metal/pkg/clients"
	vlanclient "github.com/packethost/crossplane-provider-equinix-metal/pkg/clients/vlan"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Error strings.
const (
	errManagedUpdateFailed     = "cannot update VirtualNetwork custom resource"
	errProviderSecretNil       = "cannot find Secret reference on Provider"
	errGetProviderConfig       = "cannot get ProviderConfig"
	errTrackPCUsage            = "cannot track ProviderConfig usage"
	errGetProviderConfigSecret = "cannot get ProviderConfig Secret"
	errGenObservation          = "cannot generate observation"
	errNewClient               = "cannot create new VirtualNetwork client"
	errNotVirtualNetwork       = "managed resource is not a VirtualNetwork"
	errGetVirtualNetwork       = "cannot get VirtualNetwork"
	errCreateVirtualNetwork    = "cannot create VirtualNetwork"
	errDeleteVirtualNetwork    = "cannot delete VirtualNetwork"
)

// SetupVirtualNetwork adds a controller that reconciles VirtualNetworks
func SetupVirtualNetwork(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.VirtualNetworkGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.VirtualNetworkGroupVersionKind),
		managed.WithExternalConnecter(&connecter{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &packetv1beta1.ProviderConfigUsage{}),
		}),
		managed.WithConnectionPublishers(),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.VirtualNetwork{}).
		Complete(r)
}

type connecter struct {
	kube        client.Client
	usage       resource.Tracker
	newClientFn func(ctx context.Context, credentials []byte, projectID string) (vlanclient.ClientWithDefaults, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	g, ok := mg.(*v1alpha1.VirtualNetwork)
	if !ok {
		return nil, errors.New(errNotVirtualNetwork)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	p := &packetv1beta1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: g.Spec.ProviderConfigReference.Name}, p); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	ref := p.Spec.Credentials.SecretRef
	if ref == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfigSecret)
	}
	newClientFn := vlanclient.NewClient
	if c.newClientFn != nil {
		newClientFn = c.newClientFn
	}
	client, err := newClientFn(ctx, s.Data[ref.Key], p.Spec.ProjectID)

	return &external{kube: c.kube, client: client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube   client.Client
	client vlanclient.ClientWithDefaults
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha1.VirtualNetwork)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotVirtualNetwork)
	}

	// Observe virtual network
	device, _, err := e.client.Get(meta.GetExternalName(v), nil)
	if packetclient.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetVirtualNetwork)
	}

	current := v.Spec.ForProvider.DeepCopy()
	vlanclient.LateInitialize(&v.Spec.ForProvider, device)
	if !cmp.Equal(current, &v.Spec.ForProvider) {
		if err := e.kube.Update(ctx, v); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	v.Status.AtProvider, err = vlanclient.GenerateObservation(device)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	v.Status.SetConditions(runtimev1alpha1.Available())

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: vlanclient.IsUpToDate(v, device),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha1.VirtualNetwork)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotVirtualNetwork)
	}

	v.Status.SetConditions(runtimev1alpha1.Creating())

	create := vlanclient.CreateFromVirtualNetwork(v, e.client.GetProjectID(packetclient.CredentialProjectID))
	vlan, _, err := e.client.Create(create)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateVirtualNetwork)
	}

	v.Status.AtProvider.ID = vlan.ID
	meta.SetExternalName(v, vlan.ID)
	if err := e.kube.Update(ctx, v); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errManagedUpdateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// NOTE(hasheddan): VirtualNetwork cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha1.VirtualNetwork)
	if !ok {
		return errors.New(errNotVirtualNetwork)
	}
	v.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(packetclient.IsNotFound, err), errDeleteVirtualNetwork)
}
