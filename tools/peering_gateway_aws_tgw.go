package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func PeeringGatewayAwsTgwGetAll() mcp.Tool {
	return mcp.NewTool("peering_gateway_aws_tgw_get_all",
		mcp.WithDescription("Get all AWS Transit Gateway peering gateways of the tenant network, with their CXP, AWS region, ASN and state."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func PeeringGatewayAwsTgwGetById() mcp.Tool {
	return mcp.NewTool("peering_gateway_aws_tgw_get_by_id",
		mcp.WithDescription("Get details of one AWS Transit Gateway peering gateway by its ID."),
		mcp.WithString("peeringGatewayId",
			mcp.Required(),
			mcp.Description("Peering gateway ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func PeeringGatewayAwsTgwAttachmentGetAll() mcp.Tool {
	return mcp.NewTool("peering_gateway_aws_tgw_attachment_get_all",
		mcp.WithDescription("Get all AWS Transit Gateway peering attachments of the tenant network. An AWS TGW connector uses one ACTIVE attachment, by its id (awsTgwPeeringAttachmentId); awsTgwId is the id of the attachment's peering gateway."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func PeeringGatewayAwsTgwAttachmentGetById() mcp.Tool {
	return mcp.NewTool("peering_gateway_aws_tgw_attachment_get_by_id",
		mcp.WithDescription("Get details of one AWS Transit Gateway peering attachment by its ID."),
		mcp.WithString("attachmentId",
			mcp.Required(),
			mcp.Description("Peering attachment ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
