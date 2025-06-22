package runtime

import (
	"fmt"
	"sync"

	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
)

// PathTraversal performs breadth-first traversal to discover all possible execution paths
// 
// Current limitations:
// - Control flow dependencies are not properly represented. All method calls within
//   a method are shown as sibling edges rather than showing sequential/conditional
//   relationships. For example, if a method calls A, then conditionally calls B or C,
//   this will show A, B, and C as three sibling edges rather than A followed by
//   a choice between B and C.
type PathTraversal struct {
	mu      sync.Mutex
	loader  *loader.Loader
	visited map[string]bool // Track visited (component.method) to prevent cycles
	edgeID  int64           // Counter for unique edge IDs
}

// AllPathsTraceData represents the complete execution tree (matches proto structure)
type AllPathsTraceData struct {
	TraceID string    `json:"trace_id"`
	Root    TraceNode `json:"root"`
}

// TraceNode represents a single node in the execution tree (matches proto structure)
type TraceNode struct {
	StartingTarget string      `json:"starting_target"`
	Edges          []Edge      `json:"edges"`
	Groups         []GroupInfo `json:"groups"`
}

// Edge represents a transition from one node to another (matches proto structure)
type Edge struct {
	ID            string    `json:"id"`
	NextNode      TraceNode `json:"next_node"`
	Label         string    `json:"label"`
	IsAsync       bool      `json:"is_async"`
	IsReverse     bool      `json:"is_reverse"`
	Probability   string    `json:"probability"`
	Condition     string    `json:"condition"`
	IsConditional bool      `json:"is_conditional"`
}

// GroupInfo allows flexible grouping of edges with labels (matches proto structure)
type GroupInfo struct {
	GroupStart int32  `json:"group_start"`
	GroupEnd   int32  `json:"group_end"`
	GroupLabel string `json:"group_label"`
	GroupType  string `json:"group_type"`
}

// NewPathTraversal creates a new path traversal engine
func NewPathTraversal(l *loader.Loader) *PathTraversal {
	return &PathTraversal{
		loader:  l,
		visited: make(map[string]bool),
		edgeID:  1,
	}
}

// TraceAllPaths performs breadth-first traversal starting from the given component.method
func (pt *PathTraversal) TraceAllPaths(currentCompName string, compDecl *ComponentDecl, methodName string, maxDepth int32) (*AllPathsTraceData, error) {
	// Check method exists
	methodDecl, err := compDecl.GetMethod(methodName)
	if err != nil || methodDecl == nil {
		return nil, fmt.Errorf("method '%s' not found in component '%s ( %s )'", methodName, currentCompName, compDecl.Name.Value)
	}

	// Reset state for new traversal
	pt.mu.Lock()
	pt.visited = make(map[string]bool)
	pt.edgeID = 1
	pt.mu.Unlock()

	// Build the root node by traversing all paths (here we use the type name instead of the attribute name)
	rootTarget := fmt.Sprintf("%s.%s", compDecl.Name.Value, methodName)
	rootNode, err := pt.buildTraceNode(currentCompName, compDecl, methodDecl, maxDepth, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to build trace tree: %w", err)
	}

	return &AllPathsTraceData{
		TraceID: fmt.Sprintf("%s_allpaths_%d", rootTarget, pt.edgeID),
		Root:    *rootNode,
	}, nil
}

