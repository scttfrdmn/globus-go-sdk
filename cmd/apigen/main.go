// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// Package main provides a tool for generating API signatures from Go code
// for tracking API compatibility between versions.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// APISignature represents the signature of an API component
type APISignature struct {
	Package string      `json:"package"`
	Name    string      `json:"name"`
	Type    string      `json:"type"` // "func", "type", "const", "var"
	Details interface{} `json:"details"`
}

// FuncSignature represents a function or method signature
type FuncSignature struct {
	Params       []string `json:"params"`
	Results      []string `json:"results"`
	Receiver     string   `json:"receiver,omitempty"`
	IsExported   bool     `json:"is_exported"`
	File         string   `json:"file"`
	IsDeprecated bool     `json:"is_deprecated"`
}

// TypeSignature represents a type definition
type TypeSignature struct {
	Kind         string                 `json:"kind"` // "struct", "interface", "alias", etc.
	Fields       map[string]string      `json:"fields,omitempty"`
	Methods      map[string]interface{} `json:"methods,omitempty"`
	IsExported   bool                   `json:"is_exported"`
	File         string                 `json:"file"`
	IsDeprecated bool                   `json:"is_deprecated"`
}

// ConstSignature represents a constant declaration
type ConstSignature struct {
	Type         string `json:"type"`
	Value        string `json:"value,omitempty"`
	IsExported   bool   `json:"is_exported"`
	File         string `json:"file"`
	IsDeprecated bool   `json:"is_deprecated"`
}

// VarSignature represents a variable declaration
type VarSignature struct {
	Type         string `json:"type"`
	IsExported   bool   `json:"is_exported"`
	File         string `json:"file"`
	IsDeprecated bool   `json:"is_deprecated"`
}

// APISignatures contains all the API signatures for a codebase
type APISignatures struct {
	Version    string          `json:"version"`
	Signatures []*APISignature `json:"signatures"`
}

func main() {
	// Parse command line flags
	dir := flag.String("dir", ".", "Directory to scan for Go files")
	output := flag.String("output", "api-signatures.json", "Output file for API signatures")
	version := flag.String("version", "current", "Version of the API being scanned")
	flag.Parse()

	// Collect API signatures
	signatures, err := collectAPISignatures(*dir, *version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting API signatures: %v\n", err)
		os.Exit(1)
	}

	// Write signatures to output file
	data, err := json.MarshalIndent(signatures, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling API signatures: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(*output, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing API signatures to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("API signatures written to %s\n", *output)
}

// collectAPISignatures scans a directory for Go files and extracts API signatures
func collectAPISignatures(dir, version string) (*APISignatures, error) {
	signatures := &APISignatures{
		Version:    version,
		Signatures: []*APISignature{},
	}

	// Function to process all Go files in a directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse Go file
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse file %s: %v\n", path, err)
			return nil
		}

		// Skip files in vendor, examples, or cmd directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/examples/") || strings.Contains(path, "/cmd/") {
			return nil
		}

		// Extract package path
		pkgPath := extractPackagePath(path, dir)
		if pkgPath == "" {
			return nil
		}

		// Process declarations in the file
		for _, decl := range file.Decls {
			processDeclaration(decl, file, fset, path, pkgPath, signatures)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return signatures, nil
}

// extractPackagePath determines the package path from the file path
func extractPackagePath(filePath, baseDir string) string {
	// Convert to absolute paths
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return ""
	}

	// Extract the relative path from the base directory
	relPath, err := filepath.Rel(absBaseDir, absFilePath)
	if err != nil {
		return ""
	}

	// Get the package directory
	pkgDir := filepath.Dir(relPath)
	if pkgDir == "." {
		pkgDir = ""
	}

	return pkgDir
}

