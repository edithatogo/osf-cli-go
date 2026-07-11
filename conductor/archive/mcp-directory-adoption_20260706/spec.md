# Spec: MCP.Directory Adoption

## Overview

Submit, verify, and optimize OSF MCP server discovery in MCP.Directory, using
public repo materials and recorded submission evidence.

## Functional Requirements

- Verify whether the previous submission is listed, pending, rejected, or stale.
- Prepare current copy/paste submission fields from repo-local metadata.
- Use Chrome for browser submission, search, and verification; ask the user to
  log in if the flow requires account authentication.
- Improve any quality/discoverability score or listing completeness as close to
  100% as possible.
- Record submission receipts, listing URLs, screenshots where safe, and blockers.

## Acceptance Criteria

- MCP.Directory has a verified listing/submission receipt or a precise blocker.
- Repo-local submission packet and `registry/directory-submissions.json` match
  the latest outcome.
- All validation gates pass.

## Out Of Scope

- Misrepresenting pending manual review as published.
- Storing browser credentials or private OSF data.
