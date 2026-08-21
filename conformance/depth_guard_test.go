package conformance

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// nestedContained builds a document whose resource nesting depth is exactly depth:
// depth-1 wrappers plus the innermost resource. This is the shape that made
// deserialization quadratic, since every level re-reads its own subtree.
func nestedContained(depth int) []byte {
	if depth < 1 {
		panic("depth must be at least 1")
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		b.WriteString(`{"resourceType":"Patient","contained":[`)
	}
	b.WriteString(`{"resourceType":"Patient"}`)
	for i := 0; i < depth-1; i++ {
		b.WriteString(`]}`)
	}
	return []byte(b.String())
}

// nestedBundles builds Bundle-in-Bundle nesting to exactly depth resources, the
// other path where a resource contains a resource.
func nestedBundles(depth int) []byte {
	if depth < 1 {
		panic("depth must be at least 1")
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		b.WriteString(`{"resourceType":"Bundle","entry":[{"resource":`)
	}
	b.WriteString(`{"resourceType":"Patient","id":"leaf"}`)
	for i := 0; i < depth-1; i++ {
		b.WriteString(`}]}`)
	}
	return []byte(b.String())
}

// deepQuestionnaire nests Questionnaire.item structurally without nesting any
// resources. It must be accepted: the published examples reach 28 levels of this,
// and none of it costs a re-read.
func deepQuestionnaire(depth int) []byte {
	var b strings.Builder
	b.WriteString(`{"resourceType":"Questionnaire","status":"draft","item":[`)
	for i := 0; i < depth; i++ {
		b.WriteString(`{"linkId":"g","type":"group","item":[`)
	}
	b.WriteString(`{"linkId":"leaf","type":"string"}`)
	for i := 0; i < depth; i++ {
		b.WriteString(`]}`)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

// TestDepthGuardRejectsHostileNesting is the regression test for the quadratic
// blow-up: 160 KB of nested contained used to cost ~670 MB of heap and ~4 s of CPU.
//
// It asserts on time rather than only on the error, because the point of the guard
// is that rejection happens before deserialization, in one linear pass.
func TestDepthGuardRejectsHostileNesting(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			limit := c.getMaxDepth()
			if limit <= 0 {
				t.Fatalf("MaxResourceDepth is %d; the guard is disabled by default", limit)
			}

			for _, tc := range []struct {
				name  string
				build func(int) []byte
			}{
				{"contained", nestedContained},
				{"bundles", nestedBundles},
			} {
				t.Run(tc.name, func(t *testing.T) {
					payload := tc.build(limit + 1000)

					start := time.Now()
					_, err := c.unmarshalJSON(payload)
					elapsed := time.Since(start)

					if err == nil {
						t.Fatalf("accepted %d levels of nesting with a limit of %d", limit+1000, limit)
					}
					if !errors.Is(err, c.errMaxDepth) {
						t.Errorf("rejected, but not via ErrMaxResourceDepth: %v", err)
					}
					// Generous ceiling: the scan is linear, so this only fails if
					// the document is being deserialized before being checked.
					if elapsed > 200*time.Millisecond {
						t.Errorf("rejection took %v for %d bytes; the guard should reject"+
							" before deserializing", elapsed, len(payload))
					}
					t.Logf("%d bytes rejected in %v", len(payload), elapsed.Round(time.Microsecond))
				})
			}
		})
	}
}

// reorderedContained nests resources with "contained" declared *before*
// "resourceType".
//
// JSON does not order members, so this is as valid as any other ordering — and it
// bypassed the first version of the guard completely, because that version marked
// an object as a resource at the moment it read the key. A depth-4000 payload was
// accepted and cost 7.5 seconds of CPU. The fix computes depth when each object
// closes, which is order-independent.
func reorderedContained(depth int) []byte {
	if depth < 1 {
		panic("depth must be at least 1")
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		b.WriteString(`{"contained":[`)
	}
	b.WriteString(`{"resourceType":"Patient"}`)
	for i := 0; i < depth-1; i++ {
		b.WriteString(`],"resourceType":"Patient"}`)
	}
	return []byte(b.String())
}

// escapedKeyContained spells the member name with a unicode escape.
// encoding/json decodes "resourceType" to resourceType, so the document
// deserializes normally — but a raw byte comparison against the literal misses it.
func escapedKeyContained(depth int) []byte {
	if depth < 1 {
		panic("depth must be at least 1")
	}
	const escaped = `"resourceType"`
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		b.WriteString(`{` + escaped + `:"Patient","contained":[`)
	}
	b.WriteString(`{` + escaped + `:"Patient"}`)
	for i := 0; i < depth-1; i++ {
		b.WriteString(`]}`)
	}
	return []byte(b.String())
}

