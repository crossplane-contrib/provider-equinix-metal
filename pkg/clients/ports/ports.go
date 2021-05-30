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

package ports

import (
	"context"

	"github.com/packethost/packngo"

	"github.com/packethost/crossplane-provider-equinix-metal/pkg/clients"
)

// Client implements the Equinix Metal API methods needed to interact with Ports for
// the Equinix Metal Crossplane Provider
type Client interface {
	Assign(*packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error)
	Unassign(*packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error)
	GetPortByName(string, string) (*packngo.Port, error)
}

// build-time test that the interface is implemented
var _ Client = (&packngo.Client{}).DevicePorts //nolint:staticcheck

// ClientWithDefaults is an interface that provides Port services and
// provides default values for common properties
type ClientWithDefaults interface {
	Client
	clients.DefaultGetter
}

// CredentialedClient is a credentialed client to Equinix Metal Port services
type CredentialedClient struct {
	Client
	*clients.Credentials
}

var _ ClientWithDefaults = &CredentialedClient{}

// NewClient returns a Client implementing the Equinix Metal API methods needed to
// interact with Ports for the Equinix Metal Crossplane Provider
func NewClient(ctx context.Context, credentials []byte, projectID string) (ClientWithDefaults, error) {
	client, err := clients.NewClient(ctx, credentials)
	if err != nil {
		return nil, err
	}
	portsClient := CredentialedClient{
		Client:      client.Client.DevicePorts, //nolint:staticcheck
		Credentials: client.Credentials,
	}
	portsClient.SetProjectID(projectID)
	return portsClient, nil
}
