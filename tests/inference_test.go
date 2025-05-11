package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/panyam/leetcoach/sdl/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Parser - in a real scenario, replace with your actual parser.
// For testing, we'll manually construct FileDecl for some tests if needed,
// or rely on a very simple string->FileDecl converter if you have one.
// For now, let's assume Parse(input string) (*FileDecl, error) exists.
// If your parser isn't ready, you might need to construct ASTs manually for some tests.

// Helper to find a specific expression node after parsing (example)
// You'll need more sophisticated finders based on your test needs.
func findExprInAnalyze(file *FileDecl, sysName, analyzeName string) (Expr, error) {
	if file == nil {
		return nil, fmt.Errorf("file is nil")
	}
	sys, err := file.GetSystem(sysName)
	if err != nil {
		return nil, fmt.Errorf("system %s not found: %w", sysName, err)
	}
	if sys == nil {
		return nil, fmt.Errorf("system %s is nil", sysName)
	}
	for _, item := range sys.Body {
		if ad, ok := item.(*AnalyzeDecl); ok && ad.Name.Name == analyzeName {
			return ad.Target, nil
		}
	}
	return nil, fmt.Errorf("analyze block %s not found in system %s", analyzeName, sysName)
}

func findExprInMethodReturn(file *FileDecl, compName, methodName string) (Expr, error) {
	// Simplified: assumes last statement is ExprStmt with the return value
	// or a ReturnStmt.
	comp, err := file.GetComponent(compName)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		return nil, fmt.Errorf("component %s not found", compName)
	}
	meth, err := comp.GetMethod(methodName)
	if err != nil {
		return nil, err
	}
	if meth == nil {
		return nil, fmt.Errorf("method %s not found in %s", methodName, compName)
	}
	if meth.Body == nil || len(meth.Body.Statements) == 0 {
		return nil, fmt.Errorf("method %s.%s has no body or statements", compName, methodName)
	}
	lastStmt := meth.Body.Statements[len(meth.Body.Statements)-1]
	if es, ok := lastStmt.(*ExprStmt); ok {
		return es.Expression, nil
	}
	if rs, ok := lastStmt.(*ReturnStmt); ok {
		return rs.ReturnValue, nil
	}
	return nil, fmt.Errorf("last statement in %s.%s is not an ExprStmt or ReturnStmt", compName, methodName)
}

func assertInferredType(t *testing.T, expr Expr, expectedType *Type, testName string) {
	t.Helper()
	base, err := getExprBase(expr)
	require.NoError(t, err, "[%s] Failed to get ExprBase", testName)
	require.NotNil(t, base.InferredType, "[%s] InferredType is nil for expr: %s", testName, expr.String())
	assert.True(t, expectedType.Equals(base.InferredType),
		"[%s] Type mismatch for expr '%s'. Expected: %s, Got: %s",
		testName, expr.String(), expectedType.String(), base.InferredType.String())
}

func assertErrorContains(t *testing.T, errors []error, substring string, testName string) {
	t.Helper()
	found := false
	var allErrors []string
	for _, err := range errors {
		allErrors = append(allErrors, err.Error())
		if strings.Contains(err.Error(), substring) {
			found = true
			break
		}
	}
	assert.True(t, found, "[%s] Expected error containing '%s', but not found. All errors: \n%s", testName, substring, strings.Join(allErrors, "\n"))
}

// Main test function structure
func runTypeInferenceTest(t *testing.T, dsl string, testName string, checks func(t *testing.T, file *FileDecl, errors []error)) {
	t.Run(testName, func(t *testing.T) {
		// This is where you'd call your actual parser
		// For now, we simulate or assume it.
		// file, pErr := Parse(dsl)
		// For initial testing without a live parser, you might need to construct
		// FileDecl instances manually or use a simplified test parser.
		// Let's assume for now `Parse` exists and works as described.
		_, file, pErr := parser.Parse(strings.NewReader(dsl))

		if pErr != nil {
			// If parser fails, the test setup is problematic or DSL is invalid for parser
			t.Fatalf("[%s] DSL parsing failed: %v\nDSL:\n%s", testName, pErr, dsl)
			return
		}
		require.NotNil(t, file, "[%s] Parser returned nil FileDecl", testName)

		errs := InferTypesForFile(file)
		checks(t, file, errs)
	})
}

