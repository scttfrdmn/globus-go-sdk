// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scttfrdmn/globus-go-sdk/cmd/globus-cli/auth"
	"github.com/scttfrdmn/globus-go-sdk/cmd/globus-cli/transfer"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Usage       string
	Execute     func(args []string) error
}

// Default configuration directory
var configDir string

func init() {
	// Determine the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining home directory: %v\n", err)
		os.Exit(1)
	}

	// Set the default configuration directory
	configDir = filepath.Join(homeDir, ".globus-cli")

	// Create the configuration directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating configuration directory: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Define available commands
	commands := []Command{
		{
			Name:        "login",
			Description: "Log in to Globus",
			Usage:       "globus-cli login",
			Execute:     auth.LoginCommand,
		},
		{
			Name:        "token",
			Description: "Display or manage tokens",
			Usage:       "globus-cli token [info|revoke] [token]",
			Execute:     auth.TokenCommand,
		},
		{
			Name:        "logout",
			Description: "Log out from Globus",
			Usage:       "globus-cli logout",
			Execute:     auth.LogoutCommand,
		},
		{
			Name:        "ls",
			Description: "List files on an endpoint",
			Usage:       "globus-cli ls <endpoint-id> <path>",
			Execute:     transfer.ListCommand,
		},
		{
			Name:        "transfer",
			Description: "Transfer files between endpoints",
			Usage:       "globus-cli transfer <source-endpoint-id> <source-path> <dest-endpoint-id> <dest-path>",
			Execute:     transfer.TransferCommand,
		},
		{
			Name:        "status",
			Description: "Check the status of a transfer task",
			Usage:       "globus-cli status <task-id>",
			Execute:     transfer.StatusCommand,
		},
	}

	// If no arguments are provided, show usage
	if len(os.Args) < 2 {
		showUsage(commands)
		return
	}

	// Get the command name from arguments
	cmdName := os.Args[1]

	// Handle help command
	if cmdName == "help" {
		if len(os.Args) > 2 {
			showCommandHelp(commands, os.Args[2])
		} else {
			showUsage(commands)
		}
		return
	}

	// Find and execute the command
	for _, cmd := range commands {
		if cmd.Name == cmdName {
			if err := cmd.Execute(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// If we got here, the command wasn't found
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
	showUsage(commands)
	os.Exit(1)
}

// showUsage displays usage information for all commands
func showUsage(commands []Command) {
	fmt.Println("Globus CLI - A command line interface for Globus")
	fmt.Println("\nUsage:")
	fmt.Println("  globus-cli <command> [arguments]")
	fmt.Println("\nAvailable commands:")

	for _, cmd := range commands {
		fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
	}

	fmt.Println("\nFor more information on a command, use:")
	fmt.Println("  globus-cli help <command>")
}

// showCommandHelp displays help for a specific command
func showCommandHelp(commands []Command, cmdName string) {
	for _, cmd := range commands {
		if cmd.Name == cmdName {
			fmt.Printf("Command: %s\n", cmd.Name)
			fmt.Printf("Description: %s\n", cmd.Description)
			fmt.Printf("Usage: %s\n", cmd.Usage)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
	showUsage(commands)
}
