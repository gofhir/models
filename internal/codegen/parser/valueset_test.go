package parser

// Composition is the case the registry used to ignore: an include may name other
// ValueSets instead of listing codes, and the result is their contents.
//
// These are built by hand rather than taken from the specs. The cycle guard has
// no published example — no HL7 ValueSet includes itself — so the only way to
// know it works is to write one.

import (
	"encoding/json"
	"testing"
)

// bundleOf builds the JSON bundle the registry loads, from resources given as
// literals. Order is deliberate in several tests: a reference may point at a
// ValueSet that appears later.
func bundleOf(t *testing.T, resources ...string) []byte {
	t.Helper()
	entries := make([]json.RawMessage, 0, len(resources))
	for _, r := range resources {
		entries = append(entries, json.RawMessage(r))
	}
	b, err := json.Marshal(map[string]any{
		"resourceType": "Bundle",
		"entry": func() []map[string]any {
			out := make([]map[string]any, 0, len(entries))
			for _, e := range entries {
				out = append(out, map[string]any{"resource": e})
			}
			return out
		}(),
	})
	if err != nil {
		t.Fatalf("building the bundle: %v", err)
	}
	return b
}

func loadRegistry(t *testing.T, bundle []byte) *ValueSetRegistry {
	t.Helper()
	r := NewValueSetRegistry()
	if err := r.LoadFromBundle(bundle); err != nil {
		t.Fatalf("loading: %v", err)
	}
	return r
}

func codesOf(vs *ParsedValueSet) []string {
	if vs == nil {
		return nil
	}
	out := make([]string, 0, len(vs.Codes))
	for _, c := range vs.Codes {
		out = append(out, c.Code)
	}
	return out
}

func TestComposedValueSetTakesInTheOneItIncludes(t *testing.T) {
	// The shape that broke R5: one include names another ValueSet, a second
	// include names a CodeSystem. Only the second used to resolve.
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/composed","name":"Composed",
		  "compose":{"include":[
		    {"valueSet":["http://x/parts"]},
		    {"system":"http://x/extra"}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/parts","name":"Parts",
		  "compose":{"include":[{"system":"http://x/base","concept":[
		    {"code":"a"},{"code":"b"}]}]}}`,
		`{"resourceType":"CodeSystem","url":"http://x/extra",
		  "concept":[{"code":"c"}]}`,
	)

	got := codesOf(loadRegistry(t, bundle).Get("http://x/composed"))
	want := map[string]bool{"a": true, "b": true, "c": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want a, b and c", got)
	}
	for _, c := range got {
		if !want[c] {
			t.Errorf("unexpected code %q in %v", c, got)
		}
	}
}

func TestCompositionIsFollowedThroughSeveralLevels(t *testing.T) {
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/top","name":"Top",
		  "compose":{"include":[{"valueSet":["http://x/middle"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/middle","name":"Middle",
		  "compose":{"include":[{"valueSet":["http://x/bottom"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/bottom","name":"Bottom",
		  "compose":{"include":[{"system":"http://x/s","concept":[{"code":"deep"}]}]}}`,
	)

	if got := codesOf(loadRegistry(t, bundle).Get("http://x/top")); len(got) != 1 || got[0] != "deep" {
		t.Errorf("got %v, want [deep]", got)
	}
}