// --- Test Cases ---

func TestInferLiterals(t *testing.T) {
	dsl := `
        system TestSys {
            analyze IntLit = 123;
            analyze FloatLit = 45.6;
            analyze StringLit = "hello";
            analyze BoolLit = true;
        }
    `
	runTypeInferenceTest(t, dsl, "Literals", func(t *testing.T, file *FileDecl, errors []error) {
		assert.Empty(t, errors, "Should be no errors for valid literals")

		expr, _ := findExprInAnalyze(file, "TestSys", "IntLit")
		assertInferredType(t, expr, IntType, "IntLiteral")

		expr, _ = findExprInAnalyze(file, "TestSys", "FloatLit")
		assertInferredType(t, expr, FloatType, "FloatLiteral")

		expr, _ = findExprInAnalyze(file, "TestSys", "StringLit")
		assertInferredType(t, expr, StrType, "StringLiteral")

		expr, _ = findExprInAnalyze(file, "TestSys", "BoolLit")
		assertInferredType(t, expr, BoolType, "BoolLiteral")
	})
}

func TestInferIdentifiers(t *testing.T) {
	dsl := `
        enum Color { RED, GREEN, BLUE }

        component MyComp {
            param CompParam: int = 0;
            param CompFloatParam: float = 1.0;

            method Process(methParam: string) : bool {
                let localVar: Color = Color.RED;
                // localVar; // Access local var
                // self.CompParam; // Access component param via self
                // methParam; // Access method param
                return methParam == "ok";
            }
        }

        system TestSys {
            instance c1: MyComp = { CompParam = 100 };
            let sysVar: float = 2.5;

            analyze TestLocal = c1.Process("test"); // This will trigger method inference
            analyze TestCompParamAccess = c1.CompParam;
            analyze TestSelfCompParam = { // Need to create a synthetic method to test self.
                let mc = c1; // Assume c1 is in scope
                // mc.CompParam; // This would be handled by method call or direct member access
                // For self, it must be inside a method. Process() already covers it.
                // To test 'self' explicitly, we'd need to find an expr inside Process:
                // For example, if Process had "let x = self.CompParam;" we'd find "self.CompParam"
				1; // Placeholder
            };
            analyze TestSysVar = sysVar;
            analyze TestInstance = c1;
            analyze TestEnumRef = Color.GREEN;
            analyze TestUnresolved = unresolvedVar;
        }
    `
	runTypeInferenceTest(t, dsl, "Identifiers", func(t *testing.T, file *FileDecl, errors []error) {
		// We expect one error for "unresolvedVar"
		assert.Len(t, errors, 1, "Expected 1 error for unresolvedVar")
		assertErrorContains(t, errors, "identifier 'unresolvedVar' not found", "UnresolvedIdentifier")

		// TestSysVar
		expr, _ := findExprInAnalyze(file, "TestSys", "TestSysVar")
		assertInferredType(t, expr, FloatType, "SystemLetVariable")

		// TestInstance
		expr, _ = findExprInAnalyze(file, "TestSys", "TestInstance")
		assertInferredType(t, expr, &Type{Name: "MyComp"}, "InstanceIdentifier")

		// TestEnumRef (Color.GREEN)
		expr, _ = findExprInAnalyze(file, "TestSys", "TestEnumRef")
		assertInferredType(t, expr, &Type{Name: "Color", IsEnum: true}, "EnumMemberAccess")

		// TestCompParamAccess (c1.CompParam) - member access
		expr, _ = findExprInAnalyze(file, "TestSys", "TestCompParamAccess")
		assertInferredType(t, expr, IntType, "ComponentInstanceParamAccess")

		// To test self.CompParam, methParam, localVar, we need to inspect expressions *inside* MyComp.Process
		// This requires a helper to find expressions within a method body.
		// Assuming findExprInMethodReturn or a similar helper
		// For `let localVar: Color = Color.RED;` inside Process method:
		comp, _ := file.GetComponent("MyComp")
		meth, _ := comp.GetMethod("Process")
		letStmt := meth.Body.Statements[0].(*LetStmt)
		localVarIdent := letStmt.Variables[0] // 'localVar'
		assertInferredType(t, localVarIdent, &Type{Name: "Color", IsEnum: true}, "LetLocalVar")
		assertInferredType(t, letStmt.Value, &Type{Name: "Color", IsEnum: true}, "LetLocalVarValue(Color.RED)")

		// For `self.CompParam` if it was `let x = self.CompParam;`
		// Let's modify Process method for a direct test:
		// method Process(methParam: string) : bool {
		//    let accessSelf = self.CompParam;
		//    ...
		// }
		// Then find `self.CompParam` as the value of `accessSelf`
		// For now, `c1.Process("test")` return is tested via CallExpr test.
	})
}

