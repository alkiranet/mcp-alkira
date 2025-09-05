package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetHealths() mcp.Tool {
	return mcp.NewTool("getHealths",
		mcp.WithDescription("Get health status of all resources."),
	)
}

func GetConnectorHealthById() mcp.Tool {
	return mcp.NewTool("getConnectorHealthById",
		mcp.WithDescription("Get health status of a connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetConnectorInstanceHealthById() mcp.Tool {
	return mcp.NewTool("getConnectorInstanceHealthById",
		mcp.WithDescription("Get health status of a connector instance by its ID"),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
		mcp.WithString("instanceId",
			mcp.Required(),
			mcp.Description("Instance ID."),
		),
	)
}

func GetServiceHealthById() mcp.Tool {
	return mcp.NewTool("getServiceHealthById",
		mcp.WithDescription("Get health status of a service by ID"),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetServiceInstanceHealthById() mcp.Tool {
	return mcp.NewTool("getServiceInstanceHealthById",
		mcp.WithDescription("Get health status of a service instance by ID"),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
		mcp.WithString("instanceId",
			mcp.Required(),
			mcp.Description("Instance ID."),
		),
	)
}
