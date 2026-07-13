# Spec: Claude Official Plugin Directory Adoption

## Overview

Prepare, validate, package, and submit the OSF Claude Code/Cowork plugin to the
official Claude plugin directory or record the exact external review blocker.

## Functional Requirements

- Validate the Claude plugin package, `.claude-plugin` metadata, MCP server
  config, bundled binary expectations, and install docs.
- Build or verify public ZIP/release artifacts suitable for submission.
- Use Chrome for plugin directory form/login flows; ask the user to log in if
  authentication cannot proceed.
- Improve validation and review-readiness until the package meets all published
  requirements.
- Record submission receipt, review queue URL, or blocker.

## Acceptance Criteria

- Claude plugin validates locally and is submitted or blocked with exact next
  action.
- Package docs explain install, validation, bundled binary, and OSF auth.
- All repo validation gates pass.

## Out Of Scope

- Claiming Anthropic acceptance before a receipt or listing exists.
