package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllInternetApplication = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewInternetApplication(client)
})
