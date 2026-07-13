# JASP OSF integration comparison

Reviewed 2026-07-13 against the public `development` branch of
[jasp-stats/jasp-desktop](https://github.com/jasp-stats/jasp-desktop).

## Evidence

JASP is a cross-platform desktop statistics application released under
[AGPL-3.0](https://github.com/jasp-stats/jasp-desktop), with its repository
page reporting 977 stars, 9,596 commits, and release `v0.98.1` on 2026-07-07
at the time of review. The relevant public implementation is concentrated in
the [OSF file menu](https://github.com/jasp-stats/jasp-desktop/blob/development/Desktop/components/JASP/Widgets/FileMenu/OSF.qml),
[login flow](https://github.com/jasp-stats/jasp-desktop/blob/development/Desktop/components/JASP/Widgets/FileMenu/OSFLogin.qml),
and [OSF storage node](https://github.com/jasp-stats/jasp-desktop/blob/development/Desktop/osf/onlinedatanodeosf.cpp).

The storage node follows OSF JSON:API relationships and pagination, resolves
file and folder paths, and uses provider links for download, upload, folder
creation, deletion, and file information. It also reads
`extra.hashes.md5`. The QML surface provides authenticated desktop browsing,
breadcrumbs, sorting, save/open dialogs, and login/logout controls.

## Capability comparison

| Capability | JASP | OSF CLI Go | Decision |
|---|---|---|---|
| Authenticated OSF access | Desktop username/password login and logout | Token or username/password with redaction and explicit auth modes | Existing CLI contract retained |
| Browse files and folders | Desktop file menu, breadcrumbs, sorting | `files list`, folder path traversal, table/JSON, MCP `osf_files_list` | Existing parity confirmed |
| Download and upload | Provider-backed desktop file transfer | CLI download/upload with path and conflict safeguards | Existing parity confirmed |
| Create and delete folders/files | Desktop actions backed by provider links | Explicit CLI `mkdir`, `rm`, and upload commands | Existing parity confirmed |
| Provider checksum metadata | Reads `extra.hashes.md5` | API model and CLI/MCP JSON now expose optional `md5` | Implemented in this track |
| Desktop save/open experience | QML dialogs and remembered UI state | Shell-oriented commands and `open` URL integration | Rejected as a desktop UI concern; CLI equivalents are documented |
| JASP analysis state and autosave | JASP-specific application behavior | No JASP analysis runtime or file format dependency | Rejected as outside OSF client scope |
| Cross-platform distribution | Windows, macOS, Linux desktop releases | Cross-platform CLI, MCP package, checksums, and release gates | Different surfaces; no gap identified |

## Outcome

The only material maintainable gap identified at the OSF client contract layer
was provider checksum preservation. The implementation is intentionally
metadata-only: it never downloads file content to calculate a checksum. JASP's
desktop UI and JASP-specific analysis lifecycle remain explicit non-goals,
because reproducing them would add a desktop application dependency without
improving the CLI, API, or MCP contract.
