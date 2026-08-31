package conformance

// Every resource type, empty, through all four serialization paths.
//
// The corpus exercises documents that are complete, because published examples
// are. This exercises the opposite: a resource straight from the factory, with
// nothing set. That is the shape a library user starts from, and the shape that
// exposed `"code":{}` — a required complex field emitting an empty object
// because the Go type could not represent absence.
//
// It also guards the change that fixed it. Making ~810 fields pointers means
// anything that dereferences one without a nil check panics on exactly this
// input, and a panic in a marshaler takes down the caller's process.

import (
	"encoding/json"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

// version bundles the per-package entry points this test needs. codecs() cannot
// be reused: it deals in `any` and exposes no factory or type listing.
type version struct {
	name        string
	types       []string
	newFn       func(string) (any, error)
	marshalJSON func(any) ([]byte, error)
	marshalXML  func(any) ([]byte, error)
	parseXML    func([]byte) (any, error)
}

func versions() []version {
	return []version{
		{
			name:        "r4",
			types:       r4.AllResourceTypes(),
			newFn:       func(s string) (any, error) { return r4.NewResource(s) },
			marshalJSON: r4.Marshal,
			marshalXML:  func(v any) ([]byte, error) { return r4.MarshalResourceXML(v.(r4.Resource)) },
			parseXML:    func(b []byte) (any, error) { return r4.UnmarshalResourceXML(b) },
		},
		{
			name:        "r4b",
			types:       r4b.AllResourceTypes(),
			newFn:       func(s string) (any, error) { return r4b.NewResource(s) },
			marshalJSON: r4b.Marshal,
			marshalXML:  func(v any) ([]byte, error) { return r4b.MarshalResourceXML(v.(r4b.Resource)) },
			parseXML:    func(b []byte) (any, error) { return r4b.UnmarshalResourceXML(b) },
		},
		{
			name:        "r5",
			types:       r5.AllResourceTypes(),
			newFn:       func(s string) (any, error) { return r5.NewResource(s) },
			marshalJSON: r5.Marshal,
			marshalXML:  func(v any) ([]byte, error) { return r5.MarshalResourceXML(v.(r5.Resource)) },
			parseXML:    func(b []byte) (any, error) { return r5.UnmarshalResourceXML(b) },
		},
	}
}

func TestEveryEmptyResourceSerializesCleanly(t *testing.T) {
	for _, v := range versions() {
		t.Run(v.name, func(t *testing.T) {
			if len(v.types) < 100 {
				t.Fatalf("only %d resource types listed; the registry is not being seen", len(v.types))
			}

			for _, name := range v.types {
				res, err := v.newFn(name)
				if err != nil {
					t.Errorf("%s: NewResource: %v", name, err)
					continue
				}

				// A panic here is the failure mode worth catching: it would take
				// the caller's process down, so it must not escape as a crash of
				// the whole test binary either.
				func() {
					defer func() {
						if p := recover(); p != nil {
							t.Errorf("%s: panic while serializing an empty resource: %v", name, p)
						}
					}()

					out, err := v.marshalJSON(res)
					if err != nil {
						t.Errorf("%s: marshal JSON: %v", name, err)
						return
					}

					var fields map[string]any
					if err = json.Unmarshal(out, &fields); err != nil {
						t.Errorf("%s: output is not valid JSON: %v", name, err)
						return
					}
					// An empty object in place of an element violates ele-1: an
					// element must have a value or children. Nothing was set, so
					// nothing but resourceType should be present.
					for key, val := range fields {
						if obj, ok := val.(map[string]any); ok && len(obj) == 0 {
							t.Errorf("%s: emits an empty object for %q, which is invalid FHIR: %s", name, key, out)
						}
					}

					x, err := v.marshalXML(res)
					if err != nil {
						t.Errorf("%s: marshal XML: %v", name, err)
						return
					}
					if _, err := v.parseXML(x); err != nil {
						t.Errorf("%s: our own XML does not parse back: %v", name, err)
					}
				}()
			}

			t.Logf("%d resource types serialized empty, JSON and XML", len(v.types))
		})
	}
}
