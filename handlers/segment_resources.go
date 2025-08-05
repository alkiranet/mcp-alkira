package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllSegmentResources = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewSegmentResource(client)
})

var GetAllSegmentResourceShares = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewSegmentResourceShare(client)
})
