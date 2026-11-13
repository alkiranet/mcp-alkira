package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mark3labs/mcp-go/server"
)

func newServer(get func(context.Context) (*ak.AlkiraClient, error)) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", "test")
	Register(srv, get)
	return srv
}

func noClient(context.Context) (*ak.AlkiraClient, error) {
	return nil, errors.New("no client in this test")
}

// call sends one JSON-RPC request through srv and returns the raw "result" member.
func call(t *testing.T, srv *server.MCPServer, method string, params any) json.RawMessage {
	t.Helper()
	req := map[string]any{"jsonrpc": "2.0", "id": 1, "method": method}
	if params != nil {
		req["params"] = params
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(srv.HandleMessage(context.Background(), body))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Error != nil {
		t.Fatalf("%s returned a JSON-RPC error: %s", method, env.Error)
	}
	return env.Result
}

func toolNames(t *testing.T, srv *server.MCPServer) []string {
	t.Helper()
	var res struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, srv, "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	return names
}

type toolResult struct {
	IsError bool `json:"isError"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func callTool(t *testing.T, srv *server.MCPServer, name string, args map[string]any) toolResult {
	t.Helper()
	var res toolResult
	raw := call(t, srv, "tools/call", map[string]any{"name": name, "arguments": args})
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Content) != 1 || res.Content[0].Type != "text" {
		t.Fatalf("%s: want one text content item, got %s", name, raw)
	}
	return res
}

func TestRegister_AllToolsUseProvider(t *testing.T) {
	var calls atomic.Int64
	srv := newServer(func(ctx context.Context) (*ak.AlkiraClient, error) {
		calls.Add(1)
		return nil, errors.New("provider failed")
	})
	names := toolNames(t, srv)

	t.Run("tools", func(t *testing.T) {
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				res := callTool(t, srv, name, map[string]any{})
				if !res.IsError || res.Content[0].Text != "provider failed" {
					t.Fatalf("want isError with provider error, got isError=%v text=%q", res.IsError, res.Content[0].Text)
				}
			})
		}
	})

	if got := calls.Load(); got != int64(len(names)) {
		t.Fatalf("provider calls = %d, want %d", got, len(names))
	}
}

type requestLog struct {
	mu    sync.Mutex
	paths []string
}

func (l *requestLog) add(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, path)
}

func (l *requestLog) get() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.paths...)
}

// fakePortal serves canned JSON for the given paths and records each request path.
func fakePortal(t *testing.T, routes map[string]string) (*ak.AlkiraClient, *requestLog) {
	t.Helper()
	seen := &requestLog{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.add(r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "api-key test" {
			t.Errorf("Authorization = %q", got)
		}
		body, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(ts.Close)

	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = 0
	httpClient.Logger = nil
	return &ak.AlkiraClient{
		Client:          httpClient,
		URI:             ts.URL + "/api",
		TenantNetworkId: "7",
		Authorization:   "api-key test",
	}, seen
}

func providerFor(client *ak.AlkiraClient) func(context.Context) (*ak.AlkiraClient, error) {
	return func(context.Context) (*ak.AlkiraClient, error) { return client, nil }
}

func TestRegister_SegmentGetAll_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworks/7/segments": `[{"id":1,"name":"seg-a","asn":65000},{"id":2,"name":"seg-b","asn":65001}]`,
	})
	res := callTool(t, newServer(providerFor(client)), "segment_get_all", map[string]any{})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if text := res.Content[0].Text; !strings.Contains(text, "seg-a") || !strings.Contains(text, "seg-b") {
		t.Fatalf("summary missing segments: %s", text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworks/7/segments" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_SegmentGetById_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworks/7/segments/42": `{"id":42,"name":"seg-42"}`,
	})
	res := callTool(t, newServer(providerFor(client)), "segment_get_by_id", map[string]any{"segmentId": "42"})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "seg-42") {
		t.Fatalf("unexpected body: %s", res.Content[0].Text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworks/7/segments/42" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_TenantNetworkSummary_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworksummaries": `[{"id":7,"name":"tn-7"}]`,
	})
	res := callTool(t, newServer(providerFor(client)), "tenant_network_summary", map[string]any{})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "tn-7") {
		t.Fatalf("unexpected body: %s", res.Content[0].Text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworksummaries" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_Annotations(t *testing.T) {
	var res struct {
		Tools []struct {
			Name        string `json:"name"`
			Annotations struct {
				ReadOnlyHint    *bool `json:"readOnlyHint"`
				DestructiveHint *bool `json:"destructiveHint"`
			} `json:"annotations"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, newServer(noClient), "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if hint := tool.Annotations.ReadOnlyHint; hint == nil || !*hint {
			t.Errorf("%s: readOnlyHint is not true", tool.Name)
		}
		if hint := tool.Annotations.DestructiveHint; hint == nil || *hint {
			t.Errorf("%s: destructiveHint is not false", tool.Name)
		}
	}
}

