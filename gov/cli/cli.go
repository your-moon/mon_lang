package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/base"
	codegen "github.com/your-moon/mn_compiler_go_version/code_gen"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	semanticanalysis "github.com/your-moon/mn_compiler_go_version/semantic_analysis"
	"github.com/your-moon/mn_compiler_go_version/tackygen"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Execute     func(args []string) error
}

// CLI represents the command-line interface
type CLI struct {
	commands map[string]Command
	debug    bool
	help     bool
}

// New creates a new CLI instance
func New() *CLI {
	cli := &CLI{
		commands: make(map[string]Command),
	}

	// Register commands
	cli.registerCommands()

	return cli
}

// registerCommands registers all available commands
func (c *CLI) registerCommands() {
	c.commands["lex"] = Command{
		Name:        "lex",
		Description: "Зөвхөн лексер ажиллуулах",
		Execute:     c.runLexer,
	}

	c.commands["parse"] = Command{
		Name:        "parse",
		Description: "Лексер болон парсер ажиллуулах",
		Execute:     c.runParser,
	}

	c.commands["validate"] = Command{
		Name:        "validate",
		Description: "Семантик шинжилгээ хийх",
		Execute:     c.runValidate,
	}

	c.commands["tacky"] = Command{
		Name:        "tacky",
		Description: "Лексер, парсер ажиллуулах, Tacky IR үүсгэх",
		Execute:     c.runTacky,
	}

	c.commands["compile"] = Command{
		Name:        "compile",
		Description: "Лексер, парсер ажиллуулах, Tacky үүсгэх",
		Execute:     c.runCompiler,
	}

	c.commands["gen"] = Command{
		Name:        "gen",
		Description: "Бүх алхмыг ажиллуулах, assembly үүсгэх",
		Execute:     c.runGen,
	}
}

// Run executes the CLI with the given arguments
func (c *CLI) Run(args []string) error {
	// Skip program name
	args = args[1:]

	// Check if we have at least a command
	if len(args) < 1 {
		c.printUsage()
		return fmt.Errorf("missing command")
	}

	// Extract command
	command := args[0]
	args = args[1:]

	// Parse flags for the specific command
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	fs.BoolVar(&c.debug, "debug", false, "enable debug mode")
	fs.BoolVar(&c.help, "help", false, "show detailed help information")

	// Find the position of the first flag
	fileArg := ""
	flagArgs := args
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = args[i:]
			if i > 0 {
				fileArg = args[0]
			}
			break
		}
		if i == len(args)-1 {
			fileArg = args[0]
			flagArgs = args[1:]
		}
	}

	// Parse the flags
	if err := fs.Parse(flagArgs); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	// Set debug mode
	base.Debug = c.debug

	// Check if help flag is set
	if c.help {
		c.printDetailedHelp()
		return nil
	}

	// Check if we have a file path
	if fileArg == "" {
		c.printUsage()
		return fmt.Errorf("missing file argument")
	}

	// Execute command
	if cmd, ok := c.commands[command]; ok {
		return cmd.Execute([]string{fileArg})
	}

	c.printUsage()
	return fmt.Errorf("unknown command: %s", command)
}