func TestAReferenceIsFollowedEvenWhenItsTargetComesLater(t *testing.T) {
	// Bundle order does not put a target before the ValueSet that references it,
	// which is why resolution cannot happen while loading. The referrer is first
	// here on purpose.
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/first","name":"First",
		  "compose":{"include":[{"valueSet":["http://x/second"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/second","name":"Second",
		  "compose":{"include":[{"system":"http://x/s","concept":[{"code":"late"}]}]}}`,
	)

	if got := codesOf(loadRegistry(t, bundle).Get("http://x/first")); len(got) != 1 || got[0] != "late" {
		t.Errorf("got %v, want [late]", got)
	}
}

func TestAVersionedReferenceResolves(t *testing.T) {
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/pinned","name":"Pinned",
		  "compose":{"include":[{"valueSet":["http://x/target|4.0.1"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/target","name":"Target",
		  "compose":{"include":[{"system":"http://x/s","concept":[{"code":"v"}]}]}}`,
	)

	if got := codesOf(loadRegistry(t, bundle).Get("http://x/pinned")); len(got) != 1 || got[0] != "v" {
		t.Errorf("got %v, want [v]", got)
	}
}

func TestACycleTerminates(t *testing.T) {
	// Nothing HL7 publishes has one. The guard exists because a composed ValueSet
	// is a graph, and without it this test would not return — which is the only
	// way to find out whether the guard is right.
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/a","name":"A",
		  "compose":{"include":[
		    {"system":"http://x/s","concept":[{"code":"fromA"}]},
		    {"valueSet":["http://x/b"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/b","name":"B",
		  "compose":{"include":[
		    {"system":"http://x/s","concept":[{"code":"fromB"}]},
		    {"valueSet":["http://x/a"]}]}}`,
	)

	got := codesOf(loadRegistry(t, bundle).Get("http://x/a"))
	if len(got) != 2 {
		t.Fatalf("got %v, want both codes exactly once", got)
	}
	seen := map[string]bool{}
	for _, c := range got {
		if seen[c] {
			t.Errorf("%q appears twice: %v", c, got)
		}
		seen[c] = true
	}
	if !seen["fromA"] || !seen["fromB"] {
		t.Errorf("got %v, want fromA and fromB", got)
	}
}

func TestASelfReferenceTerminates(t *testing.T) {
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/self","name":"Self",
		  "compose":{"include":[
		    {"system":"http://x/s","concept":[{"code":"only"}]},
		    {"valueSet":["http://x/self"]}]}}`,
	)

	if got := codesOf(loadRegistry(t, bundle).Get("http://x/self")); len(got) != 1 || got[0] != "only" {
		t.Errorf("got %v, want [only]", got)
	}
}

func TestARepeatedCodeIsKeptOnce(t *testing.T) {
	// One code becomes one Go constant, and the name comes from the code alone.
	// Composition makes duplicates ordinary: a set that includes another may
	// restate a code it already has, and two constants of one name do not
	// compile. The first occurrence wins, so the code keeps the system it was
	// first seen under.
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/dup","name":"Dup",
		  "compose":{"include":[
		    {"system":"http://x/first","concept":[{"code":"shared","display":"from first"}]},
		    {"valueSet":["http://x/other"]}]}}`,
		`{"resourceType":"ValueSet","url":"http://x/other","name":"Other",
		  "compose":{"include":[{"system":"http://x/second","concept":[
		    {"code":"shared","display":"from second"},{"code":"unique"}]}]}}`,
	)

	vs := loadRegistry(t, bundle).Get("http://x/dup")
	got := codesOf(vs)
	if len(got) != 2 {
		t.Fatalf("got %v, want shared once and unique", got)
	}
	for _, c := range vs.Codes {
		if c.Code == "shared" && c.System != "http://x/first" {
			t.Errorf("shared kept the system %q; the first occurrence should win", c.System)
		}
	}
}

func TestAMissingReferenceIsSkippedRatherThanFatal(t *testing.T) {
	// A bundle need not carry every ValueSet a composition names. Dropping what
	// cannot be resolved is right; failing the whole load is not.
	bundle := bundleOf(t,
		`{"resourceType":"ValueSet","url":"http://x/partial","name":"Partial",
		  "compose":{"include":[
		    {"system":"http://x/s","concept":[{"code":"present"}]},
		    {"valueSet":["http://x/nowhere"]}]}}`,
	)

	if got := codesOf(loadRegistry(t, bundle).Get("http://x/partial")); len(got) != 1 || got[0] != "present" {
		t.Errorf("got %v, want [present]", got)
	}
}
