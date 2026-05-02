# Read-Only Commands

## Objective

Implement user-facing read-only commands for projects, components, and files using the CLI and API client contracts.

## Acceptance Criteria

- `projects list` lists accessible projects.
- `projects get <guid-or-url>` shows one project.
- `components list <project-guid-or-url>` lists child components.
- `files list <project-or-component-guid>` lists OSF Storage paths.
- Each command has human and JSON output tests.
