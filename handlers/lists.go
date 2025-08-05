package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllListAsPath = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewListAsPath(client)
})

var GetAllListCommunity = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewListCommunity(client)
})

var GetAllListExtendedCommunity = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewListExtendedCommunity(client)
})

var GetAllDnsServerList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewDnsServerList(client)
})

var GetAllGlobalCidrList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewGlobalCidrList(client)
})

var GetAllUdrList = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewUdrList(client)
})

var GetAllPolicyPrefixListIndividual = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewPolicyPrefixList(client)
})

var GetAllPolicyFqdnListIndividual = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewPolicyFqdnList(client)
})
