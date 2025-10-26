package cmd

import (
	"fmt"
	"os"

	"github.com/michielvha/crossplane-docs/pkg/generator"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	showNested bool
)

// xrdCmd represents the xrd command
var xrdCmd = &cobra.Command{
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
	RunE: runXRD,
}

func init() {
	rootCmd.AddCommand(xrdCmd)

	xrdCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	xrdCmd.Flags().BoolVar(&showNested, "show-nested", true, "Show nested object structures")
}

func runXRD(cmd *cobra.Command, args []string) error {
	xrdFile := args[0]

	// Check if file exists
	if _, err := os.Stat(xrdFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", xrdFile)
	}

	// Generate documentation
	gen := generator.New()
	markdown, err := gen.GenerateFromFile(xrdFile, generator.Options{
		ShowNested: showNested,
	})
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Output
	if outputFile == "" {
		fmt.Println(markdown)
	} else {
		if err := os.WriteFile(outputFile, []byte(markdown), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Documentation generated successfully: %s\n", outputFile)
	}

	return nil
}
