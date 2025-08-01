Alkira MCP Server
---

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for Alkira.

This server provides access to Alkira system by your choice of your AI
agent. It's based on the open source MCP framework
[mcp-go](https://github.com/mark3labs/mcp-go).

> [!CAUTION]
> NOTE: This project is only for experimental right now and not ready
> for production use.


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
      "args": [],
      "env": {
        "PORTAL_URL": "http://localhost:3000",
        "GRAFANA_API_KEY": "<your service account token>"
      }
    }
  }
}
```

### CLAUDE CODE

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

```json
{
	"command": "PATH-TO-YOUR-MCP-REPO/bin/mcp-alkira",
    "args": [
      "--portal",
      "test.portal.alkira.com",
      "--key",
      "xxxxxxxxxxxxxxxxxxxxxx"
    ],
	"env": {}
}
```

E.g. my configuration block on my Mac is like this:

```json
"mcpServers": {
  "mcp-alkira": {
    "command": "/Users/spwang/mcp-alkira/bin/mcp-alkira",
    "args": [
      "--portal",
      "test.portal.alkira.com",
      "--key",
      "xxxxxxxxxxxxxxxxxxxxxx"
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
