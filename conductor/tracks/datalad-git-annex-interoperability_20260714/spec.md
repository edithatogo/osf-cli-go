# DataLad and git-annex interoperability contract

## Objective

Address issue #69 by defining a safe optional interoperability boundary. The
core CLI and MCP server remain OSF-native and must not silently create Git,
DataLad, or git-annex state.

## Contract

- Any future integration is a separately versioned companion tool or explicit
  command group.
- Repository mutations require an explicit destination, conflict policy, and
  dry-run or manifest preview.
- Git remote and git-annex special-remote behavior is validated against local
  fixtures before any opt-in live test.
- Credential-manager access remains outside the core client and is injected by
  the companion boundary; secrets never enter manifests or logs.
- Linux, macOS, and Windows behavior must be tested independently.

## Deliberate non-goal

This track does not add a partial git-remote helper to the core repository.
Protocol implementation requires maintainer agreement and executable fixtures;
the contract is complete and the implementation remains a separately scoped
future feature.
