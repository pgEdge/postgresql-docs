# PostgreSQL Documentation

[![CI](https://github.com/pgEdge/postgresql-docs/actions/workflows/ci.yml/badge.svg)](https://github.com/pgEdge/postgresql-docs/actions/workflows/ci.yml)

MkDocs Material site for the PostgreSQL documentation, converted
from the upstream SGML sources. Also supports converting pgAdmin 4
RST (reStructuredText) documentation.

The `main` branch contains the builder tooling. Additional
branches contain generated docs for each major version, named
`pg<version>` (e.g. `pg17`, `pg18`).

## Prerequisites

- Go 1.25+
- Python 3 with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/)
- PostgreSQL source tree (for SGML docs) or pgAdmin source
  (for RST docs)

## Quick Start

Place the upstream documentation source at `/doc-source` (the
default `SRC_DIR`). For PostgreSQL branches that means the
`doc/src/sgml/` directory; for pgAdmin branches the
`docs/en_US/` directory.

Build the converter and generate the docs:

```sh
make build
make convert VERSION=19devel          # PostgreSQL (SGML)
make convert-rst VERSION=9.13         # pgAdmin (RST)
```

Preview the site locally:

```sh
mkdocs serve
```

## Builder

The `builder/` directory contains a Go tool that converts
PostgreSQL's SGML/DocBook documentation and pgAdmin's RST
documentation to Markdown. It handles:

- SGML entity resolution and parsing
- DocBook-to-Markdown conversion (inline, block, tables, xrefs)
- RST parsing and conversion (headings, directives, tables,
  cross-references, toctree resolution)
- `func_table_entry` tables split into proper multi-column layout
- Image copying from the source tree
- MkDocs nav generation from document structure

### Makefile Targets

| Target     | Description                              |
| ---------- | ---------------------------------------- |
| `setup`    | Configure git hooks (run once after clone)|
| `build`    | Compile the converter to `bin/`          |
| `test`     | Run all Go tests                         |
| `lint`     | Run `gofmt` and `go vet`                |
| `convert`  | Build and run the SGML converter         |
| `convert-rst` | Build and run the RST converter       |
| `validate` | Build and run with link validation       |
| `clean`    | Remove the compiled binary               |

### Usage

```
pgdoc-converter [flags]
  -mode        Conversion mode: sgml or rst (default "sgml")
  -src         Path to source documentation directory
  -out         Output directory for .md files (default "./docs")
  -mkdocs      Path to mkdocs.yml (default "./mkdocs.yml")
  -version     Version label (e.g. "17.2" or "9.13")
  -copyright   Copyright string (RST mode)
  -pgadmin-src Path to pgAdmin source (for literalinclude)
  -validate    Run link validation after conversion
  -verbose     Verbose output
```

## TODO: Additional Component Docs Sites

The knowledgebase builder YAML already includes these
non-pgEdge components. We should investigate building
MkDocs sites for them, similar to what we've done for
PostgreSQL:

- [x] pgAdmin 4 (RST converter implemented)
- [ ] PgBouncer (1.24-1.25)
- [ ] pgBackRest (2.56-2.57)
- [ ] PostGIS (3.5.3-3.5.5)
- [ ] pgvector (0.8.0-0.8.1)
- [ ] pgAudit (16.1-18.0)
- [ ] psycopg2 (2.9.10)
- [ ] PostgREST (14.5)

## Project Structure

```
builder/          Go converter source
  shared/           Shared types (FileEntry, IDEntry, writer)
  convert/          SGML-to-Markdown conversion
  sgml/             SGML tokenizer, parser, entity resolver
  rst/              RST parser, converter, directive handlers
  nav/              MkDocs nav YAML generation
  validate/         Link validation
docs/             Generated Markdown + MkDocs support files
  img/              Site images (logo, favicon)
  overrides/        MkDocs Material template overrides
  stylesheets/      Custom CSS
mkdocs.yml        MkDocs configuration
Makefile          Build targets
```
