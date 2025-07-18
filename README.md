Alkira MCP Server
---

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for Alkira.

This server provides access to Alkira system by your choice of your AI
agent. It's based on the open source MCP framework
[mcp-go](https://github.com/mark3labs/mcp-go).

[!CAUTION]
NOTE: This project is only for experimental right now and not ready
for production use.


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

## CLAUDE CODE

This is pure command line tool, so the configuration is all done
through configuration json files. Currently, the CLI is still in its
early stage, there are lots of bugs or inconsistent behaviors.

There are multiple ways to integrate MCP:

* Using command:

```
$ claude mcp add mcp-alkira
```

which still needs manual tweak of JSON files to me.

* Directly editing configuration file, you will need to add this block
  to `mcpServers` block in `~/.claude.json`:

```
{
	"command": "PATH-TO-YOUR-MCP-REPO/bin/mcp-alkira",
	"args": [],
	"env": {
		"AK_PORTAL": "xxx",
		"AK_USERNAME": "xxx",
		"AK_PASSWORD": "xxx"
	}
}
```

E.g. mine looks like something like this:

```
"mcpServers": {
  "mcp-alkira": {
    "command": "/Users/spwang/Workspace/alkira/ak-infra/mcp-alkira/bin/mcp-alkira",
    "args": [],
    "env": {}
},
```

Once you launch `claude`, check MCP status to make sure that it's up:

```
$ claude

╰─$ claude                                                                                                                                                        127 ╭|────────────────────────────────────────────────────────────╮
│ ✻ Welcome to Claude Code!                                  │
│                                                            │
│   /help for help, /status for your current setup           │
│                                                            │
│   cwd: /Users/spwang/Workspace/alkira/ak-infra/mcp-alkira  │
╰────────────────────────────────────────────────────────────╯

 ※ Tip: Want Claude to remember something? Hit # to add preferences, tools, and instructions to Claude's memory
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Manage MCP servers                                                                                                                                                   │
│                                                                                                                                                                      │
│ ❯ 1. mcp-alkira  ✔ connected · Enter to view details                                                                                                                 │
│   2. mcp-curl    ✔ connected · Enter to view details                                                                                                                 │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
   Esc to exit

```

That's it.


AVAILABLE TOOLS
---

The MCP server provides the following tools for interacting with Alkira:

* `getAllSegments` - Get all segments
* `getAllGroups` - Get all groups
* `getAllBillingTags` - Get all billing tags
* `getAllLists` - Get all lists
* `getAllConnectors` - Get all connectors
