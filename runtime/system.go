package runtime

import (
	"log"
	"reflect"

	"github.com/panyam/sdl/decl"
)

type SystemInstance struct {
	// Where the system is defined
	File   *FileInstance
	System *SystemDecl

	// Holds the component instances and parameters
	Env *Env[Value]
}

// Initializes a new runtime System instance and its root environment
func NewSystemInstance(file *FileInstance, system *SystemDecl) *SystemInstance {
	sysinst := &SystemInstance{File: file, System: system}
	return sysinst
}

// A system declaration contains instantiations of components and other statemetns.
// Specifically in initializers it is important to not be bound by order.
// This method compiles the System into a set of statements that can be executed so that
// all components are intantiated first and then their properties/params are set.
func (s *SystemInstance) Initializer() (blockStmt *BlockStmt, err error) {
	// 1. a New Expression for constructing a componet
	// 2. a set expression for setting params/dependencies in a component - this way we can avoid the two pass approach?
	// This two passs approach needs to be repeated in every stage (eg type checking)
	var stmts []Stmt
	var instanceDecls []*InstanceDecl

	// Pass 1 - Create all the NewExprs and add the LetStmt to our list
	for _, item := range s.System.Body {
		switch it := item.(type) {
		case *InstanceDecl:
			// 1. Find the component definition in this file
			// 2. Instantiate it - it should be an expression evaluation (not overrides will be handled in pass 2)

			// if it is an instance declaration, find the component type from the environment
			// and create a new instance of the component type
			instanceDecls = append(instanceDecls, it)
			// stmts = append(stmts, &ExprStmt{ Expression: })
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: it.Name,
				Value:      &decl.NewExpr{ComponentExpr: &IdentifierExpr{Value: it.ComponentName.Value}},
			})
		case *LetStmt:
			// Add this as is
			stmts = append(stmts, it)
		default:
			log.Fatal("Invalid type: ", it, reflect.TypeOf(it))
			// i.Errorf(item.Pos(), "type inference for system body item type %T not implemented yet", item)
		}
	}

	// Pass 2 - Add SetExpr statements to enable the overrides
	for _, it := range instanceDecls {
		for _, assign := range it.Overrides {
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: &MemberAccessExpr{
					Receiver: it.Name,
					Member:   assign.Var,
				},
				Value: assign.Value,
			})
		}
	}
	return &BlockStmt{Statements: stmts}, nil
}

type InitStmt struct {
	From     *InitStmt
	Pos      Location
	Attrib   string
	CompInst *ComponentInstance // this should be From.CompInst.Attrib.  If From == nil then this is a System level component
}

// Goes through all components and gets uninitialized components so user knows what/how to set them
// This is usually called after the Initializer expression is called but before any other expressions are called.
func (s *SystemInstance) GetUninitializedComponents(env *Env[Value]) (items []*InitStmt) {
	var visit func(i *InitStmt)
	visit = func(i *InitStmt) {
		compDecl := i.CompInst.ComponentDecl
		deps, _ := compDecl.Dependencies()
		for _, dep := range deps {
			depInst, ok := i.CompInst.Get(dep.Name.Value)
			if !ok || depInst.IsNil() {
				items = append(items, &InitStmt{
					From:   i,
					Pos:    compDecl.Pos(),
					Attrib: dep.Name.Value,
				})
			} else {
				visit(&InitStmt{
					From:     i,
					Pos:      compDecl.Pos(),
					Attrib:   dep.Name.Value,
					CompInst: depInst.Value.(*ComponentInstance),
				})
			}
		}
	}

	for _, item := range s.System.Body {
		it, ok := item.(*InstanceDecl)
		if !ok {
			continue
		}

		compValue, ok := env.Get(it.Name.Value)
		if !ok {
			items = append(items, &InitStmt{
				Pos:    item.Pos(),
				Attrib: it.Name.Value,
			})
			continue
		}

		compInst := compValue.Value.(*ComponentInstance)
		visit(&InitStmt{
			Pos:      item.Pos(),
			Attrib:   it.Name.Value,
			CompInst: compInst,
		})
	}
	return
}
