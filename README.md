Alkira MCP Server
---

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for Alkira.

This server provides access to Alkira system by your choice of your AI
agent. It's based on the open source MCP framework
[mcp-go](https://github.com/mark3labs/mcp-go).

> [!CAUTION]
> This project is only for experimental right now and not
> ready for production use.


## BUILD

You will need Golang to build and you could easily build `mcp-alkira` by:

```
$ make build
```

The following targets are also supported:

* `make fmt`        - Do `gofmt`
* `make vendor`     - Shortcut command to do `go mod tidy` and `go mod vendor`
* `make superclean` - Remove all files not part of the repo (including new files)


HOW TO USE
---

The MCP server needs a tenant Portal URL and API KEY to operate. Fore
each, you could configure one MCP server for it.

The API Key for user could be generated from the portal by following
`Settings` -> `User Management`. There should be a KEY button on each
user.

The detailed configuration for some common tools are listed below:

> [!NOTE]
> In this early stage, all AI agent and tools are changing quickly,
> please refer to each one's documentation for the up-to-date
> instruction on how to configure MCP server.

### Claude Desktop

```json
{
  "mcpServers": {
    "mcp-alkira-test": {
      "command": "mcp-alkira",
      "args": [
        "--portal",
        "YOUR PORTAL URL",
        "--key",
        "YOUR KEY"
      ],
      "env": {}
    }
  }
}
```

### Claude Code

This is CLI, so the configuration is all done through configuration
JSON files. Currently, the CLI is still in its early stage, there are
lots of bugs or inconsistent behaviors.

There are multiple ways to integrate MCP server:

* Using command:

```
$ claude mcp add mcp-alkira <path_to_mcp-alkira-bin>
```

which still needs manual tweak of JSON files after the block is added
to `~/.claude.json`.

* Directly editing configuration file, you will need to add this block
  to `mcpServers` block in `~/.claude.json`:

```json
{
	"command": "PATH-TO-YOUR-MCP-REPO/bin/mcp-alkira",
    "args": [
      "--portal",
      "YOUR PORTAL URL",
      "--key",
      "YOUR API KEY"
    ],
	"env": {}
}
```

E.g. my configuration block on my Mac is like this:

```json
"mcpServers": {
  "mcp-alkira": {
    "command": "/Users/test/mcp/mcp-alkira",
    "args": [
      "--portal",
      "YOUR PORTAL URL",
      "--key",
      "OUR API KEY"
    ],
    "env": {}
},
```

Once you launch `claude`, check MCP status to make sure that it's up:

```bash
╰─$ claude

│ Manage MCP servers                                         │
│                                                            │
│ ❯ 1. mcp-alkira  ✔ connected · Enter to view details       │
│   2. mcp-k8s     ✔ connected · Enter to view details       │
╰────────────────────────────────────────────────────────────╯
   Esc to exit

```

That's it.


AVAILABLE TOOLS
---

> [!NOTE]
> The easiest way to get a list of possible tools is simply asking AI
> agent. It should always give you a list of all available tools.

In AI fashion, just ask the agent what's all available tools from
mcp-alkira.
