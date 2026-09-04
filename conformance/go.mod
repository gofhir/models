module github.com/gofhir/models/conformance

go 1.26

require (
	github.com/gofhir/models/r4/v2 v2.0.0
	github.com/gofhir/models/r4b/v2 v2.0.0
	github.com/gofhir/models/r5/v2 v2.0.0
)

replace github.com/gofhir/models/r4/v2 => ../r4

replace github.com/gofhir/models/r4b/v2 => ../r4b

replace github.com/gofhir/models/r5/v2 => ../r5
