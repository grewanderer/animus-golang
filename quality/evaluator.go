package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const evaluationReportSchemaV1 = "animus.quality.evaluation_report.v1"

type evaluationReport struct {
	Schema         string          `json:"schema"`
	EvaluationID   string          `json:"evaluation_id"`
	DatasetVersion datasetVersion  `json:"dataset_version"`
	Rule           qualityRuleInfo `json:"rule"`
	Status         string          `json:"status"`
	EvaluatedAt    time.Time       `json:"evaluated_at"`
	EvaluatedBy    string          `json:"evaluated_by"`
	Summary        summary         `json:"summary"`
	Checks         []checkResult   `json:"checks"`
	Error          string          `json:"error,omitempty"`
	RawRuleSpec    json.RawMessage `json:"raw_rule_spec,omitempty"`
}

type qualityRuleInfo struct {
	RuleID string `json:"rule_id"`
	Name   string `json:"name"`
}

type datasetVersion struct {
	VersionID     string          `json:"version_id"`
	DatasetID     string          `json:"dataset_id"`
	Ordinal       int64           `json:"ordinal"`
	ObjectKey     string          `json:"object_key"`
	ContentSHA256 string          `json:"content_sha256"`
	SizeBytes     int64           `json:"size_bytes"`
	ContentType   string          `json:"content_type,omitempty"`
	Filename      string          `json:"filename,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	QualityRuleID string          `json:"quality_rule_id,omitempty"`
}

type summary struct {
	ChecksTotal int      `json:"checks_total"`
	ChecksPass  int      `json:"checks_pass"`
	ChecksFail  int      `json:"checks_fail"`
	ChecksError int      `json:"checks_error"`
	Failing     []string `json:"failing_check_ids,omitempty"`
}

type checkResult struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Status     string         `json:"status"`
	Message    string         `json:"message,omitempty"`
	Observed   map[string]any `json:"observed,omitempty"`
	Expected   map[string]any `json:"expected,omitempty"`
	DurationMs int64          `json:"duration_ms"`
}

type objectInfo struct {
	Size        int64
	ContentType string
}

type evaluationInputs struct {
	DatasetVersion datasetVersion
	RuleID         string
	RuleName       string
	RuleSpec       RuleSpec
	RuleSpecRaw    json.RawMessage
	Object         objectInfo
}

type objectOpener func(ctx context.Context, objectKey string) (io.ReadCloser, error)

func evaluate(ctx context.Context, now time.Time, evaluatedBy string, evaluationID string, in evaluationInputs, openObject objectOpener) evaluationReport {
	report := evaluationReport{
		Schema:       evaluationReportSchemaV1,
		EvaluationID: evaluationID,
		DatasetVersion: datasetVersion{
			VersionID:     in.DatasetVersion.VersionID,
			DatasetID:     in.DatasetVersion.DatasetID,
			Ordinal:       in.DatasetVersion.Ordinal,
			ObjectKey:     in.DatasetVersion.ObjectKey,
			ContentSHA256: in.DatasetVersion.ContentSHA256,
			SizeBytes:     in.Object.Size,
			ContentType:   firstNonEmpty(in.Object.ContentType, in.DatasetVersion.ContentType),
			Filename:      in.DatasetVersion.Filename,
			Metadata:      in.DatasetVersion.Metadata,
			QualityRuleID: in.DatasetVersion.QualityRuleID,
		},
		Rule: qualityRuleInfo{
			RuleID: in.RuleID,
			Name:   in.RuleName,
		},
		EvaluatedAt: now.UTC(),
		EvaluatedBy: evaluatedBy,
		RawRuleSpec: in.RuleSpecRaw,
	}

	checks := make([]checkResult, 0, len(in.RuleSpec.Checks))
	var (
		passCount  int
		failCount  int
		errorCount int
		failingIDs []string
	)

	for _, check := range in.RuleSpec.Checks {
		start := time.Now()
		result := evaluateCheck(ctx, check, in, openObject)
		result.DurationMs = time.Since(start).Milliseconds()
		checks = append(checks, result)

		switch result.Status {
		case "pass":
			passCount++
		case "fail":
			failCount++
			failingIDs = append(failingIDs, check.ID)
		default:
			errorCount++
			failingIDs = append(failingIDs, check.ID)
		}
	}

	report.Checks = checks
	report.Summary = summary{
		ChecksTotal: len(checks),
		ChecksPass:  passCount,
		ChecksFail:  failCount,
		ChecksError: errorCount,
		Failing:     failingIDs,
	}

	switch {
	case errorCount > 0:
		report.Status = "error"
	case failCount > 0:
		report.Status = "fail"
	default:
		report.Status = "pass"
	}

	return report
}

func evaluateCheck(ctx context.Context, check CheckSpec, in evaluationInputs, openObject objectOpener) checkResult {
	kind := strings.ToLower(strings.TrimSpace(check.Type))
	result := checkResult{
		ID:   strings.TrimSpace(check.ID),
		Type: kind,
	}

	switch kind {
	case checkObjectSizeBytes:
		min := int64(-1)
		max := int64(-1)
		if check.MinBytes != nil {
			min = *check.MinBytes
		}
		if check.MaxBytes != nil {
			max = *check.MaxBytes
		}
		size := in.Object.Size

		result.Observed = map[string]any{"size_bytes": size}
		result.Expected = map[string]any{}
		if min >= 0 {
			result.Expected["min_bytes"] = min
		}
		if max >= 0 {
			result.Expected["max_bytes"] = max
		}

		if min >= 0 && size < min {
			result.Status = "fail"
			result.Message = "size below minimum"
			return result
		}
		if max >= 0 && size > max {
			result.Status = "fail"
			result.Message = "size above maximum"
			return result
		}
		result.Status = "pass"
		return result

	case checkContentTypeIn:
		allowed := trimNonEmpty(check.Allowed)
		actual := strings.TrimSpace(firstNonEmpty(in.Object.ContentType, in.DatasetVersion.ContentType))
		result.Observed = map[string]any{"content_type": actual}
		result.Expected = map[string]any{"allowed": allowed}
		for _, item := range allowed {
			if strings.EqualFold(item, actual) {
				result.Status = "pass"
				return result
			}
		}
		result.Status = "fail"
		result.Message = "content type not allowed"
		return result

	case checkFilenameSuffixIn:
		allowed := trimNonEmpty(check.Allowed)
		filename := strings.TrimSpace(in.DatasetVersion.Filename)
		result.Observed = map[string]any{"filename": filename}
		result.Expected = map[string]any{"allowed": allowed}
		for _, item := range allowed {
			if strings.HasSuffix(strings.ToLower(filename), strings.ToLower(strings.TrimSpace(item))) {
				result.Status = "pass"
				return result
			}
		}
		result.Status = "fail"
		result.Message = "filename suffix not allowed"
		return result

	case checkMetadataRequiredKeys:
		required := trimNonEmpty(check.Keys)
		meta, err := parseJSONObject(in.DatasetVersion.Metadata)
		if err != nil {
			result.Status = "error"
			result.Message = "invalid metadata json"
			return result
		}
		missing := make([]string, 0)
		for _, key := range required {
			if !hasNonEmptyKey(meta, key) {
				missing = append(missing, key)
			}
		}
		result.Expected = map[string]any{"required_keys": required}
		result.Observed = map[string]any{"missing_keys": missing}
		if len(missing) > 0 {
			result.Status = "fail"
			result.Message = "missing required metadata keys"
			return result
		}
		result.Status = "pass"
		return result

	case checkCSVHeaderHasColumns:
		cols := trimNonEmpty(check.Columns)
		delimiter := strings.TrimSpace(check.Delimiter)
		if delimiter == "" {
			delimiter = ","
		}
		if len([]rune(delimiter)) != 1 {
			result.Status = "error"
			result.Message = "invalid delimiter"
			return result
		}

		reader, err := openObject(ctx, in.DatasetVersion.ObjectKey)
		if err != nil {
			result.Status = "error"
			result.Message = "object read failed"
			return result
		}
		defer reader.Close()

		header, err := readCSVHeader(reader, rune(delimiter[0]))
		if err != nil {
			result.Status = "error"
			result.Message = "csv header parse failed"
			result.Observed = map[string]any{"error": err.Error()}
			return result
		}

		missing := make([]string, 0)
		for _, required := range cols {
			if !containsString(header, required) {
				missing = append(missing, required)
			}
		}
		result.Expected = map[string]any{"required_columns": cols, "delimiter": delimiter}
		result.Observed = map[string]any{"header": header, "missing": missing}
		if len(missing) > 0 {
			result.Status = "fail"
			result.Message = "missing required csv columns"
			return result
		}
		result.Status = "pass"
		return result

	case checkVerifyContentSHA256:
		reader, err := openObject(ctx, in.DatasetVersion.ObjectKey)
		if err != nil {
			result.Status = "error"
			result.Message = "object read failed"
			return result
		}
		defer reader.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, reader); err != nil {
			result.Status = "error"
			result.Message = "sha256 compute failed"
			return result
		}
		actual := hex.EncodeToString(hasher.Sum(nil))
		expected := strings.ToLower(strings.TrimSpace(in.DatasetVersion.ContentSHA256))
		result.Observed = map[string]any{"content_sha256": actual}
		result.Expected = map[string]any{"content_sha256": expected}
		if actual != expected {
			result.Status = "fail"
			result.Message = "sha256 mismatch"
			return result
		}
		result.Status = "pass"
		return result

	case checkContentSHA256In:
		allowed := trimNonEmpty(check.Allowed)
		actual := strings.ToLower(strings.TrimSpace(in.DatasetVersion.ContentSHA256))
		result.Observed = map[string]any{"content_sha256": actual}
		result.Expected = map[string]any{"allowed": allowed}
		for _, item := range allowed {
			if strings.EqualFold(item, actual) {
				result.Status = "pass"
				return result
			}
		}
		result.Status = "fail"
		result.Message = "sha256 not in allowlist"
		return result

	default:
		result.Status = "error"
		result.Message = "unsupported check type"
		return result
	}
}

func parseJSONObject(raw json.RawMessage) (map[string]any, error) {
	raw = bytesTrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func hasNonEmptyKey(m map[string]any, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return true
	}
}

func readCSVHeader(r io.Reader, delimiter rune) ([]string, error) {
	const maxHeaderBytes = 1 << 20
	br := bufio.NewReader(io.LimitReader(r, maxHeaderBytes))
	line, err := br.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		return nil, errors.New("empty header")
	}

	cr := csv.NewReader(strings.NewReader(line))
	cr.Comma = delimiter
	cr.FieldsPerRecord = -1
	fields, err := cr.Read()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		col := strings.TrimSpace(f)
		if col == "" {
			continue
		}
		key := strings.ToLower(col)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, col)
	}
	return out, nil
}

func containsString(in []string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	for _, item := range in {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return true
		}
	}
	return false
}

func bytesTrimSpace(in []byte) []byte {
	return []byte(strings.TrimSpace(string(in)))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
