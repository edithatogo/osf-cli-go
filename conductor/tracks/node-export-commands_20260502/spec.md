# Node Export and Snapshot Commands

## Objective

Create an `osf export` command that produces a comprehensive node snapshot including metadata, files, contributors, wiki content, component hierarchy, identifiers, and activity logs.

## Acceptance Criteria

- Single `osf export <node-id>` command captures all node data in one pass
- Exported snapshot includes: node metadata, file tree, contributor list, wiki pages, child components, DOIs/identifiers, and recent activity logs
- Output supports JSON format (default) for programmatic use
- Output supports human-readable summary mode
- Export handles partial failures gracefully (logs per-section errors without aborting the full export)

## Non-Goals

- Export-to-zip or archive file creation
- Cross-node bulk export
- Import or restore functionality
- Export scheduling or automation
