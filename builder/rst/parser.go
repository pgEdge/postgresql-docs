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
)

// Parser converts RST text into a tree of Nodes.
type Parser struct {
	lines        []string
	pos          int
	headingChars []rune // tracks heading decoration order
}

// Parse parses an RST document and returns the root node.
func Parse(text string) *Node {
	p := &Parser{
		lines: strings.Split(text, "\n"),
	}
	root := &Node{Type: DocumentNode}
	p.parseBody(root, 0)
	return root
}

// parseBody parses lines at a given indentation level into
// the parent node's children.
func (p *Parser) parseBody(parent *Node, indent int) {
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]

		// Blank line — skip
		if strings.TrimSpace(line) == "" {
			p.pos++
			continue
		}

		// Check indentation — if less than expected, return to caller
		lineIndent := countIndent(line)
		if lineIndent < indent && strings.TrimSpace(line) != "" {
			return
		}

		// Label: .. _name:
		if p.isLabel(line) {
			p.parseLabel(parent)
			continue
		}

		// Substitution definition: .. |name| directive::
		if p.isSubstitutionDef(line) {
			p.parseSubstitutionDef(parent)
			continue
		}

		// Comment: .. (not followed by directive or label)
		if p.isComment(line) {
			p.parseComment(parent)
			continue
		}

		// Directive: .. name:: [arg]
		if p.isDirective(line) {
			p.parseDirective(parent, indent)
			continue
		}

		// Heading: check if next line is a decoration line
		if p.isHeading() {
			p.parseHeading(parent)
			continue
		}

		// Transition: a line of 4+ decoration chars, preceded and
		// followed by blank lines
		if p.isTransition(line) {
			parent.Children = append(parent.Children,
				&Node{Type: TransitionNode})
			p.pos++
			continue
		}

		// Grid table
		if p.isGridTable(line) {
			p.parseGridTable(parent)
			continue
		}

		// Bullet list: starts with *, -, or +
		if p.isBulletListItem(line, indent) {
			p.parseBulletList(parent, indent)
			continue
		}

		// Enumerated list: starts with number. or (number) or #.
		if p.isEnumListItem(line, indent) {
			p.parseEnumList(parent, indent)
			continue
		}

		// Field list: :field: value
		if p.isFieldList(line, indent) {
			p.parseFieldList(parent, indent)
			continue
		}

		// Line block: lines starting with |
		if p.isLineBlock(line, indent) {
			p.parseLineBlock(parent, indent)
			continue
		}

		// Literal block: previous paragraph ended with ::
		// (handled within paragraph parsing)

		// Default: paragraph
		p.parseParagraph(parent, indent)
	}
}

// countIndent returns the number of leading spaces.
func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// isDecorationLine checks if a line consists entirely of one
// repeated RST decoration character.
func isDecorationLine(line string) (rune, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return 0, false
	}
	decorChars := "=*-~^\"#`+:._"
	ch := rune(trimmed[0])
	if !strings.ContainsRune(decorChars, ch) {
		return 0, false
	}
	for _, r := range trimmed {
		if r != ch {
			return 0, false
		}
	}
	return ch, true
}

// headingLevel returns the heading level for a decoration character,
// assigning levels in order of first appearance.
func (p *Parser) headingLevel(ch rune) int {
	for i, c := range p.headingChars {
		if c == ch {
			return i + 1
		}
	}
	p.headingChars = append(p.headingChars, ch)
	return len(p.headingChars)
}

// isHeading checks if the current line is a heading (text followed
// by or preceded by a decoration line).
func (p *Parser) isHeading() bool {
	if p.pos >= len(p.lines) {
		return false
	}

	line := p.lines[p.pos]
	trimmed := strings.TrimSpace(line)

	// Case 1: overline + title + underline
	if ch, ok := isDecorationLine(line); ok {
		if p.pos+2 < len(p.lines) {
			titleLine := strings.TrimSpace(p.lines[p.pos+1])
			if titleLine != "" {
				underline := p.lines[p.pos+2]
				if uch, uok := isDecorationLine(underline); uok && uch == ch {
					return true
				}
			}
		}
		return false
	}

	// Case 2: title + underline
	if trimmed != "" && p.pos+1 < len(p.lines) {
		nextLine := p.lines[p.pos+1]
		if _, ok := isDecorationLine(nextLine); ok {
			nextTrimmed := strings.TrimSpace(nextLine)
			if len(nextTrimmed) >= len(strings.TrimRight(trimmed, " ")) {
				return true
			}
		}
	}

	return false
}

