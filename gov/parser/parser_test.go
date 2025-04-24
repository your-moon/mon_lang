package parser

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/lexer"
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

	if program.FnDef.Token.Type != lexer.IDENT {
		t.Errorf("Expected function name to be IDENT, got %v", program.FnDef.Token.Type)
	}
	if program.FnDef.Token.Value == nil || *program.FnDef.Token.Value != "майн" {
		t.Errorf("Expected function name to be 'майн', got %v", program.FnDef.Token.Value)
	}

	if len(program.FnDef.BlockItems) != 1 {
		t.Errorf("Expected 1 statement in function body, got %d", len(program.FnDef.BlockItems))
	}

	returnStmt, ok := program.FnDef.BlockItems[0].(*ASTReturnStmt)
	if !ok {
		t.Error("Expected return statement")
	} else {
		constant, ok := returnStmt.ReturnValue.(*ASTConstant)
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

			if len(program.FnDef.BlockItems) != 1 {
				t.Errorf("Expected 1 statement in function body, got %d", len(program.FnDef.BlockItems))
			}

			returnStmt, ok := program.FnDef.BlockItems[0].(*ASTReturnStmt)
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

	if len(program.FnDef.BlockItems) != 1 {
		t.Errorf("Expected 1 statement in function body, got %d", len(program.FnDef.BlockItems))
	}

	stmt, ok := program.FnDef.BlockItems[0].(*ExpressionStmt)
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

	if len(program.FnDef.BlockItems) != 4 {
		t.Errorf("Expected 4 statements in function body, got %d", len(program.FnDef.BlockItems))
	}

	if decl, ok := program.FnDef.BlockItems[0].(*Decl); !ok {
		t.Error("Expected first statement to be declaration")
	} else if decl.Ident != "б" {
		t.Errorf("Expected first declaration name to be 'б', got %s", decl.Ident)
	}

	if decl, ok := program.FnDef.BlockItems[1].(*Decl); !ok {
		t.Error("Expected second statement to be declaration")
	} else {
		if decl.Ident != "а" {
			t.Errorf("Expected second declaration name to be 'а', got %s", decl.Ident)
		}
		if decl.Expr == nil {
			t.Error("Expected second declaration to have initialization")
		}
	}

	if stmt, ok := program.FnDef.BlockItems[2].(*ExpressionStmt); !ok {
		t.Error("Expected third statement to be assignment")
	} else {
		if _, ok := stmt.Expression.(*ASTAssignment); !ok {
			t.Error("Expected third statement to be assignment")
		}
	}

	if _, ok := program.FnDef.BlockItems[3].(*ASTReturnStmt); !ok {
		t.Error("Expected fourth statement to be return")
	}
}

func TestParseExamples(t *testing.T) {
	testDirs := []string{
		"../test",
		"../test/binary",
		"../test/spacetab",
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
					if program.FnDef.Token.Type != lexer.IDENT {
						t.Errorf("Expected function name to be IDENT in %s", file)
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
