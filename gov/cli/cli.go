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
	"github.com/your-moon/mn_compiler_go_version/unique"
)

type Command struct {
	Name        string
	Description string
	Execute     func(args []string) error
}

type CLI struct {
	commands map[string]Command
	debug    bool
	help     bool
}

func New() *CLI {
	cli := &CLI{
		commands: make(map[string]Command),
	}

	cli.registerCommands()

	return cli
}

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

func (c *CLI) Run(args []string) error {
	args = args[1:]

	if len(args) < 1 {
		c.printUsage()
		return fmt.Errorf("команд олдсонгүй")
	}

	command := args[0]
	args = args[1:]

	fs := flag.NewFlagSet(command, flag.ExitOnError)
	fs.BoolVar(&c.debug, "debug", false, "debug mode асаах")
	fs.BoolVar(&c.help, "help", false, "команд туслах харуулах")

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

	if err := fs.Parse(flagArgs); err != nil {
		return fmt.Errorf("флаг парс хийхэд алдаа гарлаа: %v", err)
	}

	base.Debug = c.debug

	if c.help {
		c.printDetailedHelp()
		return nil
	}

	if fileArg == "" {
		c.printUsage()
		return fmt.Errorf("файл-ийн ардаас аргумент оруулж өгнө үү.")
	}

	if cmd, ok := c.commands[command]; ok {
		return cmd.Execute([]string{fileArg})
	}

	c.printUsage()
	return fmt.Errorf("зөв команд оруулна уу: %s", command)
}

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

	return nil
}

func (c *CLI) runValidate(args []string) error {
	uniqueGen := unique.NewUniqueGen()
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

	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen)

	resolvedAst, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	if base.Debug && resolvedAst != nil {
		fmt.Println("AST:", resolvedAst.PrintAST(0))
	}

	return nil
}

func (c *CLI) runTacky(args []string) error {
	uniqueGen := unique.NewUniqueGen()
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	if base.Debug && node != nil {
		fmt.Println("AST:", node.PrintAST(0))
	}

	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen)

	resolvedAst, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	fmt.Println("\n---- TACKY IR ҮҮСГЭЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen(uniqueGen)
	tackyprogram := compilerx.EmitTacky(resolvedAst)

	fmt.Println("---- TACKY IR ЖАГСААЛТ ----:")
	for _, fn := range tackyprogram.FnDefs {
		fmt.Println(fmt.Sprintf("%s:", fn.Name))
		for _, ir := range fn.Instructions {
			ir.Ir()
		}
	}
	return nil
}

func (c *CLI) runCompiler(args []string) error {
	uniqueGen := unique.NewUniqueGen()
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen)
	resolvedAst, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	fmt.Println("\n---- КОМПАЙЛЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen(uniqueGen)
	tackyprogram := compilerx.EmitTacky(resolvedAst)

	fmt.Println("---- TACKY ЖАГСААЛТ ----:")
	for _, fn := range tackyprogram.FnDefs {
		fmt.Println(fn.Name)
		for _, ir := range fn.Instructions {
			ir.Ir()
		}
	}

	fmt.Println("\n---- ASSEMBLY ҮҮСГЭЖ БАЙНА ----:")
	asmgen := codegen.NewAsmGen()
	asmast := asmgen.GenASTAsm(tackyprogram)

	fmt.Println("---- ASMAST ЖАГСААЛТ ----:")
	for _, fn := range asmast.AsmFnDef {
		for _, ir := range fn.Irs {
			ir.Ir()
		}
	}

	return nil
}

func (c *CLI) runGen(args []string) error {
	uniqueGen := unique.NewUniqueGen()
	runeString := readFile(args[0])
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("парсингийн алдаа: %v", err)
	}

	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen)
	resolvedAst, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	compilerx := tackygen.NewTackyGen(uniqueGen)
	tackyprogram := compilerx.EmitTacky(resolvedAst)

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
