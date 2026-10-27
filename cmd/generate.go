package cmd

import (
	"fmt"

	"github.com/michielvha/crossplane-docs/pkg/generator"
	"github.com/spf13/cobra"
)

func newXRDCommand() *cobra.Command {
	var outputFile string
	var showNested bool
	command := &cobra.Command{
		Use:   "xrd [xrd-file]",
		Short: "Generate documentation from an XRD file",
		Long: `Generate markdown documentation from a Crossplane XRD (CompositeResourceDefinition) YAML file.

Examples:
  # Generate docs and print to stdout
  crossplane-docs xrd xrd.yaml

  # Generate docs and save to file
  crossplane-docs xrd xrd.yaml -o README.md
  
  # Hide nested object structures (if you want a flatter view)
  crossplane-docs xrd xrd.yaml --show-nested=false`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runXRD(cmd, args, outputFile, showNested)
		},
	}
	command.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	command.Flags().BoolVar(&showNested, "show-nested", true, "Show nested object structures")
	return command
}

func runXRD(cmd *cobra.Command, args []string, outputFile string, showNested bool) error {
	xrdFile := args[0]

	// Generate documentation
	gen := generator.New()
	markdown, err := gen.GenerateFromFile(xrdFile, generator.Options{
		ShowNested: showNested,
	})
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	return writeOutput(cmd, outputFile, markdown)
}
