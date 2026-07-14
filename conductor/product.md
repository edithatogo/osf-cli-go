# OSF CLI Go Product Guide

## Initial Concept

Build a fast, scriptable Go CLI for the Open Science Framework (OSF) that helps researchers, research software engineers, and evidence/package maintainers inspect, download, upload, synchronize, and export OSF project content from the terminal.

## Users

- Researchers who need a simple terminal workflow for public and private OSF projects.
- Research software engineers who need deterministic automation around OSF APIs.
- Data stewards and librarians who need to audit files, metadata, contributors, components, and registrations.
- Power users moving beyond GUI workflows in JASP, OpenSesame, Protocols.io, or custom scripts.

## Product Goals

- Provide a clean `osf` CLI with predictable commands, useful help text, and safe defaults.
- Provide opt-in, redacted operational events for troubleshooting without polluting command output or collecting telemetry.
- Treat the OSF API v2 documentation as the technical dictionary for object shapes, attributes, relationships, pagination, and allowed values.
- Treat the OSF Dictionary as the conceptual glossary for user-facing language such as projects, components, registrations, preprints, forks, embargoes, contributors, wiki, and files.
- Preserve metadata fidelity using the OSF Metadata Profile where project metadata maps to persistent identifiers, FAIR metadata, DataCite-oriented concepts, or discovery metatags.
- Learn from `osfclient`, `osfr`, and `osf-project-exporter` without cloning their UX limitations.

## MVP Scope

- Authenticate with an OSF personal access token without storing secrets in project files.
- Retrieve and display the current OSF user.
- List projects and components the user can access.
- Inspect one node by GUID, including title, category, public/private status, contributors, tags, dates, links, and relationships.
- List OSF Storage files and folders for a node.
- Download a single file or a folder tree with clear conflict handling.
- Emit JSON output for automation and table output for humans.
- Emit versioned structured operational events to an operator-selected local destination when explicitly enabled.

## Later Scope

- Upload files and directories with explicit conflict policies: error, skip, overwrite.
- Create projects and components with category validation.
- Manage tags, metadata fields, contributors, and licenses where OSF API permissions allow.
- Export project snapshots including metadata, contributor lists, file inventories, wiki content, component hierarchy, and activity logs.
- Support reproducible manifests for backup and audit workflows.
- Add provider-scoped Zenodo discovery and carefully gated transfer/publication
  workflows after the provider contract and sandbox evidence are complete.
- Add shell completions and packaged releases for Windows, macOS, and Linux.

## Reference Tools And Lessons

- OSF API v2 docs: authoritative data model, endpoints, pagination, relationships, and validation rules.
- OSF Dictionary: user-facing terminology for projects, components, registrations, preregistrations, embargoes, wikis, files, contributors, and add-ons.
- OSF Metadata Profile: metadata vocabulary and FAIR/discovery mapping reference.
- Zenodo developer documentation, sandbox guidance, terms, and repository
  policies: authoritative inputs for the post-1.0 provider roadmap.
- `osfclient`: familiar CLI concepts such as init, list, clone, fetch, upload, remove, and local config.
- `osfr`: broad OSF entity coverage, explicit conflicts behavior, and distinction between nodes, files, and users.
- `osf-project-exporter`: project export target that includes metadata, files, contributors, wiki content, and components.
- Protocols.io, JASP, and OpenSesame: adjacent research workflows that show OSF is often part of a broader reproducibility toolchain.

## Non-Goals For The First Release

- Reimplement the full OSF web application.
- Hide OSF permission failures behind optimistic behavior.
- Store passwords. Prefer personal access tokens and environment variables.
- Invent metadata terms that diverge from OSF API v2 or OSF help documentation.

## Product Principles

- Scriptable first: every command should be useful in CI, scheduled jobs, and reproducible research pipelines.
- Human-readable by default: tables and concise messages should make terminal use pleasant.
- JSON on demand: machine-readable output should be stable enough for downstream scripts.
- Permission-aware: private content and destructive writes must require explicit user intent.
- Cross-platform: Windows, macOS, and Linux behavior should be consistent.
