package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"testing"
	"time"
)

func TestEvaluate_ObjectSize(t *testing.T) {
	spec := RuleSpec{
		Schema: ruleSpecSchemaV1,
		Checks: []CheckSpec{
			{ID: "max", Type: checkObjectSizeBytes, MaxBytes: ptrInt64(10)},
		},
	}
	in := evaluationInputs{
		DatasetVersion: datasetVersion{VersionID: "v1", DatasetID: "d1", ObjectKey: "k"},
		RuleID:         "r1",
		RuleName:       "rule",
		RuleSpec:       spec,
		Object:         objectInfo{Size: 5},
	}
	report := evaluate(context.Background(), time.Unix(1700000000, 0).UTC(), "alice", "e1", in, func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("")), nil
	})
	if report.Status != "pass" {
		t.Fatalf("status=%q, want pass", report.Status)
	}

	in.Object.Size = 50
	report = evaluate(context.Background(), time.Unix(1700000000, 0).UTC(), "alice", "e2", in, func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("")), nil
	})
	if report.Status != "fail" {
		t.Fatalf("status=%q, want fail", report.Status)
	}
}

func TestEvaluate_CSVHeaderHasColumns(t *testing.T) {
	spec := RuleSpec{
		Schema: ruleSpecSchemaV1,
		Checks: []CheckSpec{
			{ID: "csv", Type: checkCSVHeaderHasColumns, Columns: []string{"a", "b"}},
		},
	}
	in := evaluationInputs{
		DatasetVersion: datasetVersion{VersionID: "v1", DatasetID: "d1", ObjectKey: "k"},
		RuleID:         "r1",
		RuleName:       "rule",
		RuleSpec:       spec,
		Object:         objectInfo{Size: 10, ContentType: "text/csv"},
	}
	report := evaluate(context.Background(), time.Unix(1700000000, 0).UTC(), "alice", "e1", in, func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("a,b,c\n1,2,3\n")), nil
	})
	if report.Status != "pass" {
		t.Fatalf("status=%q, want pass", report.Status)
	}
}

func TestEvaluate_VerifyContentSHA256(t *testing.T) {
	content := []byte("hello")
	sum := sha256.Sum256(content)
	spec := RuleSpec{
		Schema: ruleSpecSchemaV1,
		Checks: []CheckSpec{
			{ID: "sha", Type: checkVerifyContentSHA256},
		},
	}
	in := evaluationInputs{
		DatasetVersion: datasetVersion{
			VersionID:     "v1",
			DatasetID:     "d1",
			ObjectKey:     "k",
			ContentSHA256: hex.EncodeToString(sum[:]),
		},
		RuleID:   "r1",
		RuleName: "rule",
		RuleSpec: spec,
		Object:   objectInfo{Size: int64(len(content))},
	}
	report := evaluate(context.Background(), time.Unix(1700000000, 0).UTC(), "alice", "e1", in, func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(string(content))), nil
	})
	if report.Status != "pass" {
		t.Fatalf("status=%q, want pass", report.Status)
	}
}