func TestInferBinaryExpressions(t *testing.T) {
	dsl := `
        system TestSys {
            analyze AddInt = 1 + 2;
            analyze AddFloat = 1.0 + 2.5;
            analyze AddMix = 1 + 2.0; // int + float -> float
            analyze SubInt = 5 - 3;
            analyze MulFloat = 2.0 * 3.0;
            analyze DivInt = 10 / 2; // int / int -> int (or float if lang rule)
                                    // Current inferencer: int/int -> int
            analyze ModInt = 10 % 3;
            analyze ConcatStr = "a" + "b";

            analyze EqNum = 1 == 1.0; // numeric comparison
            analyze NeqBool = true != false;
            analyze LtStr = "a" < "b";
            analyze AndOp = true && false;
            analyze OrOp = true || (1 > 0);

            analyze ErrAddStrInt = "a" + 1;
            analyze ErrModFloat = 10.0 % 3.0;
            analyze ErrCompareStrBool = "a" == true;
            analyze ErrLogicInt = 1 && 2;
        }
    `
	runTypeInferenceTest(t, dsl, "BinaryExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		expectedErrorCount := 4
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors for binary ops", expectedErrorCount))
		assertErrorContains(t, errors, "cannot apply to String and Int", "ErrAddStrInt")
		assertErrorContains(t, errors, "requires two integers, got Float and Float", "ErrModFloat")
		assertErrorContains(t, errors, "cannot compare String and Bool", "ErrCompareStrBool")
		assertErrorContains(t, errors, "requires two booleans, got Int and Int", "ErrLogicInt")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "AddInt")
		assertInferredType(t, expr, IntType, "AddInt")
		expr, _ = findExprInAnalyze(file, "TestSys", "AddFloat")
		assertInferredType(t, expr, FloatType, "AddFloat")
		expr, _ = findExprInAnalyze(file, "TestSys", "AddMix")
		assertInferredType(t, expr, FloatType, "AddMix (promotion)")
		expr, _ = findExprInAnalyze(file, "TestSys", "SubInt")
		assertInferredType(t, expr, IntType, "SubInt")
		expr, _ = findExprInAnalyze(file, "TestSys", "MulFloat")
		assertInferredType(t, expr, FloatType, "MulFloat")
		expr, _ = findExprInAnalyze(file, "TestSys", "DivInt")
		assertInferredType(t, expr, IntType, "DivInt") // Assumes int division
		expr, _ = findExprInAnalyze(file, "TestSys", "ModInt")
		assertInferredType(t, expr, IntType, "ModInt")
		expr, _ = findExprInAnalyze(file, "TestSys", "ConcatStr")
		assertInferredType(t, expr, StrType, "ConcatStr")

		expr, _ = findExprInAnalyze(file, "TestSys", "EqNum")
		assertInferredType(t, expr, BoolType, "EqNum")
		expr, _ = findExprInAnalyze(file, "TestSys", "NeqBool")
		assertInferredType(t, expr, BoolType, "NeqBool")
		expr, _ = findExprInAnalyze(file, "TestSys", "LtStr")
		assertInferredType(t, expr, BoolType, "LtStr")
		expr, _ = findExprInAnalyze(file, "TestSys", "AndOp")
		assertInferredType(t, expr, BoolType, "AndOp")
		expr, _ = findExprInAnalyze(file, "TestSys", "OrOp")
		assertInferredType(t, expr, BoolType, "OrOp")
	})
}

