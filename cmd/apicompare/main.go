// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// Package main provides a tool for comparing API signatures between versions
// to detect breaking changes and compatibility issues.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

// CompatibilityLevel defines the expected level of compatibility
type CompatibilityLevel string

const (
	// PatchLevel requires exact API compatibility (no changes)
	PatchLevel CompatibilityLevel = "patch"
	// MinorLevel allows additions but no removals or changes
	MinorLevel CompatibilityLevel = "minor"
	// MajorLevel allows any changes but wants to report them
	MajorLevel CompatibilityLevel = "major"
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

// ComparisonResult contains the results of comparing two API signatures
type ComparisonResult struct {
	CompatibilityLevel CompatibilityLevel `json:"compatibility_level"`
	OldVersion         string             `json:"old_version"`
	NewVersion         string             `json:"new_version"`
	Removals           []*APIChange       `json:"removals"`
	Additions          []*APIChange       `json:"additions"`
	Changes            []*APIChange       `json:"changes"`
	BreakingChanges    []*APIChange       `json:"breaking_changes"`
}

// APIChange represents a change in the API
type APIChange struct {
	Package    string      `json:"package"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	ChangeType string      `json:"change_type"` // "added", "removed", "changed", "breaking"
	Details    interface{} `json:"details,omitempty"`
}

func main() {
	// Parse command line flags
	oldFile := flag.String("old", "", "Old API signatures file")
	newFile := flag.String("new", "", "New API signatures file")
	outputFile := flag.String("output", "", "Output file for comparison results (optional)")
	level := flag.String("level", "minor", "Compatibility level (patch, minor, major)")
	flag.Parse()

	if *oldFile == "" || *newFile == "" {
		fmt.Println("Usage: apicompare -old <old_signatures.json> -new <new_signatures.json> [-output <output.json>] [-level <patch|minor|major>]")
		os.Exit(1)
	}

	// Load API signatures
	oldSigs, err := loadAPISignatures(*oldFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading old API signatures: %v\n", err)
		os.Exit(1)
	}

	newSigs, err := loadAPISignatures(*newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading new API signatures: %v\n", err)
		os.Exit(1)
	}

	// Determine compatibility level
	compatLevel := MinorLevel
	switch *level {
	case "patch":
		compatLevel = PatchLevel
	case "minor":
		compatLevel = MinorLevel
	case "major":
		compatLevel = MajorLevel
	default:
		fmt.Fprintf(os.Stderr, "Invalid compatibility level: %s. Using 'minor' as default.\n", *level)
	}

	// Compare API signatures
	result := compareAPISignatures(oldSigs, newSigs, compatLevel)

	// Print comparison results
	printComparisonResults(result)

	// Write results to file if specified
	if *outputFile != "" {
		err := writeComparisonResults(result, *outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing comparison results: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Comparison results written to %s\n", *outputFile)
	}

	// Exit with error if breaking changes were found for patch or minor level
	if len(result.BreakingChanges) > 0 && compatLevel != MajorLevel {
		fmt.Fprintf(os.Stderr, "Breaking changes found for compatibility level %s\n", compatLevel)
		os.Exit(1)
	}
}

// loadAPISignatures loads API signatures from a file
func loadAPISignatures(file string) (*APISignatures, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var signatures APISignatures
	err = json.Unmarshal(data, &signatures)
	if err != nil {
		return nil, err
	}

	return &signatures, nil
}

// compareAPISignatures compares two sets of API signatures
func compareAPISignatures(oldSigs, newSigs *APISignatures, level CompatibilityLevel) *ComparisonResult {
	result := &ComparisonResult{
		CompatibilityLevel: level,
		OldVersion:         oldSigs.Version,
		NewVersion:         newSigs.Version,
		Removals:           []*APIChange{},
		Additions:          []*APIChange{},
		Changes:            []*APIChange{},
		BreakingChanges:    []*APIChange{},
	}

	// Create maps of old and new signatures
	oldMap := make(map[string]*APISignature)
	newMap := make(map[string]*APISignature)

	for _, sig := range oldSigs.Signatures {
		key := getSignatureKey(sig)
		oldMap[key] = sig
	}

	for _, sig := range newSigs.Signatures {
		key := getSignatureKey(sig)
		newMap[key] = sig
	}

	// Find removals (in old but not in new)
	for key, oldSig := range oldMap {
		if _, ok := newMap[key]; !ok {
			// Skip methods on types that have been removed
			if oldSig.Type == "func" && strings.Contains(key, ".") {
				typeName := strings.Split(key, ".")[0]
				typeKey := fmt.Sprintf("%s|%s|type", oldSig.Package, typeName)
				if _, ok := oldMap[typeKey]; ok {
					if _, ok := newMap[typeKey]; !ok {
						// Type was removed, so methods removal is already captured
						continue
					}
				}
			}

			change := &APIChange{
				Package:    oldSig.Package,
				Name:       oldSig.Name,
				Type:       oldSig.Type,
				ChangeType: "removed",
				Details:    oldSig.Details,
			}
			result.Removals = append(result.Removals, change)

			// Removals are always breaking changes for patch and minor levels
			if level == PatchLevel || level == MinorLevel {
				result.BreakingChanges = append(result.BreakingChanges, change)
			}
		}
	}

	// Find additions (in new but not in old)
	for key, newSig := range newMap {
		if _, ok := oldMap[key]; !ok {
			change := &APIChange{
				Package:    newSig.Package,
				Name:       newSig.Name,
				Type:       newSig.Type,
				ChangeType: "added",
				Details:    newSig.Details,
			}
			result.Additions = append(result.Additions, change)

			// Changes to interfaces can be breaking for minor levels
			if newSig.Type == "type" {
				details, ok := newSig.Details.(map[string]interface{})
				if ok && details["kind"] == "interface" {
					// Adding methods to interfaces can break implementations
					if level == MinorLevel {
						result.BreakingChanges = append(result.BreakingChanges, change)
					}
				}
			}
		}
	}

	// Find changes (in both but different)
	for key, oldSig := range oldMap {
		if newSig, ok := newMap[key]; ok {
			if !signaturesEqual(oldSig, newSig) {
				change := &APIChange{
					Package:    oldSig.Package,
					Name:       oldSig.Name,
					Type:       oldSig.Type,
					ChangeType: "changed",
					Details: map[string]interface{}{
						"old": oldSig.Details,
						"new": newSig.Details,
					},
				}
				result.Changes = append(result.Changes, change)

				// All changes are breaking for patch level
				if level == PatchLevel {
					result.BreakingChanges = append(result.BreakingChanges, change)
				} else if level == MinorLevel {
					// For minor level, only some changes are breaking
					if isBreakingChange(oldSig, newSig) {
						change.ChangeType = "breaking"
						result.BreakingChanges = append(result.BreakingChanges, change)
					}
				}
			}
		}
	}

	// Sort results for consistent output
	sort.Slice(result.Removals, func(i, j int) bool {
		return getChangeKey(result.Removals[i]) < getChangeKey(result.Removals[j])
	})
	sort.Slice(result.Additions, func(i, j int) bool {
		return getChangeKey(result.Additions[i]) < getChangeKey(result.Additions[j])
	})
	sort.Slice(result.Changes, func(i, j int) bool {
		return getChangeKey(result.Changes[i]) < getChangeKey(result.Changes[j])
	})
	sort.Slice(result.BreakingChanges, func(i, j int) bool {
		return getChangeKey(result.BreakingChanges[i]) < getChangeKey(result.BreakingChanges[j])
	})

	return result
}

// getSignatureKey generates a unique key for a signature
func getSignatureKey(sig *APISignature) string {
	// For methods, include the receiver type
	if sig.Type == "func" {
		details := sig.Details.(map[string]interface{})
		if receiver, ok := details["receiver"]; ok && receiver != "" {
			return fmt.Sprintf("%s|%s.%s|%s", sig.Package, receiver, sig.Name, sig.Type)
		}
	}
	return fmt.Sprintf("%s|%s|%s", sig.Package, sig.Name, sig.Type)
}

// getChangeKey generates a unique key for an API change
func getChangeKey(change *APIChange) string {
	return fmt.Sprintf("%s|%s|%s", change.Package, change.Name, change.Type)
}

// signaturesEqual checks if two signatures are functionally equivalent
func signaturesEqual(oldSig, newSig *APISignature) bool {
	// Different types are definitely not equal
	if oldSig.Type != newSig.Type {
		return false
	}

	// Convert details to maps for comparison
	oldDetails, ok1 := oldSig.Details.(map[string]interface{})
	newDetails, ok2 := newSig.Details.(map[string]interface{})
	if !ok1 || !ok2 {
		// If we can't compare details, assume they're different
		return false
	}

	// Remove file paths before comparison
	oldDetailsCopy := make(map[string]interface{})
	newDetailsCopy := make(map[string]interface{})
	for k, v := range oldDetails {
		if k != "file" {
			oldDetailsCopy[k] = v
		}
	}
	for k, v := range newDetails {
		if k != "file" {
			newDetailsCopy[k] = v
		}
	}

	// For functions, compare parameters and results
	if oldSig.Type == "func" {
		oldParams, ok1 := oldDetailsCopy["params"].([]interface{})
		newParams, ok2 := newDetailsCopy["params"].([]interface{})
		if ok1 && ok2 && !reflect.DeepEqual(oldParams, newParams) {
			return false
		}

		oldResults, ok1 := oldDetailsCopy["results"].([]interface{})
		newResults, ok2 := newDetailsCopy["results"].([]interface{})
		if ok1 && ok2 && !reflect.DeepEqual(oldResults, newResults) {
			return false
		}

		return true
	}

	// For types, compare kind, fields, and methods
	if oldSig.Type == "type" {
		oldKind, ok1 := oldDetailsCopy["kind"].(string)
		newKind, ok2 := newDetailsCopy["kind"].(string)
		if ok1 && ok2 && oldKind != newKind {
			return false
		}

		oldFields, ok1 := oldDetailsCopy["fields"].(map[string]interface{})
		newFields, ok2 := newDetailsCopy["fields"].(map[string]interface{})
		if ok1 && ok2 && !reflect.DeepEqual(oldFields, newFields) {
			return false
		}

		oldMethods, ok1 := oldDetailsCopy["methods"].(map[string]interface{})
		newMethods, ok2 := newDetailsCopy["methods"].(map[string]interface{})
		if ok1 && ok2 && !reflect.DeepEqual(oldMethods, newMethods) {
			return false
		}

		return true
	}

	// For constants and variables, compare type
	if oldSig.Type == "const" || oldSig.Type == "var" {
		oldType, ok1 := oldDetailsCopy["type"].(string)
		newType, ok2 := newDetailsCopy["type"].(string)
		if ok1 && ok2 && oldType != newType {
			return false
		}

		return true
	}

	// Default to deep equality check
	return reflect.DeepEqual(oldDetailsCopy, newDetailsCopy)
}

// isBreakingChange checks if a change is breaking for minor version compatibility
func isBreakingChange(oldSig, newSig *APISignature) bool {
	// Convert details to maps for comparison
	oldDetails, ok1 := oldSig.Details.(map[string]interface{})
	newDetails, ok2 := newSig.Details.(map[string]interface{})
	if !ok1 || !ok2 {
		// If we can't compare details, assume it's a breaking change
		return true
	}

	// For functions, check for breaking parameter changes
	if oldSig.Type == "func" {
		// Adding parameters is breaking
		oldParams, ok1 := oldDetails["params"].([]interface{})
		newParams, ok2 := newDetails["params"].([]interface{})
		if ok1 && ok2 && len(newParams) > len(oldParams) {
			return true
		}

		// Changing parameter types is breaking
		for i := 0; i < len(oldParams) && i < len(newParams); i++ {
			if oldParams[i] != newParams[i] {
				return true
			}
		}

		// Removing return values is breaking
		oldResults, ok1 := oldDetails["results"].([]interface{})
		newResults, ok2 := newDetails["results"].([]interface{})
		if ok1 && ok2 && len(newResults) < len(oldResults) {
			return true
		}

		// Changing return types is breaking
		for i := 0; i < len(oldResults) && i < len(newResults); i++ {
			if oldResults[i] != newResults[i] {
				return true
			}
		}
	}

	// For types, check for breaking field and method changes
	if oldSig.Type == "type" {
		// Changing kind is breaking
		oldKind, ok1 := oldDetails["kind"].(string)
		newKind, ok2 := newDetails["kind"].(string)
		if ok1 && ok2 && oldKind != newKind {
			return true
		}

		// Removing fields is breaking
		oldFields, ok1 := oldDetails["fields"].(map[string]interface{})
		newFields, ok2 := newDetails["fields"].(map[string]interface{})
		if ok1 && ok2 {
			for field := range oldFields {
				if _, ok := newFields[field]; !ok {
					return true
				}
			}
		}

		// Changing field types is breaking
		if ok1 && ok2 {
			for field, oldType := range oldFields {
				if newType, ok := newFields[field]; ok && oldType != newType {
					return true
				}
			}
		}

		// Removing methods is breaking
		oldMethods, ok1 := oldDetails["methods"].(map[string]interface{})
		newMethods, ok2 := newDetails["methods"].(map[string]interface{})
		if ok1 && ok2 {
			for method := range oldMethods {
				if _, ok := newMethods[method]; !ok {
					return true
				}
			}
		}
	}

	// For constants and variables, check for type changes
	if oldSig.Type == "const" || oldSig.Type == "var" {
		oldType, ok1 := oldDetails["type"].(string)
		newType, ok2 := newDetails["type"].(string)
		if ok1 && ok2 && oldType != newType {
			return true
		}
	}

	return false
}

// printComparisonResults prints the comparison results to stdout
func printComparisonResults(result *ComparisonResult) {
	fmt.Printf("API Comparison: %s vs %s (Compatibility level: %s)\n\n", result.OldVersion, result.NewVersion, result.CompatibilityLevel)

	fmt.Printf("Additions: %d\n", len(result.Additions))
	for _, change := range result.Additions {
		fmt.Printf("  + %s %s.%s\n", change.Type, change.Package, change.Name)
	}
	fmt.Println()

	fmt.Printf("Removals: %d\n", len(result.Removals))
	for _, change := range result.Removals {
		fmt.Printf("  - %s %s.%s\n", change.Type, change.Package, change.Name)
	}
	fmt.Println()

	fmt.Printf("Changes: %d\n", len(result.Changes))
	for _, change := range result.Changes {
		fmt.Printf("  ~ %s %s.%s\n", change.Type, change.Package, change.Name)
	}
	fmt.Println()

	fmt.Printf("Breaking changes: %d\n", len(result.BreakingChanges))
	for _, change := range result.BreakingChanges {
		fmt.Printf("  ! %s %s.%s (%s)\n", change.Type, change.Package, change.Name, change.ChangeType)
	}
	fmt.Println()

	if result.CompatibilityLevel != MajorLevel && len(result.BreakingChanges) > 0 {
		fmt.Printf("COMPATIBILITY WARNING: Found %d breaking changes for %s compatibility level\n", len(result.BreakingChanges), result.CompatibilityLevel)
	} else {
		fmt.Printf("API is compatible at %s level\n", result.CompatibilityLevel)
	}
}

// writeComparisonResults writes comparison results to a file
func writeComparisonResults(result *ComparisonResult, outputFile string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
