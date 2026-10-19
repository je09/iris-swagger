package swagger

import (
	"strings"
	"testing"
)

func TestToOpenAPI3(t *testing.T) {
	v2 := `{"swagger":"2.0","info":{"title":"t","version":"1"},"host":"api.test","basePath":"/v1",
"securityDefinitions":{"ApiKeyAuth":{"type":"apiKey","name":"Authorization","in":"header"}},
"paths":{"/x":{"post":{"consumes":["application/json"],"produces":["application/json"],
"parameters":[{"in":"body","name":"body","schema":{"$ref":"#/definitions/X"}}],
"responses":{"200":{"description":"ok","schema":{"$ref":"#/definitions/X"}}}}}},
"definitions":{"X":{"type":"object","properties":{"a":{"type":"string"}}}}}`

	out, err := toOpenAPI3(v2)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"openapi":"3.0.3"`, `"#/components/schemas/X"`, `"requestBody"`, `"url":"https://api.test/v1"`, `"ApiKeyAuth"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}

	if v3 := `{"openapi":"3.0.3"}`; mustConvert(t, v3) != v3 {
		t.Error("openapi 3 doc must pass through")
	}
	if out, err := toOpenAPI3("{bad"); err == nil || out != "{bad" {
		t.Error("invalid doc must be returned as is with an error")
	}
}

func mustConvert(t *testing.T, doc string) string {
	out, err := toOpenAPI3(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
