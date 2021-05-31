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

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/packethost/crossplane-provider-equinix-metal/apis/v1beta1"
	"github.com/packethost/crossplane-provider-equinix-metal/pkg/version"
)

// Client is a structure that embeds Credentials for the purposes of defaulting
// to those credential supplied values during Equinix Metal API usage. This
// allows for the Device resource to not require a ProjectID, for example, since
// the provider was configured with a ProjectID.
type Client struct {
	*Credentials

	Client *packngo.Client
}

// NewCredentialsFromJSON parses JSON bytes returning an Equinix Metal Credentials configuration
func NewCredentialsFromJSON(j []byte) (*Credentials, error) {
	config := &Credentials{}
	if err := json.Unmarshal(j, config); err != nil {
		return nil, err
	}
	return config, nil
}

// NewClient returns an Equinix Metal Client configured with credentials
func NewClient(ctx context.Context, config *Credentials) (*Client, error) {
	apiKey := config.GetAPIKey(CredentialAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("Invalid APIKey in credentials")
	}
	apiClient := packngo.NewClientWithAuth("crossplane", apiKey, nil)
	apiClient.UserAgent = fmt.Sprintf("crossplane-provider-equinix-metal/%s %s", version.Version, apiClient.UserAgent)

	client := &Client{
		Client:      apiClient,
		Credentials: config,
	}

	return client, nil
}

// GetAuthInfo returns the necessary authentication information that is
// necessary to use when the controller connects to Equinix Metal API in order
// to reconcile the managed resource.
func GetAuthInfo(ctx context.Context, c client.Client, mg resource.Managed) (credentials *Credentials, err error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	case mg.GetProviderReference() != nil:
		return nil, errors.New("providerRef is not supported (use providerConfigRef)")
	default:
		return nil, errors.New("no providerConfigRef given")
	}
}

// UseProviderConfig to return GCP authentication information.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (credentials *Credentials, err error) {
	pc := &v1beta1.ProviderConfig{}
	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, err
	}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, err
	}
	data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, c, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials")
	}
	config, err := NewCredentialsFromJSON(data)
	if err != nil {
		return nil, err
	}
	if pc.Spec.ProjectID != "" {
		config.SetProjectID(pc.Spec.ProjectID)
	}
	return config, err
}

// IsNotFound returns true if error is not found
func IsNotFound(err error) bool {
	if e, ok := err.(*packngo.ErrorResponse); ok && e.Response != nil {
		return e.Response.StatusCode == http.StatusNotFound
	}
	return false
}