// buildTraceNode recursively builds a trace node by analyzing the method body
func (pt *PathTraversal) buildTraceNode(currentCompName string, compDecl *decl.ComponentDecl, methodDecl *decl.MethodDecl, maxDepth int32, currentDepth int32) (*TraceNode, error) {
	if maxDepth > 0 && currentDepth >= maxDepth {
		return &TraceNode{
			StartingTarget: fmt.Sprintf("%s.%s", compDecl.Name.Value, methodDecl.Name.Value),
			Edges:          []Edge{},
			Groups:         []GroupInfo{},
		}, nil
	}

	target := fmt.Sprintf("%s.%s", compDecl.Name.Value, methodDecl.Name.Value)

	// Check for cycles - if we've already visited this target, don't recurse
	pt.mu.Lock()
	if pt.visited[target] {
		pt.mu.Unlock()
		return &TraceNode{
			StartingTarget: target,
			Edges:          []Edge{},
			Groups:         []GroupInfo{{GroupLabel: "cycle detected", GroupType: "cycle"}},
		}, nil
	}
	pt.visited[target] = true
	pt.mu.Unlock()

	node := &TraceNode{
		StartingTarget: target,
		Edges:          []Edge{},
		Groups:         []GroupInfo{},
	}

	// Analyze the method body to find all possible call paths
	if methodDecl.Body != nil {
		edges, groups, err := pt.analyzeStatement(currentCompName, compDecl, methodDecl.Body, maxDepth, currentDepth+1)
		if err != nil {
			return nil, err
		}
		node.Edges = edges
		node.Groups = groups
	}

	// Mark as no longer being processed (allow revisiting in different call stacks)
	pt.mu.Lock()
	delete(pt.visited, target)
	pt.mu.Unlock()

	return node, nil
}

// analyzeStatement analyzes a statement and returns edges and groups
func (pt *PathTraversal) analyzeStatement(currentCompName string, currentComp *decl.ComponentDecl, stmt decl.Stmt, maxDepth int32, currentDepth int32) ([]Edge, []GroupInfo, error) {
	var edges []Edge
	var groups []GroupInfo

	// fmt.Printf("DEBUG: Analyzing statement type: %T\n", stmt)
	switch s := stmt.(type) {
	case *decl.BlockStmt:
		// Analyze each statement in the block
		for _, childStmt := range s.Statements {
			childEdges, childGroups, err := pt.analyzeStatement(currentCompName, currentComp, childStmt, maxDepth, currentDepth)
			if err != nil {
				return nil, nil, err
			}
			edges = append(edges, childEdges...)
			groups = append(groups, childGroups...)
		}

	case *decl.ExprStmt:
		// Analyze the expression for method calls
		if exprEdges, err := pt.analyzeExpression(currentCompName, currentComp, s.Expression, maxDepth, currentDepth); err == nil {
			edges = append(edges, exprEdges...)
		}

	case *decl.IfStmt:
		// Analyze both branches of the if statement
		conditionStr := pt.expressionToString(s.Condition)
		groupStart := int32(len(edges))
		

		// Analyze then branch
		if s.Then != nil {
			thenEdges, thenGroups, err := pt.analyzeStatement(currentCompName, currentComp, s.Then, maxDepth, currentDepth)
			if err != nil {
				return nil, nil, err
			}
			// Mark conditional edges
			for i := range thenEdges {
				thenEdges[i].IsConditional = true
				thenEdges[i].Condition = conditionStr
			}
			edges = append(edges, thenEdges...)
			groups = append(groups, thenGroups...)
		}

		// Analyze else branch if present
		if s.Else != nil {
			elseEdges, elseGroups, err := pt.analyzeStatement(currentCompName, currentComp, s.Else, maxDepth, currentDepth)
			if err != nil {
				return nil, nil, err
			}
			// Mark conditional edges
			negCondition := fmt.Sprintf("!(%s)", conditionStr)
			for i := range elseEdges {
				elseEdges[i].IsConditional = true
				elseEdges[i].Condition = negCondition
			}
			edges = append(edges, elseEdges...)
			groups = append(groups, elseGroups...)
		}

		// Add group for the entire if statement
		if len(edges) > int(groupStart) {
			groups = append(groups, GroupInfo{
				GroupStart: groupStart,
				GroupEnd:   int32(len(edges) - 1),
				GroupLabel: fmt.Sprintf("if %s", conditionStr),
				GroupType:  "conditional",
			})
		}

	case *decl.ForStmt:
		// Analyze loop body
		groupStart := int32(len(edges))
		if s.Body != nil {
			bodyEdges, bodyGroups, err := pt.analyzeStatement(currentCompName, currentComp, s.Body, maxDepth, currentDepth)
			if err != nil {
				return nil, nil, err
			}
			edges = append(edges, bodyEdges...)
			groups = append(groups, bodyGroups...)
		}

		// Add group for the loop
		if len(edges) > int(groupStart) {
			loopCondition := "unknown loop"
			if s.Condition != nil {
				loopCondition = pt.expressionToString(s.Condition)
			}
			groups = append(groups, GroupInfo{
				GroupStart: groupStart,
				GroupEnd:   int32(len(edges) - 1),
				GroupLabel: fmt.Sprintf("for %s", loopCondition),
				GroupType:  "loop",
			})
		}

	case *decl.ReturnStmt:
		// Analyze the expression in the return statement
		if s.ReturnValue != nil {
			returnEdges, err := pt.analyzeExpression(currentCompName, currentComp, s.ReturnValue, maxDepth, currentDepth)
			if err != nil {
				return nil, nil, err
			}
			edges = append(edges, returnEdges...)
		}

	case *decl.LetStmt:
		// Analyze the value expression (right side of let statement)
		letEdges, err := pt.analyzeExpression(currentCompName, currentComp, s.Value, maxDepth, currentDepth)
		if err != nil {
			return nil, nil, err
		}
		edges = append(edges, letEdges...)
		// Debug: log what we found
		fmt.Printf("DEBUG: LetStmt found %d edges from expression: %T\n", len(letEdges), s.Value)

	case *decl.SetStmt:
		// Analyze the right side of assignment
		rightEdges, _ := pt.analyzeExpression(currentCompName, currentComp, s.Value, maxDepth, currentDepth)
		edges = append(edges, rightEdges...)

	case *decl.AssignmentStmt:
		// Analyze the right side of assignment
		rightEdges, _ := pt.analyzeExpression(currentCompName, currentComp, s.Value, maxDepth, currentDepth)
		edges = append(edges, rightEdges...)

	default:
		panic(fmt.Sprintf("Statement Type not supported: %T", s))
		// For other statement types, we may need to add support later
	}

	return edges, groups, nil
}

