//-------------------------------------------------------------------------
//
// pgEdge PostgreSQL Docs
//
// Copyright (c) 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package convert

import (
	"github.com/pgEdge/postgresql-docs/builder/shared"
)

// MarkdownWriter is an alias for the shared type.
type MarkdownWriter = shared.MarkdownWriter

// NewMarkdownWriter creates a new writer.
func NewMarkdownWriter() *MarkdownWriter {
	return shared.NewMarkdownWriter()
}
