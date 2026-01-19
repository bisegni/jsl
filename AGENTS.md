# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the CLI entry point and delegates to `cmd/` (Cobra commands).
- `cmd/` defines user-facing commands like `query`, `format`, `convert`, `stats`, and `validate`.
- `pkg/` holds core logic: parsing (`pkg/parser`), query planning/execution (`pkg/query`, `pkg/engine`), and storage abstractions (`pkg/database`).
- `examples/` contains sample JSON/JSONL inputs referenced by `EXAMPLES.md`.

## CLI Architecture (High Level)
- `main.go` bootstraps Cobra and forwards to `cmd.Execute()`.
- `cmd/root.go` handles argument routing:
  - SQL-like `SELECT ...` goes to `engine.ParseQuery` then `engine.Executor.Execute`.
  - Filter expressions go through `pkg/query` and existing `RunFilter`.
  - Path queries default to `RunQuery`.
- `pkg/engine` builds a table pipeline (input -> optional filter -> optional projection) and streams rows to JSON output.
- `pkg/query` owns filter parsing/matching and path extraction.
- `pkg/database` defines the storage abstraction and adapters (currently JSON/JSONL).

## CLI Flow (Diagram)
- `main.go` -> `cmd/root.go` -> `engine.ParseQuery` -> `database.NewJSONTable` -> `engine.Executor.Execute`
- `Executor.Execute` -> `Table.Iterate` -> optional `FilterTable` -> optional `ProjectTable` -> JSON array encoder

## Database/Storage Management
- There is no external DB; storage is an interface-driven adapter over input sources.
- `pkg/database.Table` is the core abstraction; it yields a `RowIterator` which yields `Row`.
- `pkg/database.JSONTable` is the default implementation:
  - Uses `pkg/parser.Parser` to stream JSON or JSONL via `Parser.Read()`.
  - Supports stdin (`-`) and inline JSON strings via temporary files.
  - Produces `database.JSONRow` wrapping a `parser.Record` or map.
- The engine never opens files directly; it consumes a `database.Table`.

## AI Agent Architecture Rules
- Keep CLI routing in `cmd/`; do not embed query execution logic directly in `main.go` or Cobra handlers beyond argument parsing.
- Use `pkg/database.Table` and `RowIterator` for new data sources; do not bypass the abstraction by reading files directly inside `pkg/engine`.
- Keep JSON/JSONL parsing inside `pkg/parser` and adapters in `pkg/database`.
- Preserve the executor pipeline order: input -> filter -> projection -> JSON array output.
- Avoid cross-package import cycles; `pkg/engine` should only depend on `pkg/query`, `pkg/database`, and `pkg/parser` (not `cmd/`).

## Build, Test, and Development Commands
- `go build -o jsl` builds the CLI binary locally.
- `go test ./...` runs the full unit test suite across `pkg/`.
- `go run . --help` runs the CLI without building a binary.

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`) and Go module layout.
- Indentation is tabs per Go conventions; keep lines readable and avoid long chained logic.
- File and type names follow Go norms: `snake_case` for files, `CamelCase` for exported types, `mixedCase` for locals.

## Testing Guidelines
- Tests live next to implementation files under `pkg/` and use `*_test.go` naming.
- Prefer table-driven tests for query and parser behavior.
- Run targeted tests with `go test ./pkg/query -run TestName` when iterating.

## Commit & Pull Request Guidelines
- Recent commits follow Conventional Commits (e.g., `feat:`, `docs:`); match this style when possible.
- Keep commit subjects short and scoped to one change.
- PRs should describe behavior changes, include example commands/output when relevant, and link any related issues.

## Configuration & Usage Notes
- The CLI auto-detects `.json` and `.jsonl` inputs; other extensions fall back to JSON then JSONL.
- For quick manual checks, use `examples/` files and commands from `EXAMPLES.md` (e.g., `jsl examples/users.json "SELECT name"`).
