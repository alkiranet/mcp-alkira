package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllSegments = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewSegment(client)
})

var GetAllGroups = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewGroup(client)
})

var GetAllBillingTags = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewBillingTag(client)
})

var GetAllCxps = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewInventoryCXP(client)
})
