# CLI JSON Contract

The CLI's machine-readable contract is emitted by `osf --output json`.

For the 1.x compatibility window:

- command names and their `status` values are stable;
- existing JSON fields are not renamed or retyped in minor releases;
- new fields may be added and consumers must ignore unknown fields;
- `--json` is equivalent to `--output json`;
- exit code `0` means the command completed, `1` means a runtime failure, and
  `2` means invalid usage or arguments.

The root contract is covered by an exact command-name regression test in
`internal/cli/cli_test.go`. Individual command JSON tests cover high-risk
authentication, file, search, export, validation, and registration surfaces.
Breaking changes require a major version, a migration note, and a deprecation
period where practical.
