package loader

import (
	"fmt"
	"strings"

	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/parser"
)

func ParseExpresssion(expr string) (decl.Expr, error) {
	// Expressions are not high level items so we need to wrap them in a component and a method to parse them and then
	// extract the value from them
	// Something like:
	//
	//  component TestComponet {
	//  	method TestMethod() {
	//  		let tempVar = {{ expression here }}
	//		}
	//  }

	wrapperString := fmt.Sprintf(`component TestComponent {
		method TestMethod() {
			let tempVar = %s
		}
	}`, expr)

	_, ast, err := parser.Parse(strings.NewReader(wrapperString))
	if err != nil {
		return nil, err
	}

	testComp, err := ast.GetDefinition("TestComponent")
	if err != nil {
		return nil, err
	}
	testMethod, err := testComp.(*decl.ComponentDecl).GetMethod("TestMethod")
	if err != nil {
		return nil, err
	}
	return testMethod.Body.Statements[0].(*LetStmt).Value, nil
}
