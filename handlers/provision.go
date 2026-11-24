package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/mark3labs/mcp-go/mcp"
)

// Two-step selective provisioning (00-PLAN.md D14), following the portal:
// karrawa 5d18e83fd src/views/network/provision/diffV2/index.js:378-387,
// diffV2/utils/dataTransform.js:12-226 and src/_api/topology.js:210-220.

// provisionFamily binds provision codes; a delete code never matches it.
const provisionFamily = "tenant_network_provision"

// Feature flags the portal reads for the provision body.
const (
	bestEffortCleanupFlag = "RELEASE-PROV-BEST-EFFORT-CLEANUP"
	// cleanupStandard is RETRY_METHOD.STANDARD (karrawa _core/constants/index.js:1388-1391).
	cleanupStandard = "REQUIRED"
)

// Attributions of a pending change (dataTransform.js:12-18).
const (
	byCaller    = "caller"
	byOtherUser = "other user"
	bySystem    = "system"
	byUntracked = "untracked"
)

// Change categories, in the portal's order.
const (
	catConnectorService = iota
	catPolicy
	catCSX
)

// policyTypes are the TPS change types the portal shows as policies.
var policyTypes = []string{"ROUTE_POLICY", "INTER_CXP_ROUTE_POLICY", "NAT_POLICY"}

// csxTypes are the keys of GLOBAL_DIFF_OBJ (karrawa _core/constants/index.js:2136-2167).
var csxTypes = []string{"GROUP", "SEGMENT", "PREFIX_LIST", "RULE_LIST", "COMMUNITY_LIST", "EXTENDED_COMMUNITY_LIST",
	"SEGMENT_RESOURCE", "INTERNET_APPLICATION", "SEGMENT_RESOURCE_SHARE", "RULE", "BYOIP", "NAT_RULE", "CXP", "AS_PATH_LIST",
	"GLOBAL_CIDR_LIST", "CONNECTOR_TEMPLATE", "IPSEC_TUNNEL_PROFILE", "FLOW_COLLECTOR", "INTER_CONNECTOR_COMMUNICATION_GROUP",
	"POLICY_FQDN_LIST", "DNS_SERVER_LIST", "UDR_LIST", "F5_VSERVER_ENDPOINT", "CXP_INFRA_CHANGE", "FIREWALL_ZONE", "SCALE_OPTION",
	"PEERING_GATEWAY", "PEERING_GATEWAY_ATTACHMENT", "TRANSIT_GATEWAY_SEGMENT_GROUP", "AZURE_VIRTUAL_NETWORK_MANAGER"}

// typeLabels is TYPE_LABEL_MAP (karrawa _core/constants/index.js:980-1023).
// The portal sends the label as entity.subType of a connector or service;
// TPS matches a selected change by change type, entity type and id only
// (NetworkChangesCurator.java getCuratedChanges), so the label is carried as sent.
var typeLabels = map[string]string{
	"SD_WAN": "SD-WAN", "ARUBA_EDGE_CONNECT": "Aruba EdgeConnect", "IP_SEC": "IPsec", "ADV_IP_SEC": "Advanced IPsec",
	"AWS_VPC": "AWS VPC", "AZURE_VNET": "Azure VNet", "AZURE_VNET_THIRD_PARTY": "Azure VNet Third Party Connector",
	"AZURE_VHUB": "Azure VHub", "GCP_VPC": "GCP VPC", "DIRECT_CONNECT": "AWS Direct Connect", "PAN": "PAN Firewall",
	"SAAS": "SaaS", "REMOTE_ACCESS": "Alkira Remote Access", "INT": "Internet Exit", "INB_INT": "Inbound Internet",
	"INTERNET_APPLICATION": "Internet-facing Application", "SEGMENT_RESOURCE_SHARE": "Segment Resource Share",
	"CHKPFW": "Check Point Firewall", "ZIA": "Zscaler ZIA", "OCI_VCN": "OCI VCN", "EXPRESS_ROUTE": "Azure ExpressRoute",
	"FTNTFW": "Fortinet Firewall", "AKAMAI_PROLEXIC": "Akamai Prolexic Connector", "INFOBLOX": "Infoblox Service",
	"CONNECTOR_TEMPLATE": "Remote Access Configuration", "CISCO_FTDV_FW": "Cisco Secure Firewall FTDv",
	"VMWARE_SDWAN": "VMWare Connector", "FORTINET_SDWAN": "Fortinet Secure SD-WAN Connector",
	"PRISMA_SDWAN": "Prisma SD-WAN Connector", "VERSA_SDWAN": "Versa SD-WAN Connector",
	"JUNIPER_SDWAN_SSR": "Juniper SD-WAN Connector", "AWS_TGW": "AWS TGW Connector", "MERAKI_SDWAN": "Meraki SD-WAN Connector", "F5LB": "F5 Load Balancer",
	"SERVICENOW": "SERVICENOW", "SPLUNK": "SPLUNK", "GENERIC_COLLECTOR": "GENERIC", "FORTIMONITOR": "FORTIMONITOR",
	"EMAIL_INTEGRATION": "EMAIL", "DYNATRACE": "DYNATRACE", "XMATTERS": "XMATTERS", "GCP_INTERCONNECT": "GCP Interconnect",
	"BLUECAT": "BlueCat Service", "MULTI_CLOUD_GATEWAY": "Lumen Multi-Cloud Gateway",
}

