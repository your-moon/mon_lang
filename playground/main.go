package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	compilerPath := flag.String("compiler", "", "Path to mon_lang compiler binary")
	stdlibPath := flag.String("stdlib", "", "Path to stdlib directory")
	flag.Parse()

	// Auto-detect compiler
	if *compilerPath == "" {
		// Look for mon_lang in same directory as playground binary
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		candidate := filepath.Join(dir, "mon_lang")
		if _, err := os.Stat(candidate); err == nil {
			*compilerPath = candidate
		} else {
			// Try PATH
			*compilerPath = "mon_lang"
		}
	}

	// Auto-detect stdlib
	if *stdlibPath == "" {
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		candidate := filepath.Join(dir, "stdlib")
		if _, err := os.Stat(candidate); err == nil {
			*stdlibPath = candidate
		} else {
			*stdlibPath = "../stdlib"
		}
	}

	absStdlib, err := filepath.Abs(*stdlibPath)
	if err != nil {
		log.Fatalf("stdlib path error: %v", err)
	}

	absCompiler, err := filepath.Abs(*compilerPath)
	if err != nil {
		log.Fatalf("compiler path error: %v", err)
	}

	srv := NewServer(absCompiler, absStdlib)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Mon Lang Playground starting on http://localhost%s", addr)
	log.Printf("Compiler: %s", absCompiler)
	log.Printf("Stdlib: %s", absStdlib)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