// parseHeading parses a section heading.
func (p *Parser) parseHeading(parent *Node) {
	line := p.lines[p.pos]

	// Case 1: overline + title + underline
	if ch, ok := isDecorationLine(line); ok {
		title := strings.TrimSpace(p.lines[p.pos+1])
		level := p.headingLevel(ch)
		parent.Children = append(parent.Children, &Node{
			Type:  HeadingNode,
			Text:  title,
			Level: level,
		})
		p.pos += 3
		return
	}

	// Case 2: title + underline
	title := strings.TrimSpace(line)
	ch, _ := isDecorationLine(p.lines[p.pos+1])
	level := p.headingLevel(ch)
	parent.Children = append(parent.Children, &Node{
		Type:  HeadingNode,
		Text:  title,
		Level: level,
	})
	p.pos += 2
}

// isLabel checks if a line is a cross-reference label.
func (p *Parser) isLabel(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, ".. _") &&
		strings.HasSuffix(trimmed, ":")
}

// parseLabel parses a label definition.
func (p *Parser) parseLabel(parent *Node) {
	line := strings.TrimSpace(p.lines[p.pos])
	// Extract label name from ".. _name:"
	name := line[4 : len(line)-1]
	parent.Children = append(parent.Children, &Node{
		Type:  LabelNode,
		Label: name,
	})
	p.pos++
}

// isDirective checks if a line starts an RST directive.
func (p *Parser) isDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ".. ") {
		return false
	}
	rest := trimmed[3:]
	// Must contain :: to be a directive
	idx := strings.Index(rest, "::")
	if idx < 0 {
		return false
	}
	// The part before :: must be the directive name (letters, hyphens)
	name := rest[:idx]
	if name == "" {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// parseDirective parses an RST directive.
func (p *Parser) parseDirective(parent *Node, baseIndent int) {
	line := strings.TrimSpace(p.lines[p.pos])
	// ".. name:: arg"
	rest := line[3:]
	idx := strings.Index(rest, "::")
	name := rest[:idx]
	arg := strings.TrimSpace(rest[idx+2:])

	p.pos++

	// Parse options and body
	options := make(map[string]string)
	var bodyLines []string

	// Determine directive body indent
	directiveIndent := baseIndent + 3
	if p.pos < len(p.lines) {
		nextLine := p.lines[p.pos]
		if strings.TrimSpace(nextLine) != "" {
			ni := countIndent(nextLine)
			if ni > baseIndent {
				directiveIndent = ni
			}
		}
	}

	// Parse options (lines starting with :option:)
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		if strings.TrimSpace(ln) == "" {
			// Blank line — end of options, start of body
			p.pos++
			break
		}
		trimmedLn := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmedLn, ":") &&
			strings.Contains(trimmedLn[1:], ":") {
			colonEnd := strings.Index(trimmedLn[1:], ":") + 1
			optName := trimmedLn[1:colonEnd]
			optVal := strings.TrimSpace(trimmedLn[colonEnd+1:])
			options[optName] = optVal
			p.pos++
		} else {
			// Not an option — must be body content
			break
		}
	}

	// Parse body (indented content)
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		if strings.TrimSpace(ln) == "" {
			// Check if next non-blank line is still indented
			nextNonBlank := p.peekNonBlank()
			if nextNonBlank >= 0 &&
				countIndent(p.lines[nextNonBlank]) >= directiveIndent {
				bodyLines = append(bodyLines, "")
				p.pos++
				continue
			}
			break
		}
		if countIndent(ln) < directiveIndent {
			break
		}
		// Strip the directive indentation
		if len(ln) > directiveIndent {
			bodyLines = append(bodyLines, ln[directiveIndent:])
		} else {
			bodyLines = append(bodyLines, strings.TrimSpace(ln))
		}
		p.pos++
	}

	body := strings.Join(bodyLines, "\n")

	node := &Node{
		Type:          DirectiveNode,
		DirectiveName: name,
		DirectiveArg:  arg,
		Options:       options,
		Body:          body,
	}

	// For toctree, body contains the list of entries
	// For nested directives (admonitions, warnings), parse body
	// as sub-document
	if isAdmonitionDirective(name) || name == "topic" {
		if body != "" {
			subParser := &Parser{
				lines:        strings.Split(body, "\n"),
				headingChars: p.headingChars,
			}
			subParser.parseBody(node, 0)
		}
	}

	parent.Children = append(parent.Children, node)
}

