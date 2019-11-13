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
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/packethost/packngo"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/packethost/stack-packet/apis/server/v1alpha1"
	packetv1alpha1 "github.com/packethost/stack-packet/apis/v1alpha1"
	"github.com/packethost/stack-packet/pkg/clients/device/fake"
	packettest "github.com/packethost/stack-packet/pkg/test"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
)

const (
	namespace  = "cool-namespace"
	deviceName = "my-cool-device"
	alwaysPXE  = true

	providerName       = "cool-packet"
	providerSecretName = "cool-packet-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyjson"

	connectionSecretName = "cool-connection-secret"
)

var (
	errorBoom = errors.New("boom")
)

type strange struct {
	resource.Managed
}

type deviceModifier func(*v1alpha1.Device)

func withConditions(c ...runtimev1alpha1.Condition) deviceModifier {
	return func(i *v1alpha1.Device) { i.Status.SetConditions(c...) }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) deviceModifier {
	return func(i *v1alpha1.Device) { i.Status.SetBindingPhase(p) }
}

func withProvisionPer(p float32) deviceModifier {
	return func(i *v1alpha1.Device) {
		i.Status.AtProvider.ProvisionPer = apiresource.MustParse(fmt.Sprintf("%.6f", p))
	}
}

func withState(s string) deviceModifier {
	return func(i *v1alpha1.Device) { i.Status.AtProvider.State = s }
}

func withID(d string) deviceModifier {
	return func(i *v1alpha1.Device) { i.Status.AtProvider.ID = d }
}

func device(im ...deviceModifier) *v1alpha1.Device {
	i := &v1alpha1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:       deviceName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: deviceName,
			},
		},
		Spec: v1alpha1.DeviceSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.DeviceParameters{
				AlwaysPXE: alwaysPXE,
			},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

var _ resource.ExternalClient = &external{}
var _ resource.ExternalConnecter = &connecter{}