func TestInferUnaryExpressions(t *testing.T) {
	dsl := `
        system TestSys {
            analyze NotBool = !true;
            analyze NegInt = -5;
            analyze NegFloat = -5.0;

            analyze ErrNotInt = !1;
            analyze ErrNegStr = -"abc";
        }
    `
	runTypeInferenceTest(t, dsl, "UnaryExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		expectedErrorCount := 2
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors for unary ops", expectedErrorCount))
		assertErrorContains(t, errors, "requires boolean, got Int", "ErrNotInt")
		assertErrorContains(t, errors, "requires integer or float, got String", "ErrNegStr")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "NotBool")
		assertInferredType(t, expr, BoolType, "NotBool")
		expr, _ = findExprInAnalyze(file, "TestSys", "NegInt")
		assertInferredType(t, expr, IntType, "NegInt")
		expr, _ = findExprInAnalyze(file, "TestSys", "NegFloat")
		assertInferredType(t, expr, FloatType, "NegFloat")
	})
}

func TestInferMemberAccessExpressions(t *testing.T) {
	dsl := `
        enum Status { ACTIVE, INACTIVE }
        component DataStore {
            param Name: string = "DefaultDB";
            param Size: int = 1024;
        }
        component Service {
            param Port: int = 8080;
            uses Store: DataStore;
            method GetStatus() : Status {
                return Status.ACTIVE;
            }
            method GetStoreName() : string {
                return self.Store.Name;
            }
        }
        system TestSys {
            instance ds: DataStore = { Name = "MyDS" };
            instance svc: Service = { Store = ds };
            analyze TestEnumVal = Status.INACTIVE;
            analyze TestCompParam = ds.Name;
            analyze TestCompParamViaSelf = svc.GetStoreName(); // Indirectly tests self.Store.Name
            // analyze TestLenOnList = myList.Len; // Assuming myList: List[int] is defined
            // analyze TestLenOnOutcomes = myOutcomes.Len; // Assuming myOutcomes: Outcomes[bool] is defined
            analyze TestMetricAccess = analyzeBlock.availability; // Simulate metric access

            analyze ErrNoSuchMember = ds.NonExistent;
            analyze ErrAccessOnInt = (1+2).NonExistent; // Access on non-component/enum
        }
    `
	runTypeInferenceTest(t, dsl, "MemberAccessExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		// Analyze block "TestMetricAccess" has "analyzeBlock" which is not defined.
		// This will lead to "analyzeBlock" being an unresolved identifier.
		// The test `ds.NonExistent` will also error.
		// The test `(1+2).NonExistent` will also error.
		expectedErrorCount := 3 // analyzeBlock unresolved, ds.NonExistent, (1+2).NonExistent
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors", expectedErrorCount))
		assertErrorContains(t, errors, "identifier 'analyzeBlock' not found", "UnresolvedMetricSource")
		assertErrorContains(t, errors, "member 'NonExistent' not found in component 'DataStore'", "ErrNoSuchMember")
		assertErrorContains(t, errors, "cannot access member 'NonExistent' on type Int", "ErrAccessOnInt")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "TestEnumVal")
		assertInferredType(t, expr, &Type{Name: "Status", IsEnum: true}, "TestEnumVal")

		expr, _ = findExprInAnalyze(file, "TestSys", "TestCompParam")
		assertInferredType(t, expr, StrType, "TestCompParam (ds.Name)")

		// For TestCompParamViaSelf, the analyze block's target is a CallExpr.
		// The call itself returns string.
		expr, _ = findExprInAnalyze(file, "TestSys", "TestCompParamViaSelf")
		assertInferredType(t, expr, StrType, "TestCompParamViaSelf (svc.GetStoreName() call)")
	})
}

