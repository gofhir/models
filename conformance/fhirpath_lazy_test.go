package conformance

// The FHIRPath model is built on first use rather than at package load.
//
// Its maps hold thousands of entries, and Go builds a map literal at run time
// rather than baking it into the binary — so as a package-level var, every
// program importing the package paid for all of it before main(), whether or not
// it ever evaluated a FHIRPath expression. Measured on r4: 1045 KB of heap after
// loading the package, against 281 KB once the model is lazy.
//
// sync.OnceValue keeps the accessor's signature and its concurrency behavior, so
// what these check is that the laziness is invisible: same instance every time,
// no race, and the data actually there when asked for.

import (
	"sync"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestFHIRPathModelIsASingleton(t *testing.T) {
	t.Run("r4", func(t *testing.T) {
		if a, b := r4.FHIRPathModel(), r4.FHIRPathModel(); a != b {
			t.Error("two calls returned different instances")
		}
	})
	t.Run("r4b", func(t *testing.T) {
		if a, b := r4b.FHIRPathModel(), r4b.FHIRPathModel(); a != b {
			t.Error("two calls returned different instances")
		}
	})
	t.Run("r5", func(t *testing.T) {
		if a, b := r5.FHIRPathModel(), r5.FHIRPathModel(); a != b {
			t.Error("two calls returned different instances")
		}
	})
}

// TestFHIRPathModelConcurrentFirstUse is the one that matters: the first call now
// does work, and several goroutines can make it at once. Run under -race.
func TestFHIRPathModelConcurrentFirstUse(t *testing.T) {
	const goroutines = 64

	var wg sync.WaitGroup
	seen := make([]*r4.FHIRPathModelData, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			seen[i] = r4.FHIRPathModel()
		}(i)
	}
	wg.Wait()

	for i, m := range seen {
		if m == nil {
			t.Fatalf("goroutine %d got nil", i)
		}
		if m != seen[0] {
			t.Fatalf("goroutine %d got a different instance", i)
		}
	}
}

// TestFHIRPathModelIsPopulated guards against the laziness returning an empty
// model — the failure that would make every FHIRPath lookup silently wrong
// rather than loudly broken.
func TestFHIRPathModelIsPopulated(t *testing.T) {
	for _, tt := range []struct {
		name string
		path string
		want string
	}{
		{"r4", "Patient.name", "HumanName"},
		{"r4b", "Patient.name", "HumanName"},
		{"r5", "Patient.name", "HumanName"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "r4":
				got = r4.FHIRPathModel().TypeOf(tt.path)
			case "r4b":
				got = r4b.FHIRPathModel().TypeOf(tt.path)
			case "r5":
				got = r5.FHIRPathModel().TypeOf(tt.path)
			}
			if got != tt.want {
				t.Errorf("TypeOf(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