// refFields maps a property name in a change's entity properties to the
// change type its numeric ids reference. The names are the reference fields
// of the TPS and policy-service entities as /changes returns them (live,
// tenant 239: connectors and services carry segmentIds, policies fromGroups,
// toGroups and ruleListId, segment resources groupId and groupPrefixes).
// The backend check at provision stays authoritative
// (NetworkEntityValidatorRegistry.java validateEntityReferences).
var refFields = map[string]string{
	"segmentId": "SEGMENT", "segmentIds": "SEGMENT",
	"groupId": "GROUP", "fromGroups": "GROUP", "toGroups": "GROUP", "includedGroups": "GROUP", "excludedGroups": "GROUP",
	"ruleListId": "RULE_LIST", "ruleId": "RULE", "natRuleIds": "NAT_RULE",
	"prefixListId": "PREFIX_LIST", "prefixListIds": "PREFIX_LIST", "srcPrefixListId": "PREFIX_LIST", "dstPrefixListId": "PREFIX_LIST",
	"sourceRoutesPrefixListId": "PREFIX_LIST", "advertiseCustomRoutesPrefixId": "PREFIX_LIST",
	"asPathListIds": "AS_PATH_LIST", "communityListIds": "COMMUNITY_LIST", "extendedCommunityListIds": "EXTENDED_COMMUNITY_LIST",
	"connectorId": "CONNECTOR", "serviceId": "SERVICE", "peeringGatewayId": "PEERING_GATEWAY",
}

// refNameFields reference a segment by name (route policies, NAT policies,
// segment resources, lists; connectors and services also list segments).
var refNameFields = map[string]string{"segment": "SEGMENT", "segments": "SEGMENT"}

// Validation states that block a change from provisioning
// (NetworkEntityValidatorRegistry.java:86-102).
var blockingValidation = []string{"PENDING", "IN_PROGRESS", "FAILED"}

// pendingChange is one entry of the merged pending list.
type pendingChange struct {
	Key         string
	Category    int
	SubType     string // connector or service label
	Change      ak.NetworkChange
	Props       map[string]any
	Implicit    bool
	Attribution string
	Selected    bool
}

func changeKey(c ak.NetworkChange) string { return c.Entity.Type + ":" + c.Entity.ID }

// mergeChanges merges the TPS and policy-service change lists as
// transformDiffData does: connectors and services, then policies, then the
// GLOBAL_DIFF_OBJ types, implicit groups left out. Other types are not shown
// by the portal and are not selectable. Within a category the order is by
// type and id, so the list and the body are stable across reads.
func mergeChanges(tps, pol []ak.NetworkChange) []*pendingChange {
	var out []*pendingChange
	seen := map[string]bool{}
	add := func(c ak.NetworkChange, cat int) {
		k := changeKey(c)
		if seen[k] {
			return
		}
		seen[k] = true
		props := map[string]any{}
		if len(c.Entity.Properties) > 0 {
			if v, err := decodeNumbers(c.Entity.Properties); err == nil {
				if m, ok := v.(map[string]any); ok {
					props = m
				}
			}
		}
		pc := &pendingChange{Key: k, Category: cat, Change: c, Props: props,
			Implicit: c.Entity.ImplicitReference || c.ImplicitReference}
		if cat == catConnectorService {
			t, _ := props["type"].(string)
			pc.SubType = t
			if l, ok := typeLabels[t]; ok {
				pc.SubType = l
			}
		}
		out = append(out, pc)
	}
	for _, c := range tps {
		if c.Entity.Type == "CONNECTOR" || c.Entity.Type == "SERVICE" {
			add(c, catConnectorService)
		}
	}
	for _, c := range pol {
		if c.Entity.Type == "POLICY" {
			add(c, catPolicy)
		}
	}
	for _, c := range tps {
		if slices.Contains(policyTypes, c.Entity.Type) {
			add(c, catPolicy)
		}
	}
	for _, list := range [][]ak.NetworkChange{tps, pol} {
		for _, c := range list {
			if !slices.Contains(csxTypes, c.Entity.Type) {
				continue
			}
			if c.Entity.Type == "GROUP" && implicitGroup(c) {
				continue
			}
			add(c, catCSX)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Change.Entity.Type != b.Change.Entity.Type {
			return a.Change.Entity.Type < b.Change.Entity.Type
		}
		return lessID(a.Change.Entity.ID, b.Change.Entity.ID)
	})
	return out
}

func lessID(a, b string) bool {
	x, ex := strconv.ParseInt(a, 10, 64)
	y, ey := strconv.ParseInt(b, 10, 64)
	if ex == nil && ey == nil {
		return x < y
	}
	return a < b
}

// implicitGroup: a group with properties.implicit, unless it is a USER_GROUP
// (dataTransform.js:51-55).
func implicitGroup(c ak.NetworkChange) bool {
	var p struct {
		Type     string `json:"type"`
		Implicit bool   `json:"implicit"`
	}
	_ = json.Unmarshal(c.Entity.Properties, &p)
	return p.Type != "USER_GROUP" && p.Implicit
}

// attribute sets each change's attribution (dataTransform.js:12-18). TPS
// stamps rows changed by the backend itself with lastUpdatedByUserId
// "system" (live, tenant 239), which is reported as system.
func attribute(changes []*pendingChange, caller string) {
	for _, c := range changes {
		uid := c.Change.LastUpdatedByUserID
		switch {
		case uid == "":
			if sg, _ := c.Props["systemGenerated"].(bool); sg {
				c.Attribution = bySystem
			} else {
				c.Attribution = byUntracked
			}
		case uid == "system":
			c.Attribution = bySystem
		case caller != "" && uid == caller:
			c.Attribution = byCaller
		default:
			c.Attribution = byOtherUser
		}
	}
}

