package fake

import (
	"github.com/packethost/packngo"

	"github.com/packethost/crossplane-provider-packet/pkg/clients/device"
)

var _ device.ClientWithDefaults = &MockClient{}

// MockClient is a fake implementation of packngo.Client.
type MockClient struct {
	MockCreate func(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error)
	MockUpdate func(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error)
	MockDelete func(deviceID string) (*packngo.Response, error)
	MockGet    func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error)

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
func (c *MockClient) Delete(deviceID string) (*packngo.Response, error) {
	return c.MockDelete(deviceID)
}

// Get calls the MockClient's MockGet function.
func (c *MockClient) Get(deviceID string, options *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
	return c.MockGet(deviceID, options)
}

// GetFacilityID calls the MockClient's MockGet function.
func (c *MockClient) GetFacilityID(id string) string {
	return c.MockGetFacilityID(id)
}

// GetProjectID calls the MockClient's MockGet function.
func (c *MockClient) GetProjectID(id string) string {
	return c.MockGetProjectID(id)
}
