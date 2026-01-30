// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package export

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type itemRow struct {
	Slug        string
	Title       string
	AuthorsJSON string
	Date        sql.NullString
}

// ExportIndex generates a markdown index of all research items.
func ExportIndex(db *sql.DB) (string, error) {
	var b strings.Builder
	b.WriteString("# Research Index\n\n")

	// Papers
	b.WriteString("## Papers\n")
	papers, err := fetchItems(db, "paper")
	if err != nil {
		return "", err
	}
	for _, it := range papers {
		tags, _ := fetchTags(db, it.Slug)
		line := formatLine(it, tags)
		b.WriteString(line)
	}
	b.WriteString("\n")

	// Blog posts (articles)
	b.WriteString("## Blog Posts\n")
	blogs, err := fetchItems(db, "article")
	if err != nil {
		return "", err
	}
	for _, it := range blogs {
		tags, _ := fetchTags(db, it.Slug)
		line := formatLine(it, tags)
		b.WriteString(line)
	}
	b.WriteString("\n")

	return b.String(), nil
}

func fetchItems(db *sql.DB, typ string) ([]itemRow, error) {
	rows, err := db.Query(`
		SELECT slug, title, authors_json, COALESCE(date_published, created_at)
		FROM items
		WHERE type = ?
		ORDER BY COALESCE(date_published, created_at) DESC, slug ASC`, typ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []itemRow
	for rows.Next() {
		var r itemRow
		if err := rows.Scan(&r.Slug, &r.Title, &r.AuthorsJSON, &r.Date); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func fetchTags(db *sql.DB, slug string) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.name
		FROM tags t
		JOIN item_tags it ON it.tag_id = t.id
		JOIN items i ON i.id = it.item_id
		WHERE i.slug = ?
		ORDER BY t.name ASC`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

func formatLine(it itemRow, tags []string) string {
	authors := parseAuthors(it.AuthorsJSON)
	authStr := authorDisplay(authors)
	yr := yearFrom(it.Date.String)
	return fmt.Sprintf("- [%s](%s.md) - %s (%s) [%s]\n", it.Title, it.Slug, authStr, yr, strings.Join(tags, ", "))
}

func parseAuthors(jsonStr string) []string {
	var arr []string
	_ = json.Unmarshal([]byte(jsonStr), &arr)
	return arr
}

func authorDisplay(authors []string) string {
	if len(authors) == 0 {
		return ""
	}
	if len(authors) == 1 {
		return authors[0]
	}
	return fmt.Sprintf("%s et al.", authors[0])
}

func yearFrom(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	if len(dateStr) >= 4 {
		return dateStr[:4]
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return fmt.Sprintf("%d", t.Year())
	}
	return ""
}