// selection is a parsed select argument.
type selection struct {
	mode  string   // "all", "mine" or "keys"
	items []string // sorted, for mode "keys"
}

// canonical is the form the code is bound to.
func (s selection) canonical() string {
	if s.mode != "keys" {
		return s.mode
	}
	return strings.Join(s.items, ",")
}

func parseSelect(args map[string]any) (selection, *mcp.CallToolResult) {
	switch v := args["select"].(type) {
	case string:
		if v == "all" || v == "mine" {
			return selection{mode: v}, nil
		}
	case []any:
		set := map[string]bool{}
		for _, it := range v {
			s, ok := it.(string)
			s = strings.TrimSpace(s)
			if !ok || s == "" || len(s) > 200 {
				return selection{}, invalidArgs(errors.New(`select items must be non-empty change keys TYPE:ID`))
			}
			if t, id, ok := strings.Cut(s, ":"); ok {
				s = strings.ToUpper(t) + ":" + id
			}
			set[s] = true
		}
		if len(set) > 0 {
			items := make([]string, 0, len(set))
			for s := range set {
				items = append(items, s)
			}
			sort.Strings(items)
			return selection{mode: "keys", items: items}, nil
		}
	}
	return selection{}, invalidArgs(errors.New(`select must be "all", "mine" or a non-empty list of change keys TYPE:ID`))
}

// errNoCaller: "mine" needs the caller's user id.
var errNoCaller = errors.New("the caller's user id could not be read")

// apply marks the selected changes. It returns the items that match no
// pending change and the bare ids that match more than one.
func (s selection) apply(changes []*pendingChange, caller string) (missing, ambiguous []string, err error) {
	switch s.mode {
	case "all":
		for _, c := range changes {
			c.Selected = true
		}
	case "mine":
		if caller == "" {
			return nil, nil, errNoCaller
		}
		for _, c := range changes {
			c.Selected = c.Change.LastUpdatedByUserID == caller
		}
	default:
		for _, it := range s.items {
			var hits []*pendingChange
			for _, c := range changes {
				if c.Key == it || (!strings.Contains(it, ":") && c.Change.Entity.ID == it) {
					hits = append(hits, c)
				}
			}
			switch len(hits) {
			case 0:
				missing = append(missing, it)
			case 1:
				hits[0].Selected = true
			default:
				ambiguous = append(ambiguous, it)
			}
		}
	}
	return missing, ambiguous, nil
}

// body builds the provision or validation body for the selected changes
// (getProvisionPayload, dataTransform.js:182-226): implicit references and
// remote access connectors are left out; connectors and services carry the
// label subType; the remote access template carries subType REMOTE_ACCESS;
// with bestEffort, changes that allow it carry bestEffortAllowed and the
// STANDARD cleanup type. A validation body never carries cleanup options.
func body(changes []*pendingChange, bestEffort bool) []ak.ProvisionChange {
	var out []ak.ProvisionChange
	for _, c := range changes {
		if !c.Selected || c.Implicit {
			continue
		}
		if c.Category == catConnectorService {
			if t, _ := c.Props["type"].(string); t == "REMOTE_ACCESS" {
				continue
			}
		}
		pc := ak.ProvisionChange{ChangeType: c.Change.ChangeType, Entity: ak.ProvisionChangeEntity{ID: c.Change.Entity.ID, Type: c.Change.Entity.Type}}
		if c.Category == catConnectorService {
			pc.Entity.SubType = c.SubType
		}
		if c.Category == catCSX && c.Change.Entity.SubType == "REMOTE_ACCESS" {
			pc.SubType = "REMOTE_ACCESS"
		}
		if bestEffort && c.Change.BestEffortAllowed {
			pc.BestEffortAllowed = true
			pc.CleanupType = cleanupStandard
		}
		out = append(out, pc)
	}
	return out
}

// refs returns the keys of the changes c's properties reference, at any depth.
func refs(c *pendingChange, changes []*pendingChange) map[string]bool {
	out := map[string]bool{}
	byName := map[string]string{}
	for _, o := range changes {
		if o.Change.Entity.Type == "SEGMENT" && o.Change.Entity.Name != "" {
			byName[o.Change.Entity.Name] = o.Key
		}
	}
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, x := range t {
				if typ, ok := refFields[k]; ok {
					for _, id := range numbers(x) {
						out[typ+":"+id] = true
					}
				}
				if _, ok := refNameFields[k]; ok {
					for _, n := range strs(x) {
						if key, ok := byName[n]; ok {
							out[key] = true
						}
					}
				}
				walk(x)
			}
		case []any:
			for _, x := range t {
				walk(x)
			}
		}
	}
	walk(c.Props)
	delete(out, c.Key)
	return out
}

func numbers(v any) []string {
	switch t := v.(type) {
	case json.Number:
		return []string{t.String()}
	case []any:
		var out []string
		for _, x := range t {
			if n, ok := x.(json.Number); ok {
				out = append(out, n.String())
			}
		}
		return out
	}
	return nil
}

