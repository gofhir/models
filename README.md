# FHIR Types for Go

Go structs for FHIR R4, R4B, and R5 resources.

## Installation

```bash
# For FHIR R4
go get github.com/gofhir/models/r4/v2

# For FHIR R4B
go get github.com/gofhir/models/r4b/v2

# For FHIR R5
go get github.com/gofhir/models/r5/v2
```

## Usage

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4/v2"
)

func main() {
    patient := r4.Patient{
        ResourceType: "Patient",
        Id:           r4.Ptr("123"),
        Active:       r4.Ptr(true),
        Name: []r4.HumanName{
            {
                Family: r4.Ptr("Smith"),
                Given:  []string{"John"},
            },
        },
    }

    data, err := r4.MarshalIndent(patient, "", "  ")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
}
```

Optional elements are pointers, so an absent value stays distinguishable from a
present-but-empty one — `Ptr`, `Val` and `First` cover the boilerplate that
follows from that.

Use `r4.Marshal` rather than `json.Marshal`: the standard encoder escapes `<` and
`>`, which mangles the XHTML in `text.div`.

> **Validation.** These types do not validate. Cardinality, required bindings and
> FHIRPath invariants are not checked — a `Patient` with no `id` and a gender of
> `"banana"` marshals happily. See
> [gofhir/validator](https://github.com/gofhir/validator).

## Development

The Go types are generated from the FHIR specifications, which are not committed
(~143 MB). After cloning, download them before running the generator:

```bash
./scripts/fetch-specs.sh          # download and verify all versions
cd cmd/generator && go run . r4   # regenerate one version
```

Hashes and source URLs are pinned in [specs.lock](specs.lock). See the
[contributing guide](https://gofhir.github.io/models/docs/contributing/) for the
full workflow.

## Related Projects

- [gofhir/fhirpath](https://github.com/gofhir/fhirpath) - FHIRPath evaluator for Go
- [gofhir/validator](https://github.com/gofhir/validator) - FHIR resource validator for Go

## License

MIT License - see [LICENSE](LICENSE) for details.
