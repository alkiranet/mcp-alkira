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

This is pure command line tool, so the configuration is all done
through configuration json files. Currently, the CLI is still in its
early stage, there are lots of bugs or inconsistent behaviors.

There are multiple ways to integrate MCP:

* Using command:

```
$ claude mcp add mcp-alkira <path_to_mcp-alkira-bin>
```

which still needs manual tweak of JSON files to me.

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

The MCP server provides the following tools for interacting with Alkira:

| Category | Tool Name | Description |
|----------|-----------|-------------|
| **Core** | `getAllSegments` | Get all segments |
| **Core** | `getAllGroups` | Get all groups |
| **Core** | `getAllBillingTags` | Get all billing tags |
| **Services** | `getAllServiceCheckpoint` | Get all Check Point firewall services |
| **Services** | `getAllServiceCiscoFTDv` | Get all Cisco Firepower Threat Defense virtual services |
| **Services** | `getAllServiceF5Lb` | Get all F5 Load Balancer services |
| **Services** | `getAllServiceFortinet` | Get all Fortinet firewall services |
| **Services** | `getAllServiceInfoblox` | Get all Infoblox DNS/DHCP services |
| **Services** | `getAllServicePan` | Get all Palo Alto Networks firewall services |
| **Services** | `getAllServiceZscaler` | Get all Zscaler security services |
| **Connectors** | `getAllConnectorArubaEdge` | Get all Aruba Edge Connect SD-WAN connectors |
| **Connectors** | `getAllConnectorAwsDirectConnect` | Get all AWS Direct Connect connectors |
| **Connectors** | `getAllConnectorAwsTgw` | Get all AWS Transit Gateway connectors |
| **Connectors** | `getAllConnectorAwsVpc` | Get all AWS VPC connectors |
| **Connectors** | `getAllConnectorAzureExpressRoute` | Get all Azure ExpressRoute connectors |
| **Connectors** | `getAllConnectorAzureVnet` | Get all Azure Virtual Network connectors |
| **Connectors** | `getAllConnectorCiscoSdwan` | Get all Cisco SD-WAN connectors |
| **Connectors** | `getAllConnectorFortinetSdwan` | Get all Fortinet SD-WAN connectors |
| **Connectors** | `getAllConnectorGcpInterconnect` | Get all Google Cloud Interconnect connectors |
| **Connectors** | `getAllConnectorGcpVpc` | Get all Google Cloud VPC connectors |
| **Connectors** | `getAllConnectorInternet` | Get all Internet Exit connectors |
| **Connectors** | `getAllConnectorIPSec` | Get all IPSec connectors |
| **Connectors** | `getAllConnectorAdvIPSec` | Get all Advanced IPSec connectors |
| **Connectors** | `getAllConnectorIPSecTunnelProfile` | Get all IPSec Tunnel Profile connectors |
| **Connectors** | `getAllConnectorOciVcn` | Get all Oracle Cloud Infrastructure VCN connectors |
| **Connectors** | `getAllConnectorRemoteAccessTemplate` | Get all Remote Access Template connectors |
| **Connectors** | `getAllConnectorVersaSdwan` | Get all Versa SD-WAN connectors |
| **Connectors** | `getAllConnectorVmwareSdwan` | Get all VMware SD-WAN (VeloCloud) connectors |
| **Routes** | `getRoutes` | Get routes for a tenant network with optional filtering (by segment, connector, route type, etc.) |
| **Routes** | `getRouteCount` | Get route counts for a tenant network with optional filtering |
| **Routes** | `getRouteSummary` | Get aggregated route analytics grouped by connector type, segment, or CXP |
| **Routes** | `getAllRoutes` | Get ALL routes with automatic pagination (handles large result sets efficiently) |
| **Policies** | `getAllNatPolicy` | Get all NAT policies |
| **Policies** | `getAllNatRule` | Get all NAT policy rules |
| **Policies** | `getAllRoutePolicy` | Get all route policies |
| **Policies** | `getAllTrafficPolicy` | Get all traffic policies |
| **Policies** | `getAllTrafficPolicyRule` | Get all traffic policy rules |
| **Policies** | `getAllPolicyRuleList` | Get all policy rule lists |
| **Policies** | `getAllPolicyPrefixList` | Get all policy prefix lists |
| **Policies** | `getAllPolicyFqdnList` | Get all policy FQDN lists |
| **Lists** | `getAllListAsPath` | Get all AS Path lists for BGP routing |
| **Lists** | `getAllListCommunity` | Get all BGP Community lists |
| **Lists** | `getAllListExtendedCommunity` | Get all BGP Extended Community lists |
| **Lists** | `getAllDnsServerList` | Get all DNS Server lists |
| **Lists** | `getAllGlobalCidrList` | Get all Global CIDR lists |
| **Lists** | `getAllUdrList` | Get all User Defined Route (UDR) lists |
| **Lists** | `getAllPolicyPrefixListIndividual` | Get all Policy Prefix lists individually |
| **Lists** | `getAllPolicyFqdnListIndividual` | Get all Policy FQDN lists individually |
| **Internet Applications** | `getAllInternetApplication` | Get all Internet Applications |
| **Monitoring** | `getAlerts` | Get all alerts with optional filters (status, type, priority) |
| **Monitoring** | `getAuditLogs` | Get all audit logs with optional filters (status, type) |
| **Monitoring** | `getJobs` | Get all jobs with optional filters (status, type) |
| **Monitoring** | `getAllHealth` | Get health status of all resources |
| **Monitoring** | `getHealthOfConnector` | Get health status of a connector by ID |
| **Monitoring** | `getHealthOfConnectorInstance` | Get health status of a connector instance by ID |
| **Monitoring** | `getHealthOfService` | Get health status of a service by ID |
| **Monitoring** | `getHealthOfServiceInstance` | Get health status of a service instance by ID |
| **Segment Resources** | `getAllSegmentResources` | Get all segment resources |
| **Segment Resources** | `getAllSegmentResourceShares` | Get all segment resource shares |


