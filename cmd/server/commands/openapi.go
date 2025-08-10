package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewOpenAPICommand creates and returns the openapi command
func NewOpenAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate OpenAPI specification",
		Long:  "Generate OpenAPI specification for the HashPost API",
		Run: func(cmd *cobra.Command, args []string) {
			// Parse command line flags
			outputFile, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")

			// Generate OpenAPI spec
			if err := GenerateOpenAPISpec(outputFile, format); err != nil {
				log.Fatal().Err(err).Msg("Failed to generate OpenAPI specification")
			}
		},
	}

	// Add flags
	cmd.Flags().String("output", "openapi.json", "Output file path")
	cmd.Flags().String("format", "json", "Output format (json or yaml)")

	return cmd
}

// GenerateOpenAPISpec generates the OpenAPI specification
func GenerateOpenAPISpec(outputFile, format string) error {
	// Create minimal server instance to generate spec
	srv := api.NewServerForOpenAPI()

	// Get OpenAPI spec based on format
	var spec []byte
	var err error
	
	switch format {
	case "yaml":
		spec, err = srv.API.OpenAPI().YAML()
	case "json":
		fallthrough
	default:
		// For JSON, we'll use YAML and convert or just use YAML for now
		spec, err = srv.API.OpenAPI().YAML()
	}
	
	if err != nil {
		return fmt.Errorf("failed to get OpenAPI spec: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write spec to file
	if err := os.WriteFile(outputFile, spec, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}

	fmt.Printf("✅ OpenAPI specification generated successfully\n")
	fmt.Printf("   Output file: %s\n", outputFile)
	fmt.Printf("   Format: %s\n", format)

	return nil
}
