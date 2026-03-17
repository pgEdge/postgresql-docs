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
)

// Precompiled regexes for inline RST markup.
var (
	// :role:`text` or :role:`title <target>`
	reRole = regexp.MustCompile(
		":([a-zA-Z0-9_-]+):`([^`]+)`")

	// `text`:index: (pgAdmin index pattern)
	reIndexRole = regexp.MustCompile(
		"`([^`]+)`:index:")

	// **bold**
	reBold = regexp.MustCompile(`\*\*([^*]+)\*\*`)

	// *italic* (but not inside **)
	reItalic = regexp.MustCompile(`(?:^|[^*])\*([^*]+)\*(?:[^*]|$)`)

	// ``literal``
	reLiteral = regexp.MustCompile("``([^`]+)``")

	// `text <URL>`_ or `text <URL>`__
	reExternalLink = regexp.MustCompile(
		"`([^<]+)<([^>]+)>`_{1,2}")

	// `text`_ (named reference)
	reNamedRef = regexp.MustCompile("`([^`]+)`_(?:[^_]|$)")

	// |substitution|
	reSubstitution = regexp.MustCompile(`\|([a-zA-Z0-9_-]+)\|`)
)

// ConvertInline converts RST inline markup to Markdown.
// The labelMap maps label names to (file, anchor) pairs.
// The fileMap maps RST basenames to output .md paths.
// The substitutions map holds substitution definitions.
func ConvertInline(
	text string,
	labelMap map[string]labelInfo,
	fileMap map[string]string,
	currentFile string,
	substitutions map[string]*Node,
) string {
	if text == "" {
		return ""
	}

	// Process in order to avoid double-conversion:
	// 1. Literal spans (protect from further processing)
	// 2. Roles
	// 3. External links
	// 4. Named references
	// 5. Bold/italic
	// 6. Substitutions
	// 7. Index entries

	// Step 1: Protect literal spans
	type placeholder struct {
		key     string
		replace string
	}
	var placeholders []placeholder
	phIdx := 0

	// Protect ``literal``
	text = reLiteral.ReplaceAllStringFunc(text, func(m string) string {
		sub := reLiteral.FindStringSubmatch(m)
		key := "\x00LIT" + string(rune('A'+phIdx)) + "\x00"
		phIdx++
		placeholders = append(placeholders, placeholder{
			key: key, replace: "`" + sub[1] + "`"})
		return key
	})

	// Step 2: Index entries (strip the :index: wrapper)
	text = reIndexRole.ReplaceAllString(text, "$1")

	// Step 3: Roles
	text = reRole.ReplaceAllStringFunc(text, func(m string) string {
		sub := reRole.FindStringSubmatch(m)
		role := sub[1]
		content := sub[2]
		return convertRole(role, content, labelMap, fileMap,
			currentFile)
	})

	// Step 4: External links
	text = reExternalLink.ReplaceAllStringFunc(text, func(m string) string {
		sub := reExternalLink.FindStringSubmatch(m)
		title := strings.TrimSpace(sub[1])
		url := strings.TrimSpace(sub[2])
		return "[" + title + "](" + url + ")"
	})

	// Step 5: Substitutions
	text = reSubstitution.ReplaceAllStringFunc(text, func(m string) string {
		sub := reSubstitution.FindStringSubmatch(m)
		name := sub[1]
		if def, ok := substitutions[name]; ok {
			if def.DirectiveName == "image" {
				alt := def.Options["alt"]
				if alt == "" {
					alt = name
				}
				return "<img src=\"images/" +
					strings.TrimPrefix(def.DirectiveArg, "images/") +
					"\" alt=\"" + alt + "\">"
			}
			return def.DirectiveArg
		}
		return m
	})

	// Restore literal placeholders
	for _, ph := range placeholders {
		text = strings.ReplaceAll(text, ph.key, ph.replace)
	}

	return text
}

// labelInfo holds cross-reference target information.
type labelInfo struct {
	File   string
	Anchor string
	Title  string
}

// convertRole converts an RST role to Markdown.
func convertRole(
	role, content string,
	labelMap map[string]labelInfo,
	fileMap map[string]string,
	currentFile string,
) string {
	switch role {
	case "ref":
		return convertRef(content, labelMap, currentFile)
	case "doc":
		return convertDoc(content, fileMap, currentFile)
	case "index":
		// Strip index entries — just return the content
		return content
	case "menuselection", "guilabel":
		return "**" + content + "**"
	case "kbd":
		return "`" + content + "`"
	case "file", "command", "program", "envvar", "option",
		"class", "func", "meth", "attr", "exc", "obj",
		"mod", "data", "const", "type", "term":
		return "`" + content + "`"
	case "code":
		return "`" + content + "`"
	case "abbr":
		// :abbr:`text (expansion)`
		return content
	case "sup":
		return "<sup>" + content + "</sup>"
	case "sub":
		return "<sub>" + content + "</sub>"
	default:
		return "`" + content + "`"
	}
}

// convertRef converts a :ref: role to a Markdown link.
func convertRef(
	content string,
	labelMap map[string]labelInfo,
	currentFile string,
) string {
	// :ref:`Title <target>` or :ref:`target`
	title := ""
	target := content
	if idx := strings.Index(content, "<"); idx >= 0 {
		title = strings.TrimSpace(content[:idx])
		end := strings.Index(content, ">")
		if end > idx {
			target = content[idx+1 : end]
		}
	}

	if info, ok := labelMap[target]; ok {
		if title == "" {
			title = info.Title
			if title == "" {
				title = target
			}
		}
		link := resolveLink(currentFile, info.File, info.Anchor)
		return "[" + title + "](" + link + ")"
	}

	if title == "" {
		title = target
	}
	return title
}

// convertDoc converts a :doc: role to a Markdown link.
func convertDoc(
	content string,
	fileMap map[string]string,
	currentFile string,
) string {
	title := ""
	target := content
	if idx := strings.Index(content, "<"); idx >= 0 {
		title = strings.TrimSpace(content[:idx])
		end := strings.Index(content, ">")
		if end > idx {
			target = content[idx+1 : end]
		}
	}

	// Clean target
	target = strings.TrimPrefix(target, "/")
	target = strings.TrimSuffix(target, ".rst")

	if mdPath, ok := fileMap[target]; ok {
		if title == "" {
			title = target
		}
		link := resolveLink(currentFile, mdPath, "")
		return "[" + title + "](" + link + ")"
	}

	if title == "" {
		title = target
	}
	return title
}

// resolveLink computes a relative link from currentFile to target.
func resolveLink(currentFile, targetFile, anchor string) string {
	if currentFile == targetFile {
		if anchor != "" {
			return "#" + anchor
		}
		return ""
	}

	// Simple relative path calculation
	fromParts := strings.Split(currentFile, "/")
	toParts := strings.Split(targetFile, "/")

	// Remove filename from from
	fromParts = fromParts[:len(fromParts)-1]

	// Find common prefix
	common := 0
	for common < len(fromParts) && common < len(toParts) {
		if fromParts[common] == toParts[common] {
			common++
		} else {
			break
		}
	}

	// Build relative path
	var parts []string
	for i := common; i < len(fromParts); i++ {
		parts = append(parts, "..")
	}
	for i := common; i < len(toParts); i++ {
		parts = append(parts, toParts[i])
	}

	link := strings.Join(parts, "/")
	if link == "" {
		link = targetFile
	}
	if anchor != "" {
		link += "#" + anchor
	}
	return link
}
