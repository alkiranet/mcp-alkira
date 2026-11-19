package registry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/internal/testfamily"
)

// MA-W36: no sleep, ticker, timer or retry loop anywhere, and no client write
// call (core or legacy AlkiraApi) outside the pipeline in handlers/write.go.
func TestGate_StaticChecks(t *testing.T) {
	for name, src := range map[string]string{
		"sleep":        "func f() { time.Sleep(time.Second) }",
		"ticker":       "t := time.NewTicker(d)",
		"core call":    "c, _ := ak.NewCore[X](client, d)",
		"legacy write": "_, err := ak.NewSegment(client).Create(&seg)",
		"legacy del":   "err := api.Delete(id)",
		"credential":   "api.CreateCredential(name, typ, c, 0)",
		"cred update":  "api.UpdateCredential(id, name, typ, c, 0)",
		"cred delete":  "api.DeleteCredential(id, typ)",
		"core delete":  "wr, err := core.DeleteWith(ctx, id, ak.DeleteOptions{})",
		"retry loop":   "for attempt := 0; attempt < 3; attempt++ {",
		"retry label":  "retry:\n\tgoto retry",
	} {
		if got := sourceViolations("handlers/family.go", src); len(got) == 0 {
			t.Errorf("seeded %s not caught", name)
		}
	}

	root := ".."
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".") && path != root) {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if got := sourceViolations(filepath.ToSlash(rel), string(src)); got != nil {
			t.Errorf("%s: %v", rel, got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

var (
	sleepRE = regexp.MustCompile(`time\.(Sleep|NewTicker|NewTimer|Tick|After|AfterFunc)\(`)
	coreRE  = regexp.MustCompile(`ak\.NewCore\[`)
	// Any write-verb call: Core Create, Update, Delete and DeleteWith, the
	// legacy AlkiraApi[T] writes and the credential writes.
	writeRE = regexp.MustCompile(`\.(Create|Update|Delete|DeleteWith|CreateCredential|UpdateCredential|DeleteCredential)\(`)
	retryRE = regexp.MustCompile(`(?i)\bfor\b[^{\n]*\b(retry|retries|attempt|attempts|backoff)\b|\bgoto\s+\w*(?i:retry|again)`)
)

func sourceViolations(path, src string) []string {
	var out []string
	if sleepRE.MatchString(src) {
		out = append(out, "sleep, ticker or timer")
	}
	if retryRE.MatchString(src) {
		out = append(out, "retry loop")
	}
	if path != "handlers/write.go" {
		if coreRE.MatchString(src) {
			out = append(out, "core constructed outside handlers/write.go")
		}
		if writeRE.MatchString(src) {
			out = append(out, "client write call outside handlers/write.go")
		}
	}
	return out
}

// Scenarios of the 03-TEST-PLAN.md section 3a matrix.
const (
	scCRUD              = "CRUD"
	scExternalDelete    = "ExternalDelete"
	scRetryOn429        = "RetryOn429"
	scImportNonexistent = "ImportNonexistent"
	scWaveCoalescing    = "WaveCoalescing"
	scRequestFailed     = "CreateFailed, UpdateFailed, DeleteFailed"
	scPreflightBlocked  = "PreflightBlocked"
	scMissingRequestID  = "MissingRequestID"
	scTimeoutPartial    = "Timeout and PartialSuccess"
	scSoftDeleteOff     = "SoftDeleteOff"
)

// applicableScenarios returns the section 3a scenarios a family must cover.
func applicableScenarios(d ak.Descriptor) []string {
	sc := []string{scCRUD, scExternalDelete, scRetryOn429, scImportNonexistent}
	if d.Provisionable {
		sc = append(sc, scWaveCoalescing, scRequestFailed, scPreflightBlocked, scMissingRequestID, scTimeoutPartial)
	}
	if d.MarkedRowLifecycle {
		sc = append(sc, scSoftDeleteOff)
	}
	return sc
}

// scenarioRows maps family -> scenario -> the tests or rows that cover it.
// A Test name must exist in this package; other entries name the MA, CG or
// H row and the ticket that delivers it.
// Families whose specs live in their own test files add their rows in init.
var scenarioRows = map[string]map[string]string{
	"widget": {
		scCRUD:              "TestWrite_CreateRequestShape TestWrite_UpdateMerge TestWrite_DeleteWithCode",
		scExternalDelete:    "TestWrite_DeleteNotFound TestWrite_DeleteCodeStale",
		scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
		scImportNonexistent: "TestWrite_DeleteNotFound; sibling-family id: CG-W15 (client-go)",
		scWaveCoalescing:    "MA-W13 (ST-03)",
		scRequestFailed:     "MA-W18 (ST-03)",
		scPreflightBlocked:  "TestWrite_StateRefusals; MA-W13 (ST-03)",
		scMissingRequestID:  "MA-W12 (ST-03)",
		scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		scSoftDeleteOff:     "TestWrite_StateRefusals TestWrite_DeleteCodeStale",
	},
	"segment": {
		scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K01.1-.5",
		scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteAlreadyGone TestFamilies_DeleteWithCode",
		scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
		scImportNonexistent: "TestFamilies_NotFound; K01 E5",
		scWaveCoalescing:    "MA-W13 (ST-03)",
		scRequestFailed:     "MA-W18 (ST-03); E11 segment half: TestFamilies_StateGuard",
		scPreflightBlocked:  "TestFamilies_StateGuard; MA-W13 (ST-03)",
		scMissingRequestID:  "MA-W12 (ST-03)",
		scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
	},
	"group": {
		scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K02.1-.5",
		scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteAlreadyGone TestFamilies_DeleteWithCode",
		scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
		scImportNonexistent: "TestFamilies_NotFound",
		scWaveCoalescing:    "MA-W13 (ST-03)",
		scRequestFailed:     "MA-W18 (ST-03)",
		scPreflightBlocked:  "TestFamilies_StateGuard (groups carry no state); MA-W13 (ST-03)",
		scMissingRequestID:  "MA-W12 (ST-03)",
		scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
	},
	"billing_tag": {
		scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K03.1-.5",
		scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
		scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
		scImportNonexistent: "TestFamilies_NotFound",
	},
}

var testFuncRE = regexp.MustCompile(`(?m)^func (Test\w+)\(`)

// scenarioGaps lists missing or stale scenario rows of the given families.
func scenarioGaps(families []writeFamily, rows map[string]map[string]string, tests []string) []string {
	var gaps []string
	for _, f := range families {
		for _, sc := range applicableScenarios(f.Descriptor) {
			row := strings.TrimSpace(rows[f.Name][sc])
			if row == "" {
				gaps = append(gaps, fmt.Sprintf("%s: no row for %s", f.Name, sc))
				continue
			}
			for _, w := range strings.FieldsFunc(row, func(r rune) bool { return r == ' ' || r == ';' || r == ',' }) {
				if strings.HasPrefix(w, "Test") && !slices.Contains(tests, w) {
					gaps = append(gaps, fmt.Sprintf("%s: %s names missing test %s", f.Name, sc, w))
				}
			}
		}
	}
	return gaps
}

// Section 3a checklist: every served write family has a row per applicable scenario.
func TestGate_ScenarioChecklist(t *testing.T) {
	var tests []string
	files, _ := filepath.Glob("*_test.go")
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range testFuncRE.FindAllStringSubmatch(string(src), -1) {
			tests = append(tests, m[1])
		}
	}
	families := append([]writeFamily{family(testfamily.Family)}, writeFamilies...)
	if gaps := scenarioGaps(families, scenarioRows, tests); gaps != nil {
		t.Errorf("scenario checklist:\n%s", strings.Join(gaps, "\n"))
	}

	seeded := []writeFamily{{Name: "gadget", Descriptor: ak.Descriptor{Collection: "gadgets", Provisionable: true}}}
	if gaps := scenarioGaps(seeded, scenarioRows, tests); len(gaps) != 9 {
		t.Errorf("seeded family without rows: gaps = %v", gaps)
	}
	stale := map[string]map[string]string{"widget": {}}
	for k, v := range scenarioRows["widget"] {
		stale["widget"][k] = v
	}
	stale["widget"][scCRUD] = "TestWrite_Gone"
	if gaps := scenarioGaps(families[:1], stale, tests); len(gaps) != 1 {
		t.Errorf("stale test name: gaps = %v", gaps)
	}
}

// MA-W40: breaking changes of the tools/list goldens against the baseline,
// the goldens at the mcp-alkira commit mcp-alkira-hosted pins (7428a2c3e579).
func TestGate_GoldenDiff(t *testing.T) {
	allowed := deprecations(t, "../schema/deprecations.yaml")
	for _, name := range []string{"tools_list.golden.json", "tools_list_write.golden.json"} {
		base := readTools(t, filepath.Join("testdata/baseline", name))
		cur := readTools(t, filepath.Join("testdata", name))
		for _, f := range breakingChanges(base, cur) {
			if !slices.Contains(allowed, f) {
				t.Errorf("%s: breaking change %q not listed in schema/deprecations.yaml", name, f)
			}
		}
	}

	tool := func(schema string) map[string]json.RawMessage {
		return map[string]json.RawMessage{"t": json.RawMessage(`{"name":"t","inputSchema":` + schema + `}`)}
	}
	base := tool(`{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"object","properties":{"c":{"type":"integer"}}}},"required":["a"]}`)
	for name, c := range map[string]struct {
		schema string
		want   []string
	}{
		"removed field": {`{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"object","properties":{}}},"required":["a"]}`, []string{"t b.c removed"}},
		"renamed field": {`{"type":"object","properties":{"a2":{"type":"string"},"b":{"type":"object","properties":{"c":{"type":"integer"}}}},"required":["a2"]}`, []string{"t a removed", "t a2 required"}},
		"type change":   {`{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"object","properties":{"c":{"type":"string"}}}},"required":["a"]}`, []string{"t b.c type"}},
		"new required":  {`{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"object","properties":{"c":{"type":"integer"}},"required":["c"]}},"required":["a"]}`, []string{"t b.c required"}},
		"new optional":  {`{"type":"object","properties":{"a":{"type":"string"},"d":{"type":"string"},"b":{"type":"object","properties":{"c":{"type":"integer"}}}},"required":["a"]}`, nil},
	} {
		if got := breakingChanges(base, tool(c.schema)); !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
	if got := breakingChanges(base, map[string]json.RawMessage{}); !slices.Equal(got, []string{"t - tool-removed"}) {
		t.Errorf("removed tool: %v", got)
	}
}

// readTools returns a golden's tools by name; a missing baseline file is an empty set.
func readTools(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) && strings.Contains(path, "baseline") {
		return map[string]json.RawMessage{}
	}
	if err != nil {
		t.Fatal(err)
	}
	var list []json.RawMessage
	if err := json.Unmarshal(b, &list); err != nil {
		t.Fatal(err)
	}
	out := map[string]json.RawMessage{}
	for _, raw := range list {
		var tool struct{ Name string }
		json.Unmarshal(raw, &tool)
		out[tool.Name] = raw
	}
	return out
}

type schemaNode struct {
	Type       any                   `json:"type"`
	Properties map[string]schemaNode `json:"properties"`
	Items      *schemaNode           `json:"items"`
	Required   []string              `json:"required"`
}

// breakingChanges lists "<tool> <path> <kind>" for removed tools and fields,
// type changes and fields that became required. New optional fields pass.
func breakingChanges(base, cur map[string]json.RawMessage) []string {
	var out []string
	for name, raw := range base {
		craw, ok := cur[name]
		if !ok {
			out = append(out, name+" - tool-removed")
			continue
		}
		var b, c struct {
			InputSchema schemaNode `json:"inputSchema"`
		}
		json.Unmarshal(raw, &b)
		json.Unmarshal(craw, &c)
		compareNode(name, "", b.InputSchema, c.InputSchema, &out)
	}
	sort.Strings(out)
	return out
}

func compareNode(tool, path string, b, c schemaNode, out *[]string) {
	if fmt.Sprint(b.Type) != fmt.Sprint(c.Type) {
		*out = append(*out, fmt.Sprintf("%s %s type", tool, orDash(path)))
		return
	}
	for k, bp := range b.Properties {
		cp, ok := c.Properties[k]
		if !ok {
			*out = append(*out, fmt.Sprintf("%s %s removed", tool, join(path, k)))
			continue
		}
		compareNode(tool, join(path, k), bp, cp, out)
	}
	for _, k := range c.Required {
		if !slices.Contains(b.Required, k) {
			*out = append(*out, fmt.Sprintf("%s %s required", tool, join(path, k)))
		}
	}
	if b.Items != nil && c.Items != nil {
		compareNode(tool, path+"[]", *b.Items, *c.Items, out)
	}
}

func join(path, k string) string {
	if path == "" {
		return k
	}
	return path + "." + k
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// deprecations reads the "- <tool> <path> <kind>" entries of schema/deprecations.yaml.
func deprecations(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if e, ok := strings.CutPrefix(strings.TrimSpace(line), "- "); ok {
			out = append(out, strings.Join(strings.Fields(strings.Trim(e, `"'`)), " "))
		}
	}
	return out
}
