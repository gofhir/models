---
title: "Custom Marshal"
linkTitle: "Custom Marshal"
description: "HTML-safe JSON serialization with r4.Marshal() for preserving FHIR narrative XHTML content."
weight: 3
---

The `gofhir/models` library provides custom JSON marshaling functions that solve a critical problem with Go's standard `json.Marshal`: HTML entity escaping in FHIR narrative content.

## The HTML Escaping Problem

Go's standard `json.Marshal` function escapes HTML characters (`<`, `>`, `&`) in strings to their Unicode escape sequences (`\u003c`, `\u003e`, `\u0026`). This is a security measure to prevent XSS attacks when JSON is embedded in HTML, but it breaks FHIR narrative content.

FHIR resources can include a `text.div` field that contains valid XHTML. The FHIR specification requires this content to be preserved exactly as provided. When Go's standard marshaler escapes the HTML entities, the resulting JSON is technically valid but violates the FHIR specification and can cause problems with FHIR servers and clients that expect unescaped XHTML.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("narrative-example"),
        Text: &r4.Narrative{
            Status: ptrTo(r4.NarrativeStatusGenerated),
            Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith, Male, DOB: 1990-01-15</p></div>`),
        },
    }

    // Standard json.Marshal escapes HTML entities
    data, _ := json.Marshal(patient)
    fmt.Println(string(data))
    // Output contains escaped HTML:
    // "div":"\u003cdiv xmlns=\"http://www.w3.org/1999/xhtml\"\u003e\u003cp\u003eJohn Smith, Male, DOB: 1990-01-15\u003c/p\u003e\u003c/div\u003e"
}
```

The escaped output `\u003cdiv\u003e` is valid JSON, but many FHIR implementations expect the literal HTML characters. This can lead to validation errors, display issues, or interoperability problems.

## The Solution: r4.Marshal()

The `r4.Marshal()` function uses `json.NewEncoder` with `SetEscapeHTML(false)` to produce JSON output that preserves HTML characters exactly as they appear in the Go strings:

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("narrative-example"),
        Text: &r4.Narrative{
            Status: ptrTo(r4.NarrativeStatusGenerated),
            Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>Hello</p></div>`),
        },
    }

    // Standard json.Marshal escapes HTML entities
    data1, _ := json.Marshal(patient)
    fmt.Println(string(data1))
    // {"resourceType":"Patient","id":"narrative-example","text":{"status":"generated","div":"\u003cdiv xmlns=\"http://www.w3.org/1999/xhtml\"\u003e\u003cp\u003eHello\u003c/p\u003e\u003c/div\u003e"}}

    // r4.Marshal preserves HTML as required by FHIR
    data2, _ := r4.Marshal(patient)
    fmt.Println(string(data2))
    // {"resourceType":"Patient","id":"narrative-example","text":{"status":"generated","div":"<div xmlns=\"http://www.w3.org/1999/xhtml\"><p>Hello</p></div>"}}
}
```

## Function Signatures

The library provides two custom marshal functions, defined in `marshal.go`:

### Marshal

```go
func Marshal(v interface{}) ([]byte, error)
```

Serializes any value to JSON without HTML escaping. Internally, it creates a `bytes.Buffer`, sets up a `json.NewEncoder` with `SetEscapeHTML(false)`, and strips the trailing newline that `Encode` appends.

### MarshalIndent

```go
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
```

Produces the same HTML-safe output as `Marshal` but applies indentation for human-readable output. It first calls `Marshal` to get the compact JSON, then applies `json.Indent` with the specified prefix and indent strings.

```go
data, err := r4.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Output:

```json
{
  "resourceType": "Patient",
  "id": "narrative-example",
  "text": {
    "status": "generated",
    "div": "<div xmlns=\"http://www.w3.org/1999/xhtml\"><p>Hello</p></div>"
  }
}
```

## Implementation Details

The `Marshal` function is straightforward but handles an important edge case. The `json.Encoder.Encode()` method appends a trailing newline character, which is not desirable for raw byte output. The function strips this newline:

```go
func Marshal(v interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := json.NewEncoder(&buf)
    enc.SetEscapeHTML(false)
    if err := enc.Encode(v); err != nil {
        return nil, err
    }
    b := buf.Bytes()
    if len(b) > 0 && b[len(b)-1] == '\n' {
        b = b[:len(b)-1]
    }
    return b, nil
}
```

## When to Use Each Function

| Scenario | Recommended Function |
|----------|---------------------|
| FHIR resource with narrative content | `r4.Marshal()` |
| FHIR resource for API response | `r4.Marshal()` |
| Quick debugging / logging | `json.Marshal()` or `r4.MarshalIndent()` |
| Storing FHIR resources | `r4.Marshal()` |
| Non-FHIR types (general Go structs) | `json.Marshal()` |

{{< callout type="info" >}}
The `r4.Marshal()` function works with any Go value, not just FHIR resources. However, its primary purpose is to handle the HTML escaping issue in FHIR narrative content. If your resources never contain narrative XHTML, `json.Marshal()` will produce functionally equivalent output.
{{< /callout >}}

## Using with HTTP Handlers

When serving FHIR resources over HTTP, use `r4.Marshal()` to ensure the response body contains unescaped HTML:

```go
func handlePatient(w http.ResponseWriter, r *http.Request) {
    patient := buildPatientWithNarrative()

    data, err := r4.Marshal(patient)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    w.Write(data)
}
```

Alternatively, if you prefer streaming the response, you can set up the encoder directly:

```go
func handlePatient(w http.ResponseWriter, r *http.Request) {
    patient := buildPatientWithNarrative()

    w.Header().Set("Content-Type", "application/fhir+json")
    enc := json.NewEncoder(w)
    enc.SetEscapeHTML(false)
    enc.Encode(patient)
}
```

## Version Availability

The `Marshal` and `MarshalIndent` functions are available in all version packages:

- `r4.Marshal()` / `r4.MarshalIndent()` for FHIR R4
- `r4b.Marshal()` / `r4b.MarshalIndent()` for FHIR R4B
- `r5.Marshal()` / `r5.MarshalIndent()` for FHIR R5

All implementations are identical in behavior.
