// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-index/internal/export"
	"github.com/yourorg/arc-sdk/config"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

func newRebuildCmd(cfg *config.Config, db *sql.DB) *cobra.Command {
	var opts output.OutputOptions
	var outFile string

	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild the research index from database",
		Long: `Rebuild the _INDEX.md file from database contents.

The index file contains all research items organized by type (Papers, Blog Posts)
in reverse chronological order.`,
		Example: `  # Rebuild the default research index
  arc-index rebuild

  # Write the index to a custom path
  arc-index rebuild --out-file docs/research/_INDEX.md

  # Output as JSON instead of writing file
  arc-index rebuild --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			return runRebuild(cmd, cfg, db, opts, outFile)
		},
	}

	opts.AddOutputFlags(cmd, output.OutputTable)
	cmd.Flags().StringVarP(&outFile, "out-file", "f", "", "Path for the generated index file (default: {research-root}/_INDEX.md)")

	return cmd
}

func runRebuild(cmd *cobra.Command, cfg *config.Config, db *sql.DB, opts output.OutputOptions, outFile string) error {
	// Determine output path
	outputPath := outFile
	if outputPath == "" {
		outputPath = filepath.Join(cfg.ResearchRoot, "_INDEX.md")
	}
	outputPath = config.ExpandPath(outputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	content, err := export.ExportIndex(db)
	if err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write index file to %s: %w", outputPath, err)
	}

	result := map[string]string{
		"path":     outputPath,
		"research": cfg.ResearchRoot,
		"status":   "rebuilt",
	}

	switch {
	case opts.Is(output.OutputJSON):
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	case opts.Is(output.OutputYAML):
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		defer enc.Close()
		return enc.Encode(result)
	case opts.Is(output.OutputQuiet):
		return nil
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Index rebuilt successfully: %s\n", outputPath)
		return nil
	}
}
