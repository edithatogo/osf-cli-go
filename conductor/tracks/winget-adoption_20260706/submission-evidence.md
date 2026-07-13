# WinGet Submission Evidence

## Submission

- Pull request: <https://github.com/microsoft/winget-pkgs/pull/401414>
- Submitted package: `Edithatogo.OSFCLI.Go` version `0.3.2`.
- Manifest path: `manifests/e/Edithatogo/OSFCLI/Go/0.3.2/`.
- Submission branch: `edithatogo:osf-cli-go-0.3.2`.

## Technical Fix

The first WinGet service validation failed with:

```text
RequiredFieldMissing: Required field missing. (DefaultLocale)
```

The version manifest was corrected to the canonical version-manifest shape
with `DefaultLocale: en-US`; package and installer metadata remain in their
dedicated manifests. The source packet is:
`packaging/winget/Edithatogo.OSFCLI.Go.yaml`.

The original validation run is recorded by the upstream bot at:
<https://dev.azure.com/shine-oss/8b78618a-7973-49d8-9174-4360829d979b/_build/results?buildId=364691>

## Current External State

- The upstream PR remains open for revalidation and maintainer review.
- The Microsoft CLA check is pending. The repository requires the contributor
  to explicitly accept the CLA; no legal agreement was made automatically.
- No WinGet publication or acceptance is claimed until the upstream PR merges.
- Track cleanup is intentionally deferred while the external CLA and
  maintainer-review gates remain unresolved.

## Local Validation

- All three manifests parse as YAML.
- `go test ./...`, `go test -race ./...`, `go vet ./...` passed.
- `go run ./tools/checkstubs`, `checkreviews`, `checkfeaturematrix`,
  `checkregistries`, and `checkreleasecontract` passed.
- `govulncheck ./...` and `git diff --check` passed.
