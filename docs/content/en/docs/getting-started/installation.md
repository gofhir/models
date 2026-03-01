---
title: "Installation"
linkTitle: "Installation"
description: "Install gofhir/models for FHIR R4, R4B, or R5 and configure your Go project."
weight: 1
---

## Prerequisites

- **Go 1.23** or later is required. You can check your version with:

```shell
go version
```

- Your project must use [Go modules](https://go.dev/ref/mod) (`go.mod`). If you do not have one yet, initialize it:

```shell
go mod init your-module-name
```

## Install

Each FHIR version is published as an independent Go module. Install only the version you need.

### FHIR R4 (4.0.1)

```shell
go get github.com/gofhir/models/r4
```

### FHIR R4B (4.3.0)

```shell
go get github.com/gofhir/models/r4b
```

### FHIR R5 (5.0.0)

```shell
go get github.com/gofhir/models/r5
```

### R4 Helpers (optional)

The helpers sub-package provides pre-built `CodeableConcept` values for common observation categories, condition categories, LOINC codes, and UCUM units:

```shell
go get github.com/gofhir/models/r4/helpers
```

## Import

After installing, import the package in your Go source files:

```go
import "github.com/gofhir/models/r4"
```

Or for R4B and R5:

```go
import "github.com/gofhir/models/r4b"
import "github.com/gofhir/models/r5"
```

All types, builders, functional options, code system constants, and serialization functions are exported from the version-specific package. There is no separate sub-package for builders or serialization -- everything is in one place.

## Using Multiple FHIR Versions

If your application needs to work with more than one FHIR version, you can import multiple packages in the same project. Use Go import aliases to avoid name collisions:

```go
import (
    r4 "github.com/gofhir/models/r4"
    r5 "github.com/gofhir/models/r5"
)

func main() {
    // R4 Patient
    patientR4 := r4.NewPatient(
        r4.WithPatientId("r4-patient"),
        r4.WithPatientActive(true),
    )

    // R5 Patient
    patientR5 := r5.NewPatient(
        r5.WithPatientId("r5-patient"),
        r5.WithPatientActive(true),
    )

    _, _ = r4.Marshal(patientR4)
    _, _ = r5.Marshal(patientR5)
}
```

Because each FHIR version is a separate Go module with its own `go.mod`, dependency versions are resolved independently and never conflict.

## Verifying the Installation

Create a simple test file to confirm everything is working:

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.NewPatient(
        r4.WithPatientId("test"),
    )
    data, err := r4.Marshal(patient)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
}
```

Run it:

```shell
go run main.go
```

Expected output:

```json
{"resourceType":"Patient","id":"test"}
```

## Next Steps

Continue to the [Quick Start](../quick-start) guide to see complete examples of creating, serializing, and deserializing FHIR resources.
