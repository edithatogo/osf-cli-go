# MCP quality and compatibility evaluation harness

## Overview

Build a repeatable harness for MCP tool contract quality, client compatibility,
security behavior, and release metadata consistency.

## Requirements

- Validate names, descriptions, schemas, structured output, limits, errors,
  redaction, pagination, and resource bounds offline.
- Exercise supported client configuration shapes and package manifests.
- Support opt-in live OSF checks without committed credentials or unsafe writes.
- Emit versioned reports suitable for release and registry submissions.

## Acceptance criteria

- Harness is rerunnable in CI and locally.
- Failure output identifies the exact tool, manifest, or contract mismatch.
- Release and registry gates consume the report or explicitly document why not.
