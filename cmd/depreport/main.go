// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package main provides a tool for scanning the codebase for deprecated functions
// and generating a report of all deprecated features.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DeprecatedFeature represents a deprecated feature found in the codebase
type DeprecatedFeature struct {
	Name         string
	FilePath     string
	Line         int
	DeprecatedIn string
	RemovalIn    string
	Guidance     string
}

var (
	// Regular expressions to extract information from code comments
	deprecatedRegex   = regexp.MustCompile(`(?i)Deprecated:?\s+(.+)`)
	deprecatedInRegex = regexp.MustCompile(`deprecated\s+in\s+([vV][\d\.]+)`)
	removalInRegex    = regexp.MustCompile(`(?:remove|removed|removal)\s+in\s+([vV][\d\.]+)`)

	// Command line flags
	srcDir     = flag.String("dir", ".", "Source directory to scan")
	outputFile = flag.String("o", "", "Output file for the report (defaults to stdout)")
)

func main() {
	flag.Parse()

	// Find all deprecated features
	features, err := findDeprecatedFeatures(*srcDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate the report
	report := generateReport(features)

	// Output the report
	if *outputFile == "" {
		fmt.Print(report)
	} else {
		err := os.WriteFile(*outputFile, []byte(report), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", *outputFile)
	}
}

func findDeprecatedFeatures(dir string) ([]DeprecatedFeature, error) {
	var features []DeprecatedFeature

	// Walk the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and vendor directory
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "/vendor/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse %s: %v\n", path, err)
			return nil
		}

		// Look for deprecated code comments
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				checkFuncForDeprecation(d, fset, path, &features)
			case *ast.GenDecl:
				checkGenDeclForDeprecation(d, fset, path, &features)
			}
		}

		// Check for LogWarning calls that indicate deprecation
		ast.Inspect(f, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				checkCallForDeprecation(call, fset, path, &features)
			}
			return true
		})

		return nil
	})

	return features, err
}

func checkFuncForDeprecation(fn *ast.FuncDecl, fset *token.FileSet, path string, features *[]DeprecatedFeature) {
	// Check for doc comments
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			if matches := deprecatedRegex.FindStringSubmatch(comment.Text); len(matches) > 1 {
				// Found a deprecated function
				feature := DeprecatedFeature{
					Name:     getFunctionName(fn),
					FilePath: path,
					Line:     fset.Position(fn.Pos()).Line,
				}

				// Extract version information from the comment
				feature.DeprecatedIn = extractVersionFromComment(deprecatedInRegex, comment.Text)
				feature.RemovalIn = extractVersionFromComment(removalInRegex, comment.Text)
				feature.Guidance = extractGuidance(comment.Text)

				*features = append(*features, feature)
				break
			}
		}
	}
}

func checkGenDeclForDeprecation(gd *ast.GenDecl, fset *token.FileSet, path string, features *[]DeprecatedFeature) {
	// Check for doc comments on type, var, and const declarations
	if gd.Doc != nil {
		for _, comment := range gd.Doc.List {
			if matches := deprecatedRegex.FindStringSubmatch(comment.Text); len(matches) > 1 {
				for _, spec := range gd.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						feature := DeprecatedFeature{
							Name:     s.Name.Name,
							FilePath: path,
							Line:     fset.Position(s.Pos()).Line,
						}
						feature.DeprecatedIn = extractVersionFromComment(deprecatedInRegex, comment.Text)
						feature.RemovalIn = extractVersionFromComment(removalInRegex, comment.Text)
						feature.Guidance = extractGuidance(comment.Text)
						*features = append(*features, feature)
					case *ast.ValueSpec:
						for _, name := range s.Names {
							feature := DeprecatedFeature{
								Name:     name.Name,
								FilePath: path,
								Line:     fset.Position(name.Pos()).Line,
							}
							feature.DeprecatedIn = extractVersionFromComment(deprecatedInRegex, comment.Text)
							feature.RemovalIn = extractVersionFromComment(removalInRegex, comment.Text)
							feature.Guidance = extractGuidance(comment.Text)
							*features = append(*features, feature)
						}
					}
				}
				break
			}
		}
	}
}