func TestInferCallExpressions(t *testing.T) {
	dsl := `
        component Greeter {
            param Greeting: string = "Hello";
            method SayHello(name: string) : string {
                return self.Greeting + ", " + name + "!";
            }
            method DoNothing() { // void return, implies NilType
            }
            method Add(a: int, b: int) : int {
                return a + b;
            }
            method PromoteArg(val: float) : float {
                return val + 1.0;
            }
        }
        system TestSys {
            instance g: Greeter;
            analyze CallSayHello = g.SayHello("World");
            analyze CallDoNothing = g.DoNothing();
            analyze CallAdd = g.Add(10, 20);
            analyze CallPromote = g.PromoteArg(5); // int literal 5 promoted to float arg

            analyze ErrArgCount = g.Add(10);
            analyze ErrArgType = g.SayHello(123);
            // analyze ErrCallOnInt = (1).SayHello("No"); // Parser might reject (1).Method
            analyze ErrNonMethod = g.NonExistentMethod();
        }
    `
	runTypeInferenceTest(t, dsl, "CallExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		expectedErrorCount := 3
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors", expectedErrorCount))
		assertErrorContains(t, errors, "argument count mismatch for call to 'g.Add'", "ErrArgCount")
		assertErrorContains(t, errors, "type mismatch for argument 0 of call to 'g.SayHello': expected String, got Int", "ErrArgType")
		assertErrorContains(t, errors, "method 'NonExistentMethod' not found in component 'Greeter'", "ErrNonMethod")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "CallSayHello")
		assertInferredType(t, expr, StrType, "CallSayHello")
		expr, _ = findExprInAnalyze(file, "TestSys", "CallDoNothing")
		assertInferredType(t, expr, NilType, "CallDoNothing (void return)")
		expr, _ = findExprInAnalyze(file, "TestSys", "CallAdd")
		assertInferredType(t, expr, IntType, "CallAdd")
		expr, _ = findExprInAnalyze(file, "TestSys", "CallPromote")
		assertInferredType(t, expr, FloatType, "CallPromote (arg promotion)")
	})
}

func TestInferTupleAndChainedExpressions(t *testing.T) {
	dsl := `
        system TestSys {
            analyze MyTuple = (1, "a", true);
            analyze ChainArith = 1 + 2 * 3 - 4; // Assuming standard precedence, result is int
                                             // Our inferChainedExprType is left-associative, so (1+2)*3-4
                                             // (1+2)*3-4 = 3*3-4 = 9-4 = 5 (int)
            analyze ChainLogic = true && false || true; // (true && false) || true = false || true = true (bool)
            analyze ChainCompare = 1 < 2 && "a" == "a"; // (1 < 2) && ("a" == "a") = true && true = true (bool)

            analyze ErrChainType = 1 + "a" - true;
        }
    `
	runTypeInferenceTest(t, dsl, "TupleAndChainedExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		assert.Len(t, errors, 1, "Expected 1 error")
		assertErrorContains(t, errors, "cannot apply to Int and String", "ErrChainType")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "MyTuple")
		assertInferredType(t, expr, TupleType(IntType, StrType, BoolType), "MyTuple")

		expr, _ = findExprInAnalyze(file, "TestSys", "ChainArith")
		assertInferredType(t, expr, IntType, "ChainArith")

		expr, _ = findExprInAnalyze(file, "TestSys", "ChainLogic")
		assertInferredType(t, expr, BoolType, "ChainLogic")

		expr, _ = findExprInAnalyze(file, "TestSys", "ChainCompare")
		assertInferredType(t, expr, BoolType, "ChainCompare")
	})
}

