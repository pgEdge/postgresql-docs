package md

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterDocFiles(t *testing.T) {
	input := []string{
		"README.md",
		"CHANGELOG.md",
		"CONTRIBUTING.md",
		"LICENSE.md",
		"CODE_OF_CONDUCT.md",
		"config.md",
		"usage.md",
		"frag-config-man.md",
		"frag-usage-man.md",
		"changes.md",
	}
	got := filterDocFiles(input)
	want := []string{"README.md", "config.md", "usage.md"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGithubAnchor(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Getting Started", "getting-started"},
		{"Installation & Setup", "installation--setup"},
		{"pgvector", "pgvector"},
		{"Half-Precision Vectors", "half-precision-vectors"},
		{"C++ Example", "c-example"},
		{"What's New in v2", "whats-new-in-v2"},
		{"foo_bar_baz", "foo_bar_baz"},
	}
	for _, tt := range tests {
		got := githubAnchor(tt.title)
		if got != tt.want {
			t.Errorf("githubAnchor(%q) = %q, want %q",
				tt.title, got, tt.want)
		}
	}
}

func TestSplitMarkdown(t *testing.T) {
	content := `# My Project

Intro paragraph here.

Some more intro.

## Installation

Install instructions.

### From Source

Build from source.

## Usage

Usage instructions.

## FAQ

Frequently asked questions.
`
	res := splitMarkdown(content)

	if res.title != "My Project" {
		t.Errorf("title = %q, want %q", res.title, "My Project")
	}

	if !strings.Contains(res.intro, "Intro paragraph") {
		t.Error("intro should contain intro paragraph")
	}
	if strings.Contains(res.intro, "## Installation") {
		t.Error("intro should not contain H2 sections")
	}

	if len(res.sections) != 3 {
		t.Fatalf("got %d sections, want 3", len(res.sections))
	}

	wantTitles := []string{"Installation", "Usage", "FAQ"}
	for i, s := range res.sections {
		if s.title != wantTitles[i] {
			t.Errorf("section[%d].title = %q, want %q",
				i, s.title, wantTitles[i])
		}
	}

	// Installation section should include H3 subsection
	if !strings.Contains(res.sections[0].content, "### From Source") {
		t.Error("Installation should contain subsection")
	}
}

func TestSplitMarkdownNoH2(t *testing.T) {
	content := "# Title\n\nJust a simple doc.\n"
	res := splitMarkdown(content)
	if res.title != "Title" {
		t.Errorf("title = %q, want %q", res.title, "Title")
	}
	if len(res.sections) != 0 {
		t.Errorf("got %d sections, want 0", len(res.sections))
	}
}

func TestSplitMarkdownCodeBlock(t *testing.T) {
	content := "# Title\n\n## Real Section\n\n" +
		"```markdown\n## Not A Section\ncode here\n```\n\n" +
		"After code.\n"
	res := splitMarkdown(content)
	if len(res.sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(res.sections))
	}
	if res.sections[0].title != "Real Section" {
		t.Errorf("title = %q, want %q",
			res.sections[0].title, "Real Section")
	}
	if !strings.Contains(res.sections[0].content,
		"## Not A Section") {
		t.Error("code block content should be preserved")
	}
}

func TestPromoteHeadings(t *testing.T) {
	input := "## Title\n\nText\n\n### Sub\n\n#### Deep\n"
	got := promoteHeadings(input)
	if !strings.Contains(got, "# Title\n") {
		t.Error("H2 should become H1")
	}
	if !strings.Contains(got, "## Sub\n") {
		t.Error("H3 should become H2")
	}
	if !strings.Contains(got, "### Deep\n") {
		t.Error("H4 should become H3")
	}
}

func TestPromoteHeadingsCodeBlock(t *testing.T) {
	input := "## Title\n\n```\n## Not promoted\n```\n"
	got := promoteHeadings(input)
	if !strings.Contains(got, "# Title\n") {
		t.Error("H2 should become H1")
	}
	if !strings.Contains(got, "## Not promoted") {
		t.Error("heading inside code block should not be promoted")
	}
}

func TestRewriteAnchors(t *testing.T) {
	anchorMap := map[string]string{
		"installation": "installation.md",
		"from-source":  "installation.md#from-source",
		"usage":        "usage.md",
	}

	input := "See [install](#installation) and " +
		"[build](#from-source) and [use](#usage) and " +
		"[unknown](#other)."

	got := rewriteAnchors(input, anchorMap)

	if !strings.Contains(got, "](installation.md)") {
		t.Error("should rewrite #installation")
	}
	if !strings.Contains(got,
		"](installation.md#from-source)") {
		t.Error("should rewrite #from-source")
	}
	if !strings.Contains(got, "](usage.md)") {
		t.Error("should rewrite #usage")
	}
	if !strings.Contains(got, "](#other)") {
		t.Error("should preserve unknown anchors")
	}
}

