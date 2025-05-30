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
	Name   string
	// Holds the component instances and parameters
	Env *Env[Value]
}

// Initializes a new runtime System instance and its root environment
func NewSystemInstance(file *FileInstance, system *SystemDecl) *SystemInstance {
	sysinst := &SystemInstance{File: file, System: system}
	return sysinst
}

func (s *SystemInstance) Compile() (blockStmt *BlockStmt, err error) {
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
				TargetExpr: it.NameNode,
				Value:      &decl.NewExpr{ComponentExpr: &IdentifierExpr{Name: it.ComponentType.Name}},
			})

			/*
				comp, err := NewComponentInstance(s.File, compDecl)
				if err != nil {
					// let us come back to this later
					panic(err)
				}
				cv, err := NewValue(decl.ComponentType(compDecl), comp)
				if err != nil {
					panic(err)
				}
				s.Env.Set(it.NameNode.Name, cv)
			*/
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
					Receiver: it.NameNode,
					Member:   assign.Var,
				},
				Value: assign.Value,
			})
		}
	}
	return &BlockStmt{Statements: stmts}, nil
}
