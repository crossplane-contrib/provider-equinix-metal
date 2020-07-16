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

	"github.com/packethost/crossplane-provider-packet/pkg/clients/ports"
)

var _ ports.ClientWithDefaults = &MockClient{}

// MockClient is a fake implementation of packngo.Client.
type MockClient struct {
	MockAssign        func(*packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error)
	MockUnassign      func(*packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error)
	MockGetPortByName func(string, string) (*packngo.Port, error)

	MockGetProjectID  func(string) string
	MockGetFacilityID func(string) string
}

// Assign calls the MockClient's MockAssign function.
func (c *MockClient) Assign(p *packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error) {
	return c.MockAssign(p)
}

// Unassign calls the MockClient's MockUnassign function.
func (c *MockClient) Unassign(p *packngo.PortAssignRequest) (*packngo.Port, *packngo.Response, error) {
	return c.MockUnassign(p)
}

// GetPortByName calls the MockClient's MockGetPortByName function.
func (c *MockClient) GetPortByName(deviceID string, name string) (*packngo.Port, error) {
	return c.MockGetPortByName(deviceID, name)
}

// GetFacilityID calls the MockClient's MockGet function.
func (c *MockClient) GetFacilityID(id string) string {
	return c.MockGetFacilityID(id)
}

// GetProjectID calls the MockClient's MockGet function.
func (c *MockClient) GetProjectID(id string) string {
	return c.MockGetProjectID(id)
}
