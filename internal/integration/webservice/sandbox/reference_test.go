package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
)

var updateReference = flag.Bool("update-reference", false, "regenerate reference/reference.json from reference/members.tsv and the live sandbox")

const membersSource = "reference/members.tsv"

const referenceProbe = `
const paths = %s;
const kinds = {};
const parentOf = (path) => path.includes('.') ? eval(path.slice(0, path.lastIndexOf('.'))) : globalThis;
const nameOf = (path) => path.slice(path.lastIndexOf('.') + 1);
for (const path of paths) {
  try {
    const parent = parentOf(path);
    const name = nameOf(path);
    if (parent === null || parent === undefined || !(name in Object(parent))) {
      kinds[path] = 'missing';
      continue;
    }
    if (path.includes('.to.')) {
      let target = Object(parent);
      let descriptor;
      while (target && !descriptor) {
        descriptor = Object.getOwnPropertyDescriptor(target, name);
        target = Object.getPrototypeOf(target);
      }
      kinds[path] = descriptor.get ? 'property' : typeof descriptor.value === 'function' ? 'function' : 'property';
      continue;
    }
    const value = eval(path);
    kinds[path] = typeof value === 'function' ? 'function' : value !== null && typeof value === 'object' && !Array.isArray(value) ? 'object' : 'property';
  }
  catch (error) {
    kinds[path] = 'missing: ' + error.message;
  }
}
console.log(JSON.stringify(kinds));
`

// TestScriptReferenceIsGenerated checks reference.json against members.tsv and the live sandbox; -update-reference rewrites it.
func TestScriptReferenceIsGenerated(t *testing.T) {
	members := readMembersSource(t)
	kinds := probeKinds(t, members)
	reference := dto.ScriptReference{Members: make([]dto.ScriptMember, 0, len(members))}
	for _, member := range members {
		kind := kinds[member.Path]
		if kind != "object" && kind != "function" && kind != "property" {
			t.Errorf("%s is not in the sandbox: %s", member.Path, kind)
		}
		member.Kind = kind
		reference.Members = append(reference.Members, member)
	}
	generated, err := json.MarshalIndent(reference, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	generated = append(generated, '\n')
	if *updateReference {
		if err := os.WriteFile(referenceFile, generated, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	embedded, err := referenceFiles.ReadFile(referenceFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.ReplaceAll(embedded, []byte("\r\n"), []byte("\n")), generated) && !*updateReference {
		t.Fatalf("%s is out of date: run make script-reference", referenceFile)
	}
}

func readMembersSource(t *testing.T) []dto.ScriptMember {
	t.Helper()
	content, err := os.ReadFile(membersSource)
	if err != nil {
		t.Fatal(err)
	}
	members := []dto.ScriptMember{}
	for _, line := range strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		columns := strings.Split(line, "\t")
		if len(columns) != 3 {
			t.Fatalf("%s: expected path, signature and description: %q", membersSource, line)
		}
		members = append(members, dto.ScriptMember{Path: columns[0], Signature: columns[1], Description: columns[2]})
	}
	return members
}

func probeKinds(t *testing.T, members []dto.ScriptMember) map[string]string {
	t.Helper()
	paths := make([]string, 0, len(members))
	for _, member := range members {
		paths = append(paths, member.Path)
	}
	encoded, _ := json.Marshal(paths)
	var output []string
	request := entity.NewRequest("01R", "Probe")
	request.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.NewBodyRaw("json", "{}")}
	request.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "t"}}
	scope := entity.ScriptContext{
		Request:  request,
		Response: &entity.Response{StatusCode: 200, StatusText: "OK", Headers: []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}, Body: []byte("{}"), BodyLength: 2, SizeBytes: 2},
		Info:     entity.NewScriptInfo("01R", "Probe", "Probe", []string{"Probe"}),
		Scopes:   entity.ScriptScopes{Environment: entity.NewScriptScope("dev", []entity.ScriptVariable{})},
		Console: func(log entity.ScriptLog) {
			if log.Level.String() == "log" {
				output = append(output, log.Text())
			}
		},
	}
	probe := entity.ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Request, OwnerName: "Probe", Text: strings.Replace(referenceProbe, "%s", string(encoded), 1)}
	result := testRunner(time.Minute, memoryGrowthLimit).Run(context.Background(), probe, scope, nil, entity.DefaultAppSettings())
	if result.HasFailed() || len(output) == 0 {
		t.Fatalf("probe failed: %+v %v", result.Error, output)
	}
	kinds := map[string]string{}
	if err := json.Unmarshal([]byte(output[len(output)-1]), &kinds); err != nil {
		t.Fatalf("probe output: %v", err)
	}
	return kinds
}
