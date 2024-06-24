package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	DebugLogger *log.Logger
	isDebug     bool
)

func init() {
	DebugLogger = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func EnableDebug() {
	isDebug = true
}

func Log(format string, v ...interface{}) {
	if isDebug {
		_, file, line, _ := runtime.Caller(1)
		DebugLogger.Printf("%s:%d: %s", file, line, fmt.Sprintf(format, v...))
	}
}

func CreateOrModifyFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			indentation := line[:len(line)-len(trimmed)]
			if i > 0 {
				lines[i] = indentation + trimmed
			}
		}
	}

	// Language-agnostic test file handling
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	dirname := filepath.Base(filepath.Dir(path))

	switch ext {
	case ".go":
		if strings.HasSuffix(filename, "_test.go") {
			content = ensureCorrectPackage(content, dirname, "package")
		}
	case ".py":
		if strings.HasPrefix(filename, "test_") {
			content = ensureCorrectPackage(content, dirname, "from . import")
		}
	case ".js":
		if strings.HasPrefix(filename, "test_") || strings.HasSuffix(filename, ".test.js") {
			content = ensureCorrectPackage(content, dirname, "const { ") // Assuming ES6 module syntax
		}
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func ensureCorrectPackage(content, packageName, prefix string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			lines[i] = prefix + " " + packageName
			break
		}
	}
	return strings.Join(lines, "\n")
}

func GetFileLanguage(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".py":
		return "Python"
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	// TODO: Add more language support
	default:
		return "Unknown"
	}
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