func TestInferDistributeAndSampleExpressions(t *testing.T) {
	dsl := `
        system TestSys {
            analyze DistGood = distribute {
                0.5 => 10, // int
                0.5 => 20  // int
            };
            analyze DistWithDefault = distribute {
                0.3 => "hi",
                default => "def"
            };
            analyze SampleGood = sample DistGood; // Samples from Outcomes[int], so int

            analyze ErrDistTypeMismatch = distribute {
                0.5 => 100,     // int
                0.5 => "hello"  // string
            };
            analyze ErrDistCondNotNumeric = distribute {
                true => 1, // condition not numeric
                false => 2
            };
            analyze ErrSampleNotOutcomes = sample (1 + 2); // sample from int
        }
    `
	runTypeInferenceTest(t, dsl, "DistributeAndSampleExpressions", func(t *testing.T, file *FileDecl, errors []error) {
		expectedErrorCount := 3
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors", expectedErrorCount))
		assertErrorContains(t, errors, "type mismatch in distribute expr cases: expected Int (from case 0), got String for case 1", "ErrDistTypeMismatch")
		assertErrorContains(t, errors, "condition of distribute case 0 at pos .+ must be numeric (for weight), got Bool", "ErrDistCondNotNumeric")
		assertErrorContains(t, errors, "'from' expression must be Outcomes[T], got Int", "ErrSampleNotOutcomes")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "DistGood")
		assertInferredType(t, expr, OutcomesType(IntType), "DistGood")

		expr, _ = findExprInAnalyze(file, "TestSys", "DistWithDefault")
		assertInferredType(t, expr, OutcomesType(StrType), "DistWithDefault")

		expr, _ = findExprInAnalyze(file, "TestSys", "SampleGood")
		assertInferredType(t, expr, IntType, "SampleGood")
	})
}

func TestInferStatementsAndScopes(t *testing.T) {
	dsl := `
        component Scoper {
            param Rate: float = 1.0;
            method Calculate(input: int) : float {
                let intermediate = input * 2; // intermediate is int
                let factor: float = self.Rate;  // factor is float
                if intermediate > 10 {
                    let bonus = 5.0; // bonus is float, scoped to if-block
                    return (intermediate + factor) * bonus; // (int+float)*float -> float
                } else {
                    let penalty = 0.5; // penalty is float, scoped to else-block
                    // return "wrong"; // Error: type mismatch with method return
                    return (intermediate - factor) * penalty; // (int-float)*float -> float
                }
            }
            method CheckReturnPromotion() : float {
                return 10; // int, should promote to float for return
            }
            method BadReturn() : int {
                return 10.5; // float, cannot implicitly demote to int
            }
        }
        system TestSys {
            instance sc: Scoper = { Rate = 1.5 };
            analyze Result1 = sc.Calculate(10);
            analyze Result2 = sc.Calculate(3);
            analyze ResultPromote = sc.CheckReturnPromotion();
            analyze ResultBadReturn = sc.BadReturn();

            // Instance override type checking
            instance goodOverride: Scoper = { Rate = 2.0 }; // float = float (ok)
            instance goodPromoteOverride: Scoper = { Rate = 2 }; // float = int (ok, promote)
            // instance badOverride: Scoper = { Rate = "fast" }; // Error: float = string
        }
    `
	runTypeInferenceTest(t, dsl, "StatementsAndScopes", func(t *testing.T, file *FileDecl, errors []error) {
		// Error 1: BadReturn method, returning float for int.
		// Error 2 (if uncommented): badOverride, Rate = "fast"
		expectedErrorCount := 1 // Update if badOverride is uncommented
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors", expectedErrorCount))
		assertErrorContains(t, errors, "return type mismatch for method 'BadReturn'.*expected Int, got Float", "BadReturnMismatch")
		// if expectedErrorCount > 1 {
		//    assertErrorContains(t, errors, "type mismatch for override param 'Rate' in instance 'badOverride'.*expected Float, got String", "BadInstanceOverride")
		// }

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "Result1")
		assertInferredType(t, expr, FloatType, "sc.Calculate(10)")
		expr, _ = findExprInAnalyze(file, "TestSys", "Result2")
		assertInferredType(t, expr, FloatType, "sc.Calculate(3)")
		expr, _ = findExprInAnalyze(file, "TestSys", "ResultPromote")
		assertInferredType(t, expr, FloatType, "sc.CheckReturnPromotion (return promotion)")

		// Check types inside Scoper.Calculate method:
		comp, _ := file.GetComponent("Scoper")
		meth, _ := comp.GetMethod("Calculate")

		// let intermediate = input * 2;
		letIntermediateStmt := meth.Body.Statements[0].(*LetStmt)
		intermediateIdent := letIntermediateStmt.Variables[0]
		assertInferredType(t, intermediateIdent, IntType, "let intermediate (ident)")
		assertInferredType(t, letIntermediateStmt.Value, IntType, "let intermediate (value: input*2)")

		// let factor: float = self.Rate;
		letFactorStmt := meth.Body.Statements[1].(*LetStmt)
		factorIdent := letFactorStmt.Variables[0]
		assertInferredType(t, factorIdent, FloatType, "let factor (ident)") // Should use declared type
		assertInferredType(t, letFactorStmt.Value, FloatType, "let factor (value: self.Rate)")

		// if intermediate > 10
		ifStmt := meth.Body.Statements[2].(*IfStmt)
		assertInferredType(t, ifStmt.Condition, BoolType, "if condition (intermediate > 10)")

		// Inside 'then' block: let bonus = 5.0;
		thenBlock := ifStmt.Then
		letBonusStmt := thenBlock.Statements[0].(*LetStmt)
		bonusIdent := letBonusStmt.Variables[0]
		assertInferredType(t, bonusIdent, FloatType, "let bonus (ident)")
	})
}

