---
title: "Real-World Usage"
linkTitle: "Real-World Usage"
description: "Production patterns for FHIR servers, API clients, format conversion, and type-safe resource routing."
weight: 2
---

This page demonstrates production-ready patterns for using `gofhir/models` in real applications. All examples use the R4 package.

## Helper Function

```go
func ptrTo[T any](v T) *T {
    return &v
}
```

## 1. FHIR Server Resource Creation Endpoint

An HTTP handler that accepts any FHIR resource, validates it, assigns an ID, and stores it:

```go
import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/google/uuid"
    "github.com/gofhir/models/r4"
)

func handleCreateResource(w http.ResponseWriter, req *http.Request) {
    if req.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(req.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }

    // Peek at the resource type before full deserialization
    resourceType, err := r4.GetResourceType(body)
    if err != nil {
        http.Error(w, "Invalid FHIR resource: missing resourceType", http.StatusBadRequest)
        return
    }

    if !r4.IsKnownResourceType(resourceType) {
        http.Error(w, fmt.Sprintf("Unknown resource type: %s", resourceType), http.StatusBadRequest)
        return
    }

    // Deserialize into the correct Go struct
    resource, err := r4.UnmarshalResource(body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to parse %s: %v", resourceType, err), http.StatusBadRequest)
        return
    }

    // Assign a server-generated ID
    resource.SetId(uuid.New().String())

    // Set metadata
    resource.SetMeta(&r4.Meta{
        VersionId:   ptrTo("1"),
        LastUpdated: ptrTo("2024-06-15T12:00:00Z"),
    })

    // Store the resource (implementation-specific)
    if err := store.Save(resource); err != nil {
        http.Error(w, "Failed to store resource", http.StatusInternalServerError)
        return
    }

    // Return the created resource
    responseData, err := r4.Marshal(resource)
    if err != nil {
        http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/fhir+json")
    w.Header().Set("Location", fmt.Sprintf("/%s/%s", resourceType, *resource.GetId()))
    w.WriteHeader(http.StatusCreated)
    w.Write(responseData)
}
```

## 2. Deserializing a Bundle from an API Response

Parsing a search result Bundle from a FHIR server API and extracting individual resources:

```go
import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gofhir/models/r4"
)

func fetchPatients(baseURL string) ([]*r4.Patient, error) {
    resp, err := http.Get(baseURL + "/Patient?_count=50")
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Deserialize as a Bundle
    resource, err := r4.UnmarshalResource(body)
    if err != nil {
        return nil, fmt.Errorf("failed to parse Bundle: %w", err)
    }

    bundle, ok := resource.(*r4.Bundle)
    if !ok {
        return nil, fmt.Errorf("expected Bundle, got %s", resource.GetResourceType())
    }

    // Extract Patient resources from Bundle entries
    var patients []*r4.Patient
    for _, entry := range bundle.Entry {
        if entry.Resource == nil {
            continue
        }
        if patient, ok := entry.Resource.(*r4.Patient); ok {
            patients = append(patients, patient)
        }
    }

    fmt.Printf("Found %d patients (total: %d)\n", len(patients), safeTotal(bundle.Total))
    return patients, nil
}

func safeTotal(t *uint32) uint32 {
    if t != nil {
        return *t
    }
    return 0
}
```

## 3. Converting Between JSON and XML Format

The library supports both FHIR JSON and FHIR XML. You can convert between formats by deserializing from one and serializing to the other:

```go
import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

// JSON to XML conversion
func jsonToXML(jsonData []byte) ([]byte, error) {
    // Parse JSON into a Resource
    resource, err := r4.UnmarshalResource(jsonData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    // Serialize to XML
    xmlData, err := r4.MarshalResourceXMLIndent(resource, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal XML: %w", err)
    }

    return xmlData, nil
}

// XML to JSON conversion
func xmlToJSON(xmlData []byte) ([]byte, error) {
    // Parse XML into a Resource
    resource, err := r4.UnmarshalResourceXML(xmlData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse XML: %w", err)
    }

    // Serialize to JSON (HTML-safe)
    jsonData, err := r4.MarshalIndent(resource, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal JSON: %w", err)
    }

    return jsonData, nil
}
```

Usage example:

