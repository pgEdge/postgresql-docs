//-------------------------------------------------------------------------
//
// pgEdge PostgreSQL Docs
//
// Copyright (c) 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package rst

import (
	"os"
	"path/filepath"
	"strings"
)

// CollectAllLabels scans all RST files in srcDir for label definitions,
// building a complete label-to-location map. This is a comprehensive
// scan that catches labels not found via toctree walking.
func CollectAllLabels(
	srcDir string,
	fileMap map[string]string,
	labelMap map[string]labelInfo,
) []string {
	var warnings []string

	for rstName, outputPath := range fileMap {
		rstPath := filepath.Join(srcDir, rstName+".rst")
		data, err := os.ReadFile(rstPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, ".. _") ||
				!strings.HasSuffix(trimmed, ":") {
				continue
			}

			label := trimmed[4 : len(trimmed)-1]
			if label == "" {
				continue
			}

			// Find the next heading after this label
			title := findNextHeading(lines, i+1)
			if title == "" {
				title = label
			}

			labelMap[label] = labelInfo{
				File:   outputPath,
				Anchor: label,
				Title:  title,
			}
		}
	}

	return warnings
}

// findNextHeading looks for the next heading after line index start.
func findNextHeading(lines []string, start int) string {
	// Skip blank lines
	i := start
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	if i >= len(lines) {
		return ""
	}

	// Check for overline + title + underline pattern
	line := lines[i]
	if ch, ok := isDecorationLine(line); ok && ch != 0 {
		// Overline pattern
		if i+1 < len(lines) {
			title := strings.TrimSpace(lines[i+1])
			return cleanTitle(title)
		}
	}

	// Check for title + underline pattern
	title := strings.TrimSpace(line)
	if title != "" && i+1 < len(lines) {
		nextLine := lines[i+1]
		if _, ok := isDecorationLine(nextLine); ok {
			return cleanTitle(title)
		}
	}

	return ""
}