// analyzeExpression analyzes an expression for method calls and returns edges
func (pt *PathTraversal) analyzeExpression(currentCompName string, currentComp *decl.ComponentDecl, expr decl.Expr, maxDepth int32, currentDepth int32) ([]Edge, error) {
	var edges []Edge

	switch e := expr.(type) {
	case *decl.CallExpr:
		// This is a method call - create an edge
		edge, err := pt.createEdgeFromCall(currentCompName, currentComp, e, maxDepth, currentDepth)
		if err == nil && edge != nil {
			edges = append(edges, *edge)
		}

	case *decl.BinaryExpr:
		// Analyze both sides of binary expression
		leftEdges, _ := pt.analyzeExpression(currentCompName, currentComp, e.Left, maxDepth, currentDepth)
		rightEdges, _ := pt.analyzeExpression(currentCompName, currentComp, e.Right, maxDepth, currentDepth)
		edges = append(edges, leftEdges...)
		edges = append(edges, rightEdges...)

	case *decl.LiteralExpr:
		// these tyeps can be skipped as they wont have any Calls in them
		break

	default:
		// For other expression types, continue analyzing
		panic(fmt.Sprintf("Expression Type not supported: %T", e))
	}

	return edges, nil
}

// createEdgeFromCall creates an edge from a method call expression
func (pt *PathTraversal) createEdgeFromCall(currentCompName string, currentComp *decl.ComponentDecl, call *decl.CallExpr, maxDepth int32, currentDepth int32) (*Edge, error) {
	// Extract component and method from the call
	receiverName, methodName, err := pt.extractCallTarget(call)
	if err != nil {
		return nil, err
	}

	// Handle different types of calls
	var targetCompName string
	var targetCompDecl *decl.ComponentDecl
	var targetMethodDecl *decl.MethodDecl

	if receiverName == "self" {
		// Self call - use current component
		targetCompName = currentCompName
		targetCompDecl = currentComp
		targetMethodDecl, _ = targetCompDecl.GetMethod(methodName)
	} else if receiverName != "" {
		// Call to a dependency - find the component type from the uses declarations
		deps, _ := currentComp.Dependencies()
		for _, dep := range deps {
			if dep.Name.Value == receiverName {
				// Found the dependency, get its component type
				targetCompName = dep.Name.Value
				targetCompDecl = dep.ResolvedComponent
				if targetCompDecl == nil {
					panic("Shouldnt be here. Components should have been resolved")
				}
				targetMethodDecl, _ = targetCompDecl.GetMethod(methodName)
				break
			}
		}
	} else {
		// Direct function call (e.g., delay()) - skip these as they're not component calls
		return nil, nil
	}

	// If we couldn't find the component or method, create a simple edge
	if targetCompDecl == nil || targetMethodDecl == nil {
		return &Edge{
			ID:    pt.nextEdgeID(),
			Label: "calls",
			NextNode: TraceNode{
				StartingTarget: fmt.Sprintf("%s.%s", receiverName, methodName),
				Edges:          []Edge{},
				Groups:         []GroupInfo{},
			},
		}, nil
	}

	// Recursively build the next node
	nextNode, err := pt.buildTraceNode(targetCompName, targetCompDecl, targetMethodDecl, maxDepth, currentDepth)
	if err != nil {
		return nil, err
	}

	return &Edge{
		ID:       pt.nextEdgeID(),
		Label:    "calls",
		NextNode: *nextNode,
	}, nil
}

