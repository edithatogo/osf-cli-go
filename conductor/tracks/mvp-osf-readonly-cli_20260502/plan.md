# Plan: MVP OSF Read-Only CLI

## Phase 1: CLI Foundation

- [x] Task: Create Go module and baseline `osf` command scaffold
- [ ] Task: Add command router with help, version, and consistent error handling
- [ ] Task: Add table and JSON output primitives

## Phase 2: OSF API Client

- [ ] Task: Implement typed OSF API client with base URL, auth header, context, and status-aware errors
- [ ] Task: Implement JSON:API pagination traversal
- [ ] Task: Add fixtures and offline tests for users, nodes, contributors, and files

## Phase 3: Read-Only Commands

- [ ] Task: Add `auth whoami` using `OSF_TOKEN`
- [ ] Task: Add `projects list`
- [ ] Task: Add `nodes get <guid-or-url>`
- [ ] Task: Add `files list <node-guid>`

## Phase 4: Download

- [ ] Task: Resolve OSF Storage file and folder paths
- [ ] Task: Download a single file with fail/skip/overwrite conflict policies
- [ ] Task: Download a folder tree with manifest output

## Phase 5: Documentation And Release Readiness

- [ ] Task: Document install, auth, command examples, and safety defaults
- [ ] Task: Add integration-test instructions gated by environment variables
- [ ] Task: Add release checklist for Windows, macOS, and Linux binaries