// isAdmonitionDirective returns true for admonition-type directives.
func isAdmonitionDirective(name string) bool {
	switch name {
	case "note", "warning", "tip", "caution", "important",
		"danger", "error", "hint", "attention", "admonition":
		return true
	}
	return false
}

// isComment checks if a line is a comment (not a directive or label).
func (p *Parser) isComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ".. ") && trimmed != ".." {
		return false
	}
	// Not a directive or label
	return !p.isDirective(line) && !p.isLabel(line) &&
		!p.isSubstitutionDef(line)
}

// parseComment skips a comment block.
func (p *Parser) parseComment(parent *Node) {
	p.pos++
	// Skip indented continuation lines
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		if strings.TrimSpace(ln) == "" {
			p.pos++
			continue
		}
		if countIndent(ln) > 0 {
			p.pos++
			continue
		}
		break
	}
	parent.Children = append(parent.Children, &Node{Type: CommentNode})
}

// isSubstitutionDef checks for substitution definitions.
func (p *Parser) isSubstitutionDef(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, ".. |") &&
		strings.Contains(trimmed, "| ")
}

// parseSubstitutionDef parses a substitution definition.
func (p *Parser) parseSubstitutionDef(parent *Node) {
	line := strings.TrimSpace(p.lines[p.pos])
	// ".. |name| directive:: arg"
	pipeEnd := strings.Index(line[4:], "|")
	if pipeEnd < 0 {
		p.pos++
		return
	}
	name := line[4 : 4+pipeEnd]
	rest := strings.TrimSpace(line[4+pipeEnd+1:])

	// rest is like "image:: images/sm_icon.png"
	dirName := ""
	dirArg := ""
	if idx := strings.Index(rest, "::"); idx >= 0 {
		dirName = strings.TrimSpace(rest[:idx])
		dirArg = strings.TrimSpace(rest[idx+2:])
	}

	p.pos++

	// Parse options
	options := make(map[string]string)
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		trimmedLn := strings.TrimSpace(ln)
		if trimmedLn == "" {
			break
		}
		if countIndent(ln) > 0 && strings.HasPrefix(trimmedLn, ":") {
			colonEnd := strings.Index(trimmedLn[1:], ":")
			if colonEnd >= 0 {
				optName := trimmedLn[1 : colonEnd+1]
				optVal := strings.TrimSpace(trimmedLn[colonEnd+2:])
				options[optName] = optVal
				p.pos++
				continue
			}
		}
		break
	}

	parent.Children = append(parent.Children, &Node{
		Type:             SubstitutionDefNode,
		SubstitutionName: name,
		DirectiveName:    dirName,
		DirectiveArg:     dirArg,
		Options:          options,
	})
}

// isBulletListItem checks if the line starts a bullet list item.
func (p *Parser) isBulletListItem(line string, indent int) bool {
	trimmed := strings.TrimLeft(line, " ")
	li := countIndent(line)
	if li < indent {
		return false
	}
	if len(trimmed) < 2 {
		return false
	}
	return (trimmed[0] == '*' || trimmed[0] == '-' ||
		trimmed[0] == '+') && trimmed[1] == ' '
}

