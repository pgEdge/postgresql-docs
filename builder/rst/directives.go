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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgEdge/postgresql-docs/builder/shared"
)

// DirectiveHandler converts a directive node to Markdown.
type DirectiveHandler func(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error

// directiveHandlers maps directive names to their handlers.
// Populated by initDirectiveHandlers to avoid initialization cycles.
var directiveHandlers map[string]DirectiveHandler

func initDirectiveHandlers() {
	if directiveHandlers != nil {
		return
	}
	directiveHandlers = map[string]DirectiveHandler{
		"image":          handleImage,
		"figure":         handleFigure,
		"code-block":     handleCodeBlock,
		"sourcecode":     handleCodeBlock,
		"note":           handleAdmonition,
		"warning":        handleAdmonition,
		"tip":            handleAdmonition,
		"caution":        handleAdmonition,
		"important":      handleAdmonition,
		"danger":         handleAdmonition,
		"hint":           handleAdmonition,
		"attention":      handleAdmonition,
		"admonition":     handleAdmonition,
		"toctree":        handleToctreeDirective,
		"csv-table":      handleCSVTable,
		"table":          handleTableDirective,
		"topic":          handleTopic,
		"youtube":        handleYouTube,
		"literalinclude": handleLiteralInclude,
		"deprecated":     handleDeprecated,
		"versionadded":   handleVersionChanged,
		"versionchanged": handleVersionChanged,
		"seealso":        handleSeeAlso,
		"contents":       handleSkipDirective,
		"raw":            handleRaw,
		"only":           handleOnly,
		"highlight":      handleSkipDirective,
		"index":          handleSkipDirective,
	}
}

// handleImage converts an image directive to Markdown.
func handleImage(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	imgPath := node.DirectiveArg
	alt := node.Options["alt"]
	if alt == "" {
		alt = "image"
	}

	w.BlankLine()
	w.WriteString(fmt.Sprintf("![%s](%s)\n", alt, imgPath))

	// Copy image file
	ctx.copyImage(imgPath)

	return nil
}

// handleFigure converts a figure directive to Markdown.
func handleFigure(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	imgPath := node.DirectiveArg
	alt := node.Options["alt"]
	if alt == "" {
		alt = "image"
	}

	w.BlankLine()
	w.WriteString(fmt.Sprintf("![%s](%s)\n", alt, imgPath))
	ctx.copyImage(imgPath)

	// Caption is in the body
	if node.Body != "" {
		w.BlankLine()
		w.WriteString("*" + convertInlineCtx(ctx, node.Body) + "*\n")
	}

	return nil
}

// handleCodeBlock converts a code-block directive.
func handleCodeBlock(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	lang := node.DirectiveArg
	w.StartCodeBlock(lang)
	w.WriteString(node.Body)
	w.EndCodeBlock()
	return nil
}

// handleAdmonition converts note/warning/tip etc. to MkDocs admonitions.
func handleAdmonition(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	kind := node.DirectiveName
	title := node.DirectiveArg

	w.BlankLine()
	if title != "" {
		w.WriteString(fmt.Sprintf("!!! %s \"%s\"\n\n", kind, title))
	} else {
		w.WriteString(fmt.Sprintf("!!! %s\n\n", kind))
	}

	// Convert children (parsed body) with 4-space indent
	content := convertAdmonitionBody(ctx, node)

	for _, line := range strings.Split(content, "\n") {
		if line == "" {
			w.WriteString("\n")
		} else {
			w.WriteString("    " + line + "\n")
		}
	}
	w.EnsureNewline()
	return nil
}

// convertAdmonitionBody converts the children of an admonition node.
func convertAdmonitionBody(ctx *ConvertContext, node *Node) string {
	if len(node.Children) > 0 {
		subW := shared.NewMarkdownWriter()
		for _, child := range node.Children {
			convertNode(ctx, child, subW)
		}
		return strings.TrimSpace(subW.String())
	}
	// Fallback: use raw body with inline conversion
	if node.Body != "" {
		return convertInlineCtx(ctx, node.Body)
	}
	return ""
}

// handleToctreeDirective is a no-op — toctrees are resolved separately.
func handleToctreeDirective(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	return nil
}

// handleCSVTable converts a csv-table directive to a Markdown table.
func handleCSVTable(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	w.BlankLine()

	header := node.Options["header"]
	body := node.Body

	// Parse header
	var headerCells []string
	if header != "" {
		headerCells = parseCSVRow(header)
	}

	// Parse body rows
	var rows [][]string
	if body != "" {
		// CSV table body: each logical row may span multiple lines
		// when quoted values contain newlines
		csvRows := splitCSVRows(body)
		for _, row := range csvRows {
			cells := parseCSVRow(row)
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		}
	}

	// Determine column count
	numCols := len(headerCells)
	for _, row := range rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}
	if numCols == 0 {
		return nil
	}

	// Write header
	if len(headerCells) > 0 {
		w.WriteString("|")
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(headerCells) {
				cell = convertInlineCtx(ctx, headerCells[i])
			}
			w.WriteString(" " + cell + " |")
		}
		w.WriteString("\n")
	} else {
		// Generate empty header
		w.WriteString("|")
		for i := 0; i < numCols; i++ {
			w.WriteString("   |")
		}
		w.WriteString("\n")
	}

	// Write separator
	w.WriteString("|")
	for i := 0; i < numCols; i++ {
		_ = i
		w.WriteString("---|")
	}
	w.WriteString("\n")

	// Write rows
	for _, row := range rows {
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

	return nil
}

