package decl

import (
	"github.com/panyam/leetcoach/sdl/components"
)

type MockDisk struct {
	InstanceName string
	Profile      string
	ReadLatency  float64 // Example parameter
}

func NewMockDiskComponent(instanceName string) (ComponentRuntime, error) {
	disk := &MockDisk{
		Profile:     components.ProfileSSD, // Default
		ReadLatency: 0.0001,                // Default
	}
	return &NativeComponent{
		InstanceName: instanceName,
		GoInstance:   disk,
	}, nil
}

// Mock Service for testing 'uses'
type MockSvc struct {
	InstanceName string
	DB           any // Field to hold the injected dependency
	Timeout      int64
}

func NewMockSvcComponent(instanceName string) (ComponentRuntime, error) {
	svc := &MockSvc{
		Timeout: 1000, // Default
	}
	return &NativeComponent{
		InstanceName: instanceName,
		GoInstance:   svc,
	}, nil
}
