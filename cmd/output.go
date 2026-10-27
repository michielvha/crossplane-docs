package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func writeOutput(command *cobra.Command, filename, markdown string) error {
	if filename == "" {
		_, err := io.WriteString(command.OutOrStdout(), markdown)
		return err
	}
	if err := os.WriteFile(filename, []byte(markdown), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	_, err := fmt.Fprintf(command.ErrOrStderr(), "Documentation generated successfully: %s\n", filename)
	return err
}
