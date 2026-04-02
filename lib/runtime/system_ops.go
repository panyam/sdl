package runtime

import (
	"github.com/panyam/sdl/lib/components"
)

// TraceAllPaths performs BFS path discovery from a component method in the system.
func TraceAllPaths(sys *SystemInstance, componentName, methodName string, maxDepth int) (*AllPathsTraceData, error) {
	compInst := sys.FindComponent(componentName)
	if compInst == nil {
		return nil, nil
	}

	pt := NewPathTraversal(sys.File.Runtime.Loader)
	return pt.TraceAllPaths(componentName, compInst.ComponentDecl, methodName, int32(maxDepth))
}

// ComponentUtilization holds utilization info for a single component.
type ComponentUtilization struct {
	Component string
	Infos     []components.UtilizationInfo
}

// GetSystemUtilization returns utilization info for all components in a system.
func GetSystemUtilization(sys *SystemInstance) []*ComponentUtilization {
	if sys == nil || sys.Env == nil {
		return nil
	}

	var results []*ComponentUtilization
	for varName, value := range sys.Env.All() {
		if varName == "self" {
			continue
		}
		if compInst, ok := value.Value.(*ComponentInstance); ok {
			infos := compInst.GetUtilizationInfo()
			if len(infos) > 0 {
				results = append(results, &ComponentUtilization{
					Component: varName,
					Infos:     infos,
				})
			}
		}
	}
	return results
}