// processDeclaration extracts API signatures from a declaration
func processDeclaration(decl ast.Decl, file *ast.File, fset *token.FileSet, filePath, pkgPath string, signatures *APISignatures) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		// Process function/method declaration
		if !d.Name.IsExported() {
			return
		}

		// Create function signature
		funcSig := &FuncSignature{
			Params:     extractFieldList(d.Type.Params),
			Results:    extractFieldList(d.Type.Results),
			IsExported: true,
			File:       filePath,
		}

		// Add receiver for methods
		if d.Recv != nil && len(d.Recv.List) > 0 {
			funcSig.Receiver = exprToString(d.Recv.List[0].Type)
		}

		// Check for deprecation comment
		funcSig.IsDeprecated = isDeprecated(d.Doc)

		// Add to signatures
		signatures.Signatures = append(signatures.Signatures, &APISignature{
			Package: pkgPath,
			Name:    d.Name.Name,
			Type:    "func",
			Details: funcSig,
		})

	case *ast.GenDecl:
		// Process type, const, or var declaration
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				// Process type declaration
				if !s.Name.IsExported() {
					continue
				}

				// Create type signature
				typeSig := &TypeSignature{
					IsExported:   true,
					File:         filePath,
					IsDeprecated: isDeprecated(d.Doc),
				}

				// Process different kinds of types
				switch t := s.Type.(type) {
				case *ast.StructType:
					typeSig.Kind = "struct"
					typeSig.Fields = extractStructFields(t)
				case *ast.InterfaceType:
					typeSig.Kind = "interface"
					typeSig.Methods = extractInterfaceMethods(t)
				case *ast.Ident:
					typeSig.Kind = "alias"
					// typeSig.AliasedType = t.Name
				default:
					typeSig.Kind = reflect.TypeOf(t).Elem().Name()
				}

				// Add to signatures
				signatures.Signatures = append(signatures.Signatures, &APISignature{
					Package: pkgPath,
					Name:    s.Name.Name,
					Type:    "type",
					Details: typeSig,
				})

			case *ast.ValueSpec:
				// Process const or var declaration
				for _, name := range s.Names {
					if !name.IsExported() {
						continue
					}

					// Determine if this is a const or var
					declType := "var"
					if d.Tok == token.CONST {
						declType = "const"
					}

					// Create signature
					valSig := &ConstSignature{
						Type:         exprToString(s.Type),
						IsExported:   true,
						File:         filePath,
						IsDeprecated: isDeprecated(d.Doc),
					}

					// Add to signatures
					signatures.Signatures = append(signatures.Signatures, &APISignature{
						Package: pkgPath,
						Name:    name.Name,
						Type:    declType,
						Details: valSig,
					})
				}
			}
		}
	}
}

// extractFieldList converts an ast.FieldList to a slice of strings
func extractFieldList(fields *ast.FieldList) []string {
	if fields == nil || len(fields.List) == 0 {
		return []string{}
	}

	var result []string
	for _, field := range fields.List {
		fieldType := exprToString(field.Type)
		if len(field.Names) == 0 {
			result = append(result, fieldType)
		} else {
			for range field.Names {
				result = append(result, fieldType)
			}
		}
	}
	return result
}

// extractStructFields extracts field definitions from a struct type
func extractStructFields(structType *ast.StructType) map[string]string {
	fields := make(map[string]string)
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		return fields
	}

	for _, field := range structType.Fields.List {
		fieldType := exprToString(field.Type)
		if len(field.Names) == 0 {
			// Embedded field
			fields["embedded"] = fieldType
		} else {
			for _, name := range field.Names {
				if name.IsExported() {
					fields[name.Name] = fieldType
				}
			}
		}
	}
	return fields
}

// extractInterfaceMethods extracts method signatures from an interface type
func extractInterfaceMethods(interfaceType *ast.InterfaceType) map[string]interface{} {
	methods := make(map[string]interface{})
	if interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		return methods
	}

	for _, method := range interfaceType.Methods.List {
		switch t := method.Type.(type) {
		case *ast.FuncType:
			if len(method.Names) > 0 && method.Names[0].IsExported() {
				methods[method.Names[0].Name] = map[string][]string{
					"params":  extractFieldList(t.Params),
					"results": extractFieldList(t.Results),
				}
			}
		case *ast.Ident:
			// Embedded interface
			methods["embedded"] = t.Name
		}
	}
	return methods
}

// exprToString converts an ast.Expr to a string representation
func exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan"
	case *ast.StructType:
		return "struct{}"
	case *ast.BasicLit:
		return t.Value
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// isDeprecated checks if a doc comment contains a deprecation notice
func isDeprecated(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, comment := range doc.List {
		if strings.Contains(comment.Text, "DEPRECATED") || strings.Contains(comment.Text, "Deprecated") {
			return true
		}
	}
	return false
}
