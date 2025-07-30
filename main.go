//go:generate peg grammar.peg

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var (
		version   string
		schemaDir string
	)

	rootCmd := &cobra.Command{
		Use:   "mcheck <json-file>",
		Short: "Validate Minecraft datapack JSON files against mcdoc schemas",
		Long: `mcheck is a tool for validating Minecraft datapack JSON files against
mcdoc schemas with version-specific constraints.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonPath := args[0]

			// Parse the target version
			targetVersion, err := parseVersion(version)
			if err != nil {
				return fmt.Errorf("invalid version format: %w", err)
			}

			// Find schema directory if not provided
			if schemaDir == "" {
				// Look for vanilla-mcdoc directory
				if _, err := os.Stat("vanilla-mcdoc"); err == nil {
					schemaDir = "vanilla-mcdoc"
				} else {
					return fmt.Errorf("schema directory not found, please specify with --schema-dir")
				}
			}

			// Create PEG-based validator and validate
			validator := NewPEGMCDocValidator(targetVersion, schemaDir)
			return validator.ValidateJSON(jsonPath)
		},
	}

	rootCmd.Flags().StringVarP(&version, "version", "v", "1.20.1", "Target Minecraft version")
	rootCmd.Flags().StringVarP(&schemaDir, "schema-dir", "s", "", "Path to vanilla-mcdoc directory")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}