func checkCallForDeprecation(call *ast.CallExpr, fset *token.FileSet, path string, features *[]DeprecatedFeature) {
	// Check for calls to LogWarning or LogFeatureWarning
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			// Check if this is a call to the deprecation package
			if x.Name == "deprecation" &&
				(sel.Sel.Name == "LogWarning" || sel.Sel.Name == "LogFeatureWarning") {

				if len(call.Args) >= 4 { // LogWarning has at least 4 args
					// Try to extract literal arguments
					if len(call.Args) >= 2 {
						featureName := extractStringLiteral(call.Args[1])
						if featureName != "" {
							feature := DeprecatedFeature{
								Name:     featureName,
								FilePath: path,
								Line:     fset.Position(call.Pos()).Line,
							}

							if len(call.Args) >= 3 {
								feature.DeprecatedIn = extractStringLiteral(call.Args[2])
							}

							if len(call.Args) >= 4 {
								feature.RemovalIn = extractStringLiteral(call.Args[3])
							}

							if len(call.Args) >= 5 {
								feature.Guidance = extractStringLiteral(call.Args[4])
							}

							*features = append(*features, feature)
						}
					}
				}
			}
		}
	}
}

func getFunctionName(fn *ast.FuncDecl) string {
	if fn.Recv == nil {
		// Regular function
		return fn.Name.Name
	}

	// Method
	if len(fn.Recv.List) > 0 {
		typ := fn.Recv.List[0].Type
		var recvType string

		// Handle pointer receiver
		if star, ok := typ.(*ast.StarExpr); ok {
			if ident, ok := star.X.(*ast.Ident); ok {
				recvType = "*" + ident.Name
			}
		} else if ident, ok := typ.(*ast.Ident); ok {
			recvType = ident.Name
		}

		if recvType != "" {
			return fmt.Sprintf("%s.%s", recvType, fn.Name.Name)
		}
	}

	return fn.Name.Name
}

func extractVersionFromComment(regex *regexp.Regexp, comment string) string {
	matches := regex.FindStringSubmatch(comment)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractGuidance(comment string) string {
	// Look for guidance after "Use", "Instead", "Replace", etc.
	patterns := []string{
		`[Uu]se\s+([^\.]+)`,
		`[Ii]nstead[,\s]+use\s+([^\.]+)`,
		`[Rr]eplace\s+with\s+([^\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(comment)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func extractStringLiteral(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		// Remove surrounding quotes
		return strings.Trim(lit.Value, `"'`)
	}
	return ""
}

func generateReport(features []DeprecatedFeature) string {
	var sb strings.Builder

	sb.WriteString("# Deprecated Features Report\n\n")
	sb.WriteString("This report was generated automatically and lists all deprecated features found in the codebase.\n\n")

	// Group by removal version
	byRemovalVersion := make(map[string][]DeprecatedFeature)
	noRemovalVersion := []DeprecatedFeature{}

	for _, feature := range features {
		if feature.RemovalIn != "" {
			byRemovalVersion[feature.RemovalIn] = append(byRemovalVersion[feature.RemovalIn], feature)
		} else {
			noRemovalVersion = append(noRemovalVersion, feature)
		}
	}

	// Features with removal version
	sb.WriteString("## Features with Planned Removal\n\n")

	if len(byRemovalVersion) == 0 {
		sb.WriteString("No features with planned removal date found.\n\n")
	} else {
		for version, versionFeatures := range byRemovalVersion {
			sb.WriteString(fmt.Sprintf("### To be removed in %s\n\n", version))
			sb.WriteString("| Feature | File | Deprecated In | Guidance |\n")
			sb.WriteString("|---------|------|--------------|----------|\n")

			for _, feature := range versionFeatures {
				relPath, _ := filepath.Rel(*srcDir, feature.FilePath)
				sb.WriteString(fmt.Sprintf("| `%s` | %s:%d | %s | %s |\n",
					feature.Name,
					relPath,
					feature.Line,
					feature.DeprecatedIn,
					feature.Guidance,
				))
			}

			sb.WriteString("\n")
		}
	}

	// Features without removal version
	sb.WriteString("## Features without Planned Removal\n\n")

	if len(noRemovalVersion) == 0 {
		sb.WriteString("No features without planned removal date found.\n\n")
	} else {
		sb.WriteString("| Feature | File | Deprecated In | Guidance |\n")
		sb.WriteString("|---------|------|--------------|----------|\n")

		for _, feature := range noRemovalVersion {
			relPath, _ := filepath.Rel(*srcDir, feature.FilePath)
			sb.WriteString(fmt.Sprintf("| `%s` | %s:%d | %s | %s |\n",
				feature.Name,
				relPath,
				feature.Line,
				feature.DeprecatedIn,
				feature.Guidance,
			))
		}

		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("Total deprecated features: %d\n", len(features)))
	sb.WriteString(fmt.Sprintf("Features with planned removal: %d\n", len(features)-len(noRemovalVersion)))
	sb.WriteString(fmt.Sprintf("Features without planned removal: %d\n", len(noRemovalVersion)))

	return sb.String()
}
