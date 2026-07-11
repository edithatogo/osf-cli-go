# Glama release evidence

Validated on 2026-07-11 against the claimed Glama listing at
<https://glama.ai/mcp/servers/edithatogo/osf-cli-go>.

- Repository revision: `046866e590042d9eb23fc9a5309d6181e528e749`
- Glama build test: `019f5050-5755-7541-aaee-7ca854fcf4bc`
- Build result: successful in 44 seconds
- Glama image: `registry.glama.ai/mcp-qmag27edub:zd2yqd5raa`
- Glama platform release label: `0.1.0` (assigned by Glama)
- Embedded server version: `0.3.0`
- MCP protocol observed: `2025-11-25`
- Runtime evidence: structured JSON startup logs and successful enumeration of
  all six advertised tools

The browser inspector sandbox was requested without credentials to avoid
transmitting a personal OSF token. Deployment remained pending during the
validation window, so no tool invocation is represented as successful usage.
The six tools remain available in Glama's generated tool index.

The listing cross-references FYI MCP, Healthpoint MCP, and SourceRight through
Glama related-server metadata and the README. The README also links the Glama
OSF connector by pipeworx-io and the Paperclip OSF ecosystem project.