func strs(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// dependency is an unselected pending change a selected one needs.
type dependency struct {
	Need *pendingChange
	By   []string
}

// unselectedDependencies lists the pending changes the selection references
// but leaves out: a selected add or modify needs the pending adds and
// modifies it references; a selected delete needs the pending deletes of
// the objects that reference it (validateDeletedReferences).
func unselectedDependencies(changes []*pendingChange) []dependency {
	byKey := map[string]*pendingChange{}
	for _, c := range changes {
		byKey[c.Key] = c
	}
	need := map[string][]string{}
	for _, c := range changes {
		if !c.Selected || c.Implicit {
			continue
		}
		if c.Change.ChangeType != "DELETED" {
			for k := range refs(c, changes) {
				if d := byKey[k]; d != nil && !d.Selected && !d.Implicit && d.Change.ChangeType != "DELETED" {
					need[k] = append(need[k], c.Key)
				}
			}
			continue
		}
		for _, o := range changes {
			if o.Selected || o.Implicit || o.Change.ChangeType != "DELETED" {
				continue
			}
			if refs(o, changes)[c.Key] {
				need[o.Key] = append(need[o.Key], c.Key)
			}
		}
	}
	var out []dependency
	for k, by := range need {
		sort.Strings(by)
		out = append(out, dependency{Need: byKey[k], By: by})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Need.Key < out[j].Need.Key })
	return out
}

// pendingHash hashes the merged pending list; any added, removed or
// re-edited change changes it.
func pendingHash(changes []*pendingChange) [32]byte {
	lines := make([]string, 0, len(changes))
	for _, c := range changes {
		lc, _ := c.Props["lastConfigUpdatedAt"].(json.Number)
		lines = append(lines, strings.Join([]string{c.Key, c.Change.ChangeType, strconv.FormatInt(c.Change.Entity.LastModifiedAt, 10),
			lc.String(), strconv.FormatBool(c.Implicit), strconv.FormatBool(c.Change.BestEffortAllowed), c.Change.LastUpdatedByUserID}, "|"))
	}
	sort.Strings(lines)
	return sha256.Sum256([]byte(strings.Join(lines, "\n")))
}

// codeState binds a code to the pending list, the exact body and the newest
// provision request. The pending list stays the same until a provision
// succeeds, so the request binding voids the code once any provision,
// this one included, is filed after the preview.
func codeState(in *planInput) [32]byte {
	h := pendingHash(in.Changes)
	raw, _ := json.Marshal(in.Body)
	return sha256.Sum256(append(append(h[:], raw...), "\n"+in.NewestRequest...))
}

// planInput is everything one preview, provision or validate reads.
type planInput struct {
	Changes      []*pendingChange
	Caller       string
	CallerErr    error
	Flags        []string
	FlagsErr     error
	Missing      []string
	Ambiguous    []string
	Body         []ak.ProvisionChange
	Dependencies []dependency
	Blocked      []ak.ValidationState
	Failed       []*pendingChange // selected changes whose entity is FAILED (Q-A15)
	// NewestRequest is the id of the newest provision request, "" when there is none.
	NewestRequest string
}

// planReads says what plan reads besides both change lists and the caller id.
type planReads struct {
	// flags holds the tenant flags when the caller has read them; plan reads
	// them only when haveFlags is false.
	flags     []string
	haveFlags bool
	// forValidation builds the body of a dry-run validation request.
	forValidation bool
	// requests reads the newest provision request, which a provision code binds.
	requests bool
	// validation reads the open validation states that block a provision.
	validation bool
}

// previewReads are the reads of the provision preview and the provision.
var previewReads = planReads{requests: true, validation: true}

// plan reads the flags, both change lists, the caller id and what reads
// names, and applies sel. GETs only.
func plan(ctx context.Context, client *ak.AlkiraClient, sel selection, reads planReads) (*planInput, error) {
	in := &planInput{}
	if reads.haveFlags {
		in.Flags = reads.flags
	} else {
		in.Flags, in.FlagsErr = client.EnabledFeatures(ctx)
	}
	stale := slices.Contains(in.Flags, ak.StaleNetworkForChangesFlag)
	tps, err := client.GetNetworkChanges(ctx, stale)
	if err != nil {
		return nil, fmt.Errorf("network changes: %w", err)
	}
	pol, err := client.GetNetworkPolicyChanges(ctx, stale)
	if err != nil {
		return nil, fmt.Errorf("policy changes: %w", err)
	}
	in.Changes = mergeChanges(tps, pol)
	if reads.requests {
		newest, _, err := client.ListProvisionRequestsPage(ctx, "", 1)
		if err != nil {
			return nil, fmt.Errorf("provision requests: %w", err)
		}
		if len(newest) > 0 {
			in.NewestRequest = newest[0].Id
		}
	}
	in.Caller, in.CallerErr = client.GetSelfUserID(ctx)
	attribute(in.Changes, in.Caller)
	in.Missing, in.Ambiguous, err = sel.apply(in.Changes, in.Caller)
	if err != nil {
		return nil, err
	}
	in.Body = body(in.Changes, !reads.forValidation && slices.Contains(in.Flags, bestEffortCleanupFlag))
	in.Dependencies = unselectedDependencies(in.Changes)
	var states []ak.ValidationState
	if reads.validation {
		if states, err = client.ListValidationStates(ctx, blockingValidation...); err != nil {
			return nil, fmt.Errorf("validation states: %w", err)
		}
	}
	for _, c := range in.Changes {
		if !c.Selected {
			continue
		}
		for _, vs := range states {
			if vs.NetworkEntityType == c.Change.Entity.Type && vs.NetworkEntityID == c.Change.Entity.ID && slices.Contains(blockingValidation, vs.State) {
				in.Blocked = append(in.Blocked, vs)
			}
		}
		st, _ := c.Props["state"].(string)
		if c.Change.Entity.ProvisionedState == ak.ProvisionFailed || st == ak.ProvisionFailed {
			in.Failed = append(in.Failed, c)
		}
	}
	return in, nil
}