func TestBuildAnchorMap(t *testing.T) {
	sections := []section{
		{
			title: "Installation",
			slug:  "installation",
			content: "## Installation\n\n### From Source\n\n" +
				"Build steps.\n",
		},
		{
			title:   "Usage",
			slug:    "usage",
			content: "## Usage\n\nUse it.\n",
		},
	}
	m := buildAnchorMap(sections)

	if m["installation"] != "installation.md" {
		t.Errorf("installation = %q", m["installation"])
	}
	if m["from-source"] != "installation.md#from-source" {
		t.Errorf("from-source = %q", m["from-source"])
	}
	if m["usage"] != "usage.md" {
		t.Errorf("usage = %q", m["usage"])
	}
}

func TestConvertAlerts(t *testing.T) {
	input := `Before.

> [!NOTE]
> This is a note.
> Second line.

After.

> [!WARNING]
> Be careful.

> Normal blockquote.
`
	got := convertAlerts(input)

	if !strings.Contains(got, "!!! note") {
		t.Error("should convert NOTE alert")
	}
	if !strings.Contains(got, "    This is a note.") {
		t.Error("should indent note body")
	}
	if !strings.Contains(got, "    Second line.") {
		t.Error("should include continuation lines")
	}
	if !strings.Contains(got, "!!! warning") {
		t.Error("should convert WARNING alert")
	}
	if !strings.Contains(got, "> Normal blockquote.") {
		t.Error("should preserve normal blockquotes")
	}
}

func TestConvertAlertsCaution(t *testing.T) {
	input := "> [!CAUTION]\n> Danger zone.\n"
	got := convertAlerts(input)
	if !strings.Contains(got, "!!! danger") {
		t.Error("CAUTION should map to danger")
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		content  string
		filename string
		want     string
	}{
		{"# My Title\n\nBody.", "file.md", "My Title"},
		{"No heading here.", "config.md", "config"},
		{"## Only H2\n\nBody.", "readme.md", "readme"},
	}
	for _, tt := range tests {
		got := extractTitle(tt.content, tt.filename)
		if got != tt.want {
			t.Errorf("extractTitle(%q, %q) = %q, want %q",
				tt.content[:20], tt.filename, got, tt.want)
		}
	}
}

func TestConverterSplitFile(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	readme := `# Test Project

Introduction.

## Installation

Install steps.

## Usage

Usage info.
`
	if err := os.WriteFile(
		filepath.Join(srcDir, "README.md"),
		[]byte(readme), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewConverter(srcDir, outDir, "Test v1.0", false)
	if err := c.Convert(); err != nil {
		t.Fatal(err)
	}

	files := c.Files()
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}

	// Check index.md exists
	indexData, err := os.ReadFile(
		filepath.Join(outDir, "index.md"))
	if err != nil {
		t.Fatal("index.md not created")
	}
	if !strings.Contains(string(indexData), "# Test Project") {
		t.Error("index.md should contain title")
	}

	// Check section files
	instData, err := os.ReadFile(
		filepath.Join(outDir, "installation.md"))
	if err != nil {
		t.Fatal("installation.md not created")
	}
	if !strings.Contains(string(instData), "# Installation") {
		t.Error("installation.md should have promoted H1")
	}

	usageData, err := os.ReadFile(
		filepath.Join(outDir, "usage.md"))
	if err != nil {
		t.Fatal("usage.md not created")
	}
	if !strings.Contains(string(usageData), "# Usage") {
		t.Error("usage.md should have promoted H1")
	}
}

func TestConverterCopyFiles(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	os.WriteFile(filepath.Join(srcDir, "config.md"),
		[]byte("# Configuration\n\nConfig docs.\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "usage.md"),
		[]byte("# Usage\n\nUsage docs.\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "frag-config-man.md"),
		[]byte("Fragment content.\n"), 0644)

	c := NewConverter(srcDir, outDir, "PgBouncer 1.25", false)
	if err := c.Convert(); err != nil {
		t.Fatal(err)
	}

	files := c.Files()
	// 2 doc files + generated index = 3
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}

	// Fragment should not be copied
	if _, err := os.Stat(
		filepath.Join(outDir, "frag-config-man.md")); err == nil {
		t.Error("fragment file should not be copied")
	}

	// Index should be generated
	indexData, err := os.ReadFile(
		filepath.Join(outDir, "index.md"))
	if err != nil {
		t.Fatal("index.md not created")
	}
	if !strings.Contains(string(indexData), "PgBouncer 1.25") {
		t.Error("index should contain project name")
	}
}

func TestConverterCopyFilesWithREADME(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()

	os.WriteFile(filepath.Join(srcDir, "README.md"),
		[]byte("# Home\n\nWelcome.\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "guide.md"),
		[]byte("# Guide\n\nGuide content.\n"), 0644)

	c := NewConverter(srcDir, outDir, "Test v1", false)
	if err := c.Convert(); err != nil {
		t.Fatal(err)
	}

	// README.md should become index.md
	if _, err := os.Stat(
		filepath.Join(outDir, "index.md")); err != nil {
		t.Error("README.md should be renamed to index.md")
	}

	files := c.Files()
	// README→index + guide = 2 (no generated index needed)
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
}

func TestFindMarkdownFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"),
		[]byte("# Hi\n"), 0644)
	os.WriteFile(filepath.Join(dir, "GUIDE.MD"),
		[]byte("# Guide\n"), 0644)
	os.WriteFile(filepath.Join(dir, "Makefile"),
		[]byte("all:\n"), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	files, err := findMarkdownFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(files), files)
	}
}
