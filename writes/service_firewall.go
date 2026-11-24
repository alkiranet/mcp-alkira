package writes

import (
	"encoding/json"
	"strings"

	ak "github.com/alkiranet/client-go/tenant"
)

// serviceInstancesNote is the instances update rule the four firewall service families share.
const serviceInstancesNote = "A patch on instances must list every stored instance by its id; an instance without id is added. " +
	"remove.instances drops a stored instance."

// serviceDeleteNote is the delete fact the four firewall service families share.
const serviceDeleteNote = "A service used by a policy rule is refused as in use. " +
	"The credentials the service references are deleted within seconds of the backend removing it, so a new service needs new credentials. " +
	"The instances of a deployed service that are not deployed yet are removed at once, with their credentials; its deployed instances keep their credentials until the deletion is provisioned from the Alkira portal."

// serviceCredentials names, for the delete preview, the credentials a stored
// firewall service references; the backend deletes them with the service. A
// deployed service is marked for deletion and removed by the next provision
// (TPS ServiceLogicImpl.java:1128-1130, :1185-1202).
func serviceCredentials(_ string, raw []byte) string {
	var s struct {
		CredentialID             string `json:"credentialId"`
		RegistrationCredentialID string `json:"registrationCredentialId"`
		MasterKeyCredentialID    string `json:"masterKeyCredentialId"`
		ManagementServer         struct {
			CredentialID string `json:"credentialId"`
		} `json:"managementServer"`
		Instances []struct {
			CredentialID string `json:"credentialId"`
		} `json:"instances"`
	}
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	var ids []string
	for _, id := range []string{s.CredentialID, s.RegistrationCredentialID, s.MasterKeyCredentialID, s.ManagementServer.CredentialID} {
		if id != "" {
			ids = append(ids, id)
		}
	}
	for _, in := range s.Instances {
		if in.CredentialID != "" {
			ids = append(ids, in.CredentialID)
		}
	}
	if len(ids) == 0 {
		return ""
	}
	list := "the credentials it references (" + strings.Join(ids, ", ") + ")"
	if deployed, err := ak.Deployed(ak.ServicePanDescriptor(), raw); err == nil && deployed {
		return list + ", once a provision removes it; instances not yet deployed lose theirs at once"
	}
	return list
}
