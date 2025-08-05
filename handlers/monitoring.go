package handlers

import (
	"github.com/alkiranet/alkira-client-go/alkira"
)

var GetAlerts = CreateHandlerWithOptionalParams(func(client *alkira.AlkiraClient, params map[string]string) (string, error) {
	return client.GetAlerts(params["status"], params["type"], params["priority"])
}, "status", "type", "priority")

var GetAuditLogs = CreateHandlerWithOptionalParams(func(client *alkira.AlkiraClient, params map[string]string) (string, error) {
	return client.GetAuditLogs(params["status"], params["type"])
}, "status", "type")

var GetJobs = CreateHandlerWithOptionalParams(func(client *alkira.AlkiraClient, params map[string]string) (string, error) {
	return client.GetJobs(params["status"], params["type"])
}, "status", "type")

var GetAllHealth = CreateSimpleHandler(func(client *alkira.AlkiraClient) (string, error) {
	return client.GetHealthAll()
})

var GetHealthOfConnector = CreateHandlerWithRequiredParams(func(client *alkira.AlkiraClient, params ...string) (string, error) {
	return client.GetHealthOfConnector(params[0])
}, "id")

var GetHealthOfConnectorInstance = CreateHandlerWithRequiredParams(func(client *alkira.AlkiraClient, params ...string) (string, error) {
	return client.GetHealthOfConnectorInstance(params[0], params[1])
}, "id", "instanceId")

var GetHealthOfService = CreateHandlerWithRequiredParams(func(client *alkira.AlkiraClient, params ...string) (string, error) {
	return client.GetHealthOfService(params[0])
}, "id")

var GetHealthOfServiceInstance = CreateHandlerWithRequiredParams(func(client *alkira.AlkiraClient, params ...string) (string, error) {
	return client.GetHealthOfServiceInstance(params[0], params[1])
}, "id", "instanceId")
