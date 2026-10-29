package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/michielvha/crossplane-docs/internal/engine"
	"github.com/spf13/cobra"
)

func newXRDCommand() *cobra.Command { return newGenerateCommand("xrd", "CompositeResourceDefinition") }

func newCompositionCommand() *cobra.Command { return newGenerateCommand("composition", "Composition") }
func newBatchCommand() *cobra.Command       { return newGenerateCommand("generate", "") }
func newGenerateCommand(name, kind string) *cobra.Command {
	var configFile, format, output, version, templateFile, outputMode, rendered, extensions string
	var allVersions, check, render, validate, nested, patches, recursive bool
	var schemas, include, exclude []string
	runtime := engine.RenderOptions{Timeout: 2 * time.Minute, Name: "example", CrossplaneVersion: "v2.4.3", CrossplaneImage: engine.DefaultCrossplaneImage}
	command := &cobra.Command{Use: name + " PATH", Short: "Generate configurable documentation from Crossplane manifests", Args: cobra.ExactArgs(1)}
	var htmlLayout string
	flags := command.Flags()
	flags.StringVar(&configFile, "config", "", "Configuration file (default: .crossplane-docs.yml)")
	flags.StringVar(&htmlLayout, "html-layout", "reference", "HTML layout: reference, explorer or compact")
	flags.StringVar(&format, "format", "markdown", "Output: markdown, html, asciidoc, json or yaml")
	flags.StringVarP(&output, "output", "o", "", "Output file (default: stdout)")
	flags.StringVar(&outputMode, "output-mode", "replace", "replace or inject")
	flags.StringVar(&version, "api-version", "", "Select an XRD version")
	flags.BoolVar(&allVersions, "all-versions", false, "Document every XRD version")
	flags.BoolVar(&check, "check", false, "Check for documentation drift without writing")
	flags.BoolVar(&nested, "show-nested", true, "Expand nested object, array and map schemas")
	flags.BoolVar(&patches, "show-patches", true, "Show patch mappings and transformations")
	flags.BoolVar(&recursive, "recursive", true, "Discover YAML files recursively")
	flags.StringSliceVar(&include, "include", nil, "Included YAML path or filename patterns")
	flags.StringSliceVar(&exclude, "exclude", nil, "Excluded YAML path or filename patterns")
	flags.StringSliceVar(&schemas, "schema", nil, "Explicit local files allowed for schema references")
	flags.StringVar(&templateFile, "template", "", "Custom output template file")
	flags.BoolVar(&render, "render", false, "Execute the function pipeline using the Crossplane CLI")
	flags.StringVar(&rendered, "rendered", "", "Read saved rendered YAML")
	flags.StringVar(&runtime.Binary, "crossplane", "", "Crossplane CLI executable (default: crossplane)")
	flags.StringVar(&runtime.XR, "xr", "", "Sample composite resource for rendering")
	flags.StringVar(&runtime.Composition, "composition", "", "Composition file for rendering a batch")
	flags.StringVar(&runtime.Functions, "functions", "", "Function definitions for rendering")
	flags.StringVar(&runtime.Observed, "observed-resources", "", "Mock observed resources")
	flags.StringArrayVar(&runtime.Required, "required-resources", nil, "Mock required resource files (repeatable)")
	flags.StringVar(&runtime.Schemas, "required-schemas", "", "Schema directory supplied to functions")
	flags.StringArrayVar(&runtime.Context, "context-values", nil, "Pipeline context KEY=JSON (repeatable)")
	flags.StringVar(&runtime.Name, "scenario", "example", "Render scenario name")
	flags.StringVar(&runtime.CrossplaneImage, "crossplane-image", engine.DefaultCrossplaneImage, "Pinned Crossplane engine image")
	flags.StringVar(&runtime.CrossplaneVersion, "crossplane-version", "v2.4.3", "Pinned Crossplane engine version")
	flags.DurationVar(&runtime.Timeout, "timeout", 2*time.Minute, "Maximum renderer/validator duration")
	flags.BoolVar(&validate, "validate", false, "Validate XR or rendered output against supplied schemas")
	flags.StringVar(&extensions, "extensions", "", "XRD/CRD schema file or directory for validation")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		config, err := engine.LoadConfig(args[0], configFile)
		if err != nil {
			return err
		}
		if flags.Changed("html-layout") {
			config.HTMLLayout = htmlLayout
		}
		if flags.Changed("format") {
			config.Formatter = format
		}
		if flags.Changed("output") {
			config.Output.File = output
		}
		if flags.Changed("output-mode") {
			config.Output.Mode = outputMode
		}
		if flags.Changed("show-nested") {
			config.Settings.Nested = nested
		}
		if flags.Changed("show-patches") {
			config.Settings.Patches = patches
		}
		if flags.Changed("recursive") {
			config.Recursive = recursive
		}
		if flags.Changed("include") {
			config.Include = include
		}
		if flags.Changed("exclude") {
			config.Exclude = exclude
		}
		if flags.Changed("template") {
			config.Template = templateFile
		}
		if err := config.Validate(); err != nil {
			return err
		}
		if allVersions && version != "" {
			return fmt.Errorf("--all-versions and --api-version are mutually exclusive")
		}
		if flags.Changed("crossplane-version") && !flags.Changed("crossplane-image") {
			runtime.CrossplaneImage = ""
		}
		if render && rendered != "" {
			return fmt.Errorf("choose --render or --rendered")
		}
		inputs, err := engine.Load(args[0], config)
		if err != nil {
			return err
		}
		if kind != "" && (len(inputs) != 1 || engine.String(inputs[0].Object["kind"]) != kind) {
			return fmt.Errorf("expected one %s document", kind)
		}
		if runtime.Functions != "" {
			extra, err := engine.Load(runtime.Functions, engine.Defaults())
			if err != nil {
				return err
			}
			inputs = append(inputs, extra...)
		}
		doc, err := engine.Build(inputs, engine.Options{APIVersion: version, AllVersions: allVersions, Config: config, RefFiles: schemas})
		if err != nil {
			return err
		}
		var result []byte
		if render {
			if runtime.Composition == "" {
				if kind != "Composition" {
					return fmt.Errorf("batch rendering requires --composition")
				}
				runtime.Composition = args[0]
			}
			result, err = engine.Render(cmd.Context(), runtime)
			if err != nil {
				return err
			}
		}
		if rendered != "" {
			result, err = os.ReadFile(rendered)
			if err != nil {
				return err
			}
		}
		if render || rendered != "" {
			scenario, err := engine.ParseScenario(result, runtime.Name, runtime.XR, doc.Compositions...)
			if err != nil {
				return err
			}
			scenario.Inputs = engine.ScenarioInputs(runtime, rendered)
			if validate {
				if err := engine.Validate(cmd.Context(), runtime, extensions, result); err != nil {
					return err
				}
				scenario.Validated = true
			}
			doc.Scenarios = append(doc.Scenarios, scenario)
		} else if validate {
			if runtime.XR == "" {
				return fmt.Errorf("validation requires --xr or rendered output")
			}
			data, err := os.ReadFile(runtime.XR)
			if err != nil {
				return err
			}
			if err := engine.Validate(cmd.Context(), runtime, extensions, data); err != nil {
				return err
			}
		}
		data, err := engine.Format(doc, config)
		if err != nil {
			return err
		}
		if config.Output.File != "" || check {
			if err := engine.Write(data, config, check); err != nil {
				return err
			}
			if !check {
				_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Documentation generated successfully: %s\n", filepath.Clean(config.Output.File))
			}
			return err
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	}
	return command
}
