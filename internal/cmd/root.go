// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"database/sql"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-sdk/config"
)

// NewRootCmd creates the root command for arc-index.
func NewRootCmd(cfg *config.Config, db *sql.DB) *cobra.Command {
	root := &cobra.Command{
		Use:   "arc-index",
		Short: "Index and export utilities for research content",
		Long: `Build and manage research indexes from database contents.

Arc-index generates index files (like _INDEX.md) from your research database,
organizing papers, blog posts, and other items by type and date.`,
	}

	root.AddCommand(
		newRebuildCmd(cfg, db),
		newStatsCmd(db),
	)

	return root
}