```go
jsonInput := []byte(`{
    "resourceType": "Patient",
    "id": "example",
    "name": [{"family": "Smith", "given": ["John"]}]
}`)

xmlOutput, err := jsonToXML(jsonInput)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(xmlOutput))
// <?xml version="1.0" encoding="UTF-8"?>
// <Patient xmlns="http://hl7.org/fhir">
//   <id value="example"/>
//   <name>
//     <family value="Smith"/>
//     <given value="John"/>
//   </name>
// </Patient>
```

## 4. Type-Safe Resource Routing Using the Registry

A resource router that dispatches to type-specific handlers using the registry:

```go
import (
    "fmt"
    "net/http"

    "github.com/gofhir/models/r4"
)

type ResourceHandler func(r4.Resource) error

type FHIRRouter struct {
    handlers map[string]ResourceHandler
}

func NewFHIRRouter() *FHIRRouter {
    return &FHIRRouter{
        handlers: make(map[string]ResourceHandler),
    }
}

func (router *FHIRRouter) Handle(resourceType string, handler ResourceHandler) {
    if !r4.IsKnownResourceType(resourceType) {
        panic(fmt.Sprintf("unknown resource type: %s", resourceType))
    }
    router.handlers[resourceType] = handler
}

func (router *FHIRRouter) Dispatch(data []byte) error {
    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        return fmt.Errorf("failed to unmarshal: %w", err)
    }

    handler, ok := router.handlers[resource.GetResourceType()]
    if !ok {
        return fmt.Errorf("no handler registered for %s", resource.GetResourceType())
    }

    return handler(resource)
}
```

Register handlers for specific resource types:

```go
router := NewFHIRRouter()

router.Handle("Patient", func(res r4.Resource) error {
    patient := res.(*r4.Patient)
    fmt.Printf("Processing patient: %s\n", safeName(patient))
    // Validate, store, transform, etc.
    return nil
})

router.Handle("Observation", func(res r4.Resource) error {
    obs := res.(*r4.Observation)
    fmt.Printf("Processing observation: %s\n", safeCode(obs))
    // Process the observation
    return nil
})

// Dispatch incoming resources
err := router.Dispatch(incomingJSON)
```

### Listing All Supported Resource Types

You can use `AllResourceTypes` to auto-register default handlers or generate API documentation:

```go
import (
    "sort"
    "github.com/gofhir/models/r4"
)

// Generate OpenAPI paths for all resource types
types := r4.AllResourceTypes()
sort.Strings(types)

for _, rt := range types {
    fmt.Printf("/%s:\n", rt)
    fmt.Printf("  get:\n    summary: Search %s resources\n", rt)
    fmt.Printf("  post:\n    summary: Create a new %s\n", rt)
}
```

{{< callout type="info" >}}
The registry functions use an internal factory map that is populated at compile time. There is no runtime registration step and no initialization required beyond importing the package. All registry functions are safe for concurrent use.
{{< /callout >}}

## Putting It All Together

A minimal FHIR facade that combines several patterns:

```go
import (
    "fmt"
    "io"
    "net/http"

    "github.com/gofhir/models/r4"
)

func main() {
    mux := http.NewServeMux()

    // Accept any FHIR resource via POST
    mux.HandleFunc("/fhir/", func(w http.ResponseWriter, req *http.Request) {
        body, _ := io.ReadAll(req.Body)

        // Parse content type to determine format
        contentType := req.Header.Get("Content-Type")

        var resource r4.Resource
        var err error

        switch contentType {
        case "application/fhir+xml":
            resource, err = r4.UnmarshalResourceXML(body)
        default:
            resource, err = r4.UnmarshalResource(body)
        }

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Respond in the requested format
        accept := req.Header.Get("Accept")
        switch accept {
        case "application/fhir+xml":
            data, _ := r4.MarshalResourceXMLIndent(resource, "", "  ")
            w.Header().Set("Content-Type", "application/fhir+xml")
            w.Write(data)
        default:
            data, _ := r4.Marshal(resource)
            w.Header().Set("Content-Type", "application/fhir+json")
            w.Write(data)
        }
    })

    fmt.Println("FHIR server listening on :8080")
    http.ListenAndServe(":8080", mux)
}
```
