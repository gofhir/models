# FHIR Types for Go

Go structs for FHIR R4, R4B, and R5 resources.

## Installation

```bash
# For FHIR R4
go get github.com/gofhir/models/r4

# For FHIR R4B
go get github.com/gofhir/models/r4b

# For FHIR R5
go get github.com/gofhir/models/r5
```

## Usage

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.Patient{
        ResourceType: "Patient",
        Id:           r4.String("123"),
        Active:       r4.Boolean(true),
        Name: []r4.HumanName{
            {
                Family: r4.String("Smith"),
                Given:  []r4.String{r4.String("John")},
            },
        },
    }

    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))
}
```

## Related Projects

- [gofhir/fhirpath](https://github.com/gofhir/fhirpath) - FHIRPath evaluator for Go
- [gofhir/validator](https://github.com/gofhir/validator) - FHIR resource validator for Go

## License

MIT License - see [LICENSE](LICENSE) for details.
