package parser

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler/lexer"
)

func TestParseSimple(t *testing.T) {
	source := []int32("функц майн() -> тоо { буц 5; }")
	p := NewParser(source)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if len(p.Errors()) > 0 {
		t.Errorf("Parse errors: %v", p.Errors())
	}
	if program == nil {
		t.Fatal("Program is nil")
	}

	decl := program.Decls[0]
	fnDecl, ok := decl.(*FnDecl)
	if !ok {
		t.Errorf("Expected function declaration, got %v", fnDecl)
	}

	if fnDecl.Token.Type != lexer.FN {
		t.Errorf("Expected function name to be FN, got %v", fnDecl.Token.Type)
	}

	if len(fnDecl.Body.BlockItems) != 1 {
		t.Errorf("Expected 1 statement in function body, got %d", len(fnDecl.Body.BlockItems))
	}

	returnStmt, ok := fnDecl.Body.BlockItems[0].(*ASTReturnStmt)
	if !ok {
		t.Error("Expected return statement")
	} else {
		constant, ok := returnStmt.ReturnValue.(*ASTConstInt)
		if !ok {
			t.Error("Expected constant in return statement")
		} else if constant.Value != 5 {
			t.Errorf("Expected return value to be 5, got %d", constant.Value)
		}
	}
}

func TestParseBinaryOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"plus", "функц майн() -> тоо { буц 5 + 5; }"},
		{"minus", "функц майн() -> тоо { буц 5 - 5; }"},
		{"multiply", "функц майн() -> тоо { буц 5 * 5; }"},
		{"divide", "функц майн() -> тоо { буц 5 / 5; }"},
		{"and", "функц майн() -> тоо { буц 1 && 1; }"},
		{"or", "функц майн() -> тоо { буц 1 || 1; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []int32(tt.source)
			p := NewParser(source)
			program, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.name, err)
			}
			if len(p.Errors()) > 0 {
				t.Errorf("Parse errors in %s: %v", tt.name, p.Errors())
			}
			if program == nil {
				t.Fatalf("Program is nil for %s", tt.name)
			}

			decl := program.Decls[0]
			fnDecl, ok := decl.(*FnDecl)
			if !ok {
				t.Errorf("Expected function declaration, got %v", fnDecl)
			}

			if len(fnDecl.Body.BlockItems) != 1 {
				t.Errorf("Expected 1 statement in function body, got %d", len(fnDecl.Body.BlockItems))
			}

			returnStmt, ok := fnDecl.Body.BlockItems[0].(*ASTReturnStmt)
			if !ok {
				t.Error("Expected return statement")
			} else {
				_, ok := returnStmt.ReturnValue.(*ASTBinary)
				if !ok {
					t.Errorf("Expected binary operation in return statement for %s", tt.name)
				}
			}
		})
	}
}

func TestParseConditional(t *testing.T) {
	source := []int32("функц үндсэн() -> тоо { а = 1 ? 2 ? 4*5 : 5/3 : 3*3; }")
	p := NewParser(source)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Failed to parse conditional: %v", err)
	}
	if len(p.Errors()) > 0 {
		t.Errorf("Parse errors: %v", p.Errors())
	}
	if program == nil {
		t.Fatal("Program is nil")
	}

	decl := program.Decls[0]
	fnDecl, ok := decl.(*FnDecl)
	if !ok {
		t.Errorf("Expected function declaration, got %v", fnDecl)
	}

	if len(fnDecl.Body.BlockItems) != 1 {
		t.Errorf("Expected 1 statement in function body, got %d", len(fnDecl.Body.BlockItems))
	}

	stmt, ok := fnDecl.Body.BlockItems[0].(*ExpressionStmt)
	if !ok {
		t.Error("Expected expression statement")
	} else {
		assign, ok := stmt.Expression.(*ASTAssignment)
		if !ok {
			t.Error("Expected assignment")
		} else {
			_, ok := assign.Right.(*ASTConditional)
			if !ok {
				t.Error("Expected conditional in assignment")
			}
		}
	}
}

