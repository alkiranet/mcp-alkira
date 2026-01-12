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

### Tool Categories

The MCP server provides **~120 tools** organized into the following categories:

#### Tenant & Network
| Tool | Description |
|------|-------------|
| `TenantNetworkSummary` | Get tenant network overview |
| `TenantResourceUsages` | Current resource usage |
| `TenantResourceLimits` | Resource limits |

#### Billing & CXPs
| Tool | Description |
|------|-------------|
| `BillingTagGetAll` | List all billing tags |
| `BillingTagGetById` | Get billing tag by ID |
| `BillingTagGetByName` | Get billing tag by name |
| `BillingTagGetTotal` | Get total billing tag count |
| `CxpGetAll` | List all CXPs |

#### Segments & Groups
| Tool | Description |
|------|-------------|
| `SegmentGetAll` | List all segments |
| `SegmentGetById` | Get segment by ID |
| `SegmentGetByName` | Get segment by name |
| `GroupGetById` | Get group by ID |
| `GroupGetByName` | Get group by name |
| `GroupGetTotal` | Get total group count |
| `SegmentResourceGetAll/ById/ByName/Total` | Segment resource operations |
| `SegmentResourceShareGetAll/ById/ByName/Total` | Resource share operations |

#### Connectors

Each connector type supports: `GetAll`, `GetById`, `GetByName`, `GetTotal`

| Category | Connector Types |
|----------|-----------------|
| **Cloud** | AWS VPC, Azure VNet, GCP VPC, OCI VCN |
| **WAN** | AWS Direct Connect, Azure ExpressRoute, GCP Interconnect |
| **SD-WAN** | Cisco, Fortinet, Aruba Edge, Versa, VMware (VeloCloud) |
| **Branch** | IPSec, IPSec Advanced, Remote Access |
| **Transit** | AWS TGW |
| **Other** | Internet |

#### Services

Each service type supports: `GetAll`, `GetById`, `GetByName`, `GetTotal`

| Service | Description |
|---------|-------------|
| PAN | Palo Alto Networks firewall |
| Fortinet | FortiGate firewall |
| Check Point | Check Point firewall |
| Cisco FTDv | Cisco Firepower Threat Defense |
| Zscaler | Zscaler Internet Access |
| F5 LB | F5 Load Balancer |
| Infoblox | DNS/DHCP/IPAM |

#### Routes
| Tool | Description |
|------|-------------|
| `GetRoutes` | Query routes with filters |
| `GetRouteCount` | Count routes matching criteria |
| `GetRouteSummary` | Route summary by segment |
| `GetAllRoutes` | Get all routes (use with caution) |

#### Policies

**NAT Policies:**
- `PolicyNatGetAll/ById/ByName/Total` - NAT policy operations
- `PolicyNatRuleGetAll/ById/ByName/Total` - NAT rule operations

**Route Policies:**
- `PolicyRouteGetAll/ById/ByName/Total` - Route policy operations

**Traffic Policies:**
- `PolicyTrafficGetAll/ById/ByName` - Traffic policy operations
- `PolicyTrafficRuleGetAll/ById/ByName` - Traffic rule operations
- `PolicyTrafficRuleListGetAll/ById/ByName` - Traffic rule list operations

#### Lists
| List Type | Tools |
|-----------|-------|
| AS Path | `ListAsPathGetAll/ById/ByName/Total` |
| BGP Community | `ListCommunityGetAll/ById/ByName/Total` |
| Extended Community | `ListExtendedCommunityGetAll/ById/ByName/Total` |
| DNS Server | `ListDnsServerGetAll/ById/ByName/Total` |
| Global CIDR | `ListGlobalCidrGetAll/ById/ByName/Total` |
| UDR | `ListUdrGetAll/ById/ByName/Total` |
| Prefix | `ListPrefixGetById/ByName/ByPrefix` |
| Policy FQDN | `ListPolicyFqdnGetAll/ById/ByName/Total` |

#### Internet Applications
| Tool | Description |
|------|-------------|
| `InternetApplicationGetAll` | List all internet applications |
| `InternetApplicationGetById` | Get by ID |
| `InternetApplicationGetByName` | Get by name |
| `InternetApplicationGetTotal` | Get total count |

#### Health & Monitoring
| Tool | Description |
|------|-------------|
| `HealthConnectorGetById` | Get connector health status |
| `HealthConnectorInstanceGetById` | Get connector instance health |
| `HealthServiceGetById` | Get service health status |
| `HealthServiceInstanceGetById` | Get service instance health |
| `Alerts` | Get alerts |
| `AuditLogs` | Get audit logs |
| `Jobs` | Get jobs |


COMMAND LINE OPTIONS
---

| Option | Default | Description |
|--------|---------|-------------|
| `--portal` | (required) | Alkira Portal URL |
| `--key` | (required) | API key |
| `--mode` | `stdio` | Server mode: `stdio` or `sse` |
| `--port` | `8081` | Port for SSE mode |
| `--maxToken` | `0` | Max token payload size |


PROMPTS
---

| Prompt | Description |
|--------|-------------|
| `summary` | Generate a quick tenant summary |
