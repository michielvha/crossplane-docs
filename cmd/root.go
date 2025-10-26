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
	Use:   "crossplane-xrd-docs",
	Short: "Generate documentation for Crossplane XRDs",
	Long: `crossplane-xrd-docs is a CLI tool that generates terraform-docs style 
markdown documentation from Crossplane XRD (CompositeResourceDefinition) files.

It parses the OpenAPI v3 schema from XRD YAML files and outputs formatted
markdown tables with field names, types, descriptions, defaults, and validation rules.`,
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
