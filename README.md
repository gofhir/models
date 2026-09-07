# FHIR Types for Go

Go types for **FHIR R4, R4B and R5** — generated from the published
StructureDefinitions, with fluent builders, conformant JSON and XML, and a
conformance corpus that round-trips every example HL7 publishes.

```go
patient := r4.NewPatientBuilder().
    SetId("p1").
    SetGender(r4.AdministrativeGenderFemale).
    AddName(r4.NewHumanNameBuilder().
        SetFamily("Smith").
        AddGiven("Jane").
        Build()).
    Build()

data, _ := json.Marshal(patient)
// {"resourceType":"Patient","id":"p1","name":[{"family":"Smith","given":["Jane"]}],"gender":"female"}
```

## Install

```bash
go get github.com/gofhir/models/r4/v2     # FHIR R4  (4.0.1)
go get github.com/gofhir/models/r4b/v2    # FHIR R4B (4.3.0)
go get github.com/gofhir/models/r5/v2     # FHIR R5  (5.0.0)
```

Each version is its own module, so a program that only speaks R4 does not carry
R5. They can be used together when a system has to bridge versions.

## What it does

**Round-trips the published corpus.** Every example HL7 publishes, read and
written back without losing or changing a value: **8757/8757 in JSON** and
**3653/3653 in XML**, across all three versions. The suite runs as a ratchet — it
fails on a regression *and* on unrecorded progress — so the number cannot quietly
drift.

**Fluent builders all the way down.** 663 builders in R4, 681 in R4B, 834 in R5:
resources, datatypes, backbone elements, and the `_field` companions that carry
extensions on primitives. The chain does not stop partway and leave you writing
struct literals with hand-made pointers.

**Absent stays distinguishable from empty.** Optional elements are pointers,
because an omitted `Patient.active` is not the same statement as `active: false`.
`Ptr`, `PtrSlice`, `Val` and `First` keep that from being tedious.

**Decimals keep their precision.** `2.00` is not `2.0` in FHIR — the trailing
zero is significant — so decimals are a dedicated type rather than a `float64`.

**Extensions are findable.** `GetExtensionByURL` on every type that can hold one,
including `Extension` itself, which is how a complex extension is walked.

**A resource from a newer server stays readable.** A `resourceType` this version
does not define is preserved verbatim rather than failing the document, in JSON
and XML alike — so an R5 server's answer does not break an R4 client over one
unrecognised entry.

**Enums know what they are.** `Display()`, `System()` and `Coding()` on 206 code
system types in R4, so building a `CodeableConcept` does not mean copying system
URLs by hand.

## What it does not do

**It does not validate.** These are types, not a conformance engine: the structs
will happily hold a document FHIR would reject. Validation lives in
[`gofhir/validator`](https://github.com/gofhir/validator), which works on the raw
document and is the right place for that question.

**It does not talk to a server.** No client, no search, no terminology service.

## Usage

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4/v2"
)

func main() {
    // Read anything, including a Bundle whose entries are different types.
    resource, err := r4.UnmarshalResource(data)
    if err != nil {
        panic(err)
    }

    switch res := resource.(type) {
    case *r4.Patient:
        fmt.Println("patient", r4.Val(res.Id))
    case *r4.Observation:
        fmt.Println("observation", res.Status.Display())
    case *r4.UnknownResource:
        fmt.Println("a type this version does not model:", res.Type)
    }

    // Write it back.
    out, err := json.Marshal(resource)
    fmt.Println(string(out), err)
}
```

XML works the same way, through `MarshalResourceXML` and `UnmarshalResourceXML`.

## Documentation

Full documentation, in English and Spanish, at
**[gofhir.github.io/models](https://gofhir.github.io/models/)** — including the
[v1 to v2 migration guide](https://gofhir.github.io/models/docs/migration/v1-to-v2/)
with the old→new mapping for every renamed type.

API reference on [pkg.go.dev](https://pkg.go.dev/github.com/gofhir/models/r4/v2).

## Versions

The import path carries the `/v2` suffix Go requires; the package keeps its FHIR
name:

```go
import "github.com/gofhir/models/r4/v2"

var p r4.Patient   // the package is r4, not v2
```

> `v2.0.0` cannot be installed — the Go checksum database holds a different hash
> for that tag than the repository produces, and that database is append-only, so
> it cannot be corrected. **v2.1.0 and later are unaffected**, and plain
> `go get .../v2` resolves to the latest.

## Contributing

The Go files are generated: edit the templates in `internal/codegen`, not the
output. `go run cmd/generator/main.go r4` regenerates one version. The conformance
corpus is fetched separately with `scripts/fetch-examples.sh`.

## License

See [LICENSE](LICENSE).