// preflight reads the network's last provision state. A FAILED network
// returns a refusal carrying the failed requests' errorDetails; a running
// provision refuses too, as the portal disables Provision while the network
// is IN_PROGRESS (karrawa 5d18e83fd diffV2/index.js:90-104).
func preflight(ctx context.Context, client *ak.AlkiraClient, action string) (string, *mcp.CallToolResult) {
	state, provisioned, err := client.GetLastProvisionState(ctx)
	if err != nil {
		return "", refuse("backend_error", fmt.Sprintf("The network's last provision state could not be read, so nothing was %s: %v", action, err))
	}
	if !provisioned {
		return "never provisioned", nil
	}
	if state == ak.ProvisionInProgress {
		return state, refuse("provision_running", "A provision is running on this network (last provision state IN_PROGRESS), so nothing was "+action+
			". Wait for it to finish, then run tenant_network_provision_preview again.")
	}
	if state != ak.ProvisionFailed {
		return state, nil
	}
	return state, refuse("network_failed", "The network's last provision FAILED, so nothing was "+action+": a FAILED network blocks provisioning for every user "+
		"until an operator repairs it in the portal. "+failedDetails(ctx, client))
}

// maxFailedDetails bounds the failed requests read for their errorDetails.
const maxFailedDetails = 3

// failedDetails lists the errorDetails of the newest FAILED provision
// requests from one page (TPS lists the newest first). List rows carry no
// errorDetails, so each is read by id.
func failedDetails(ctx context.Context, client *ak.AlkiraClient) string {
	failed, hits, err := client.ListProvisionRequestsPage(ctx, ak.ProvisionFailed, maxFailedDetails)
	if err != nil {
		return "The failed requests could not be listed: " + err.Error()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "FAILED provision requests: %d.", hits)
	for _, f := range failed {
		pr, err := client.GetProvisionRequest(ctx, f.Id)
		switch {
		case err != nil:
			fmt.Fprintf(&b, " Request %s: details could not be read (%v).", f.Id, err)
		case pr.ErrorDetails == nil:
			fmt.Fprintf(&b, " Request %s: no error details.", f.Id)
		default:
			fmt.Fprintf(&b, " Request %s: %s.", f.Id, errorDetailsText(pr.ErrorDetails))
		}
	}
	return b.String()
}

func errorDetailsText(e *ak.ProvisionErrorDetails) string {
	msg := strconv.Quote(e.Message)
	for _, fe := range e.Errors {
		msg += fmt.Sprintf("; %s: %s", fe.Field, fe.Message)
	}
	if e.Code != "" {
		msg += " (code " + e.Code + ")"
	}
	if e.ContactSupport() {
		msg += " [contact support: this failure needs escalation, not a retry]"
	}
	return msg
}

// planRefusal reports a plan that could not be read; nothing names what the
// tool did not do, as in "listed", "previewed" or "sent".
func planRefusal(err error, nothing string) *mcp.CallToolResult {
	return refuse("backend_error", fmt.Sprintf("The pending changes could not be read, so nothing was %s: %v", nothing, err))
}

// changeLine renders one pending change.
func changeLine(c *pendingChange) string {
	mark := " "
	if c.Selected {
		mark = "*"
	}
	line := fmt.Sprintf("%s %s %s %q", mark, c.Key, c.Change.ChangeType, c.Change.Entity.Name)
	if c.SubType != "" {
		line += " (" + c.SubType + ")"
	}
	switch c.Attribution {
	case byOtherUser:
		line += ", by another user"
		if c.Change.LastUpdatedBy != "" {
			line += " " + c.Change.LastUpdatedBy
		}
		line += " [OTHER USER]"
	default:
		line += ", by " + c.Attribution
	}
	if c.Implicit {
		line += ", implicit (the backend includes it as needed; never sent)"
	}
	return line
}

func flagsNote(in *planInput) string {
	if in.FlagsErr != nil {
		return fmt.Sprintf("Tenant feature flags could not be read (%v); they were treated as off.\n", in.FlagsErr)
	}
	return ""
}

// selectionProblems refuses unknown or ambiguous select items.
func selectionProblems(in *planInput) *mcp.CallToolResult {
	var parts []string
	if len(in.Missing) > 0 {
		parts = append(parts, "not pending: "+strings.Join(in.Missing, ", "))
	}
	if len(in.Ambiguous) > 0 {
		parts = append(parts, "more than one pending change has the id "+strings.Join(in.Ambiguous, ", ")+"; use TYPE:ID")
	}
	if parts == nil {
		return nil
	}
	return refuse("invalid_selection", "Select does not match the pending changes ("+strings.Join(parts, "; ")+"). List them with tenant_network_changes_get.")
}

func blockedText(in *planInput) string {
	var parts []string
	for _, vs := range in.Blocked {
		parts = append(parts, fmt.Sprintf("%s:%s is %s (validation state id %s:%s:%s)", vs.NetworkEntityType, vs.NetworkEntityID, vs.State,
			vs.NetworkEntityType, vs.NetworkEntitySubType, vs.NetworkEntityID))
	}
	return strings.Join(parts, "; ")
}

func failedText(in *planInput) string {
	var keys []string
	for _, c := range in.Failed {
		keys = append(keys, c.Key)
	}
	return strings.Join(keys, ", ")
}

