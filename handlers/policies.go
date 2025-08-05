package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllNatPolicy = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewNatPolicy(client)
})

var GetAllNatRule = CreatePaginatedGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewNatRule(client)
})

var GetAllRoutePolicy = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewRoutePolicy(client)
})

var GetAllTrafficPolicy = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewTrafficPolicy(client)
})

var GetAllTrafficPolicyRule = CreatePaginatedGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewTrafficPolicyRule(client)
})

var GetAllPolicyRuleList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewPolicyRuleList(client)
})

var GetAllPolicyPrefixList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewPolicyPrefixList(client)
})

var GetAllPolicyFqdnList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewPolicyFqdnList(client)
})