// parseBulletList parses a bullet list.
func (p *Parser) parseBulletList(parent *Node, indent int) {
	list := &Node{Type: BulletListNode}

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "" {
			// Check if list continues after blank
			next := p.peekNonBlank()
			if next >= 0 && p.isBulletListItem(p.lines[next], indent) {
				p.pos++
				continue
			}
			break
		}
		if !p.isBulletListItem(line, indent) {
			// Could be continuation of current item
			li := countIndent(line)
			itemIndent := indent + 2
			if li >= itemIndent {
				// Continuation — append to last item
				if len(list.Children) > 0 {
					last := list.Children[len(list.Children)-1]
					last.Text += "\n" + line[itemIndent:]
				}
				p.pos++
				continue
			}
			break
		}

		// New list item
		trimmed := strings.TrimLeft(line, " ")
		text := trimmed[2:] // skip "* " or "- "
		item := &Node{Type: ListItemNode, Text: text}
		list.Children = append(list.Children, item)
		p.pos++

		// Collect continuation lines
		itemIndent := countIndent(line) + 2
		for p.pos < len(p.lines) {
			ln := p.lines[p.pos]
			if strings.TrimSpace(ln) == "" {
				// Check if continuation follows after blank
				next := p.peekNonBlank()
				if next >= 0 && countIndent(p.lines[next]) >= itemIndent &&
					!p.isBulletListItem(p.lines[next], indent) {
					item.Text += "\n"
					p.pos++
					continue
				}
				break
			}
			if countIndent(ln) >= itemIndent {
				if len(ln) > itemIndent {
					item.Text += "\n" + ln[itemIndent:]
				} else {
					item.Text += "\n" + strings.TrimSpace(ln)
				}
				p.pos++
			} else {
				break
			}
		}
	}

	if len(list.Children) > 0 {
		parent.Children = append(parent.Children, list)
	}
}

// isEnumListItem checks if the line starts an enumerated list item.
func (p *Parser) isEnumListItem(line string, indent int) bool {
	trimmed := strings.TrimLeft(line, " ")
	li := countIndent(line)
	if li < indent {
		return false
	}
	// Patterns: "1. ", "#. ", "(1) ", "1) "
	if len(trimmed) < 3 {
		return false
	}
	// Check for "#. " or digit+". "
	if trimmed[0] == '#' && len(trimmed) > 1 && trimmed[1] == '.' &&
		trimmed[2] == ' ' {
		return true
	}
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	if i > 0 && i < len(trimmed) {
		if trimmed[i] == '.' && i+1 < len(trimmed) && trimmed[i+1] == ' ' {
			return true
		}
		if trimmed[i] == ')' && i+1 < len(trimmed) && trimmed[i+1] == ' ' {
			return true
		}
	}
	return false
}

// parseEnumList parses an enumerated list.
func (p *Parser) parseEnumList(parent *Node, indent int) {
	list := &Node{Type: EnumListNode}

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "" {
			next := p.peekNonBlank()
			if next >= 0 && p.isEnumListItem(p.lines[next], indent) {
				p.pos++
				continue
			}
			break
		}
		if !p.isEnumListItem(line, indent) {
			break
		}

		trimmed := strings.TrimLeft(line, " ")
		// Find the text after the marker
		markerEnd := 0
		if trimmed[0] == '#' {
			markerEnd = 3 // "#. "
		} else {
			i := 0
			for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
				i++
			}
			if trimmed[i] == '.' {
				markerEnd = i + 2
			} else {
				markerEnd = i + 2 // "N) "
			}
		}
		if markerEnd > len(trimmed) {
			markerEnd = len(trimmed)
		}
		text := trimmed[markerEnd:]

		item := &Node{Type: ListItemNode, Text: text}
		list.Children = append(list.Children, item)
		p.pos++

		// Collect continuation lines
		itemIndent := countIndent(line) + markerEnd
		for p.pos < len(p.lines) {
			ln := p.lines[p.pos]
			if strings.TrimSpace(ln) == "" {
				next := p.peekNonBlank()
				if next >= 0 && countIndent(p.lines[next]) >= itemIndent &&
					!p.isEnumListItem(p.lines[next], indent) {
					item.Text += "\n"
					p.pos++
					continue
				}
				break
			}
			if countIndent(ln) >= itemIndent {
				if len(ln) > itemIndent {
					item.Text += "\n" + ln[itemIndent:]
				} else {
					item.Text += "\n" + strings.TrimSpace(ln)
				}
				p.pos++
			} else {
				break
			}
		}
	}

	if len(list.Children) > 0 {
		parent.Children = append(parent.Children, list)
	}
}

// isFieldList checks if a line starts a field list.
func (p *Parser) isFieldList(line string, indent int) bool {
	trimmed := strings.TrimLeft(line, " ")
	li := countIndent(line)
	if li < indent {
		return false
	}
	if len(trimmed) < 3 || trimmed[0] != ':' {
		return false
	}
	// Must have closing colon with text between
	end := strings.Index(trimmed[1:], ":")
	if end < 1 {
		return false
	}
	// Must be followed by a space or end of line
	afterColon := end + 2
	if afterColon < len(trimmed) && trimmed[afterColon] != ' ' {
		return false
	}
	return true
}