// parseCSVRow splits a CSV row into cells, handling quoting.
func parseCSVRow(line string) []string {
	var cells []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' {
			if inQuote {
				// Check for escaped quote
				if i+1 < len(line) && line[i+1] == '"' {
					current.WriteByte('"')
					i++
				} else {
					inQuote = false
				}
			} else {
				inQuote = true
			}
		} else if ch == ',' && !inQuote {
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	cells = append(cells, strings.TrimSpace(current.String()))
	return cells
}

// splitCSVRows splits CSV body text into logical rows, handling
// multi-line quoted values.
func splitCSVRows(body string) []string {
	var rows []string
	var current strings.Builder
	inQuote := false

	for _, line := range strings.Split(body, "\n") {
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(line)

		// Count quotes to track state
		for _, ch := range line {
			if ch == '"' {
				inQuote = !inQuote
			}
		}

		if !inQuote {
			row := strings.TrimSpace(current.String())
			if row != "" {
				rows = append(rows, row)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		row := strings.TrimSpace(current.String())
		if row != "" {
			rows = append(rows, row)
		}
	}

	return rows
}

// handleTableDirective converts a .. table:: directive with a grid
// table body.
func handleTableDirective(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	title := node.DirectiveArg
	if title != "" {
		w.BlankLine()
		w.WriteString("**" + convertInlineCtx(ctx, title) + "**\n")
	}

	// The body contains a grid table — parse and render it
	if node.Body != "" {
		subRoot := Parse(node.Body)
		for _, child := range subRoot.Children {
			convertNode(ctx, child, w)
		}
	}

	return nil
}

// handleTopic converts a topic directive to a blockquote.
func handleTopic(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	title := node.DirectiveArg

	w.BlankLine()
	if title != "" {
		w.WriteString("**" + convertInlineCtx(ctx, title) + "**\n")
	}
	w.BlankLine()

	// Render body as blockquote
	content := convertAdmonitionBody(ctx, node)
	for _, line := range strings.Split(content, "\n") {
		w.WriteString("> " + line + "\n")
	}
	w.EnsureNewline()
	return nil
}

// handleYouTube converts a youtube directive to an embedded video.
func handleYouTube(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	videoID := node.DirectiveArg
	width := node.Options["width"]
	if width == "" {
		width = "560"
	}

	w.BlankLine()
	w.WriteString(fmt.Sprintf(
		"<div style=\"text-align: center;\">\n"+
			"<iframe width=\"%s\" height=\"315\" "+
			"src=\"https://www.youtube.com/embed/%s\" "+
			"frameborder=\"0\" allowfullscreen></iframe>\n"+
			"</div>\n",
		width, videoID))
	return nil
}

// handleLiteralInclude includes a file as a code block.
func handleLiteralInclude(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	filePath := node.DirectiveArg
	lang := node.Options["language"]
	if lang == "" {
		lang = node.Options["lang"]
	}

	// Try to read from pgAdmin source
	if ctx.PgAdminSrcDir != "" {
		fullPath := filepath.Join(ctx.PgAdminSrcDir, filePath)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			w.StartCodeBlock(lang)
			w.WriteString(string(data))
			w.EndCodeBlock()
			return nil
		}
	}

	// Fallback: show as a reference
	w.BlankLine()
	w.WriteString(fmt.Sprintf(
		"*See source file: `%s`*\n", filePath))
	return nil
}

// handleDeprecated converts a deprecated directive.
func handleDeprecated(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	w.BlankLine()
	version := node.DirectiveArg
	w.WriteString(fmt.Sprintf(
		"!!! warning \"Deprecated since version %s\"\n\n", version))
	if node.Body != "" {
		for _, line := range strings.Split(
			convertInlineCtx(ctx, node.Body), "\n") {
			if line == "" {
				w.WriteString("\n")
			} else {
				w.WriteString("    " + line + "\n")
			}
		}
	}
	return nil
}

// handleVersionChanged converts versionadded/versionchanged.
func handleVersionChanged(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	version := node.DirectiveArg
	label := "Changed in version"
	if node.DirectiveName == "versionadded" {
		label = "New in version"
	}
	w.BlankLine()
	w.WriteString(fmt.Sprintf("*%s %s.*", label, version))
	if node.Body != "" {
		w.WriteString(" " + convertInlineCtx(ctx, node.Body))
	}
	w.WriteString("\n")
	return nil
}

// handleSeeAlso converts a seealso directive.
func handleSeeAlso(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	w.BlankLine()
	w.WriteString("!!! tip \"See Also\"\n\n")
	if node.Body != "" {
		for _, line := range strings.Split(
			convertInlineCtx(ctx, node.Body), "\n") {
			if line == "" {
				w.WriteString("\n")
			} else {
				w.WriteString("    " + line + "\n")
			}
		}
	}
	return nil
}

// handleSkipDirective silently skips a directive.
func handleSkipDirective(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	return nil
}

// handleRaw passes through raw HTML content.
func handleRaw(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	format := node.DirectiveArg
	if format == "html" && node.Body != "" {
		w.BlankLine()
		w.WriteString(node.Body + "\n")
	}
	return nil
}

// handleOnly handles the .. only:: directive (conditional content).
func handleOnly(
	ctx *ConvertContext,
	node *Node,
	w *shared.MarkdownWriter,
) error {
	// Include all content regardless of condition
	if node.Body != "" {
		subRoot := Parse(node.Body)
		for _, child := range subRoot.Children {
			convertNode(ctx, child, w)
		}
	}
	return nil
}