func TestParseDeclarations(t *testing.T) {
	source := []int32("функц үндсэн() -> тоо { зарла б:тоо; зарла а:тоо = 10 + 3; б = а * 2; буц б; }")
	p := NewParser(source)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Failed to parse declarations: %v", err)
	}
	if len(p.Errors()) > 0 {
		t.Errorf("Parse errors: %v", p.Errors())
	}
	if program == nil {
		t.Fatal("Program is nil")
	}

	decl := program.Decls[0]
	fnDecl, ok := decl.(*FnDecl)
	if !ok {
		t.Errorf("Expected function declaration, got %v", fnDecl)
	}

	if len(fnDecl.Body.BlockItems) != 4 {
		t.Errorf("Expected 4 statements in function body, got %d", len(fnDecl.Body.BlockItems))
	}

	if decl, ok := fnDecl.Body.BlockItems[0].(*VarDecl); !ok {
		t.Error("Expected first statement to be declaration")
	} else if decl.Ident != "б" {
		t.Errorf("Expected first declaration name to be 'б', got %s", decl.Ident)
	}

	if decl, ok := fnDecl.Body.BlockItems[1].(*VarDecl); !ok {
		t.Error("Expected second statement to be declaration")
	} else {
		if decl.Ident != "а" {
			t.Errorf("Expected second declaration name to be 'а', got %s", decl.Ident)
		}
		if decl.Expr == nil {
			t.Error("Expected second declaration to have initialization")
		}
	}

	if stmt, ok := fnDecl.Body.BlockItems[2].(*ExpressionStmt); !ok {
		t.Error("Expected third statement to be assignment")
	} else {
		if _, ok := stmt.Expression.(*ASTAssignment); !ok {
			t.Error("Expected third statement to be assignment")
		}
	}

	if _, ok := fnDecl.Body.BlockItems[3].(*ASTReturnStmt); !ok {
		t.Error("Expected fourth statement to be return")
	}
}

func TestParseExamples(t *testing.T) {
	testDirs := []string{
		"../test",
		"../test/binary",
		"../test/spacetab",
		"../test/precedence",
	}

	for _, dir := range testDirs {
		t.Run(dir, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join(dir, "*.mn"))
			if err != nil {
				t.Fatalf("Failed to read test directory %s: %v", dir, err)
			}

			for _, file := range files {
				t.Run(filepath.Base(file), func(t *testing.T) {
					content, err := os.ReadFile(file)
					if err != nil {
						t.Fatalf("Failed to read test file %s: %v", file, err)
					}

					source := convertToRuneArray(string(content))

					p := NewParser(source)

					program, err := p.ParseProgram()
					if err != nil {
						t.Fatalf("Failed to parse %s: %v", file, err)
					}

					if len(p.Errors()) > 0 {
						t.Errorf("Parse errors in %s: %v", file, p.Errors())
					}

					if program == nil {
						t.Errorf("Program is nil for %s", file)
					}
				})
			}
		})
	}
}