func TestInferSystemItems(t *testing.T) {
	dsl := `
        component Producer {
            param DataRate: int = 100;
            method Produce() : string { return "data"; }
        }
        component Consumer {
            uses Source: Producer;
            param BufferSize: int = 64;
            method Consume() : int {
                // let x = self.Source.Produce(); // x is string
                return self.BufferSize;
            }
        }
        system DataFlow {
            instance prod: Producer = { DataRate = 200 };
            instance cons: Consumer = {
                Source = prod;          // Correct: Producer = Producer
                BufferSize = 128;
            };
            // instance badCons: Consumer = { Source = cons }; // Error: Producer = Consumer

            let sysLevelVar: string = prod.Produce();

            analyze ResultData = cons.Consume();
            analyze CheckSysVar = sysLevelVar;

            analyze MetricsProvider = prod.Produce(); // Target for expect
            // Expect block
            expect MetricsProvider {       // Analyze block name is 'MetricsProvider'
                                           // Its type is 'string' from prod.Produce()
                // MetricsProvider.P99 < 100;  // Error: P99 not on string
            }
        }
    `
	runTypeInferenceTest(t, dsl, "SystemItems", func(t *testing.T, file *FileDecl, errors []error) {
		// Error 1 (if badCons uncommented): Type mismatch for 'Source' dependency
		// Error 2 (if MetricsProvider.P99 uncommented): Accessing P99 on string.
		// Error 3: MetricsProvider is the name of the *analyze block*, not the *value*.
		// The expression in expect is implicitly "MetricsProvider.<metric>".
		// Current InferExprType will try to resolve "MetricsProvider" as an identifier in expectScope,
		// which should be the type of `prod.Produce()` (string).
		// Then it will try MemberAccess on string.
		// So only one error from the expect block for P99 on string.

		// Let's make the expect test valid for type inference, even if semantically flawed for metrics:
		// expect MetricsProvider { MetricsProvider.Len == 4; } // Assuming .Len on string
		// Current `inferMemberAccessExprType` might not have .Len for basic strings.
		// Let's assume the expect block is empty or has a valid check based on existing rules.
		// For now, with the P99 error, we expect 1 error if badCons is commented.
		expectedErrorCount := 0 // Change if uncommenting errors in DSL
		// If MetricsProvider.P99 is active:
		// expectedErrorCount = 1
		// assertErrorContains(t, errors, "cannot access member 'P99' on type String", "ExpectMetricError")

		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors", expectedErrorCount))

		var expr Expr
		expr, _ = findExprInAnalyze(file, "DataFlow", "ResultData")
		assertInferredType(t, expr, IntType, "cons.Consume()")

		expr, _ = findExprInAnalyze(file, "DataFlow", "CheckSysVar")
		assertInferredType(t, expr, StrType, "sysLevelVar")

		// Check instance types in scope
		sysDecl, _ := file.GetSystem("DataFlow")
		require.NotNil(t, sysDecl)
		// Manually create a scope to test
		rootScope := NewRootTypeScope(file)
		systemScope := rootScope.Push(nil, nil)
		// Populate scope as inferTypesForSystemDeclBodyItem would
		for _, item := range sysDecl.Body {
			inferTypesForSystemDeclBodyItem(item, systemScope) // This will populate systemScope
		}

		prodType, ok := systemScope.Get("prod")
		require.True(t, ok, "'prod' not in system scope")
		assert.True(t, (&Type{Name: "Producer"}).Equals(prodType), "Type of 'prod' instance")

		consType, ok := systemScope.Get("cons")
		require.True(t, ok, "'cons' not in system scope")
		assert.True(t, (&Type{Name: "Consumer"}).Equals(consType), "Type of 'cons' instance")

		// Check analyze block's identifier type
		analyzeTargetExpr, _ := findExprInAnalyze(file, "DataFlow", "MetricsProvider")
		base, _ := getExprBase(analyzeTargetExpr)
		analyzeBlockIdentType := base.InferredType // Type of prod.Produce()

		sys, _ := file.GetSystem("DataFlow")
		var analyzeDeclNode *AnalyzeDecl
		for _, item := range sys.Body {
			if ad, ok := item.(*AnalyzeDecl); ok && ad.Name.Name == "MetricsProvider" {
				analyzeDeclNode = ad
				break
			}
		}
		require.NotNil(t, analyzeDeclNode)
		baseIdent, _ := getExprBase(analyzeDeclNode.Name)
		assert.True(t, analyzeBlockIdentType.Equals(baseIdent.InferredType), "Analyze block name 'MetricsProvider' should have inferred type of its target")
	})
}

