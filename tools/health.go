package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func HealthConnectorGetById() mcp.Tool {
	return mcp.NewTool("health_connector_get_by_id",
		mcp.WithDescription("Get health status of a connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func HealthConnectorInstanceGetById() mcp.Tool {
	return mcp.NewTool("health_connector_instance_get_by_id",
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

func HealthServiceGetById() mcp.Tool {
	return mcp.NewTool("health_service_get_by_id",
		mcp.WithDescription("Get health status of a service by ID"),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func HealthServiceInstanceGetById() mcp.Tool {
	return mcp.NewTool("health_service_instance_get_by_id",
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
