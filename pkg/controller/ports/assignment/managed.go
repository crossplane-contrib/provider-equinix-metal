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

package assignment

import (
	"context"
	"strings"

	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/packethost/crossplane-provider-equinix-metal/apis/ports/v1alpha1"
	packetv1alpha2 "github.com/packethost/crossplane-provider-equinix-metal/apis/v1alpha2"
	packetclient "github.com/packethost/crossplane-provider-equinix-metal/pkg/clients"
	portsclient "github.com/packethost/crossplane-provider-equinix-metal/pkg/clients/ports"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Error strings.
const (
	errProviderSecretNil = "cannot find Secret reference on Provider"
	errGetProvider       = "cannot get Provider"
	errGetProviderSecret = "cannot get Provider Secret"
	errNewClient         = "cannot create new Assignment client"
	errNotAssignment     = "managed resource is not a Assignment"
	errGetPort           = "cannot get Port"
	errCreateAssignment  = "cannot create Assignment"
	errDeleteAssignment  = "cannot delete Assignment"
)

// SetupAssignment adds a controller that reconciles Assignments
func SetupAssignment(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.AssignmentGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.AssignmentGroupVersionKind),
		managed.WithExternalConnecter(&connecter{kube: mgr.GetClient()}),
		managed.WithInitializers(),
		managed.WithConnectionPublishers(),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Assignment{}).
		Complete(r)
}

type connecter struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, projectID string) (portsclient.ClientWithDefaults, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	g, ok := mg.(*v1alpha1.Assignment)
	if !ok {
		return nil, errors.New(errNotAssignment)
	}

	p := &packetv1alpha2.Provider{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: g.Spec.ProviderReference.Name}, p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	newClientFn := portsclient.NewClient
	if c.newClientFn != nil {
		newClientFn = c.newClientFn
	}
	client, err := newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.ProjectID)

	return &external{kube: c.kube, client: client}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube   client.Client
	client portsclient.ClientWithDefaults
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	a, ok := mg.(*v1alpha1.Assignment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAssignment)
	}

	// Observe port
	port, err := e.client.GetPortByName(a.Spec.ForProvider.DeviceID, a.Spec.ForProvider.Name)
	if packetclient.IsNotFound(err) {
		return managed.ExternalObservation{}, errors.New("port does not exist")
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPort)
	}

	o := managed.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}

	for _, net := range port.AttachedVirtualNetworks {
		if strings.TrimPrefix(net.Href, "/virtual-networks/") == a.Spec.ForProvider.VirtualNetworkID {
			a.Status.SetConditions(runtimev1alpha1.Available())
			o.ResourceExists = true
		}
	}

	meta.SetExternalName(a, port.ID)
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	a, ok := mg.(*v1alpha1.Assignment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAssignment)
	}
	a.Status.SetConditions(runtimev1alpha1.Creating())
	_, _, err := e.client.Assign(&packngo.PortAssignRequest{PortID: meta.GetExternalName(a), VirtualNetworkID: a.Spec.ForProvider.VirtualNetworkID})
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateAssignment)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// NOTE(hasheddan): Assignment cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	a, ok := mg.(*v1alpha1.Assignment)
	if !ok {
		return errors.New(errNotAssignment)
	}
	a.SetConditions(runtimev1alpha1.Deleting())
	_, _, err := e.client.Unassign(&packngo.PortAssignRequest{PortID: meta.GetExternalName(a), VirtualNetworkID: a.Spec.ForProvider.VirtualNetworkID})
	return errors.Wrap(resource.Ignore(packetclient.IsNotFound, err), errDeleteAssignment)
}
