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

	"github.com/packethost/crossplane-provider-equinix-metal/pkg/clients/device"
)

var _ device.ClientWithDefaults = &MockClient{}

// MockClient is a fake implementation of packngo.Client.
type MockClient struct {
	// mock the Client for Devices

	MockCreate func(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error)
	MockUpdate func(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error)
	MockDelete func(deviceID string, force bool) (*packngo.Response, error)
	MockGet    func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error)

	// mock the PortsClient

	MockDeviceToNetworkType func(deviceID string, networkType string) (*packngo.Device, error)
	MockDeviceNetworkType   func(deviceID string) (string, error)
	MockConvertDevice       func(*packngo.Device, string) error

	MockGetProjectID  func(string) string
	MockGetFacilityID func(string) string
}

// Create calls the MockClient's MockCreate function.
func (c *MockClient) Create(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error) {
	return c.MockCreate(createRequest)
}

// Update calls the MockClient's MockUpdate function.
func (c *MockClient) Update(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error) {
	return c.MockUpdate(deviceID, createRequest)
}

// Delete calls the MockClient's MockDelete function.
func (c *MockClient) Delete(deviceID string, force bool) (*packngo.Response, error) {
	return c.MockDelete(deviceID, false)
}

// Get calls the MockClient's MockGet function.
func (c *MockClient) Get(deviceID string, options *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
	return c.MockGet(deviceID, options)
}

// DeviceToNetworkType calls the MockClient's MockDeviceToNetworkType function.
func (c *MockClient) DeviceToNetworkType(deviceID string, networkType string) (*packngo.Device, error) {
	return c.MockDeviceToNetworkType(deviceID, networkType)
}

// DeviceNetworkType calls the MockClient's MockDeviceNetworkType function.
func (c *MockClient) DeviceNetworkType(deviceID string) (string, error) {
	return c.MockDeviceNetworkType(deviceID)
}

// GetFacilityID calls the MockClient's MockGet function.
func (c *MockClient) GetFacilityID(id string) string {
	return c.MockGetFacilityID(id)
}

// GetProjectID calls the MockClient's MockGet function.
func (c *MockClient) GetProjectID(id string) string {
	return c.MockGetProjectID(id)
}

// ConvertDevice calls the MockClient's MockConvertDevice function.
func (c *MockClient) ConvertDevice(d *packngo.Device, networkType string) error {
	return c.MockConvertDevice(d, networkType)
}
