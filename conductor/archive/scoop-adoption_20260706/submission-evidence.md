# Scoop Submission Evidence

## Project-Owned Bucket

- Bucket: <https://github.com/edithatogo/scoop-osf>
- Branch: `main`
- Manifest: `bucket/osf-cli-go.json`
- Bucket commit: `70743f258c995d8f47252c8ef5c0da3daf4834c2`
- Install command:
  `scoop bucket add osf https://github.com/edithatogo/scoop-osf`
- Package route: `scoop install osf/osf-cli-go`
- Release: `v0.3.2`

## Validation

- Manifest parses as JSON.
- x64 and arm64 URLs point to the public v0.3.2 release.
- x64 and arm64 hashes match the published Windows checksum files.
- The manifest includes `osf.exe` installation and GitHub checkver/autoupdate
  metadata.

## Main Bucket Boundary

Scoop Main PR [#8293](https://github.com/ScoopInstaller/Main/pull/8293) was
closed because the project did not meet the bucket's popularity threshold.
The project-owned bucket is the supported distribution route. No Scoop Main
acceptance is claimed.
