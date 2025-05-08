#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# API compatibility verification script for the Globus Go SDK
# This script compares two versions of the SDK to detect potentially breaking API changes

set -e

print_help() {
    echo "Usage: $0 <old_version> <new_version> [-level=<level>] [-verbose]"
    echo ""
    echo "Options:"
    echo "  -level=<level>   Compatibility level to check: patch, minor, or major (default: minor)"
    echo "  -verbose         Show detailed output for all checks"
    echo ""
    echo "Examples:"
    echo "  $0 v0.9.10 v0.9.11 -level=patch    # Check for unexpected changes in a patch release"
    echo "  $0 v0.9.0 v0.10.0 -level=minor     # Check for breaking changes in a minor release"
    echo "  $0 v0.9.0 v1.0.0 -level=major      # Generate compatibility report for a major release"
    echo ""
    exit 1
}

# Process command line arguments
if [ $# -lt 2 ]; then
    print_help
fi

OLD_VERSION=$1
NEW_VERSION=$2
shift 2

LEVEL="minor"
VERBOSE=0

for arg in "$@"; do
    case $arg in
        -level=*)
            LEVEL="${arg#*=}"
            ;;
        -verbose)
            VERBOSE=1
            ;;
        *)
            echo "Unknown argument: $arg"
            print_help
            ;;
    esac
done

# Validate level
if [[ ! "$LEVEL" =~ ^(patch|minor|major)$ ]]; then
    echo "Invalid level: $LEVEL. Must be one of: patch, minor, major"
    exit 1
fi

# Setup working directories
WORK_DIR=$(mktemp -d)
OLD_DIR="$WORK_DIR/old"
NEW_DIR="$WORK_DIR/new"
REPORT_DIR="$WORK_DIR/report"

echo "Creating working directory: $WORK_DIR"
mkdir -p "$OLD_DIR" "$NEW_DIR" "$REPORT_DIR"

cleanup() {
    echo "Cleaning up temporary directory: $WORK_DIR"
    rm -rf "$WORK_DIR"
}

trap cleanup EXIT

echo "Cloning old version ($OLD_VERSION)..."
git clone -q --depth 1 --branch "$OLD_VERSION" https://github.com/scttfrdmn/globus-go-sdk.git "$OLD_DIR" || {
    echo "Error: Failed to clone repository at version $OLD_VERSION"
    echo "Verify that the tag exists: git ls-remote --tags https://github.com/scttfrdmn/globus-go-sdk.git"
    exit 1
}

echo "Cloning new version ($NEW_VERSION)..."
if [[ "$NEW_VERSION" == "HEAD" ]]; then
    git clone -q --depth 1 https://github.com/scttfrdmn/globus-go-sdk.git "$NEW_DIR"
else
    git clone -q --depth 1 --branch "$NEW_VERSION" https://github.com/scttfrdmn/globus-go-sdk.git "$NEW_DIR" || {
        echo "Error: Failed to clone repository at version $NEW_VERSION"
        echo "Verify that the tag exists: git ls-remote --tags https://github.com/scttfrdmn/globus-go-sdk.git"
        exit 1
    }
fi

# Generate API signatures for both versions
echo "Generating API signatures for $OLD_VERSION..."
(cd "$OLD_DIR" && go list ./... | grep -v internal | grep -v cmd | grep -v test | grep -v examples) > "$REPORT_DIR/old_packages.txt"

echo "Generating API signatures for $NEW_VERSION..."
(cd "$NEW_DIR" && go list ./... | grep -v internal | grep -v cmd | grep -v test | grep -v examples) > "$REPORT_DIR/new_packages.txt"

