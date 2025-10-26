package cmd

import (
	"fmt"
	"os"

	"github.com/michielvha/crossplane-docs/pkg/composition"
	"github.com/spf13/cobra"
)

var (
	compOutputFile string
	showPatches    bool
)

// compositionCmd represents the composition command
var compositionCmd = &cobra.Command{
	Use:   "composition [composition-file]",
	Short: "Generate documentation from a Composition file",
	Long: `Generate markdown documentation from a Crossplane Composition YAML file.

Shows what managed resources are created, field mappings, patches, and transformations.

Examples:
  # Generate docs and print to stdout
  crossplane-docs composition composition.yaml

  # Generate docs and save to file
  crossplane-docs composition composition.yaml -o COMPOSITION.md
  
  # Hide patch details
  crossplane-docs composition composition.yaml --show-patches=false`,
	Args: cobra.ExactArgs(1),
	RunE: runComposition,
}

func init() {
	rootCmd.AddCommand(compositionCmd)

	compositionCmd.Flags().StringVarP(&compOutputFile, "output", "o", "", "Output file (default: stdout)")
	compositionCmd.Flags().BoolVar(&showPatches, "show-patches", true, "Show patch details and transformations")
}

func runComposition(cmd *cobra.Command, args []string) error {
	compositionFile := args[0]

	// Check if file exists
	if _, err := os.Stat(compositionFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", compositionFile)
	}

	// Generate documentation
	gen := composition.New()
	markdown, err := gen.GenerateFromFile(compositionFile, composition.Options{
		ShowPatches: showPatches,
	})
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Output
	if compOutputFile == "" {
		fmt.Println(markdown)
	} else {
		if err := os.WriteFile(compOutputFile, []byte(markdown), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Documentation generated successfully: %s\n", compOutputFile)
	}

	return nil
}
