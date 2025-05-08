// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package config

import (
	"reflect"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// TestConfigHasVersionCheckField ensures that the Config struct has a VersionCheck field
// This test would have caught the issue where VersionCheck was referenced in api_version.go
// but was missing from the Config struct
func TestConfigHasVersionCheckField(t *testing.T) {
	config := &Config{}
	configType := reflect.TypeOf(config).Elem()

	field, ok := configType.FieldByName("VersionCheck")
	if !ok {
		t.Fatal("Config struct missing VersionCheck field")
	}

	// Verify the field is of the correct type
	if field.Type != reflect.TypeOf((*core.VersionCheck)(nil)) {
		t.Fatalf("VersionCheck field should be of type *core.VersionCheck, got %v", field.Type)
	}
}

// TestConfigHasVersionCheckAccessors ensures that the Config struct has
// getter and setter methods for the VersionCheck field
func TestConfigHasVersionCheckAccessors(t *testing.T) {
	config := &Config{}
	configType := reflect.TypeOf(config)

	// Test for GetVersionCheck method
	_, ok := configType.MethodByName("GetVersionCheck")
	if !ok {
		t.Fatal("Config struct missing GetVersionCheck method")
	}

	// Test for SetVersionCheck method
	_, ok = configType.MethodByName("SetVersionCheck")
	if !ok {
		t.Fatal("Config struct missing SetVersionCheck method")
	}
}

// TestVersionCheckAccessorsWork ensures that the getter and setter methods
// actually work correctly
func TestVersionCheckAccessorsWork(t *testing.T) {
	config := &Config{}

	// Initially should be nil
	if config.GetVersionCheck() != nil {
		t.Fatal("New Config should have nil VersionCheck")
	}

	// Set a VersionCheck
	vc := core.NewVersionCheck()
	config.SetVersionCheck(vc)

	// Verify it was set correctly
	if config.GetVersionCheck() != vc {
		t.Fatal("VersionCheck not properly set or retrieved")
	}
}

// TestAPIVersionMethods ensures that the WithAPIVersionCheck and
// WithCustomAPIVersion methods work properly by using the accessors
func TestAPIVersionMethods(t *testing.T) {
	config := &Config{}

	// Test WithAPIVersionCheck enables version checking
	config.WithAPIVersionCheck(true)
	vc := config.GetVersionCheck()
	if vc == nil {
		t.Fatal("WithAPIVersionCheck should create a VersionCheck if one doesn't exist")
	}
	if !vc.Enabled() {
		t.Fatal("WithAPIVersionCheck(true) should enable version checking")
	}

	// Test that WithAPIVersionCheck(false) disables version checking
	config.WithAPIVersionCheck(false)
	if vc.Enabled() {
		t.Fatal("WithAPIVersionCheck(false) should disable version checking")
	}

	// Test WithCustomAPIVersion
	service, version := "test-service", "v1.0"
	config.WithCustomAPIVersion(service, version)
	customVersion, ok := vc.GetCustomVersion(service)
	if !ok {
		t.Fatalf("WithCustomAPIVersion should have set custom version for %s", service)
	}
	if customVersion != version {
		t.Fatalf("Expected custom version %s, got %s", version, customVersion)
	}
}
