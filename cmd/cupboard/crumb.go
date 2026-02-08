// Crumb subcommand group for managing work items.
// Implements: prd003-crumbs-interface (Crumb entity operations);
//
//	docs/ARCHITECTURE § CLI.
package main

import (
	"github.com/spf13/cobra"
)

// jsonOutput controls whether to output JSON instead of human-readable format.
var jsonOutput bool

var crumbCmd = &cobra.Command{
	Use:   "crumb",
	Short: "Manage crumbs (work items)",
	Long: `Crumb provides commands for managing work items in the cupboard.

Crumbs are individual work items with names, states, and properties.

Commands:
  add     Create a new crumb
  get     Retrieve a crumb by ID
  list    List all crumbs
  delete  Delete a crumb by ID`,
}

func init() {
	// Add --json flag to crumb parent command (inherited by subcommands)
	crumbCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Add subcommands
	crumbCmd.AddCommand(crumbAddCmd)
	crumbCmd.AddCommand(crumbGetCmd)
	crumbCmd.AddCommand(crumbListCmd)
	crumbCmd.AddCommand(crumbDeleteCmd)

	// Register crumb command with root
	rootCmd.AddCommand(crumbCmd)
}
