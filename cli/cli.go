package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/base"
	codegen "github.com/your-moon/mn_compiler_go_version/code_gen"
	"github.com/your-moon/mn_compiler_go_version/code_gen/asmsymbol"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/linker"
	"github.com/your-moon/mn_compiler_go_version/parser"
	semanticanalysis "github.com/your-moon/mn_compiler_go_version/semantic_analysis"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/tackygen"
	"github.com/your-moon/mn_compiler_go_version/util"
	"github.com/your-moon/mn_compiler_go_version/util/unique"
)

type Command struct {
	Name        string
	Description string
	Execute     func(args []string) error
}

type CLI struct {
	commands   map[string]Command
	debug      bool
	help       bool
	genAsm     bool
	genObj     bool
	run        bool
	outputFile string
}

type Options struct {
	InputFile  string
	OutputFile string
	GenAsm     bool
	GenObj     bool
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
	fs.BoolVar(&c.genAsm, "asm", false, "assembly файл үүсгэх")
	fs.BoolVar(&c.genObj, "obj", false, "object файл үүсгэх")
	fs.BoolVar(&c.run, "run", false, "компиляц хийгээд ажиллуулах")
	fs.StringVar(&c.outputFile, "o", "", "гаралтын файлын нэр)")

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
	fmt.Println("  --asm      Assembly файл үүсгэх")
	fmt.Println("  --obj      Object файл үүсгэх")
	fmt.Println("  --run      Compile and run the program")
	fmt.Println("  -o         Гаралтын файлын нэр")
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
	fmt.Println("\n  --asm     Assembly файл үүсгэх")
	fmt.Println("            Жишээ: compiler gen input.mn --asm")
	fmt.Println("\n  --obj     Object файл үүсгэх")
	fmt.Println("            Жишээ: compiler gen input.mn --obj")
	fmt.Println("\n  --run     Компиляцын ард програм ажиллуулах")
	fmt.Println("            Жишээ: compiler gen input.mn --run")
	fmt.Println("\n  -o        Гаралтын файлын нэр")
	fmt.Println("            Жишээ: compiler gen input.mn -o output")

	fmt.Println("\nЖишээ:")
	fmt.Println("  compiler lex input.mn")
	fmt.Println("  compiler parse input.mn --debug")
	fmt.Println("  compiler gen input.mn -o myprogram")
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
			fmt.Println(fmt.Sprintf("[debug] %v", token.Type))
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

	if base.Debug {
		fmt.Println("\n---- ПАРСИНГ ----:")
	}
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

	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen, table)

	resolvedAst, _, err := resolver.Analyze(node)
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

	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen, table)

	resolvedAst, _, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	fmt.Println("\n---- TACKY IR ҮҮСГЭЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen(uniqueGen, table)
	tackyprogram := compilerx.EmitTacky(resolvedAst)

	fmt.Println("---- TACKY IR ЖАГСААЛТ ----:")
	for _, fn := range tackyprogram.ExternDefs {
		fn.Ir()
	}

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

	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen, table)
	resolvedAst, symbolTable, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	fmt.Println("\n---- КОМПАЙЛЖ БАЙНА ----:")
	compilerx := tackygen.NewTackyGen(uniqueGen, table)

	tackyprogram := compilerx.EmitTacky(resolvedAst)

	fmt.Println("---- TACKY ЖАГСААЛТ ----:")
	for _, fn := range tackyprogram.FnDefs {
		fmt.Println(fn.Name)
		for _, ir := range fn.Instructions {
			ir.Ir()
		}
	}

	asmTable := asmsymbol.NewAsmSymbolTable()
	fmt.Println("\n---- ASSEMBLY ҮҮСГЭЖ БАЙНА ----:")
	asmgen := codegen.NewAsmGen(table)
	asmast := asmgen.GenASTAsm(tackyprogram, symbolTable, asmTable)

	fmt.Println("---- ASMAST ЖАГСААЛТ ----:")
	for _, fn := range asmast.AsmFnDef {
		for _, ir := range fn.Irs {
			ir.Ir()
		}
	}

	asmBuffer := new(bytes.Buffer)
	asmWriter := codegen.NewGenASM(asmBuffer, util.GetOsType())
	asmWriter.GenAsm(asmast)

	err = os.WriteFile("debug_out.asm", asmBuffer.Bytes(), 0644)
	if err != nil {
		fmt.Println("Failed to write debug_out.asm:", err)
	}

	// linker := linker.NewLinker("out")
	// linker.SetAssemblyContent(asmBuffer.String())
	// if err := linker.Link(); err != nil {
	// 	fmt.Println("Error linking:", err)
	// 	os.Exit(1)
	// }

	// if err := linker.MakeExecutable(); err != nil {
	// 	fmt.Println("Error making executable:", err)
	// 	os.Exit(1)
	// }

	// fmt.Println("Created executable: out")
	return nil
}

