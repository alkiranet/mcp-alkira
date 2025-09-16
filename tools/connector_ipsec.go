package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorIPSecGetAll() mcp.Tool {
	return mcp.NewTool("connector_ipsec_get_all",
		mcp.WithDescription("Get all IPSec connectors. By default, 50 " +
			"connectors will be returned once. limit will indicate the number " +
			"of resouces that will returned once. Offset could be used to get " +
			"the next batch of the resource. Hits will tell the total number of" +
			"the resources."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}

func ConnectorIPSecGetById() mcp.Tool {
	return mcp.NewTool("connector_ipsec_get_by_id",
		mcp.WithDescription("Get details of one IPSec connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorIPSecGetByName() mcp.Tool {
	return mcp.NewTool("connector_ipsec_get_by_name",
		mcp.WithDescription("Get details of one IPSec connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorIPSecGetTotal() mcp.Tool {
	return mcp.NewTool("connector_ipsec_get_total",
		mcp.WithDescription("Get total numbers of IPSec connectors."),
	)
}

func ConnectorIPSecAdvGetAll() mcp.Tool {
	return mcp.NewTool("connector_ipsec_adv_get_all",
		mcp.WithDescription("Get all Advanced IPSec connectors."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}

func ConnectorIPSecAdvGetById() mcp.Tool {
	return mcp.NewTool("connector_ipsec_adv_get_by_id",
		mcp.WithDescription("Get details of one Advanced IPSec connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorIPSecAdvGetByName() mcp.Tool {
	return mcp.NewTool("connector_ipsec_adv_get_by_name",
		mcp.WithDescription("Get details of one Advanced IPSec connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorIPSecAdvGetTotal() mcp.Tool {
	return mcp.NewTool("connector_ipsec_adv_get_total",
		mcp.WithDescription("Get total numbers of Advanced IPSec connectors."),
	)
}

func ConnectorIPSecTunnelProfileGetAll() mcp.Tool {
	return mcp.NewTool("connector_ipsec_tunnel_profile_get_all",
		mcp.WithDescription("Get all IPSec Tunnel Profiles"),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}

func ConnectorIPSecTunnelProfileGetTotal() mcp.Tool {
	return mcp.NewTool("connector_ipsec_tunnel_profile_get_total",
		mcp.WithDescription("Get total numbers of IPSec Tunnel Profiles."),
	)
}