// TenantNetworkChangesGet returns the tenant_network_changes_get handler.
func TenantNetworkChangesGet(client *ak.AlkiraClient) ToolHandler {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, res := toolArgs(req, nil, nil); res != nil {
			return res, nil
		}
		in, err := plan(ctx, client, selection{mode: "keys"}, planReads{})
		if err != nil {
			return planRefusal(err, "listed"), nil
		}
		var out strings.Builder
		out.WriteString(flagsNote(in))
		writeChanges(&out, in)
		return mcp.NewToolResultText(out.String()), nil
	}
}

func writeChanges(out *strings.Builder, in *planInput) {
	others := 0
	for _, c := range in.Changes {
		if c.Attribution == byOtherUser {
			others++
		}
	}
	if in.CallerErr != nil {
		fmt.Fprintf(out, "The caller's user id could not be read (%v); the caller's own changes show as another user's.\n", in.CallerErr)
	}
	fmt.Fprintf(out, "Pending changes: %d. Changes by other users: %d.\n", len(in.Changes), others)
	for _, c := range in.Changes {
		out.WriteString(changeLine(c))
		out.WriteByte('\n')
	}
}

// NewPreviewPipeline returns the pipeline Register serves
// tenant_network_provision_preview with. Unlike NewWritePipeline it never
// draws a random key: a code signed with a per-process key is refused by
// tenant_network_provision on every other replica and in every pipeline
// built with the shared key. Without WithConfirmKey, or with WithoutConfirm,
// it has no signer and the preview answers confirm_unavailable with no
// request.
func NewPreviewPipeline(opts ...WriteOption) (*WritePipeline, error) {
	p := &WritePipeline{timeout: writeTimeout(), now: time.Now}
	for _, o := range opts {
		if err := o(p); err != nil {
			return nil, err
		}
	}
	if p.key == nil || p.noConfirm {
		return p, nil
	}
	s, err := confirm.NewSigner(p.key, confirm.WithClock(p.now))
	if err != nil {
		return nil, err
	}
	p.confirm = s
	return p, nil
}

// ProvisionPreview returns the tenant_network_provision_preview handler.
// It sends GETs only and issues a code signed with p's key.
func ProvisionPreview(p *WritePipeline) func(*ak.AlkiraClient) ToolHandler {
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"select"}, []string{"select"})
			if res != nil {
				return res, nil
			}
			sel, res := parseSelect(args)
			if res != nil {
				return res, nil
			}
			if p.confirm == nil {
				return refuse(RefusalConfirmUnavailable, "Provisioning is unavailable: no confirmation key is configured, so no confirmation code can be issued; nothing was read."), nil
			}
			ctx, cancel := context.WithTimeout(ctx, p.timeout)
			defer cancel()
			netState, res := preflight(ctx, client, "previewed")
			if res != nil {
				return res, nil
			}
			in, err := plan(ctx, client, sel, previewReads)
			if errors.Is(err, errNoCaller) {
				return refuse("caller_unknown", `select "mine" needs the caller's user id, which could not be read; select change keys or "all".`), nil
			}
			if err != nil {
				return planRefusal(err, "previewed"), nil
			}
			if res := selectionProblems(in); res != nil {
				return res, nil
			}

			var out strings.Builder
			out.WriteString("Provision preview. Nothing was provisioned.\n")
			fmt.Fprintf(&out, "Network last provision state: %s.\n", netState)
			out.WriteString(flagsNote(in))
			writeChanges(&out, in)
			fmt.Fprintf(&out, "Selected (*): %d; sent: %d.\n", countSelected(in.Changes), len(in.Body))
			for _, d := range in.Dependencies {
				fmt.Fprintf(&out, "Required by the backend, not selected: %s %q (needed by %s).\n", d.Need.Key, d.Need.Change.Entity.Name, strings.Join(d.By, ", "))
			}
			if len(in.Failed) > 0 {
				fmt.Fprintf(&out, "Warning: selected changes whose objects are FAILED: %s. Provisioning retries them as saved.\n", failedText(in))
			}
			body, _ := json.Marshal(bodyOrEmpty(in.Body))
			fmt.Fprintf(&out, "Request body: %s\n", body)
			switch {
			case len(in.Blocked) > 0:
				fmt.Fprintf(&out, "No confirmCode: selected changes are not validated: %s. Check them with %s, then run the preview again.\n", blockedText(in), ValidationStateTool)
			case len(in.Body) == 0:
				out.WriteString("No confirmCode: nothing to provision; the selection holds no change the portal would send.\n")
			default:
				code, exp := p.confirm.Issue(tenantOf(client), provisionFamily, sel.canonical(), codeState(in))
				fmt.Fprintf(&out, "To provision exactly these changes, call tenant_network_provision with select %s and confirmCode %s before %s. "+
					"Requests filed within about 10 s by other users of the network are deployed in the same run.\n",
					selectText(sel), code, exp.UTC().Format(time.RFC3339))
				fmt.Fprintf(&out, "confirmCode: %s\nexpires: %s\n", code, exp.UTC().Format(time.RFC3339))
			}
			return mcp.NewToolResultText(out.String()), nil
		}
	}
}

func countSelected(changes []*pendingChange) int {
	n := 0
	for _, c := range changes {
		if c.Selected {
			n++
		}
	}
	return n
}

func selectText(sel selection) string {
	if sel.mode != "keys" {
		return strconv.Quote(sel.mode)
	}
	b, _ := json.Marshal(sel.items)
	return string(b)
}