func TestConnect(t *testing.T) {
	provider := packetv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: packetv1alpha1.ProviderSpec{
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      providerSecretName,
				},
				Key: providerSecretKey,
			},
		},
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}

	type strange struct {
		resource.Managed
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		conn resource.ExternalConnecter
		args args
		want want
	}{
		"Connected": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*packetv1alpha1.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newClientFn: func(_ context.Context, _ []byte) (packngo.DeviceService, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				err: nil,
			},
		},
		"NotDevice": {
			conn: &connecter{},
			args: args{ctx: context.Background(), mg: &strange{}},
			want: want{err: errors.New(errNotDevice)},
		},
		"FailedToGetProvider": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errorBoom
				}},
			},
			args: args{ctx: context.Background(), mg: device()},
			want: want{err: errors.Wrap(errorBoom, errGetProvider)},
		},
		"FailedToGetProviderSecret": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*packetv1alpha1.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return errorBoom
					}
					return nil
				}},
			},
			args: args{ctx: context.Background(), mg: device()},
			want: want{err: errors.Wrap(errorBoom, errGetProviderSecret)},
		},
		"FailedToCreateDevice": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*packetv1alpha1.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = secret
					}
					return nil
				}},
				newClientFn: func(_ context.Context, _ []byte) (packngo.DeviceService, error) { return nil, errorBoom },
			},
			args: args{ctx: context.Background(), mg: device()},
			want: want{err: errors.Wrap(errorBoom, errNewClient)},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.conn.Connect(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg          resource.Managed
		observation resource.ExternalObservation
		err         error
	}

	cases := map[string]struct {
		client resource.ExternalClient
		args   args
		want   want
	}{
		"ObservedDeviceAvailableNoUpdateNeeded": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{
						DeviceRaw: packngo.DeviceRaw{
							State:        v1alpha1.StateActive,
							ProvisionPer: float32(100),
							AlwaysPXE:    alwaysPXE,
						},
					}, nil, nil
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
					withProvisionPer(float32(100)),
					withState(v1alpha1.StateActive)),
				observation: resource.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"ObservedDeviceAvailableUpdateNeeded": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{
						DeviceRaw: packngo.DeviceRaw{
							State:        v1alpha1.StateActive,
							ProvisionPer: float32(100),
							AlwaysPXE:    !alwaysPXE,
						},
					}, nil, nil
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
					withProvisionPer(float32(100)),
					withState(v1alpha1.StateActive)),
				observation: resource.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"ObservedDeviceCreating": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{
						DeviceRaw: packngo.DeviceRaw{
							State:        v1alpha1.StateProvisioning,
							ProvisionPer: float32(50),
							AlwaysPXE:    alwaysPXE,
						},
					}, nil, nil
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(
					withConditions(runtimev1alpha1.Creating()),
					withProvisionPer(float32(50)),
					withState(v1alpha1.StateProvisioning)),
				observation: resource.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"ObservedDeviceQueued": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{
						DeviceRaw: packngo.DeviceRaw{
							State:        v1alpha1.StateQueued,
							ProvisionPer: float32(50),
							AlwaysPXE:    alwaysPXE,
						},
					}, nil, nil
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(
					withConditions(runtimev1alpha1.Unavailable()),
					withProvisionPer(float32(50)),
					withState(v1alpha1.StateQueued)),
				observation: resource.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"ObservedDeviceDoesNotExist": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return nil, nil, &packngo.ErrorResponse{
						Response: &http.Response{
							StatusCode: http.StatusNotFound,
						},
					}
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg:          device(),
				observation: resource.ExternalObservation{ResourceExists: false},
			},
		},
		"NotDevice": {
			client: &external{},
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotDevice),
			},
		},
		"FailedToGetDevice": {
			client: &external{client: &fake.MockClient{
				MockGet: func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
					return nil, nil, errorBoom
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg:  device(),
				err: errors.Wrap(errorBoom, errGetDevice),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.client.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.observation, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Observe(): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Observe(): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions(), packettest.EquateQuantities()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg       resource.Managed
		creation resource.ExternalCreation
		err      error
	}

	cases := map[string]struct {
		client resource.ExternalClient
		args   args
		want   want
	}{
		"CreatedInstance": {
			client: &external{client: &fake.MockClient{
				MockCreate: func(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{
						DeviceRaw: packngo.DeviceRaw{
							ID: deviceName,
						},
					}, nil, nil
				}},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(
					withConditions(runtimev1alpha1.Creating()),
					withID(deviceName)),
			},
		},
		"NotDevice": {
			client: &external{},
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotDevice),
			},
		},
		"FailedToCreateDevice": {
			client: &external{client: &fake.MockClient{
				MockCreate: func(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error) {
					return nil, nil, errorBoom
				},
			}},

			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg:  device(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errorBoom, errCreateDevice),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.client.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.creation, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Create(): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Create(): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions(), packettest.EquateQuantities()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg     resource.Managed
		update resource.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		client resource.ExternalClient
		args   args
		want   want
	}{
		"UpdatedInstance": {
			client: &external{client: &fake.MockClient{
				MockUpdate: func(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error) {
					return &packngo.Device{}, nil, nil
				},
			}},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(withConditions()),
			},
		},
		"NotCloudMemorystoreInstance": {
			client: &external{},
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotDevice),
			},
		},
		"FailedToUpdateInstance": {
			client: &external{client: &fake.MockClient{
				MockUpdate: func(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error) {
					return nil, nil, errorBoom
				},
			}},

			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg:  device(),
				err: errors.Wrap(errorBoom, errUpdateDevice),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.client.Update(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.update, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Update(): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Update(): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions(), packettest.EquateQuantities()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		client resource.ExternalClient
		args   args
		want   want
	}{
		"DeletedInstance": {
			client: &external{client: &fake.MockClient{
				MockDelete: func(deviceID string) (*packngo.Response, error) {
					return nil, nil
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg: device(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"NotCloudMemorystoreInstance": {
			client: &external{},
			args: args{
				ctx: context.Background(),
				mg:  &strange{},
			},
			want: want{
				mg:  &strange{},
				err: errors.New(errNotDevice),
			},
		},
		"FailedToDeleteInstance": {
			client: &external{client: &fake.MockClient{
				MockDelete: func(deviceID string) (*packngo.Response, error) {
					return nil, errorBoom
				},
			}},

			args: args{
				ctx: context.Background(),
				mg:  device(),
			},
			want: want{
				mg:  device(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errorBoom, errDeleteDevice),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.client.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.client.Delete(): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions(), packettest.EquateQuantities()); diff != "" {
				t.Errorf("resource.Managed: -want, +got:\n%s", diff)
			}
		})
	}
}
