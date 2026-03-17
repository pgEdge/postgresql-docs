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

// ExportSlugify exposes slugify for use by other packages.
func ExportSlugify(s string) string {
	return shared.Slugify(s)
}
