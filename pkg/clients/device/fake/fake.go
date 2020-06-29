package fake

import (
	"github.com/packethost/packngo"
)

var _ packngo.DeviceService = &MockClient{}

// MockClient is a fake implementation of packngo.Client.
type MockClient struct {
	MockCreate          func(createRequest *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error)
	MockUpdate          func(deviceID string, createRequest *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error)
	MockDelete          func(deviceID string) (*packngo.Response, error)
	MockGet             func(deviceID string, getOpt *packngo.GetOptions) (*packngo.Device, *packngo.Response, error)
	MockList            func(projectID string, listOpt *packngo.ListOptions) ([]packngo.Device, *packngo.Response, error)
	MockReboot          func(deviceID string) (*packngo.Response, error)
	MockPowerOff        func(deviceID string) (*packngo.Response, error)
	MockPowerOn         func(deviceID string) (*packngo.Response, error)
	MockLock            func(deviceID string) (*packngo.Response, error)
	MockUnlock          func(deviceID string) (*packngo.Response, error)
	MockListBGPSessions func(deviceID string, listOpt *packngo.ListOptions) ([]packngo.BGPSession, *packngo.Response, error)
	MockListEvents      func(deviceID string, listOpt *packngo.ListOptions) ([]packngo.Event, *packngo.Response, error)
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

// List calls the MockClient's MockList function.
func (c *MockClient) List(projectID string, options *packngo.ListOptions) ([]packngo.Device, *packngo.Response, error) {
	return c.MockList(projectID, options)
}

// Reboot calls the MockClient's MockReboot function.
func (c *MockClient) Reboot(deviceID string) (*packngo.Response, error) {
	return c.MockReboot(deviceID)
}

// PowerOff calls the MockClient's MockPowerOff function.
func (c *MockClient) PowerOff(deviceID string) (*packngo.Response, error) {
	return c.MockPowerOff(deviceID)
}

// PowerOn calls the MockClient's MockPowerOn function.
func (c *MockClient) PowerOn(deviceID string) (*packngo.Response, error) {
	return c.MockPowerOn(deviceID)
}

// Lock calls the MockClient's MockLock function.
func (c *MockClient) Lock(deviceID string) (*packngo.Response, error) {
	return c.MockLock(deviceID)
}

// Unlock calls the MockClient's MockUnlock function.
func (c *MockClient) Unlock(deviceID string) (*packngo.Response, error) {
	return c.MockUnlock(deviceID)
}

// ListBGPSessions calls the MockClient's MockListBGPSessions function.
func (c *MockClient) ListBGPSessions(deviceID string, options *packngo.ListOptions) ([]packngo.BGPSession, *packngo.Response, error) {
	return c.MockListBGPSessions(deviceID, options)
}

// ListEvents calls the MockClient's MockListEvents function.
func (c *MockClient) ListEvents(deviceID string, options *packngo.ListOptions) ([]packngo.Event, *packngo.Response, error) {
	return c.MockListEvents(deviceID, options)
}
