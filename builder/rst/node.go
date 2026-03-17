//-------------------------------------------------------------------------
//
// pgEdge PostgreSQL Docs
//
// Copyright (c) 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

// Package rst parses reStructuredText documents and converts them
// to Markdown suitable for MkDocs Material.
package rst

// NodeType identifies the kind of RST node.
type NodeType int

const (
	// DocumentNode is the root of a parsed RST file.
	DocumentNode NodeType = iota
	// HeadingNode is a section heading.
	HeadingNode
	// ParagraphNode is a paragraph of text.
	ParagraphNode
	// DirectiveNode is an RST directive (.. directive::).
	DirectiveNode
	// BulletListNode is an unordered list.
	BulletListNode
	// EnumListNode is an ordered/numbered list.
	EnumListNode
	// ListItemNode is a single list item.
	ListItemNode
	// BlockQuoteNode is an indented block quote.
	BlockQuoteNode
	// LiteralBlockNode is a :: literal block.
	LiteralBlockNode
	// CommentNode is an RST comment.
	CommentNode
	// LabelNode is a cross-reference label (.. _name:).
	LabelNode
	// SubstitutionDefNode is a substitution definition (.. |name| ...).
	SubstitutionDefNode
	// FieldListNode is a field list.
	FieldListNode
	// FieldNode is a single field in a field list.
	FieldNode
	// GridTableNode is a grid table.
	GridTableNode
	// TransitionNode is a horizontal rule/transition.
	TransitionNode
	// LineBlockNode is a line block (prefixed with |).
	LineBlockNode
)

// Node represents a single element in the RST document tree.
type Node struct {
	Type     NodeType
	Children []*Node

	// Text is the raw text content (for paragraphs, list items, etc.).
	Text string

	// Level is the heading level (1-based).
	Level int

	// DirectiveName is the directive type (e.g., "image", "code-block").
	DirectiveName string
	// DirectiveArg is the argument after the directive name.
	DirectiveArg string
	// Options holds directive options (e.g., :alt:, :align:).
	Options map[string]string
	// Body is the directive body content.
	Body string

	// Label is the label name for LabelNode.
	Label string

	// SubstitutionName is the name for SubstitutionDefNode.
	SubstitutionName string

	// FieldName is the field name for FieldNode.
	FieldName string
	// FieldBody is the field value for FieldNode.
	FieldBody string

	// TableRows holds parsed grid table data.
	// Each row is a slice of cell strings.
	TableRows [][]string
	// TableHeader indicates whether the first row is a header.
	TableHeader bool
}
