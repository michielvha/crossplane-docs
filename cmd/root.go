package cmd

import (
	"fmt"
	"os"

	"github.com/michielvha/crossplane-docs/internal/engine"
	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// NewRootCommand returns a fresh command tree for each invocation.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "crossplane-docs",
		Short:         "Generate documentation for Crossplane resources",
		Long:          `Generate Markdown reference tables from Crossplane XRDs and static Composition definitions.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newXRDCommand(), newCompositionCommand(), newBatchCommand())
	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code := 1
		if cmd, _, findErr := NewRootCommand().Find(os.Args[1:]); findErr == nil && cmd != nil {
			for _, arg := range os.Args[1:] {
				if arg == "--check" || arg == "--check=true" {
					if !engine.IsDrift(err) {
						code = 2
					}
					break
				}
			}
		}
		os.Exit(code)
	}
}
