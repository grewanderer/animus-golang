package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	ruleSpecSchemaV1 = "animus.quality.rule.v1"

	checkObjectSizeBytes      = "object_size_bytes"
	checkContentTypeIn        = "content_type_in"
	checkFilenameSuffixIn     = "filename_suffix_in"
	checkMetadataRequiredKeys = "metadata_required_keys"
	checkCSVHeaderHasColumns  = "csv_header_has_columns"
	checkVerifyContentSHA256  = "verify_content_sha256"
	checkContentSHA256In      = "content_sha256_in"
)

type RuleSpec struct {
	Schema string      `json:"schema"`
	Checks []CheckSpec `json:"checks"`
}

type CheckSpec struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	MinBytes *int64 `json:"min_bytes,omitempty"`
	MaxBytes *int64 `json:"max_bytes,omitempty"`

	Allowed []string `json:"allowed,omitempty"`
	Keys    []string `json:"keys,omitempty"`

	Columns   []string `json:"columns,omitempty"`
	Delimiter string   `json:"delimiter,omitempty"`
}

func (s RuleSpec) Validate() error {
	if strings.TrimSpace(s.Schema) != ruleSpecSchemaV1 {
		return fmt.Errorf("spec.schema must be %q", ruleSpecSchemaV1)
	}
	if len(s.Checks) == 0 {
		return errors.New("spec.checks must be non-empty")
	}

	seen := make(map[string]struct{}, len(s.Checks))
	for i, check := range s.Checks {
		id := strings.TrimSpace(check.ID)
		if id == "" {
			return fmt.Errorf("spec.checks[%d].id is required", i)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("spec.checks[%d].id must be unique (duplicate %q)", i, id)
		}
		seen[id] = struct{}{}

		kind := strings.ToLower(strings.TrimSpace(check.Type))
		if kind == "" {
			return fmt.Errorf("spec.checks[%d].type is required", i)
		}

		switch kind {
		case checkObjectSizeBytes:
			if check.MinBytes == nil && check.MaxBytes == nil {
				return fmt.Errorf("spec.checks[%d] object_size_bytes requires min_bytes or max_bytes", i)
			}
			if check.MinBytes != nil && *check.MinBytes < 0 {
				return fmt.Errorf("spec.checks[%d].min_bytes must be >= 0", i)
			}
			if check.MaxBytes != nil && *check.MaxBytes < 0 {
				return fmt.Errorf("spec.checks[%d].max_bytes must be >= 0", i)
			}
			if check.MinBytes != nil && check.MaxBytes != nil && *check.MinBytes > *check.MaxBytes {
				return fmt.Errorf("spec.checks[%d].min_bytes must be <= max_bytes", i)
			}
		case checkContentTypeIn, checkFilenameSuffixIn:
			allowed := trimNonEmpty(check.Allowed)
			if len(allowed) == 0 {
				return fmt.Errorf("spec.checks[%d] %s requires allowed", i, kind)
			}
		case checkMetadataRequiredKeys:
			keys := trimNonEmpty(check.Keys)
			if len(keys) == 0 {
				return fmt.Errorf("spec.checks[%d] metadata_required_keys requires keys", i)
			}
		case checkCSVHeaderHasColumns:
			cols := trimNonEmpty(check.Columns)
			if len(cols) == 0 {
				return fmt.Errorf("spec.checks[%d] csv_header_has_columns requires columns", i)
			}
			if strings.TrimSpace(check.Delimiter) != "" && len([]rune(strings.TrimSpace(check.Delimiter))) != 1 {
				return fmt.Errorf("spec.checks[%d].delimiter must be a single character", i)
			}
		case checkVerifyContentSHA256:
		case checkContentSHA256In:
			allowed := trimNonEmpty(check.Allowed)
			if len(allowed) == 0 {
				return fmt.Errorf("spec.checks[%d] content_sha256_in requires allowed", i)
			}
			for _, item := range allowed {
				value := strings.ToLower(strings.TrimSpace(item))
				if len(value) != 64 {
					return fmt.Errorf("spec.checks[%d] content_sha256_in allowed items must be 64-char hex", i)
				}
				if _, err := hex.DecodeString(value); err != nil {
					return fmt.Errorf("spec.checks[%d] content_sha256_in allowed items must be hex: %w", i, err)
				}
			}
		default:
			return fmt.Errorf("spec.checks[%d].type unsupported: %q", i, kind)
		}
	}
	return nil
}

func trimNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, item := range in {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}
