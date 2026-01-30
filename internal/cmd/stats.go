// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

type indexStats struct {
	Papers   int `json:"papers" yaml:"papers"`
	Articles int `json:"articles" yaml:"articles"`
	Total    int `json:"total" yaml:"total"`
	Tags     int `json:"tags" yaml:"tags"`
}

func newStatsCmd(db *sql.DB) *cobra.Command {
	var opts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show index statistics",
		Long:  `Display counts of indexed items by type.`,
		Example: `  # Show stats in table format
  arc-index stats

  # Output as JSON
  arc-index stats --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			return runStats(cmd, db, opts)
		},
	}

	opts.AddOutputFlags(cmd, output.OutputTable)
	return cmd
}

func runStats(cmd *cobra.Command, db *sql.DB, opts output.OutputOptions) error {
	stats := indexStats{}

	row := db.QueryRow(`SELECT COUNT(*) FROM items WHERE type = 'paper'`)
	_ = row.Scan(&stats.Papers)

	row = db.QueryRow(`SELECT COUNT(*) FROM items WHERE type = 'article'`)
	_ = row.Scan(&stats.Articles)

	row = db.QueryRow(`SELECT COUNT(*) FROM items`)
	_ = row.Scan(&stats.Total)

	row = db.QueryRow(`SELECT COUNT(*) FROM tags`)
	_ = row.Scan(&stats.Tags)

	switch {
	case opts.Is(output.OutputJSON):
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(stats)
	case opts.Is(output.OutputYAML):
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		defer enc.Close()
		return enc.Encode(stats)
	case opts.Is(output.OutputQuiet):
		return nil
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Papers:   %d\n", stats.Papers)
		fmt.Fprintf(cmd.OutOrStdout(), "Articles: %d\n", stats.Articles)
		fmt.Fprintf(cmd.OutOrStdout(), "Total:    %d\n", stats.Total)
		fmt.Fprintf(cmd.OutOrStdout(), "Tags:     %d\n", stats.Tags)
		return nil
	}
}