func convertToRuneArray(dataString string) []int32 {
	var runeString []int32
	for len(dataString) > 0 {
		r, size := utf8.DecodeRuneInString(dataString)
		runeString = append(runeString, r)
		dataString = dataString[size:]
	}
	runeString = append(runeString, 0)
	return runeString
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		source string
		check  func(t *testing.T, expr ASTExpression)
	}{
		{
			name:   "multiplication before addition",
			source: "функц майн() -> тоо { буц 10 + 3 * 5; }",
			check: func(t *testing.T, expr ASTExpression) {
				bin, ok := expr.(*ASTBinary)
				if !ok {
					t.Fatal("Expected binary expression")
				}
				if bin.Op != ASTBinOp(A_PLUS) {
					t.Errorf("Expected root operator to be +, got %v", bin.Op)
				}
				left, ok := bin.Left.(*ASTConstInt)
				if !ok || left.Value != 10 {
					t.Error("Expected left operand to be constant 10")
				}
				right, ok := bin.Right.(*ASTBinary)
				if !ok || right.Op != ASTBinOp(A_MUL) {
					t.Error("Expected right operand to be multiplication")
				}
				rightLeft, ok := right.Left.(*ASTConstInt)
				if !ok || rightLeft.Value != 3 {
					t.Error("Expected first multiplicand to be 3")
				}
				rightRight, ok := right.Right.(*ASTConstInt)
				if !ok || rightRight.Value != 5 {
					t.Error("Expected second multiplicand to be 5")
				}
			},
		},
		{
			name:   "complex arithmetic precedence",
			source: "функц майн() -> тоо { буц 10 + 3 * 5 + 2 * 4; }",
			check: func(t *testing.T, expr ASTExpression) {
				// Should parse as ((10 + (3 * 5)) + (2 * 4))
				bin, ok := expr.(*ASTBinary)
				if !ok || bin.Op != ASTBinOp(A_PLUS) {
					t.Fatal("Expected top-level addition")
				}

				left, ok := bin.Left.(*ASTBinary)
				if !ok || left.Op != ASTBinOp(A_PLUS) {
					t.Error("Expected left to be addition")
				}

				leftLeft, ok := left.Left.(*ASTConstInt)
				if !ok || leftLeft.Value != 10 {
					t.Error("Expected leftmost operand to be 10")
				}

				leftRight, ok := left.Right.(*ASTBinary)
				if !ok || leftRight.Op != ASTBinOp(A_MUL) {
					t.Error("Expected 3 * 5 multiplication")
				}

				right, ok := bin.Right.(*ASTBinary)
				if !ok || right.Op != ASTBinOp(A_MUL) {
					t.Error("Expected 2 * 4 multiplication")
				}
			},
		},
		{
			name:   "parentheses override precedence",
			source: "функц майн() -> тоо { буц (10 + 3) * 5; }",
			check: func(t *testing.T, expr ASTExpression) {
				bin, ok := expr.(*ASTBinary)
				if !ok || bin.Op != ASTBinOp(A_MUL) {
					t.Fatal("Expected multiplication at root")
				}

				left, ok := bin.Left.(*ASTBinary)
				if !ok || left.Op != ASTBinOp(A_PLUS) {
					t.Error("Expected addition in parentheses")
				}

				right, ok := bin.Right.(*ASTConstInt)
				if !ok || right.Value != 5 {
					t.Error("Expected right operand to be 5")
				}
			},
		},
		{
			name:   "logical operators precedence",
			source: "функц майн() -> тоо { буц 1 > 0 && 2 > 1 || 3 > 2; }",
			check: func(t *testing.T, expr ASTExpression) {
				// Should parse as ((1 > 0 && 2 > 1) || 3 > 2)
				bin, ok := expr.(*ASTBinary)
				if !ok || bin.Op != ASTBinOp(A_OR) {
					t.Fatal("Expected OR at root")
				}

				left, ok := bin.Left.(*ASTBinary)
				if !ok || left.Op != ASTBinOp(A_AND) {
					t.Error("Expected AND as left child")
				}

				right, ok := bin.Right.(*ASTBinary)
				if !ok || right.Op != ASTBinOp(A_GREATERTHAN) {
					t.Error("Expected > comparison as right child")
				}
			},
		},
		{
			name:   "comparison and arithmetic precedence",
			source: "функц майн() -> тоо { буц 10 + 3 > 5 * 2; }",
			check: func(t *testing.T, expr ASTExpression) {
				// Should parse as ((10 + 3) > (5 * 2))
				bin, ok := expr.(*ASTBinary)
				if !ok || bin.Op != ASTBinOp(A_GREATERTHAN) {
					t.Fatal("Expected > at root")
				}

				left, ok := bin.Left.(*ASTBinary)
				if !ok || left.Op != ASTBinOp(A_PLUS) {
					t.Error("Expected addition as left child")
				}

				right, ok := bin.Right.(*ASTBinary)
				if !ok || right.Op != ASTBinOp(A_MUL) {
					t.Error("Expected multiplication as right child")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []int32(tt.source)
			p := NewParser(source)
			program, err := p.ParseProgram()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.name, err)
			}
			if len(p.Errors()) > 0 {
				t.Errorf("Parse errors in %s: %v", tt.name, p.Errors())
			}
			if program == nil {
				t.Fatalf("Program is nil for %s", tt.name)
			}

			returnStmt, ok := program.Decls[0].(*FnDecl).Body.BlockItems[0].(*ASTReturnStmt)
			if !ok {
				t.Fatalf("Expected return statement for %s", tt.name)
			}

			tt.check(t, returnStmt.ReturnValue)
		})
	}
}
