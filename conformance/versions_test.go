// Package conformance round-trips the official FHIR example corpora through the
// generated packages.
//
// It is a separate module so the published r4/r4b/r5 modules carry neither the
// suite nor its dependencies, and so one runner covers all three versions instead
// of being copied three times.
package conformance

import (
	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

// codec adapts one generated package to the runner. The generated Resource types
// differ per package with no common interface across modules, so the adapters
// deal in `any` and the runner never needs to know which version it is driving.
type codec struct {
	name string

	unmarshalJSON func([]byte) (any, error)
	marshalJSON   func(any) ([]byte, error)

	unmarshalXML func([]byte) (any, error)
	marshalXML   func(any) ([]byte, error)

	// maxResourceDepth reads and writes the package's depth guard, so the guard
	// can be exercised and restored without the runner importing each package.
	getMaxDepth func() int
	setMaxDepth func(int)
	errMaxDepth error
}

// codecs is the set of versions under test, in a stable order.
func codecs() []codec {
	return []codec{
		{
			name: "r4",
			unmarshalJSON: func(b []byte) (any, error) {
				return r4.UnmarshalResource(b)
			},
			marshalJSON: r4.Marshal,
			unmarshalXML: func(b []byte) (any, error) {
				return r4.UnmarshalResourceXML(b)
			},
			marshalXML: func(v any) ([]byte, error) {
				res, ok := v.(r4.Resource)
				if !ok {
					return nil, errNotResource
				}
				return r4.MarshalResourceXML(res)
			},
			getMaxDepth: func() int { return r4.MaxResourceDepth },
			setMaxDepth: func(n int) { r4.MaxResourceDepth = n },
			errMaxDepth: r4.ErrMaxResourceDepth,
		},
		{
			name: "r4b",
			unmarshalJSON: func(b []byte) (any, error) {
				return r4b.UnmarshalResource(b)
			},
			marshalJSON: r4b.Marshal,
			unmarshalXML: func(b []byte) (any, error) {
				return r4b.UnmarshalResourceXML(b)
			},
			marshalXML: func(v any) ([]byte, error) {
				res, ok := v.(r4b.Resource)
				if !ok {
					return nil, errNotResource
				}
				return r4b.MarshalResourceXML(res)
			},
			getMaxDepth: func() int { return r4b.MaxResourceDepth },
			setMaxDepth: func(n int) { r4b.MaxResourceDepth = n },
			errMaxDepth: r4b.ErrMaxResourceDepth,
		},
		{
			name: "r5",
			unmarshalJSON: func(b []byte) (any, error) {
				return r5.UnmarshalResource(b)
			},
			marshalJSON: r5.Marshal,
			unmarshalXML: func(b []byte) (any, error) {
				return r5.UnmarshalResourceXML(b)
			},
			marshalXML: func(v any) ([]byte, error) {
				res, ok := v.(r5.Resource)
				if !ok {
					return nil, errNotResource
				}
				return r5.MarshalResourceXML(res)
			},
			getMaxDepth: func() int { return r5.MaxResourceDepth },
			setMaxDepth: func(n int) { r5.MaxResourceDepth = n },
			errMaxDepth: r5.ErrMaxResourceDepth,
		},
	}
}