ROUTES TOOLS
---

The Routes tools provide comprehensive access to Alkira's routing information with advanced filtering and analytics capabilities.

### Core Route Tools

**`getRoutes`** - Retrieve detailed route information with flexible filtering
- Get received routes (learned by Alkira) or advertised routes (sent to connectors)
- Filter by segments, connectors, route types, or search terms
- Support for multiple output formats (JSON, table, CSV, summary)
- Optimized pagination with smart defaults

**`getRouteCount`** - Get route counts matching criteria
- More efficient than getRoutes for count-only queries
- Useful for checking route table sizes before pagination
- Supports all the same filtering options as getRoutes

**`getRouteSummary`** - Aggregated route analytics and insights
- Group routes by connector type, segment, or CXP (Cloud Exchange Point)
- Route distribution analysis and connector utilization metrics
- Optional detailed breakdowns and IP prefix analysis

**`getAllRoutes`** - Complete route retrieval with automatic pagination
- Handles large result sets efficiently with safety limits
- Automatic pagination with optimal batch sizes
- Built-in filtering for segments, connector types, and CXPs

### Common Usage Examples

```bash
# Get received routes for a specific segment
getRoutes tenantNetworkId="48" type="received" segmentName="Corporate"

# Find all AWS VPC routes
getRoutes tenantNetworkId="48" type="received" connectorTypes="AWS_VPC" limit=100

# Search for routes containing specific terms
getRoutes tenantNetworkId="48" type="received" search="customer" 

# Get route count by connector type
getRouteSummary tenantNetworkId="48" type="received" groupBy="connectorType"

# Find active routes (non-suppressed)
getRoutes tenantNetworkId="48" type="received" routeStatus="active"

# Export routes to CSV format
getRoutes tenantNetworkId="48" type="received" outputFormat="csv" limit=1000
```

### Key Filtering Options

- **`type`**: `received` (learned by Alkira) | `advertised` (sent to connectors) | `overlap` (conflicting routes)
- **`segmentName`**: Filter by network segment (e.g., 'Corporate', 'DMZ')  
- **`connectorTypes`**: AWS_VPC, AZURE_VNET, GCP_VPC, IP_SEC, etc.
- **`routeStatus`**: `active`, `suppressed`, `overlap`
- **`search`**: Search across all fields for specific terms
- **`prefixType`**: `LOCAL` (segment-only) | `SHARED` (cross-segment)
- **`outputFormat`**: `json`, `table`, `csv`, `summary`

For advertised routes, specify either `segmentName` or `cxp` parameter for optimal performance.