# Analysis functions
analyze_function_signatures() {
    local old_pkg=$1
    local new_pkg=$2
    local pkg_name=${old_pkg##*/}
    
    echo "Analyzing function signatures in package $pkg_name..."
    
    # Extract function signatures
    (cd "$OLD_DIR" && go doc -all "$old_pkg" | grep -E "^func " | sort) > "$REPORT_DIR/${pkg_name}_old_funcs.txt"
    (cd "$NEW_DIR" && go doc -all "$new_pkg" | grep -E "^func " | sort) > "$REPORT_DIR/${pkg_name}_new_funcs.txt"
    
    # Compare
    if [ "$LEVEL" = "patch" ]; then
        # In patch releases, function signatures should not change
        diff "$REPORT_DIR/${pkg_name}_old_funcs.txt" "$REPORT_DIR/${pkg_name}_new_funcs.txt" > "$REPORT_DIR/${pkg_name}_func_diff.txt" || {
            echo "WARNING: Function signatures changed in patch release for package $pkg_name!"
            if [ $VERBOSE -eq 1 ]; then
                cat "$REPORT_DIR/${pkg_name}_func_diff.txt"
            else
                echo "Run with -verbose to see details"
            fi
            return 1
        }
    elif [ "$LEVEL" = "minor" ]; then
        # In minor releases, we can add functions but not remove or change existing ones
        grep -Fxvf "$REPORT_DIR/${pkg_name}_new_funcs.txt" "$REPORT_DIR/${pkg_name}_old_funcs.txt" > "$REPORT_DIR/${pkg_name}_removed_funcs.txt" || true
        
        if [ -s "$REPORT_DIR/${pkg_name}_removed_funcs.txt" ]; then
            echo "WARNING: Functions were removed in minor release for package $pkg_name!"
            if [ $VERBOSE -eq 1 ]; then
                cat "$REPORT_DIR/${pkg_name}_removed_funcs.txt"
            else
                echo "Run with -verbose to see details"
            fi
            return 1
        fi
    elif [ "$LEVEL" = "major" ]; then
        # In major releases, we document all changes
        grep -Fxvf "$REPORT_DIR/${pkg_name}_new_funcs.txt" "$REPORT_DIR/${pkg_name}_old_funcs.txt" > "$REPORT_DIR/${pkg_name}_removed_funcs.txt" || true
        grep -Fxvf "$REPORT_DIR/${pkg_name}_old_funcs.txt" "$REPORT_DIR/${pkg_name}_new_funcs.txt" > "$REPORT_DIR/${pkg_name}_added_funcs.txt" || true
        
        echo "Functions removed in package $pkg_name: $(wc -l < "$REPORT_DIR/${pkg_name}_removed_funcs.txt")"
        echo "Functions added in package $pkg_name: $(wc -l < "$REPORT_DIR/${pkg_name}_added_funcs.txt")"
        
        if [ $VERBOSE -eq 1 ] && [ -s "$REPORT_DIR/${pkg_name}_removed_funcs.txt" ]; then
            echo "Removed functions:"
            cat "$REPORT_DIR/${pkg_name}_removed_funcs.txt"
        fi
    fi
    
    return 0
}

analyze_type_definitions() {
    local old_pkg=$1
    local new_pkg=$2
    local pkg_name=${old_pkg##*/}
    
    echo "Analyzing type definitions in package $pkg_name..."
    
    # Extract type definitions
    (cd "$OLD_DIR" && go doc -all "$old_pkg" | grep -E "^type " | sort) > "$REPORT_DIR/${pkg_name}_old_types.txt"
    (cd "$NEW_DIR" && go doc -all "$new_pkg" | grep -E "^type " | sort) > "$REPORT_DIR/${pkg_name}_new_types.txt"
    
    # Compare
    if [ "$LEVEL" = "patch" ]; then
        # In patch releases, type definitions should not change
        diff "$REPORT_DIR/${pkg_name}_old_types.txt" "$REPORT_DIR/${pkg_name}_new_types.txt" > "$REPORT_DIR/${pkg_name}_type_diff.txt" || {
            echo "WARNING: Type definitions changed in patch release for package $pkg_name!"
            if [ $VERBOSE -eq 1 ]; then
                cat "$REPORT_DIR/${pkg_name}_type_diff.txt"
            else
                echo "Run with -verbose to see details"
            fi
            return 1
        }
    elif [ "$LEVEL" = "minor" ]; then
        # In minor releases, we can add types but not remove existing ones
        grep -Fxvf "$REPORT_DIR/${pkg_name}_new_types.txt" "$REPORT_DIR/${pkg_name}_old_types.txt" > "$REPORT_DIR/${pkg_name}_removed_types.txt" || true
        
        if [ -s "$REPORT_DIR/${pkg_name}_removed_types.txt" ]; then
            echo "WARNING: Types were removed in minor release for package $pkg_name!"
            if [ $VERBOSE -eq 1 ]; then
                cat "$REPORT_DIR/${pkg_name}_removed_types.txt"
            else
                echo "Run with -verbose to see details"
            fi
            return 1
        fi
    elif [ "$LEVEL" = "major" ]; then
        # In major releases, we document all changes
        grep -Fxvf "$REPORT_DIR/${pkg_name}_new_types.txt" "$REPORT_DIR/${pkg_name}_old_types.txt" > "$REPORT_DIR/${pkg_name}_removed_types.txt" || true
        grep -Fxvf "$REPORT_DIR/${pkg_name}_old_types.txt" "$REPORT_DIR/${pkg_name}_new_types.txt" > "$REPORT_DIR/${pkg_name}_added_types.txt" || true
        
        echo "Types removed in package $pkg_name: $(wc -l < "$REPORT_DIR/${pkg_name}_removed_types.txt")"
        echo "Types added in package $pkg_name: $(wc -l < "$REPORT_DIR/${pkg_name}_added_types.txt")"
        
        if [ $VERBOSE -eq 1 ] && [ -s "$REPORT_DIR/${pkg_name}_removed_types.txt" ]; then
            echo "Removed types:"
            cat "$REPORT_DIR/${pkg_name}_removed_types.txt"
        fi
    fi
    
    return 0
}

analyze_constants() {
    local old_pkg=$1
    local new_pkg=$2
    local pkg_name=${old_pkg##*/}
    
    echo "Analyzing constants in package $pkg_name..."
    
    # Extract constants (this is a simplified approach)
    (cd "$OLD_DIR" && go doc -all "$old_pkg" | grep -E "^const " -A 100 | grep -E "^\t" | sort) > "$REPORT_DIR/${pkg_name}_old_consts.txt" || true
    (cd "$NEW_DIR" && go doc -all "$new_pkg" | grep -E "^const " -A 100 | grep -E "^\t" | sort) > "$REPORT_DIR/${pkg_name}_new_consts.txt" || true
    
    # Compare
    if [ "$LEVEL" = "patch" ] || [ "$LEVEL" = "minor" ]; then
        # In patch and minor releases, constants should not change value
        # (This is a simplified check - in reality we'd want to check values not just names)
        diff "$REPORT_DIR/${pkg_name}_old_consts.txt" "$REPORT_DIR/${pkg_name}_new_consts.txt" > "$REPORT_DIR/${pkg_name}_const_diff.txt" || {
            echo "WARNING: Constants changed in package $pkg_name!"
            if [ $VERBOSE -eq 1 ]; then
                cat "$REPORT_DIR/${pkg_name}_const_diff.txt"
            else
                echo "Run with -verbose to see details"
            fi
            return 1
        }
    elif [ "$LEVEL" = "major" ]; then
        # In major releases, we document all changes
        if [ $VERBOSE -eq 1 ]; then
            diff "$REPORT_DIR/${pkg_name}_old_consts.txt" "$REPORT_DIR/${pkg_name}_new_consts.txt" || echo "Constants changed in package $pkg_name"
        fi
    fi
    
    return 0
}

# Process each package
errors=0
while read -r old_pkg; do
    # Find corresponding package in new version
    pkg_name=${old_pkg#github.com/scttfrdmn/globus-go-sdk/}
    new_pkg=$(grep "${pkg_name}$" "$REPORT_DIR/new_packages.txt" || echo "")
    
    if [ -z "$new_pkg" ]; then
        echo "WARNING: Package $old_pkg no longer exists in $NEW_VERSION!"
        if [ "$LEVEL" = "patch" ] || [ "$LEVEL" = "minor" ]; then
            echo "ERROR: Removing packages is not allowed in $LEVEL releases!"
            ((errors++))
        else
            echo "Note: Package removal is allowed in major releases but should be documented"
        fi
        continue
    fi
    
    analyze_function_signatures "$old_pkg" "$new_pkg" || ((errors++))
    analyze_type_definitions "$old_pkg" "$new_pkg" || ((errors++))
    analyze_constants "$old_pkg" "$new_pkg" || ((errors++))
done < "$REPORT_DIR/old_packages.txt"

# Check for new packages
while read -r new_pkg; do
    pkg_name=${new_pkg#github.com/scttfrdmn/globus-go-sdk/}
    old_pkg=$(grep "${pkg_name}$" "$REPORT_DIR/old_packages.txt" || echo "")
    
    if [ -z "$old_pkg" ]; then
        echo "INFO: New package added: $new_pkg"
    fi
done < "$REPORT_DIR/new_packages.txt"

# Generate summary report
echo ""
echo "API Compatibility Summary Report"
echo "================================"
echo "Old version: $OLD_VERSION"
echo "New version: $NEW_VERSION"
echo "Compatibility level: $LEVEL"
echo ""

if [ $errors -eq 0 ]; then
    echo "✅ No compatibility issues found at $LEVEL level!"
    exit 0
else
    echo "❌ Found $errors potential compatibility issues!"
    if [ "$LEVEL" = "major" ]; then
        echo "Note: These issues are acceptable for a major release but should be documented."
        exit 0
    else
        echo "These issues should be fixed or documented before release."
        exit 1
    fi
fi