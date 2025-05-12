package stdlib

import (
	"os"
	"path/filepath"
	"strings"
)

// StandardLibrary represents the collection of standard library functions
type StandardLibrary struct {
	Functions map[string]string // Map of function names to their implementations
}

// New creates a new StandardLibrary instance
func New() *StandardLibrary {
	return &StandardLibrary{
		Functions: make(map[string]string),
	}
}

// Load loads all standard library files from the stdlib directory
func (sl *StandardLibrary) Load() error {
	files, err := os.ReadDir("stdlib")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mn") {
			content, err := os.ReadFile(filepath.Join("stdlib", file.Name()))
			if err != nil {
				return err
			}
			sl.parseFile(string(content))
		}
	}

	return nil
}

// parseFile parses a standard library file and extracts function definitions
func (sl *StandardLibrary) parseFile(content string) {
	lines := strings.Split(content, "\n")
	var currentFunc strings.Builder
	var currentFuncName string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Check for function declaration
		if strings.HasPrefix(line, "функц") {
			// If we were building a function, save it
			if currentFuncName != "" {
				sl.Functions[currentFuncName] = currentFunc.String()
				currentFunc.Reset()
			}

			// Extract function name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentFuncName = parts[1]
				// Remove the opening parenthesis and everything after it
				currentFuncName = strings.Split(currentFuncName, "(")[0]
			}
		}

		// Add line to current function
		if currentFuncName != "" {
			currentFunc.WriteString(line)
			currentFunc.WriteString("\n")
		}
	}

	// Save the last function
	if currentFuncName != "" {
		sl.Functions[currentFuncName] = currentFunc.String()
	}
}

// GetFunction returns the implementation of a standard library function
func (sl *StandardLibrary) GetFunction(name string) (string, bool) {
	impl, exists := sl.Functions[name]
	return impl, exists
}

// IsStdLibFunction checks if a function name is part of the standard library
func (sl *StandardLibrary) IsStdLibFunction(name string) bool {
	_, exists := sl.Functions[name]
	return exists
}