func TestInferLetTupleDestructuring(t *testing.T) {
	dsl := `
        system TestSys {
            analyze DestructureGood = {
                let a, b = (10, "ok");
                a + 1; // Should be int
            };
            analyze ReturnTuple = {
                let a, b = (10, "ok");
                (b, a); // Should be (string, int)
            };
            analyze ErrDestructureMismatchCount = {
                let x, y, z = (1, "two"); // Error: 3 vars, 2 tuple elements
                x;
            };
            analyze ErrDestructureNonTuple = {
                let x, y = 100; // Error: assigning int to 2 vars
                x;
            };
        }
    `
	runTypeInferenceTest(t, dsl, "LetTupleDestructuring", func(t *testing.T, file *FileDecl, errors []error) {
		expectedErrorCount := 2
		assert.Len(t, errors, expectedErrorCount, fmt.Sprintf("Expected %d errors for tuple destructuring", expectedErrorCount))
		assertErrorContains(t, errors, "assigns to multiple variables, but value type Tuple[Int, String] is not a matching tuple", "ErrDestructureMismatchCount") // Error is "not a matching tuple" due to length
		assertErrorContains(t, errors, "assigns to multiple variables, but value type Int is not a matching tuple", "ErrDestructureNonTuple")

		var expr Expr
		expr, _ = findExprInAnalyze(file, "TestSys", "DestructureGood") // This is a block, result is a+1
		assertInferredType(t, expr, IntType, "DestructureGood (a+1)")

		expr, _ = findExprInAnalyze(file, "TestSys", "ReturnTuple") // This is a block, result is (b,a)
		assertInferredType(t, expr, TupleType(StrType, IntType), "ReturnTuple (b,a)")

		// Verify types of a and b within the scope of DestructureGood block
		// This needs more precise node finding. Let's assume we can find the 'let a,b = ...'
		// and then check IdentifierExpr 'a' and 'b'.
		// For now, the fact that `a + 1` typed correctly to Int implies 'a' was Int.
	})
}

// Add more tests for:
// - Component param default value type checking
// - Deeper scoping rules if applicable
// - Error reporting quality (positions, messages)
// - List.Len, Outcomes.Len (if you add them to MemberAccess)
// - More complex Distribute/Sample interactions
