package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crossplane-docs",
	Short: "Generate documentation for Crossplane resources",
	Long: `crossplane-docs generates terraform-docs style markdown documentation 
for Crossplane resources including XRDs, Compositions, and more.

Parse OpenAPI schemas and resource definitions to create clean, readable
documentation tables with field names, types, descriptions, defaults, and validations.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}
