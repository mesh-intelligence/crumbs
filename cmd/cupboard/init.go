// Init command for the cupboard CLI.
// Implements: prd009-cupboard-cli R2.1, R10; prd010-configuration-directories R2, R5.
package main

import (
	"fmt"
	"os"

	"github.com/mesh-intelligence/crumbs/internal/sqlite"
	"github.com/mesh-intelligence/crumbs/pkg/types"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize cupboard storage",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := resolveDataDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "init:", err)
			os.Exit(exitSysError)
		}

		cfg := types.Config{
			Backend: "sqlite",
			DataDir: dataDir,
		}

		backend := sqlite.NewBackend()
		if err := backend.Attach(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "init:", err)
			os.Exit(exitSysError)
		}
		defer backend.Detach()

		fmt.Println("Cupboard initialized successfully")
		return nil
	},
}
