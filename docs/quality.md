# Quality Gates

## Rule Spec (`animus.quality.rule.v1`)

Rules are immutable JSON documents with a schema identifier and a list of checks:

```json
{
  "schema": "animus.quality.rule.v1",
  "checks": [
    { "id": "size", "type": "object_size_bytes", "max_bytes": 10485760 },
    { "id": "ctype", "type": "content_type_in", "allowed": ["text/csv"] },
    { "id": "cols", "type": "csv_header_has_columns", "columns": ["id", "label"], "delimiter": "," }
  ]
}
```

Supported check types:

- `object_size_bytes`: enforce `min_bytes` and/or `max_bytes` using object store size.
- `content_type_in`: enforce allowed content types.
- `filename_suffix_in`: enforce allowed filename suffixes (from dataset version metadata).
- `metadata_required_keys`: require metadata keys to be present and non-empty.
- `csv_header_has_columns`: parse the first line as CSV header and require columns.
- `verify_content_sha256`: stream the object and verify it matches `dataset_versions.content_sha256`.
- `content_sha256_in`: enforce an allowlist for `dataset_versions.content_sha256`.

## Evaluations

Evaluations are immutable records linked to a dataset version and rule.
Each evaluation produces a JSON report stored in the MinIO `artifacts` bucket.

## Enforcement

Dataset downloads are blocked unless the latest evaluation status for the dataset version’s
`quality_rule_id` is `pass`.