// parseFieldList parses a field list.
func (p *Parser) parseFieldList(parent *Node, indent int) {
	list := &Node{Type: FieldListNode}

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "" {
			p.pos++
			continue
		}
		if !p.isFieldList(line, indent) {
			break
		}

		trimmed := strings.TrimLeft(line, " ")
		end := strings.Index(trimmed[1:], ":")
		fieldName := trimmed[1 : end+1]
		fieldBody := ""
		if end+2 < len(trimmed) {
			fieldBody = strings.TrimSpace(trimmed[end+2:])
		}
		p.pos++

		// Continuation lines
		for p.pos < len(p.lines) {
			ln := p.lines[p.pos]
			if strings.TrimSpace(ln) == "" {
				break
			}
			if countIndent(ln) > indent {
				fieldBody += " " + strings.TrimSpace(ln)
				p.pos++
			} else {
				break
			}
		}

		list.Children = append(list.Children, &Node{
			Type:      FieldNode,
			FieldName: fieldName,
			FieldBody: strings.TrimSpace(fieldBody),
		})
	}

	if len(list.Children) > 0 {
		parent.Children = append(parent.Children, list)
	}
}

// isGridTable checks if a line starts a grid table.
func (p *Parser) isGridTable(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "+") &&
		strings.HasSuffix(trimmed, "+") &&
		(strings.Contains(trimmed, "-") || strings.Contains(trimmed, "="))
}

// parseGridTable parses an RST grid table.
func (p *Parser) parseGridTable(parent *Node) {
	var allLines []string
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		// Table lines start with + or |
		if !strings.HasPrefix(trimmed, "+") &&
			!strings.HasPrefix(trimmed, "|") {
			break
		}
		allLines = append(allLines, line)
		p.pos++
	}

	if len(allLines) < 3 {
		// Not enough lines for a table
		parent.Children = append(parent.Children, &Node{
			Type: ParagraphNode,
			Text: strings.Join(allLines, "\n"),
		})
		return
	}

	node := &Node{Type: GridTableNode}

	// Find column boundaries from the first separator line
	colBounds := findColumnBounds(allLines[0])

	// Group data rows and detect header separator
	hasHeader := false
	var dataGroups [][]string
	var currentGroup []string

	for _, line := range allLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "+") {
			// Separator line
			if strings.Contains(trimmed, "=") {
				hasHeader = true
			}
			if len(currentGroup) > 0 {
				dataGroups = append(dataGroups, currentGroup)
				currentGroup = nil
			}
		} else if strings.HasPrefix(trimmed, "|") {
			currentGroup = append(currentGroup, line)
		}
	}
	if len(currentGroup) > 0 {
		dataGroups = append(dataGroups, currentGroup)
	}

	// Extract cell contents
	for _, group := range dataGroups {
		row := extractRowCells(group, colBounds)
		node.TableRows = append(node.TableRows, row)
	}
	node.TableHeader = hasHeader

	parent.Children = append(parent.Children, node)
}

// findColumnBounds returns the start positions of each column
// from a separator line like "+---+---+---+".
func findColumnBounds(sep string) []int {
	var bounds []int
	for i, ch := range sep {
		if ch == '+' {
			bounds = append(bounds, i)
		}
	}
	return bounds
}

// extractRowCells extracts cell text from data lines using column
// bounds. Multi-line cells are joined with newlines.
func extractRowCells(dataLines []string, colBounds []int) []string {
	numCols := len(colBounds) - 1
	if numCols <= 0 {
		return nil
	}

	cells := make([]string, numCols)
	for _, line := range dataLines {
		for i := 0; i < numCols; i++ {
			start := colBounds[i] + 1
			end := colBounds[i+1]
			if end > len(line) {
				end = len(line)
			}
			if start >= len(line) {
				continue
			}
			cellText := ""
			if end > start {
				cellText = strings.TrimSpace(line[start:end])
			}
			if cells[i] != "" && cellText != "" {
				cells[i] += " " + cellText
			} else if cellText != "" {
				cells[i] = cellText
			}
		}
	}
	return cells
}