// Provision returns the tenant_network_provision handler.
func Provision(p *WritePipeline) func(*ak.AlkiraClient) ToolHandler {
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"select", "confirmCode", "description"}, []string{"select", "confirmCode"})
			if res != nil {
				return res, nil
			}
			sel, res := parseSelect(args)
			if res != nil {
				return res, nil
			}
			code, _ := args["confirmCode"].(string)
			if code == "" {
				return invalidArgs(errors.New("confirmCode must be a non-empty string from tenant_network_provision_preview")), nil
			}
			desc, isString := args["description"].(string)
			if _, set := args["description"]; set && !isString {
				return invalidArgs(errors.New("description must be a string")), nil
			}
			if n := ak.DescriptionLength(desc); n > ak.MaxProvisionDescription {
				return invalidArgs(fmt.Errorf("description is %d characters long, at most %d allowed", n, ak.MaxProvisionDescription)), nil
			}
			if p.confirm == nil {
				return refuse(RefusalConfirmUnavailable, "Provisioning is unavailable: no confirmation key is configured, so no confirmation code can be checked; nothing was sent."), nil
			}
			tenant := tenantOf(client)
			if err := p.confirm.CheckSignature(code, tenant, provisionFamily, sel.canonical()); err != nil {
				return provisionCodeRefusal(err), nil
			}
			ctx, cancel := context.WithTimeout(ctx, p.timeout)
			defer cancel()
			if err := client.CheckWritesAllowed(ctx); err != nil {
				return writeGateRefusal(err), nil
			}
			_, res = preflight(ctx, client, "sent")
			if res != nil {
				return res, nil
			}
			in, err := plan(ctx, client, sel, previewReads)
			if errors.Is(err, errNoCaller) {
				return refuse("caller_unknown", `select "mine" needs the caller's user id, which could not be read; nothing was sent.`), nil
			}
			if err != nil {
				return planRefusal(err, "sent"), nil
			}
			if len(in.Missing) > 0 || len(in.Ambiguous) > 0 {
				return provisionCodeRefusal(confirm.ErrStale), nil
			}
			if err := p.confirm.Check(code, tenant, provisionFamily, sel.canonical(), codeState(in)); err != nil {
				return provisionCodeRefusal(err), nil
			}
			if len(in.Blocked) > 0 {
				return refuse("validation_pending", fmt.Sprintf("Nothing was sent: selected changes are not validated: %s. Check them with %s.", blockedText(in), ValidationStateTool)), nil
			}
			if len(in.Body) == 0 {
				return refuse("nothing_selected", "Nothing was sent: the selection holds no change the portal would send."), nil
			}

			id, wr, err := client.ProvisionTenantNetworkWith(ctx, ak.ProvisionOptions{Description: desc, Changes: in.Body})
			var out strings.Builder
			switch {
			case errors.Is(err, ak.ErrNoProvisionRequest) && wr != nil:
				// The backend answered 2xx without x-provision-request-id, so
				// the call cannot tell whether a provision request was filed.
				fmt.Fprintf(&out, "The provision was accepted (status %d) but no provision request id was returned, so it cannot be followed by id. "+
					"Do not send it again: run tenant_network_provision_preview, which reports the network's last provision state and the pending changes, to check whether a provision request runs.\n", wr.Status)
			case err != nil:
				return provisionError(err), nil
			default:
				fmt.Fprintf(&out, "Provision requested for %d changes; request id %s. It runs in the background; follow it with tenant_network_provision_request_get.\n", len(in.Body), id)
			}
			if len(in.Failed) > 0 {
				fmt.Fprintf(&out, "Warning: selected changes whose objects were FAILED: %s.\n", failedText(in))
			}
			for _, d := range in.Dependencies {
				fmt.Fprintf(&out, "Not selected, referenced by %s: %s %q.\n", strings.Join(d.By, ", "), d.Need.Key, d.Need.Change.Entity.Name)
			}
			return mcp.NewToolResultText(out.String()), nil
		}
	}
}

func provisionCodeRefusal(err error) *mcp.CallToolResult {
	switch {
	case errors.Is(err, confirm.ErrExpired):
		return refuse(RefusalConfirmExpired, "confirmation code expired; nothing was sent. Run tenant_network_provision_preview again for a new code.")
	case errors.Is(err, confirm.ErrStale):
		return refuse(RefusalConfirmStale, "pending changes changed since the preview; run the preview again. Nothing was sent.")
	}
	return refuse(RefusalConfirmMismatch, "confirmation code does not match this tenant and select; nothing was sent. Use the select of the preview that issued it.")
}

// writeGateRefusal maps the kill-switch and flag-read refusals.
func writeGateRefusal(err error) *mcp.CallToolResult {
	switch {
	case errors.Is(err, ak.ErrWritesDisabled):
		return refuse("writes_disabled", "Changes through MCP are disabled for this tenant; nothing was sent. Make the change in the portal, or ask a tenant administrator to turn MCP changes back on.")
	case errors.Is(err, ak.ErrFlagsUnavailable):
		return refuse("flags_unavailable", capital(ak.ErrFlagsUnavailable.Error())+". Nothing was sent.")
	}
	return refuse("error", "Nothing was sent: "+err.Error())
}

