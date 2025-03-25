package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/your-moon/mn_compiler_go_version/base"
	codegen "github.com/your-moon/mn_compiler_go_version/code_gen"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/tackygen"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "compiler",
	Short: "Mongolian language compiler",
	Long:  `A compiler for the Mongolian programming language.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		base.Debug = debug
	},
}

var lexCmd = &cobra.Command{
	Use:   "lex [file]",
	Short: "Run only the lexer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLexer(args[0])
	},
}

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Run lexer and parser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runParser(args[0])
	},
}

var tackyCmd = &cobra.Command{
	Use:   "tacky [file]",
	Short: "Run lexer, parser, and generate tacky IR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTacky(args[0])
	},
}

var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Run lexer, parser, and tacky generation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCompiler(args[0])
	},
}

var genCmd = &cobra.Command{
	Use:   "gen [file]",
	Short: "Run all steps including assembly generation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGen(args[0])
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.compiler.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	rootCmd.AddCommand(lexCmd, parseCmd, tackyCmd, compileCmd, genCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".compiler")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runLexer(filePath string) error {
	runeString := readFile(filePath)
	scanner := lexer.NewScanner(runeString)
	fmt.Println("---- LEXING ----:")
	for {
		token, err := scanner.Scan()
		if err != nil {
			return fmt.Errorf("lexing error: %v", err)
		}
		if token.Type == lexer.EOF {
			break
		}
		fmt.Printf("Token: %v\n", token)
	}
	return nil
}

func runParser(filePath string) error {
	runeString := readFile(filePath)
	if err := runLexer(filePath); err != nil {
		return err
	}

	fmt.Println("\n---- PARSING ----:")
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	if len(parsed.Errors()) > 0 {
		return fmt.Errorf("parser errors: %v", parsed.Errors())
	}

	if base.Debug && node != nil {
		fmt.Println("AST:", node.PrintAST(0))
	}
	return nil
}

func runCompiler(filePath string) error {
	runeString := readFile(filePath)
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	fmt.Println("\n---- COMPILING ----:")
	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("---- TACKY LIST ----:")
	for _, ir := range tackyprogram.FnDef.Instructions {
		fmt.Println(ir)
		ir.Ir()
	}
	return nil
}

func runTacky(filePath string) error {
	runeString := readFile(filePath)
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	fmt.Println("\n---- GENERATING TACKY IR ----:")
	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("---- TACKY IR LIST ----:")
	for _, ir := range tackyprogram.FnDef.Instructions {
		ir.Ir()
	}
	return nil
}

func runGen(filePath string) error {
	runeString := readFile(filePath)
	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("\n---- GENERATING ASSEMBLY ----:")
	outfile := "out.asm"
	openFile, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	defer openFile.Close()

	asmastgen := codegen.NewAsmGen()
	asmast := asmastgen.GenASTAsm(tackyprogram)

	asmgen := codegen.NewGenASM(openFile, codegen.Aarch64)
	asmgen.GenAsm(asmast)
	fmt.Printf("Assembly generated in %s\n", outfile)
	return nil
}

func readFile(filePath string) []int32 {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	return convertToRuneArray(string(data))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type FileCheck struct {
	File   string
	Status bool
	Error  error
}

func processFile(filePath string) FileCheck {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return FileCheck{
			File:   filePath,
			Status: false,
		}
	}

	runeString := convertToRuneArray(string(data))
	scanner := lexer.NewScanner(runeString)

	for {
		token, err := scanner.Scan()
		if err != nil {
			return FileCheck{
				File:   filePath,
				Status: false,
				Error:  err,
			}
		}
		if token.Type == lexer.EOF {
			return FileCheck{
				File:   filePath,
				Status: true,
			}
		}
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

func canLex() []FileCheck {
	dir := "./test/ch1/"
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return nil
	}

	var fileChecks []FileCheck
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		fileChecks = append(fileChecks, processFile(filePath))
	}

	return fileChecks
}
