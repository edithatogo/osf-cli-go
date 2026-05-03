# Plan: Node Export and Snapshot Commands

## Phase 1: Export Data Types and Collector

- [ ] Task: Define `NodeExport` struct with fields for metadata, files, contributors, wikis, components, identifiers, and logs
- [ ] Task: Implement `Collector` that calls each relevant OSF API endpoint and populates the struct
- [ ] Task: Handle partial failures per-section (collect what is available, record errors per section)
- [ ] Task: Add fixture-backed tests for the collector

## Phase 2: JSON Export Output

- [ ] Task: Implement JSON serialization of the `NodeExport` struct with pretty-print option
- [ ] Task: Add write-to-file support with timestamped default filename
- [ ] Task: Add fixture-backed tests for export output correctness

## Phase 3: CLI `export` Command with `--output` Modes

- [ ] Task: Add `osf export <node-id>` command with `--output json|summary` flag
- [ ] Task: Implement human-readable summary mode showing key stats (files count, contributors, wikis, etc.)
- [ ] Task: Add `--file` flag to write output to a file path
- [ ] Task: Run quality gates and `$conductor-review`, apply fixes, and write phase review evidence
