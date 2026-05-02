// osf-cli-go is a command-line client for the Open Science Framework (OSF).
//
// # Overview
//
// osf-cli-go provides read-only access to OSF projects, components, and files
// through a Cobra-based CLI. It supports both JSON and human-readable output.
//
// # Commands
//
//	osf auth whoami                  — Show the authenticated OSF account
//	osf projects list                — List accessible projects
//	osf projects get <guid-or-url>   — Show one project
//	osf components list <project>    — List child components
//	osf files list <node>            — List OSF Storage files
//	osf files download --file <id>   — Download a single file
//	             --tree <node> <dest> — Download a folder tree
//	osf completion bash|zsh|fish|powershell — Shell completions
//
// # Authentication
//
// Set the OSF_TOKEN environment variable to a valid OSF personal access token.
// Without a token, only public node metadata and public storage files are accessible.
//
// # Package Layout
//
//	cmd/osf/            — CLI entry point
//	internal/auth/      — Token loading and redaction
//	internal/cli/       — Cobra command definitions and routing
//	internal/download/  — Safe file download with conflict handling
//	internal/osfapi/    — OSF API v2 JSON:API client
//	internal/output/    — JSON and table output helpers
package osfcli
