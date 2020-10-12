/*
Copyright 2020 The Crossplane Authors.

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

package fake

import (
	"github.com/packethost/packngo"

	"github.com/packethost/crossplane-provider-equinix-metal/pkg/clients/vlan"
)

var _ vlan.ClientWithDefaults = &MockClient{}

// MockClient is a fake implementation of packngo.Client.
type MockClient struct {
	MockList   func(projectID string, listOpt *packngo.ListOptions) (*packngo.VirtualNetworkListResponse, *packngo.Response, error)
	MockCreate func(createRequest *packngo.VirtualNetworkCreateRequest) (*packngo.VirtualNetwork, *packngo.Response, error)
	MockGet    func(vlanID string, getOpt *packngo.GetOptions) (*packngo.VirtualNetwork, *packngo.Response, error)
	MockDelete func(virtualNetworkID string) (*packngo.Response, error)

	MockGetProjectID  func(string) string
	MockGetFacilityID func(string) string
}

// List calls the MockClient's MockList function.
func (c *MockClient) List(projectID string, listOpt *packngo.ListOptions) (*packngo.VirtualNetworkListResponse, *packngo.Response, error) {
	return c.MockList(projectID, listOpt)
}

// Create calls the MockClient's MockCreate function.
func (c *MockClient) Create(createRequest *packngo.VirtualNetworkCreateRequest) (*packngo.VirtualNetwork, *packngo.Response, error) {
	return c.MockCreate(createRequest)
}

// Delete calls the MockClient's MockDelete function.
func (c *MockClient) Delete(virtualNetworkID string) (*packngo.Response, error) {
	return c.MockDelete(virtualNetworkID)
}

// Get calls the MockClient's MockGet function.
func (c *MockClient) Get(vlanID string, getOpt *packngo.GetOptions) (*packngo.VirtualNetwork, *packngo.Response, error) {
	return c.MockGet(vlanID, getOpt)
}

// GetFacilityID calls the MockClient's MockGet function.
func (c *MockClient) GetFacilityID(id string) string {
	return c.MockGetFacilityID(id)
}

// GetProjectID calls the MockClient's MockGet function.
func (c *MockClient) GetProjectID(id string) string {
	return c.MockGetProjectID(id)
}
