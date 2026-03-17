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
	"strings"

	"github.com/pgEdge/postgresql-docs/builder/shared"
)

// convertGridTable converts a GridTableNode to Markdown.
func convertGridTable(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) {
	if len(node.TableRows) == 0 {
		return
	}

	w.BlankLine()

	// Determine number of columns
	numCols := 0
	for _, row := range node.TableRows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}
	if numCols == 0 {
		return
	}

	// Check if any cell contains multi-line content that needs HTML
	needsHTML := false
	for _, row := range node.TableRows {
		for _, cell := range row {
			if strings.Contains(cell, "\n") {
				needsHTML = true
				break
			}
		}
		if needsHTML {
			break
		}
	}

	if needsHTML {
		writeHTMLTable(ctx, node, w, numCols)
	} else {
		writeMarkdownTable(ctx, node, w, numCols)
	}
}

// writeMarkdownTable writes a simple Markdown pipe table.
func writeMarkdownTable(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
	numCols int,
) {
	startRow := 0

	if node.TableHeader && len(node.TableRows) > 0 {
		// First row is the header
		row := node.TableRows[0]
		w.WriteString("|")
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = convertInlineCtx(ctx, row[i])
			}
			w.WriteString(" " + cell + " |")
		}
		w.WriteString("\n")
		startRow = 1
	} else {
		// No header — write empty header
		w.WriteString("|")
		for i := 0; i < numCols; i++ {
			_ = i
			w.WriteString("   |")
		}
		w.WriteString("\n")
	}

	// Separator
	w.WriteString("|")
	for i := 0; i < numCols; i++ {
		_ = i
		w.WriteString("---|")
	}
	w.WriteString("\n")

	// Data rows
	for _, row := range node.TableRows[startRow:] {
		w.WriteString("|")
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = convertInlineCtx(ctx, row[i])
			}
			w.WriteString(" " + cell + " |")
		}
		w.WriteString("\n")
	}
}

// writeHTMLTable writes an HTML table for complex grid tables.
func writeHTMLTable(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
	numCols int,
) {
	w.WriteString("<table>\n")

	startRow := 0
	if node.TableHeader && len(node.TableRows) > 0 {
		w.WriteString("<thead>\n<tr>\n")
		row := node.TableRows[0]
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = convertInlineCtx(ctx, row[i])
			}
			w.WriteString("  <th>" + cell + "</th>\n")
		}
		w.WriteString("</tr>\n</thead>\n")
		startRow = 1
	}

	w.WriteString("<tbody>\n")
	for _, row := range node.TableRows[startRow:] {
		w.WriteString("<tr>\n")
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = convertInlineCtx(ctx, row[i])
			}
			w.WriteString("  <td>" + cell + "</td>\n")
		}
		w.WriteString("</tr>\n")
	}
	w.WriteString("</tbody>\n</table>\n")
}