// printUsage prints the usage information
func (c *CLI) printUsage() {
	fmt.Println("Хэрэглээ: compiler <команд> <файл> [сонголтууд]")
	fmt.Println("\nКомандууд:")
	for _, cmd := range c.commands {
		fmt.Printf("  %-10s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println("\nСонголтууд:")
	fmt.Println("  --debug    Debug горим идэвхжүүлэх")
	fmt.Println("  --help     Дэлгэрэнгүй тусламж харуулах")
}

// printDetailedHelp prints detailed help information
func (c *CLI) printDetailedHelp() {
	fmt.Println("Монгол хэлний компилятор - Дэлгэрэнгүй тусламж")
	fmt.Println("=============================================")
	fmt.Println("\nХэрэглээ: compiler <команд> <файл> [сонголтууд]")

	fmt.Println("\nКомандууд:")
	fmt.Println("  lex       Зөвхөн лексер ажиллуулах")
	fmt.Println("            Жишээ: compiler lex input.mn")
	fmt.Println("\n  parse     Лексер болон парсер ажиллуулах")
	fmt.Println("            Жишээ: compiler parse input.mn")
	fmt.Println("\n  validate  Семантик шинжилгээ хийх")
	fmt.Println("            Жишээ: compiler validate input.mn")
	fmt.Println("\n  tacky     Лексер, парсер ажиллуулах, Tacky IR үүсгэх")
	fmt.Println("            Жишээ: compiler tacky input.mn")
	fmt.Println("\n  compile   Лексер, парсер ажиллуулах, Tacky үүсгэх")
	fmt.Println("            Жишээ: compiler compile input.mn")
	fmt.Println("\n  gen       Бүх алхмыг ажиллуулах, assembly үүсгэх")
	fmt.Println("            Жишээ: compiler gen input.mn")

	fmt.Println("\nСонголтууд:")
	fmt.Println("  --debug   Debug горим идэвхжүүлэх")
	fmt.Println("            Жишээ: compiler lex input.mn --debug")
	fmt.Println("\n  --help    Дэлгэрэнгүй тусламж харуулах")
	fmt.Println("            Жишээ: compiler --help")

	fmt.Println("\nЖишээ:")
	fmt.Println("  compiler lex input.mn")
	fmt.Println("  compiler parse input.mn --debug")
	fmt.Println("  compiler gen input.mn")
}

// Helper functions for each command
func (c *CLI) runLexer(args []string) error {
	runeString := readFile(args[0])
	scanner := lexer.NewScanner(runeString)
	for {
		token, err := scanner.Scan()
		if err != nil {
			return fmt.Errorf("лексер алдаа: %v", err)
		}
		if c.debug {
			if token.Value != nil {
				fmt.Printf("[debug] Токен: %s, Утга: %s\n", token.Type, *token.Value)
			} else {
				fmt.Printf("[debug] Токен: %s\n", token.Type)
			}
		}
		if token.Type == lexer.EOF {
			break
		}
	}
	return nil
}

func (c *CLI) runParser(args []string) error {
	runeString := readFile(args[0])
	if err := c.runLexer(args); err != nil {
		return err
	}

	fmt.Println("\n---- ПАРСИНГ ----:")
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	if len(parsed.Errors()) > 0 {
		return fmt.Errorf("парсерын алдаанууд: %v", parsed.Errors()[0].Error())
	}

	if base.Debug && node != nil {
		fmt.Println("AST:", node.PrintAST(0))
	}

	resolver := semanticanalysis.New(runeString)

	_, err = resolver.Resolve(node)
	if err != nil {
		return err
	}

	return nil
}

func (c *CLI) runValidate(args []string) error {
	runeString := readFile(args[0])
	if err := c.runLexer(args); err != nil {
		return err
	}

	fmt.Println("\n---- ПАРСИНГ ----:")
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	if len(parsed.Errors()) > 0 {
		return fmt.Errorf("парсерын алдаанууд: %v", parsed.Errors()[0].Error())
	}

	if base.Debug && node != nil {
		fmt.Println("AST:", node.PrintAST(0))
	}

	resolver := semanticanalysis.New(runeString)

	_, err = resolver.Resolve(node)
	if err != nil {
		return err
	}

	return nil
}

func (c *CLI) runTacky(args []string) error {
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	if base.Debug && node != nil {
		fmt.Println("AST:", node.PrintAST(0))
	}

	resolver := semanticanalysis.New(runeString)

	_, err = resolver.Resolve(node)
	if err != nil {
		return err
	}

	fmt.Println("\n---- TACKY IR ҮҮСГЭЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("---- TACKY IR ЖАГСААЛТ ----:")
	for _, ir := range tackyprogram.FnDef.Instructions {
		ir.Ir()
	}
	return nil
}

func (c *CLI) runCompiler(args []string) error {
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	fmt.Println("\n---- КОМПАЙЛЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	resolver := semanticanalysis.New(runeString)

	_, err = resolver.Resolve(node)
	if err != nil {
		return err
	}

	fmt.Println("---- TACKY ЖАГСААЛТ ----:")
	for _, ir := range tackyprogram.FnDef.Instructions {
		fmt.Println(ir)
		ir.Ir()
	}
	return nil
}

func (c *CLI) runGen(args []string) error {
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("\n---- ASSEMBLY ҮҮСГЭЖ БАЙНА ----:")
	outfile := "out.asm"
	openFile, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("гаралтын файл нээхэд алдаа гарлаа: %v", err)
	}
	defer openFile.Close()

	asmastgen := codegen.NewAsmGen()
	asmast := asmastgen.GenASTAsm(tackyprogram)

	asmgen := codegen.NewGenASM(openFile, codegen.Aarch64)
	asmgen.GenAsm(asmast)
	fmt.Printf("Assembly файл %s дотор үүслээ\n", outfile)
	return nil
}

// Helper functions
func readFile(filePath string) []int32 {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Файл уншихад алдаа гарлаа: %v\n", err)
		os.Exit(1)
	}
	return convertToRuneArray(string(data))
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
