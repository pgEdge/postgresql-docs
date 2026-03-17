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
	"regexp"
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
				cell = inlineToHTML(convertInlineCtx(ctx, row[i]))
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
				cell = cellToHTML(ctx, row[i])
			}
			w.WriteString("  <td>" + cell + "</td>\n")
		}
		w.WriteString("</tr>\n")
	}
	w.WriteString("</tbody>\n</table>\n")
}

// cellToHTML converts a grid-table cell (which may contain
// paragraphs and bullet lists) to HTML suitable for a <td>.
func cellToHTML(ctx *ConvertContext, cell string) string {
	if !strings.Contains(cell, "\n") {
		return inlineToHTML(convertInlineCtx(ctx, cell))
	}

	// Split into individual lines and classify them into
	// paragraph text and bullet items.  Consecutive bullet
	// items are grouped into a single <ul>.
	var result strings.Builder
	lines := strings.Split(cell, "\n")
	inList := false
	var paraLines []string

	flushPara := func() {
		if len(paraLines) > 0 {
			text := strings.Join(paraLines, " ")
			result.WriteString(inlineToHTML(convertInlineCtx(ctx, text)))
			paraLines = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			// Blank line — don't close a bullet list yet,
			// just flush any paragraph text.
			if !inList {
				flushPara()
			}
			continue
		}

		isBullet := len(trimmed) >= 2 &&
			(trimmed[0] == '*' || trimmed[0] == '-' ||
				trimmed[0] == '+') && trimmed[1] == ' '

		if isBullet {
			flushPara()
			if !inList {
				result.WriteString("<ul>")
				inList = true
			}
			result.WriteString("<li>" +
				inlineToHTML(convertInlineCtx(ctx, trimmed[2:])) +
				"</li>")
		} else {
			if inList {
				result.WriteString("</ul>")
				inList = false
			}
			paraLines = append(paraLines, trimmed)
		}
	}

	if inList {
		result.WriteString("</ul>")
	}
	flushPara()

	// Remove unused variable warning
	_ = reHTMLItalic

	return result.String()
}

// Precompiled patterns for Markdown-to-HTML conversion.
var (
	reHTMLBold   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reHTMLItalic = regexp.MustCompile(`(?:^|[^*])\*([^*]+)\*(?:[^*]|$)`)
	reHTMLCode   = regexp.MustCompile("`([^`]+)`")
	reHTMLLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// inlineToHTML converts text that has already been through RST
// inline conversion (producing Markdown) into HTML so it renders
// correctly inside <table> cells where Markdown is not processed.
func inlineToHTML(md string) string {
	// Order matters: bold before italic (both use *)
	s := md
	s = reHTMLBold.ReplaceAllString(s, "<strong>$1</strong>")
	// Italic: careful not to match inside <strong> tags.
	// Use a simple single-pass replacement.
	s = replaceItalic(s)
	s = reHTMLCode.ReplaceAllString(s, "<code>$1</code>")
	s = reHTMLLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	return s
}

// replaceItalic replaces *text* with <em>text</em>, avoiding
// ** (bold) and text inside HTML tags.
func replaceItalic(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '*' {
			// Skip ** (bold already converted to <strong>)
			if i+1 < len(s) && s[i+1] == '*' {
				result.WriteByte(s[i])
				result.WriteByte(s[i+1])
				i += 2
				continue
			}
			// Find closing *
			end := strings.Index(s[i+1:], "*")
			if end > 0 {
				// Check it's not ** at the end
				content := s[i+1 : i+1+end]
				if !strings.Contains(content, "*") &&
					len(content) > 0 {
					result.WriteString("<em>")
					result.WriteString(content)
					result.WriteString("</em>")
					i = i + 1 + end + 1
					continue
				}
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}