// parseParagraph parses a paragraph of text.
func (p *Parser) parseParagraph(parent *Node, indent int) {
	var lines []string
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "" {
			p.pos++
			break
		}
		// Check if the next line starts something new
		if p.pos > 0 && len(lines) > 0 {
			if p.isDirective(line) || p.isLabel(line) ||
				p.isSubstitutionDef(line) || p.isComment(line) {
				break
			}
			if p.isHeading() {
				break
			}
			if p.isGridTable(line) {
				break
			}
			li := countIndent(line)
			if li < indent {
				break
			}
			// Check for bullet/enum starting at this indent
			if p.isBulletListItem(line, indent) && len(lines) > 0 {
				break
			}
			if p.isEnumListItem(line, indent) && len(lines) > 0 {
				break
			}
		}
		lines = append(lines, strings.TrimSpace(line))
		p.pos++
	}

	if len(lines) == 0 {
		return
	}

	text := strings.Join(lines, " ")

	// Check for literal block (paragraph ending with ::)
	if strings.HasSuffix(text, "::") {
		// Emit paragraph (with :: replaced)
		paraText := text
		if text == "::" {
			paraText = ""
		} else if strings.HasSuffix(text, " ::") {
			paraText = text[:len(text)-3] + ":"
		} else {
			paraText = text[:len(text)-1]
		}
		if paraText != "" {
			parent.Children = append(parent.Children, &Node{
				Type: ParagraphNode,
				Text: paraText,
			})
		}

		// Parse literal block
		p.parseLiteralBlock(parent)
		return
	}

	parent.Children = append(parent.Children, &Node{
		Type: ParagraphNode,
		Text: text,
	})
}

// parseLiteralBlock parses an indented literal block after "::".
func (p *Parser) parseLiteralBlock(parent *Node) {
	// Skip blank lines
	for p.pos < len(p.lines) && strings.TrimSpace(p.lines[p.pos]) == "" {
		p.pos++
	}

	if p.pos >= len(p.lines) {
		return
	}

	blockIndent := countIndent(p.lines[p.pos])
	var lines []string

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "" {
			// Check if literal block continues
			next := p.peekNonBlank()
			if next >= 0 && countIndent(p.lines[next]) >= blockIndent {
				lines = append(lines, "")
				p.pos++
				continue
			}
			break
		}
		if countIndent(line) < blockIndent {
			break
		}
		if len(line) > blockIndent {
			lines = append(lines, line[blockIndent:])
		} else {
			lines = append(lines, "")
		}
		p.pos++
	}

	if len(lines) > 0 {
		parent.Children = append(parent.Children, &Node{
			Type: LiteralBlockNode,
			Text: strings.Join(lines, "\n"),
		})
	}
}

// isTransition checks if a line is a transition marker.
func (p *Parser) isTransition(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 4 {
		return false
	}
	ch, ok := isDecorationLine(line)
	if !ok {
		return false
	}
	// Must be preceded by a blank line
	if p.pos > 0 && strings.TrimSpace(p.lines[p.pos-1]) != "" {
		return false
	}
	_ = ch
	return true
}

// isLineBlock checks if a line starts a line block.
func (p *Parser) isLineBlock(line string, indent int) bool {
	trimmed := strings.TrimLeft(line, " ")
	li := countIndent(line)
	if li < indent {
		return false
	}
	return strings.HasPrefix(trimmed, "| ")
}

// parseLineBlock parses a line block.
func (p *Parser) parseLineBlock(parent *Node, indent int) {
	var lines []string
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimLeft(line, " ")
		if strings.TrimSpace(line) == "" {
			break
		}
		if !strings.HasPrefix(trimmed, "| ") {
			break
		}
		lines = append(lines, trimmed[2:])
		p.pos++
	}

	if len(lines) > 0 {
		parent.Children = append(parent.Children, &Node{
			Type: LineBlockNode,
			Text: strings.Join(lines, "\n"),
		})
	}
}

// peekNonBlank returns the index of the next non-blank line,
// or -1 if none exists.
func (p *Parser) peekNonBlank() int {
	for i := p.pos; i < len(p.lines); i++ {
		if strings.TrimSpace(p.lines[i]) != "" {
			return i
		}
	}
	return -1
}
