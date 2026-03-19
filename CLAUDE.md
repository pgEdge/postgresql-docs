# PostgreSQL Docs - Development Guidelines

This document provides guidelines for Claude Code when working
on this project.

## Project Overview

This project converts upstream documentation from multiple
PostgreSQL ecosystem projects into MkDocs Material sites. A
single Go tool (`pgdoc-converter`) handles five source formats
(SGML, XML, RST, Markdown, and pgBackRest's custom XML),
producing Markdown output and updating `mkdocs.yml` nav
sections.

## Git Branching Model

This project uses an **unusual branching model** that is
critical to understand:

- **`main` branch** contains ONLY the converter tooling
  (`builder/`), `Makefile`, `build-all.sh`, `branches.yml`,
  skeleton `mkdocs.yml`, and MkDocs support files
  (`docs/img/`, `docs/overrides/`, `docs/stylesheets/`).
  There is NO documentation content on `main`.
- **Product branches** (e.g. `pg17`, `pgbackrestmaster`,
  `postgrest145`) each contain the tooling from `main` PLUS
  the generated Markdown documentation for one product at
  one version.
- `build-all.sh` orchestrates this: for each branch defined
  in `branches.yml`, it checks out the branch, merges
  tooling from `main`, runs the converter against the
  upstream source, and commits the result.

**Key implications:**

- Tooling changes (anything under `builder/`, `Makefile`,
  etc.) MUST be committed to `main`. They propagate to
  product branches on the next build.
- Generated docs are NEVER committed to `main`.
- When working on a product branch, the `builder/` code may
  be present but should not be modified there — changes will
  be overwritten on next merge from `main`.
- If you're asked to fix a converter bug while on a product
  branch, switch to `main` first, make the fix, then
  regenerate.

## Builder Architecture

The converter (`builder/main.go`) dispatches to mode-specific
packages based on the `-mode` flag:

| Mode | Package | Source Format | Products |
|------|---------|---------------|----------|
| `sgml` | `convert/` | SGML/DocBook | PostgreSQL |
| `xml` | `convert/` | XML/DocBook | PostGIS |
| `rst` | `rst/` | reStructuredText | pgAdmin, PostgREST, psycopg2 |
| `md` | `md/` | Markdown | PgBouncer, pgvector, pgAudit |
| `backrest` | `backrest/` | Custom XML | pgBackRest |

Each converter implements a common pattern:

- `Convert()` — run the full pipeline
- `Files()` — return `[]*shared.FileEntry` for nav generation
- `Warnings()` — return accumulated warnings
- `ProjectName()` — return display name for `site_name`

After conversion, `nav.BuildNav()` and `nav.UpdateMkdocsYML()`
generate the nav section and update the mkdocs config.

### Shared infrastructure (`shared/`, `nav/`, `validate/`)

- `shared.FileEntry` — path, title, order, nav parent
- `shared.MarkdownWriter` — buffered writer with heading,
  code block, admonition, and indentation helpers
- `nav.BuildNav()` — builds nav tree from file entries
- `nav.UpdateMkdocsYML()` — replaces nav section, sets
  site_name, ensures `md_in_html` extension
- `validate/` — post-conversion link checker

### pgBackRest converter notes

pgBackRest uses its own XML DTD (NOT DocBook). Key
differences from the other converters:

- Uses Go's `encoding/xml` (not the SGML parser) because
  pgBackRest element names like `<host>`, `<code>`, `<p>`
  conflict with the SGML parser's HTML element filter.
- `{[key]}` variable substitution (not XML entities) for
  project name, version, etc. Multi-pass resolution needed
  for chained references.
- `<!ENTITY name SYSTEM "path">` for release note includes
  — manually resolved before XML parsing.
- Documents with many top-level `<section>` elements are
  split into multi-page directories; small docs (FAQ,
  metrics, coding, contributing, etc.) stay single-page.
- Nav titles use subtitle when the title is just `{[project]}`
  (which resolves to "pgBackRest" for most pages).
- `command.xml` and `configuration.xml` are auto-generated
  by building pgBackRest and live in `output/xml/` — they
  won't exist unless pgBackRest has been built. Missing-file
  link warnings for these are expected.

## Code Style

- Use **4 spaces** for indentation (not tabs) in Go code
- Use **tabs** for indentation in Go code (standard `gofmt`)
  — note: `gofmt` enforces tabs; the 4-space rule applies
  to non-Go files (Markdown, YAML, etc.)
- Always run `make lint` before committing Go changes

## Testing

- Always add tests for new functionality
- Run `make test` and `make lint` to validate
- Run full test output — never tail or trim stdout/stderr
- Only modify tests when changes are **expected** to cause
  failures; investigate unexpected failures first
- Clean up temporary test files

## Documentation Standards

- All tooling documentation belongs in the top-level
  `README.md`
- Content under `docs/` (except support files) is entirely
  generated — never hand-edit generated Markdown
- Wrap Markdown at **79 characters**; exceptions for URLs,
  code samples, and tables
- Leave a **blank line** before the first item in any list
- Use UPPERCASE for root-level Markdown files (`README.md`,
  `CLAUDE.md`)

## Build & Validation Checklist

- [ ] Code uses correct indentation
- [ ] `make test` passes
- [ ] `make lint` passes (includes `gofmt` and `go vet`)
- [ ] Changes committed to the correct branch (`main` for
      tooling, product branch for generated docs)
- [ ] `README.md` updated if adding new modes or features
- [ ] No temporary files left behind