// extractCallTarget extracts component and method names from a call expression
func (pt *PathTraversal) extractCallTarget(call *decl.CallExpr) (string, string, error) {
	switch fn := call.Function.(type) {
	case *decl.MemberAccessExpr:
		// This is a method call like "component.method" or "self.component.method"
		// Handle nested member access like self.pool.Acquire()
		if memberAccess, ok := fn.Receiver.(*decl.MemberAccessExpr); ok {
			// This is like self.pool, where Receiver is self and Member is pool
			if selfIdent, ok := memberAccess.Receiver.(*decl.IdentifierExpr); ok && selfIdent.Value == "self" {
				// Pattern: self.component.method
				componentName := memberAccess.Member.Value
				methodName := fn.Member.Value
				return componentName, methodName, nil
			}
		}
		
		// Simple case: component.method
		if recv, ok := fn.Receiver.(*decl.IdentifierExpr); ok {
			method := fn.Member
			return recv.Value, method.Value, nil
		}
	case *decl.IdentifierExpr:
		// This is a direct function call
		return "", fn.Value, nil
	}

	return "", "", fmt.Errorf("unable to extract call target from expression")
}

// expressionToString converts an expression to a string representation
func (pt *PathTraversal) expressionToString(expr decl.Expr) string {
	if expr == nil {
		return "nil"
	}

	switch e := expr.(type) {
	case *decl.IdentifierExpr:
		return e.Value
	case *decl.BinaryExpr:
		left := pt.expressionToString(e.Left)
		right := pt.expressionToString(e.Right)
		return fmt.Sprintf("%s %s %s", left, e.Operator, right)
	case *decl.MemberAccessExpr:
		recv := pt.expressionToString(e.Receiver)
		member := pt.expressionToString(e.Member)
		return fmt.Sprintf("%s.%s", recv, member)
	default:
		return "unknown"
	}
}

// nextEdgeID generates the next unique edge ID
func (pt *PathTraversal) nextEdgeID() string {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	id := fmt.Sprintf("edge_%d", pt.edgeID)
	pt.edgeID++
	return id
}
