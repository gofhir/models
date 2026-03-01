---
title: "Summary Fields"
linkTitle: "Summary Fields"
description: "Pre-computed summary field lists for implementing the FHIR _summary=true search parameter behavior on a FHIR server."
weight: 2
---

The `SummaryFields` variable is a package-level map that lists, for every resource type, the fields that the FHIR specification marks with `isSummary=true`. This data is essential for implementing the `_summary=true` search parameter on a FHIR server.

## What Are Summary Fields?

The FHIR specification defines a `_summary` search parameter that allows clients to request abbreviated versions of resources. When a client sends `_summary=true`, the server must return only the fields marked as summary fields in the StructureDefinition for that resource type, plus a few mandatory elements (`id`, `meta`, `resourceType`).

This is useful for reducing payload sizes in search results where the client only needs key identifiers and attributes rather than the full resource.

## The SummaryFields Map

Each version package exports a `SummaryFields` variable:

```go
var SummaryFields = map[string][]string{
    "Patient": {
        "active",
        "address",
        "birthDate",
        "communication",
        "gender",
        "generalPractitioner",
        "id",
        "identifier",
        "implicitRules",
        "link",
        "managingOrganization",
        "meta",
        "name",
        "telecom",
    },
    // ... all other resource types
}
```

The map is keyed by resource type name (e.g., `"Patient"`, `"Observation"`) and the values are sorted slices of field names that should be included in a summary response.

## Usage

### Basic Lookup

```go
import "github.com/gofhir/models/r4"

fields := r4.SummaryFields["Patient"]
// Returns: ["active", "address", "birthDate", "communication", "gender",
//           "generalPractitioner", "id", "identifier", "implicitRules",
//           "link", "managingOrganization", "meta", "name", "telecom"]
```

### Checking if a Field is a Summary Field

```go
import "github.com/gofhir/models/r4"

func isSummaryField(resourceType, fieldName string) bool {
    fields, ok := r4.SummaryFields[resourceType]
    if !ok {
        return false
    }
    for _, f := range fields {
        if f == fieldName {
            return true
        }
    }
    return false
}

isSummaryField("Patient", "name")      // true
isSummaryField("Patient", "photo")     // false
isSummaryField("Observation", "code")  // true
isSummaryField("Observation", "note")  // false
```

### FHIR Server Implementation

A typical FHIR server uses `SummaryFields` to filter resource fields before returning search results:

```go
import (
    "encoding/json"
    "github.com/gofhir/models/r4"
)

func applySummary(resourceType string, data []byte) ([]byte, error) {
    summaryFields := r4.SummaryFields[resourceType]
    if summaryFields == nil {
        return data, nil // unknown type, return as-is
    }

    // Build a set for fast lookup
    allowed := make(map[string]bool, len(summaryFields))
    for _, f := range summaryFields {
        allowed[f] = true
    }
    // Always include resourceType
    allowed["resourceType"] = true

    // Parse, filter, and re-serialize
    var full map[string]json.RawMessage
    if err := json.Unmarshal(data, &full); err != nil {
        return nil, err
    }

    filtered := make(map[string]json.RawMessage)
    for key, val := range full {
        if allowed[key] {
            filtered[key] = val
        }
    }

    return json.Marshal(filtered)
}
```

## Coverage

The `SummaryFields` map includes entries for every resource type defined in the FHIR version. In R4, this covers all 148 resource types. Each entry is generated directly from the `isSummary` flag in the official FHIR StructureDefinitions.

{{< callout type="info" >}}
The summary field lists are generated from the FHIR specification and always stay in sync with the resource struct definitions. If a field is added or removed from the summary set in a new FHIR version, the regenerated code will reflect the change automatically.
{{< /callout >}}

## Comparison with _elements

The `_summary=true` parameter is a coarser mechanism than the `_elements` parameter. With `_elements`, clients specify exactly which fields they want. With `_summary=true`, the server returns the specification-defined summary set. `SummaryFields` supports the latter case; implementing `_elements` requires a different approach (typically JSON field filtering based on the client-provided field list).