// TestDepthGuardResistsEvasion covers the ways a payload can hide its nesting from
// a naive scanner. Both of these were live bypasses in the first implementation,
// and both were invisible to the tests above because those builders emit
// "resourceType" first and unescaped — the shape a well-behaved encoder produces,
// not the shape an attacker chooses.
func TestDepthGuardResistsEvasion(t *testing.T) {
	evasions := []struct {
		name  string
		build func(int) []byte
	}{
		{"contained before resourceType", reorderedContained},
		{"escaped member name", escapedKeyContained},
		{"both at once", func(d int) []byte {
			if d < 1 {
				panic("depth must be at least 1")
			}
			const escaped = `"resourceType"`
			var b strings.Builder
			for i := 0; i < d-1; i++ {
				b.WriteString(`{"contained":[`)
			}
			b.WriteString(`{` + escaped + `:"Patient"}`)
			for i := 0; i < d-1; i++ {
				b.WriteString(`],` + escaped + `:"Patient"}`)
			}
			return []byte(b.String())
		}},
	}

	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			limit := c.getMaxDepth()
			for _, ev := range evasions {
				t.Run(ev.name, func(t *testing.T) {
					payload := ev.build(limit + 1000)

					start := time.Now()
					_, err := c.unmarshalJSON(payload)
					elapsed := time.Since(start)

					if err == nil {
						t.Fatalf("accepted %d levels disguised as %q (%d bytes, %v)",
							limit+1000, ev.name, len(payload), elapsed)
					}
					if !errors.Is(err, c.errMaxDepth) {
						t.Errorf("rejected, but not via ErrMaxResourceDepth: %v", err)
					}
					if elapsed > 200*time.Millisecond {
						t.Errorf("took %v to reject: the evasion delayed the guard"+
							" into deserializing", elapsed)
					}
				})
			}

			// The same tricks must not cause false rejections at legal depths.
			for _, ev := range evasions {
				if _, err := c.unmarshalJSON(ev.build(3)); err != nil {
					t.Errorf("%s: rejected a legal depth of 3: %v", ev.name, err)
				}
			}
		})
	}
}

// TestDepthGuardAcceptsLegitimateDocuments is the other half, and the reason the
// guard counts resources instead of JSON braces: a limit on braces would have to
// sit above 28 to accommodate real Questionnaires, which would leave the expensive
// case barely constrained.
func TestDepthGuardAcceptsLegitimateDocuments(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			// Deepest resource nesting in the published corpora is 3.
			if _, err := c.unmarshalJSON(nestedBundles(3)); err != nil {
				t.Errorf("rejected 3 levels of Bundle nesting, which occurs in the"+
					" published examples: %v", err)
			}

			// Structural nesting far beyond anything published, but zero nested
			// resources.
			if _, err := c.unmarshalJSON(deepQuestionnaire(100)); err != nil {
				t.Errorf("rejected a Questionnaire with 100 levels of item nesting;"+
					" structural depth must not count toward the resource limit: %v", err)
			}
		})
	}
}

// TestDepthGuardBoundary pins the limit itself, so a change to MaxResourceDepth is
// a deliberate act rather than a side effect.
func TestDepthGuardBoundary(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			limit := c.getMaxDepth()

			if _, err := c.unmarshalJSON(nestedContained(limit)); err != nil {
				t.Errorf("rejected exactly %d levels, which is the limit: %v", limit, err)
			}
			if _, err := c.unmarshalJSON(nestedContained(limit + 1)); err == nil {
				t.Errorf("accepted %d levels, one past the limit of %d", limit+1, limit)
			}
		})
	}
}

// TestDepthGuardConfigurable checks the escape hatch: a caller with unusual data
// can raise or disable the limit.
func TestDepthGuardConfigurable(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			original := c.getMaxDepth()
			t.Cleanup(func() { c.setMaxDepth(original) })

			payload := nestedContained(original + 5)

			if _, err := c.unmarshalJSON(payload); err == nil {
				t.Fatal("expected rejection at the default limit")
			}

			c.setMaxDepth(original + 100)
			if _, err := c.unmarshalJSON(payload); err != nil {
				t.Errorf("raising the limit did not allow the document: %v", err)
			}

			c.setMaxDepth(0) // disabled
			if _, err := c.unmarshalJSON(payload); err != nil {
				t.Errorf("disabling the guard did not allow the document: %v", err)
			}
		})
	}
}

// BenchmarkDepthGuardOverhead measures what the scan costs on a document shaped
// like real traffic. The guard is only defensible if it is close to free here.
func BenchmarkDepthGuardOverhead(b *testing.B) {
	c := codecs()[0] // r4

	// A Bundle of 50 Patients, roughly the shape of a search result.
	var sb strings.Builder
	sb.WriteString(`{"resourceType":"Bundle","type":"searchset","entry":[`)
	for i := 0; i < 50; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"resource":{"resourceType":"Patient","id":"p","active":true,` +
			`"name":[{"family":"Smith","given":["John","Robert"]}],` +
			`"telecom":[{"system":"phone","value":"555-0100"}]}}`)
	}
	sb.WriteString(`]}`)
	payload := []byte(sb.String())

	original := c.getMaxDepth()
	defer c.setMaxDepth(original)

	b.Run("guard-on", func(b *testing.B) {
		c.setMaxDepth(original)
		b.SetBytes(int64(len(payload)))
		for i := 0; i < b.N; i++ {
			if _, err := c.unmarshalJSON(payload); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("guard-off", func(b *testing.B) {
		c.setMaxDepth(0)
		b.SetBytes(int64(len(payload)))
		for i := 0; i < b.N; i++ {
			if _, err := c.unmarshalJSON(payload); err != nil {
				b.Fatal(err)
			}
		}
	})
}
