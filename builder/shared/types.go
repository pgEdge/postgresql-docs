//-------------------------------------------------------------------------
//
// pgEdge PostgreSQL Docs
//
// Copyright (c) 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

// Package shared provides types and utilities used by both
// the SGML and RST conversion pipelines.
package shared

import (
	"strings"
)

// IDEntry records where a document ID maps to in the output.
type IDEntry struct {
	// File is the output .md file path relative to docs/.
	File string
	// Anchor is the anchor name within the file.
	Anchor string
	// Title is the generated link text for cross-references.
	Title string
	// Type is the element type (chapter, sect1, etc.).
	Type string
}

// FileEntry tracks a single output file being generated.
type FileEntry struct {
	// Path relative to docs/ (e.g., "tutorial/start.md").
	Path string
	// Title for the nav entry.
	Title string
	// NavParent is the parent nav path for hierarchy.
	NavParent string
	// Order preserves document order within the nav.
	Order int
}

// Slugify converts a string to a URL-friendly slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ' || r == '/' || r == '.':
			b.WriteRune('-')
		}
	}
	// Collapse multiple hyphens
	result := b.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}
