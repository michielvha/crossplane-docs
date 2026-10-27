package cmd

import (
	"fmt"

	"github.com/michielvha/crossplane-docs/pkg/composition"
	"github.com/spf13/cobra"
)

func newCompositionCommand() *cobra.Command {
	var compOutputFile string
	var showPatches bool
	command := &cobra.Command{
		Use:   "composition [composition-file]",
		Short: "Generate documentation from a Composition file",
		Long: `Generate markdown documentation from a Crossplane Composition YAML file.

Shows static resource definitions, field mappings, patches, and transformations.

Examples:
  # Generate docs and print to stdout
  crossplane-docs composition composition.yaml

  # Generate docs and save to file
  crossplane-docs composition composition.yaml -o COMPOSITION.md
  
  # Hide patch details
  crossplane-docs composition composition.yaml --show-patches=false`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComposition(cmd, args, compOutputFile, showPatches)
		},
	}
	command.Flags().StringVarP(&compOutputFile, "output", "o", "", "Output file (default: stdout)")
	command.Flags().BoolVar(&showPatches, "show-patches", true, "Show patch details and transformations")
	return command
}

func runComposition(cmd *cobra.Command, args []string, compOutputFile string, showPatches bool) error {
	compositionFile := args[0]

	// Generate documentation
	gen := composition.New()
	markdown, err := gen.GenerateFromFile(compositionFile, composition.Options{
		ShowPatches: showPatches,
	})
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	return writeOutput(cmd, compOutputFile, markdown)
}