// provisionError maps a failed provision POST (01-api-surface.md section 3).
func provisionError(err error) *mcp.CallToolResult {
	var ae *ak.APIError
	errors.As(err, &ae)
	switch {
	case errors.Is(err, ak.ErrNotSent):
		return refuse("not_sent", fmt.Sprintf("The provision was not sent, nothing changed: %v", err))
	case errors.Is(err, ak.ErrPostSend), errors.Is(err, context.DeadlineExceeded),
		ae != nil && ae.Status == 500 && strings.Contains(ae.Message, "Request failed"):
		return refuse("outcome_unknown", "The provision request may have been filed: the response was lost. Do not send it again. A request filed by this call "+
			"is newer than the call in the portal's provision request list, and while it runs the network's last provision state is IN_PROGRESS, which "+
			"tenant_network_provision_preview reports. Pending changes stay pending until a provision succeeds, so they do not tell whether it was filed."+ids(ae))
	case ae == nil:
		return refuse("error", "The provision failed: "+err.Error())
	case ae.Status == 400 && strings.Contains(ae.Message, "validation status is"):
		return refuse("validation_pending", fmt.Sprintf("%s. Check it with %s.%s", ae.Message, ValidationStateTool, ids(ae)))
	case ae.Status == 400 && (strings.Contains(ae.Message, "is not included") || strings.Contains(ae.Message, "referenced by") || strings.HasPrefix(ae.Message, "De-provision ")):
		return refuse("dependency_not_selected", fmt.Sprintf("The backend refused the selection: %s. Select the changes it names as well, or deprovision the object it names first, then run the preview again.%s", ae.Message, ids(ae)))
	case ae.Status == 400 && strings.Contains(ae.Message, "Invalid change"):
		return refuse(RefusalConfirmStale, "pending changes changed since the preview; run the preview again."+ids(ae))
	case ae.Status == 403:
		return refuse("permission_denied", fmt.Sprintf("Provisioning was refused: %s. Provisioning is not enabled for this tenant, or the API key lacks network provisioning permission.%s", ae.Message, ids(ae)))
	case ae.Status == 409:
		return refuse("conflict", fmt.Sprintf("%s; wait for the running provision to finish.%s", ae.Message, ids(ae)))
	case errors.Is(err, ak.ErrRateLimited):
		return refuse("rate_limited", "Rate limit reached; retry later."+ids(ae))
	case ae.Status >= 500:
		return refuse("backend_error", fmt.Sprintf("The provision failed with status %d%s.", ae.Status, ids(ae)))
	}
	return refuse("backend_error", fmt.Sprintf("The provision failed with status %d: %s%s", ae.Status, ae.Message, ids(ae)))
}

// Validate returns the tenant_network_validate handler.
func Validate(p *WritePipeline) func(*ak.AlkiraClient) ToolHandler {
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"select"}, []string{"select"})
			if res != nil {
				return res, nil
			}
			sel, res := parseSelect(args)
			if res != nil {
				return res, nil
			}
			ctx, cancel := context.WithTimeout(ctx, p.timeout)
			defer cancel()
			if err := client.CheckWritesAllowed(ctx); err != nil {
				return writeGateRefusal(err), nil
			}
			flags, err := client.EnabledFeatures(ctx)
			if err != nil {
				return writeGateRefusal(fmt.Errorf("%w: %w", ak.ErrFlagsUnavailable, err)), nil
			}
			if !slices.Contains(flags, ak.DryRunValidationFlag) {
				return refuse("not_enabled", "Dry-run validation is not enabled for this tenant (tenant feature "+ak.DryRunValidationFlag+" is off); nothing was sent."), nil
			}
			in, err := plan(ctx, client, sel, planReads{flags: flags, haveFlags: true, forValidation: true})
			if errors.Is(err, errNoCaller) {
				return refuse("caller_unknown", `select "mine" needs the caller's user id, which could not be read; nothing was sent.`), nil
			}
			if err != nil {
				return planRefusal(err, "sent"), nil
			}
			if res := selectionProblems(in); res != nil {
				return res, nil
			}
			if len(in.Body) == 0 {
				return refuse("nothing_selected", "Nothing was sent: the selection holds no change the portal would send."), nil
			}
			id, _, err := client.CreateValidationRequest(ctx, in.Body)
			var ae *ak.APIError
			errors.As(err, &ae)
			switch {
			case err == nil:
			case ae != nil && ae.Status == 409:
				return refuse("conflict", "A dry-run validation is already running for this tenant."+ids(ae)), nil
			case ae != nil && (ae.Status == 429 || ae.Status == 503):
				return refuse("unavailable", "Dry-run validation is unavailable now; retry later."+ids(ae)), nil
			case errors.Is(err, ak.ErrNotSent):
				return refuse("not_sent", fmt.Sprintf("The validation request was not sent: %v", err)), nil
			case ae != nil:
				return refuse("backend_error", fmt.Sprintf("The validation request failed with status %d: %s%s", ae.Status, ae.Message, ids(ae))), nil
			default:
				return refuse("error", "The validation request failed: "+err.Error()), nil
			}
			body, _ := json.Marshal(bodyOrEmpty(in.Body))
			return mcp.NewToolResultText(fmt.Sprintf("Dry-run validation started for %d changes; validation request id %s. Nothing is saved or provisioned. "+
				"Follow it with tenant_network_validation_request_get.\nRequest body: %s\n", len(in.Body), id, body)), nil
		}
	}
}

// bodyOrEmpty renders a nil body as [] in tool output.
func bodyOrEmpty(b []ak.ProvisionChange) []ak.ProvisionChange {
	if b == nil {
		return []ak.ProvisionChange{}
	}
	return b
}
