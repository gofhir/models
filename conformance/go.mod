module github.com/gofhir/models/conformance

go 1.23

require (
	github.com/gofhir/models/r4 v0.0.0
	github.com/gofhir/models/r4b v0.0.0
	github.com/gofhir/models/r5 v0.0.0
)

replace github.com/gofhir/models/r4 => ../r4

replace github.com/gofhir/models/r4b => ../r4b

replace github.com/gofhir/models/r5 => ../r5
