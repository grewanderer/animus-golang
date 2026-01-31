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

	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	requiredKeys := []string{
		"$schema",
		"$id",
		"title",
		"type",
		"properties",
		"$defs",
	}
	for _, key := range requiredKeys {
		if _, ok := doc[key]; !ok {
			t.Fatalf("missing top-level key %q", key)
		}
	}

	properties, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties is not a map")
	}

	requiredProps := []string{
		"apiVersion",
		"kind",
		"specVersion",
		"spec",
	}
	for _, key := range requiredProps {
		if _, ok := properties[key]; !ok {
			t.Fatalf("missing properties key %q", key)
		}
	}
}