var update = flag.Bool("update", false, "rewrite testdata/tools_list.golden.json from the current tools/list output")

const toolsListGolden = "testdata/tools_list.golden.json"

// TestRegister_ToolsListGolden fails on any change to a tool's name, description, schema or
// annotations. Regenerate with: go test ./registry -run ToolsListGolden -update
func TestRegister_ToolsListGolden(t *testing.T) {
	var res struct {
		Tools json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(call(t, newServer(noClient), "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := json.Indent(&got, res.Tools, "", "  "); err != nil {
		t.Fatal(err)
	}
	got.WriteByte('\n')

	if *update {
		if err := os.MkdirAll(filepath.Dir(toolsListGolden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(toolsListGolden, got.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(toolsListGolden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("tools/list differs from %s; run with -update if the change is intended", toolsListGolden)
	}
}

func TestRegister_NilClientIsToolError(t *testing.T) {
	res := callTool(t, newServer(providerFor(nil)), "segment_get_all", map[string]any{})
	if !res.IsError || res.Content[0].Text != "no Alkira client available for this request" {
		t.Fatalf("want a nil-client tool error, got isError=%v text=%q", res.IsError, res.Content[0].Text)
	}
}

// connector_ipsec_tunnel_profile_get_all treats offset and limit as optional and omits them from the query.
func TestRegister_ConnectorIPSecTunnelProfileGetAll_NoArgs_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/v1/tenantnetworks/7/ipsec-tunnel-profiles": `{"data":[{"id":3,"name":"tp-3"}]}`,
	})
	res := callTool(t, newServer(providerFor(client)), "connector_ipsec_tunnel_profile_get_all", map[string]any{})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "tp-3") {
		t.Fatalf("unexpected body: %s", res.Content[0].Text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/v1/tenantnetworks/7/ipsec-tunnel-profiles" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_ToolNames(t *testing.T) {
	got := toolNames(t, newServer(noClient))
	slices.Sort(got)
	if !slices.Equal(got, wantToolNames) {
		t.Fatalf("tool names differ from the expected set:\ngot  %v\nwant %v", got, wantToolNames)
	}
}

// wantToolNames is the sorted set of tool names Register adds.
var wantToolNames = []string{
	"billing_tag_get_all",
	"billing_tag_get_by_id",
	"billing_tag_get_by_name",
	"billing_tag_get_total",
	"byoip_get_all",
	"byoip_get_by_id",
	"byoip_get_by_name",
	"byoip_get_total",
	"connector_aruba_edge_get_all",
	"connector_aruba_edge_get_by_id",
	"connector_aruba_edge_get_by_name",
	"connector_aruba_edge_get_total",
	"connector_aws_direct_connect_get_all",
	"connector_aws_direct_connect_get_by_id",
	"connector_aws_direct_connect_get_by_name",
	"connector_aws_direct_connect_get_total",
	"connector_aws_tgw_get_all",
	"connector_aws_tgw_get_by_id",
	"connector_aws_tgw_get_by_name",
	"connector_aws_tgw_get_total",
	"connector_aws_vpc_get_all",
	"connector_aws_vpc_get_by_id",
	"connector_aws_vpc_get_by_name",
	"connector_aws_vpc_get_total",
	"connector_azure_express_route_get_all",
	"connector_azure_express_route_get_by_id",
	"connector_azure_express_route_get_by_name",
	"connector_azure_express_route_get_total",
	"connector_azure_vnet_get_all",
	"connector_azure_vnet_get_by_id",
	"connector_azure_vnet_get_by_name",
	"connector_azure_vnet_get_total",
	"connector_cisco_sdwan_get_all",
	"connector_cisco_sdwan_get_by_id",
	"connector_cisco_sdwan_get_by_name",
	"connector_cisco_sdwan_get_total",
	"connector_fortinet_sdwan_get_all",
	"connector_fortinet_sdwan_get_by_id",
	"connector_fortinet_sdwan_get_by_name",
	"connector_fortinet_sdwan_get_total",
	"connector_gcp_interconnect_get_all",
	"connector_gcp_interconnect_get_by_id",
	"connector_gcp_interconnect_get_by_name",
	"connector_gcp_interconnect_get_total",
	"connector_gcp_vpc_get_all",
	"connector_gcp_vpc_get_by_id",
	"connector_gcp_vpc_get_by_name",
	"connector_gcp_vpc_get_total",
	"connector_internet_get_all",
	"connector_internet_get_by_id",
	"connector_internet_get_by_name",
	"connector_internet_get_total",
	"connector_ipsec_adv_get_all",
	"connector_ipsec_adv_get_by_id",
	"connector_ipsec_adv_get_by_name",
	"connector_ipsec_adv_get_total",
	"connector_ipsec_get_all",
	"connector_ipsec_get_by_id",
	"connector_ipsec_get_by_name",
	"connector_ipsec_get_total",
	"connector_ipsec_tunnel_profile_get_all",
	"connector_oci_vcn_get_all",
	"connector_oci_vcn_get_by_id",
	"connector_oci_vcn_get_by_name",
	"connector_oci_vcn_get_total",
	"connector_remote_access_get_all",
	"connector_remote_access_get_by_id",
	"connector_remote_access_get_by_name",
	"connector_remote_access_get_total",
	"connector_versa_sdwan_get_all",
	"connector_versa_sdwan_get_by_id",
	"connector_versa_sdwan_get_by_name",
	"connector_versa_sdwan_get_total",
	"connector_vmware_sdwan_get_all",
	"connector_vmware_sdwan_get_by_id",
	"connector_vmware_sdwan_get_by_name",
	"connector_vmware_sdwan_get_total",
	"cxp_get_all",
	"group_get_by_id",
	"group_get_by_name",
	"group_get_total",
	"health_connector_get_by_id",
	"health_connector_instance_get_by_id",
	"health_service_get_by_id",
	"health_service_instance_get_by_id",
	"internet_application_get_all",
	"internet_application_get_by_id",
	"internet_application_get_by_name",
	"internet_application_get_total",
	"list_as_path_get_all",
	"list_as_path_get_by_id",
	"list_as_path_get_by_name",
	"list_as_path_get_total",
	"list_community_get_all",
	"list_community_get_by_id",
	"list_community_get_by_name",
	"list_community_get_total",
	"list_dns_server_get_all",
	"list_dns_server_get_by_id",
	"list_dns_server_get_by_name",
	"list_dns_server_get_total",
	"list_extended_community_get_all",
	"list_extended_community_get_by_id",
	"list_extended_community_get_by_name",
	"list_extended_community_get_total",
	"list_global_cidr_get_all",
	"list_global_cidr_get_by_id",
	"list_global_cidr_get_by_name",
	"list_global_cidr_get_total",
	"list_policy_fqdn_get_all",
	"list_policy_fqdn_get_by_id",
	"list_policy_fqdn_get_by_name",
	"list_policy_fqdn_get_total",
	"list_prefix_get_by_id",
	"list_prefix_get_by_name",
	"list_prefix_get_by_prefix",
	"list_udr_get_all",
	"list_udr_get_by_id",
	"list_udr_get_by_name",
	"list_udr_get_total",
	"notification_alerts",
	"notification_auditLogs",
	"notification_jobs",
	"policy_nat_get_all",
	"policy_nat_get_by_id",
	"policy_nat_get_by_name",
	"policy_nat_get_total",
	"policy_nat_rule_get_all",
	"policy_nat_rule_get_by_id",
	"policy_nat_rule_get_by_name",
	"policy_nat_rule_get_total",
	"policy_route_get_all",
	"policy_route_get_by_id",
	"policy_route_get_by_name",
	"policy_route_get_total",
	"policy_traffic_get_all",
	"policy_traffic_get_by_id",
	"policy_traffic_get_by_name",
	"policy_traffic_rule_get_all",
	"policy_traffic_rule_get_by_id",
	"policy_traffic_rule_get_by_name",
	"policy_traffic_rule_list_get_all",
	"policy_traffic_rule_list_get_by_id",
	"policy_traffic_rule_list_get_by_name",
	"route_get",
	"route_get_all",
	"route_get_count",
	"route_get_summary",
	"segment_get_all",
	"segment_get_by_id",
	"segment_get_by_name",
	"segment_resource_get_all",
	"segment_resource_get_by_id",
	"segment_resource_get_by_name",
	"segment_resource_get_total",
	"segment_resource_share_get_all",
	"segment_resource_share_get_by_id",
	"segment_resource_share_get_by_name",
	"segment_resource_share_get_total",
	"service_check_point_get_all",
	"service_check_point_get_by_id",
	"service_check_point_get_by_name",
	"service_check_point_get_total",
	"service_cisco_ftdv_get_all",
	"service_cisco_ftdv_get_by_id",
	"service_cisco_ftdv_get_by_name",
	"service_cisco_ftdv_get_total",
	"service_f5_lb_get_all",
	"service_f5_lb_get_by_id",
	"service_f5_lb_get_by_name",
	"service_f5_lb_get_total",
	"service_fortinet_get_all",
	"service_fortinet_get_by_id",
	"service_fortinet_get_by_name",
	"service_fortinet_get_total",
	"service_infoblox_get_all",
	"service_infoblox_get_by_id",
	"service_infoblox_get_by_name",
	"service_infoblox_get_total",
	"service_pan_get_all",
	"service_pan_get_by_id",
	"service_pan_get_by_name",
	"service_pan_get_total",
	"service_zscaler_get_all",
	"service_zscaler_get_by_id",
	"service_zscaler_get_by_name",
	"service_zscaler_get_total",
	"tenant_network_firewall_zones",
	"tenant_network_summary",
	"tenant_resource_limits",
	"tenant_resource_usages",
}
