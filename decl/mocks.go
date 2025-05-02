package decl

import (
	"fmt"

	"github.com/panyam/leetcoach/sdl/components"
)

type MockDisk struct {
	InstanceName string
	Profile      string
	ReadLatency  float64 // Example parameter
}

func NewMockDiskComponent(instanceName string, params map[string]any) (any, error) {
	disk := &MockDisk{
		InstanceName: instanceName,
		Profile:      components.ProfileSSD, // Default
		ReadLatency:  0.0001,                // Default
	}
	if profileVal, ok := params["ProfileName"]; ok {
		if profileStr, okStr := profileVal.(string); okStr {
			disk.Profile = profileStr
		} else {
			return nil, fmt.Errorf("invalid type for 'ProfileName' override: expected string, got %T", profileVal)
		}
	}
	if latVal, ok := params["ReadLatency"]; ok {
		if latFloat, okFloat := latVal.(float64); okFloat {
			disk.ReadLatency = latFloat
		} else if latInt, okInt := latVal.(int64); okInt {
			// Allow int64 literals for duration/float params for convenience?
			disk.ReadLatency = float64(latInt)
		} else {
			return nil, fmt.Errorf("invalid type for 'ReadLatency' override: expected float64 or int64, got %T", latVal)
		}
	}
	// Add more param checks as needed
	return disk, nil
}

// Mock Service for testing 'uses'
type MockSvc struct {
	InstanceName string
	DB           any // Field to hold the injected dependency
	Timeout      int64
}

func NewMockSvcComponent(instanceName string, params map[string]any) (any, error) {
	svc := &MockSvc{
		InstanceName: instanceName,
		Timeout:      1000, // Default
	}
	if timeoutVal, ok := params["Timeout"]; ok {
		if toInt, okInt := timeoutVal.(int64); okInt {
			svc.Timeout = toInt
		} else {
			return nil, fmt.Errorf("invalid type for 'Timeout': %T", timeoutVal)
		}
	}
	return svc, nil
}
