package specvalidator

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPipelineSpecSchemaLoads(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "api", "pipeline_spec.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	var doc map[string]interface{}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	requiredKeys := []string{
		"$schema",
		"$id",
		"apiVersion",
		"kind",
		"specVersion",
		"spec",
	}
	for _, key := range requiredKeys {
		if _, ok := doc[key]; !ok {
			t.Fatalf("missing top-level key %q", key)
		}
	}
}
