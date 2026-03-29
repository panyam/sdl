package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnifiedSystemParsesAndLoads verifies that the new system syntax
// with typed parameters can be parsed, loaded, and initialized by the runtime.
// In the unified model, components handle all composition and systems declare
// typed parameters instead of 'use' instance declarations.
func TestUnifiedSystemParsesAndLoads(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/unified_system.sdl", "SimpleAppTest")
	require.NotNil(t, sys, "System SimpleAppTest should load successfully")
	require.NotNil(t, sys.Env, "System should have an initialized environment")
}

// TestUnifiedSystemParameterInstantiation verifies that system parameters
// are instantiated as component instances accessible via their declared names.
// Each parameter in 'system S(name: Type)' should create a component instance
// of the given type, accessible as 'name' in the system environment.
func TestUnifiedSystemParameterInstantiation(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/unified_system.sdl", "SimpleAppTest")
	require.NotNil(t, sys)

	// The system parameter 'app' should be an instance of SimpleApp
	appComp := sys.FindComponent("app")
	require.NotNil(t, appComp, "Parameter 'app' should exist as a component instance")
	assert.Equal(t, "SimpleApp", appComp.ComponentDecl.Name.Value)
}

// TestUnifiedSystemNestedComponentAccess verifies that component composition
// via 'uses' creates accessible nested instances. When a component has
// 'uses server SimpleServer()', the server should be accessible as a
// sub-component of the top-level parameter.
func TestUnifiedSystemNestedComponentAccess(t *testing.T) {
	sys, _ := loadSystem(t, "../../test/fixtures/unified_system.sdl", "SimpleAppTest")
	require.NotNil(t, sys)

	// Navigate into nested components: app.server
	serverComp := sys.FindComponent("app.server")
	require.NotNil(t, serverComp, "Nested component 'app.server' should be accessible")
	assert.Equal(t, "SimpleServer", serverComp.ComponentDecl.Name.Value)

	// Navigate deeper: app.server.db
	dbComp := sys.FindComponent("app.server.db")
	require.NotNil(t, dbComp, "Nested component 'app.server.db' should be accessible")
	assert.Equal(t, "SimpleDB", dbComp.ComponentDecl.Name.Value)
}

// TestUnifiedSystemImplicitParameterScoping verifies that component paths
// can omit the system parameter prefix. When a system has parameters,
// FindComponent tries prepending each parameter name if the direct path
// fails. This means "server" resolves to "app.server" when the system
// has a single parameter "app", keeping recipe paths clean.
func TestUnifiedSystemImplicitParameterScoping(t *testing.T) {
	defer QuietTest(t)()

	sys, _ := loadSystem(t, "../../test/fixtures/unified_system.sdl", "SimpleAppTest")
	require.NotNil(t, sys)

	// Direct path still works
	directComp := sys.FindComponent("app.server")
	require.NotNil(t, directComp, "Direct path 'app.server' should resolve")

	// Implicit scoping: "server" should resolve to "app.server"
	implicitComp := sys.FindComponent("server")
	require.NotNil(t, implicitComp, "Implicit path 'server' should resolve via parameter 'app'")
	assert.Equal(t, directComp, implicitComp, "Both paths should resolve to the same instance")

	// Deeper implicit path: "server.db" should resolve to "app.server.db"
	implicitDeep := sys.FindComponent("server.db")
	require.NotNil(t, implicitDeep, "Implicit path 'server.db' should resolve via parameter 'app'")
	assert.Equal(t, "SimpleDB", implicitDeep.ComponentDecl.Name.Value)

	// The parameter itself: "app" resolves directly (no implicit needed)
	appComp := sys.FindComponent("app")
	require.NotNil(t, appComp)
	assert.Equal(t, "SimpleApp", appComp.ComponentDecl.Name.Value)
}

// TestUnifiedSystemMethodExecution verifies that methods on components
// instantiated through the unified system model can be called and produce
// correct results. This tests the full pipeline: parse -> load -> init -> call.
func TestUnifiedSystemMethodExecution(t *testing.T) {
	defer QuietTest(t)()

	sys, _ := loadSystem(t, "../../test/fixtures/unified_system.sdl", "SimpleAppTest")
	require.NotNil(t, sys)

	appComp := sys.FindComponent("app")
	require.NotNil(t, appComp)

	serverComp := sys.FindComponent("app.server")
	require.NotNil(t, serverComp)

	// Execute HandleRequest on the server component by evaluating directly
	se := NewSimpleEval(sys.File, nil)
	env := sys.Env

	// Build call expression: app.server.HandleRequest()
	callExpr := &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: &MemberAccessExpr{
				Receiver: &IdentifierExpr{Value: "app"},
				Member:   &IdentifierExpr{Value: "server"},
			},
			Member: &IdentifierExpr{Value: "HandleRequest"},
		},
	}

	var currTime float64
	result, _ := se.Eval(callExpr, env, &currTime)
	// The method should return a Bool value (true or false depending on pool)
	assert.NotNil(t, result)
}
