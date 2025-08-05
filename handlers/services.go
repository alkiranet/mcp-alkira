package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAllServiceCheckpoint = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceCheckpoint(client)
})

var GetAllServiceCiscoFTDv = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceCiscoFTDv(client)
})

var GetAllServiceF5Lb = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceF5Lb(client)
})

var GetAllServiceFortinet = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceFortinet(client)
})

var GetAllServiceInfoblox = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceInfoblox(client)
})

var GetAllServicePan = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServicePan(client)
})

var GetAllServiceZscaler = CreateGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
	return alkira.NewServiceZscaler(client)
})