func (c *CLI) runGen(args []string) error {
	uniqueGen := unique.NewUniqueGen()
	runeString := readFile(args[0])
	if err := c.runLexer(args); err != nil {
		return err
	}

	if base.Debug {
		fmt.Println("\n---- ПАРСИНГ ----:")
	}
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

	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen, table)

	resolvedAst, symbolTable, err := resolver.Analyze(node)
	if err != nil {
		return fmt.Errorf("семантик шинжилгээний алдаа: %v", err)
	}

	if base.Debug {
		fmt.Println("\n---- RESOLVED AST ----:")
		fmt.Println(resolvedAst.PrintAST(0))
	}

	tackyGen := tackygen.NewTackyGen(uniqueGen, table)
	tackyProgram := tackyGen.EmitTacky(resolvedAst)

	if base.Debug {
		fmt.Println("\n---- TACKY IR ----:")
		tackyGen.PrettyPrint(tackyProgram)
	}

	asmTable := asmsymbol.NewAsmSymbolTable()

	asmGen := codegen.NewAsmGen(table)
	asmProgram := asmGen.GenASTAsm(tackyProgram, symbolTable, asmTable)

	asmBuffer := new(bytes.Buffer)
	asmWriter := codegen.NewGenASM(asmBuffer, util.GetOsType())
	asmWriter.GenAsm(asmProgram)

	if base.Debug {
		fmt.Println("\n---- ASMAST ----:")
		fmt.Println(asmBuffer.String())
	}

	outputFile := c.outputFile
	if outputFile == "" {
		outputFile = filepath.Base(strings.TrimSuffix(args[0], ".mn"))
	}

	if dir := filepath.Dir(outputFile); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	linker := linker.NewLinker(outputFile)
	linker.SetAssemblyContent(asmBuffer.String())
	linker.SetGenerateAsm(c.genAsm)
	linker.SetGenerateObj(c.genObj)

	if err := linker.Link(); err != nil {
		return fmt.Errorf("Error linking: %v", err)
	}

	if !c.genAsm && !c.genObj {
		if err := linker.MakeExecutable(); err != nil {
			return fmt.Errorf("Error making executable: %v", err)
		}

		if c.run {
			if err := linker.Run(); err != nil {
				return fmt.Errorf("Error running program: %v", err)
			}
		}
	}

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

func ParseArgs() (*Options, error) {
	options := &Options{}

	flag.StringVar(&options.OutputFile, "o", "", "Output file name (default: input file name)")
	flag.BoolVar(&options.GenAsm, "asm", false, "Generate assembly file")
	flag.BoolVar(&options.GenObj, "obj", false, "Generate object file")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 || args[0] != "gen" {
		return nil, fmt.Errorf("usage: %s gen <input_file>", os.Args[0])
	}

	options.InputFile = args[1]

	if options.OutputFile == "" {
		base := filepath.Base(options.InputFile)
		ext := filepath.Ext(base)
		options.OutputFile = base[:len(base)-len(ext)]
	}

	return options, nil
}
