package main

import "testing"

func TestRuleSpecValidate(t *testing.T) {
	spec := RuleSpec{
		Schema: ruleSpecSchemaV1,
		Checks: []CheckSpec{
			{ID: "size", Type: checkObjectSizeBytes, MaxBytes: ptrInt64(10)},
			{ID: "ctype", Type: checkContentTypeIn, Allowed: []string{"text/csv"}},
		},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("Validate() err=%v", err)
	}

	invalid := spec
	invalid.Schema = "bad"
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected schema error")
	}
}

func ptrInt64(v int64) *int64 { return &v }
