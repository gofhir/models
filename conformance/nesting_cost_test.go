package conformance

// The depth guard exists because deserializing a nested resource re-reads its
// subtree: the dispatcher extracts it as raw JSON and hands it to a fresh
// json.Unmarshal, which walks it again. Cost is size times depth, so an unbounded
// document turns a small request into an arbitrarily expensive one.
//
// TestDepthGuardRejectsHostileNesting already covers that such a document is
// refused. What it cannot show is that refusing it is *cheap* — a guard that
// rejects only after doing the quadratic work would pass every correctness test in
// the suite while leaving the denial of service exactly where it was.
//
// So this measures. The plan was explicit that the reproducer must be written
// against nesting depth rather than wide arrays, because a wide array is linear:
//
//	flat  n=128000 -> 182 ms    a test written this way passes against unfixed code
//	nested d=2000  -> 999 ms    the blowup is here
//
// Timing in a test is a blunt instrument, so the numbers here are the measured
// ones rather than estimates. At 4,000 levels, with MaxResourceDepth raised so the
// guard does not fire, the document takes 3.8 s; guarded, it is refused in about
// 170 µs. The budget sits between those, far enough from both that neither noise
// nor a slow machine reaches it.

import (
	"strings"
	"testing"
	"time"

	"github.com/gofhir/models/r4/v2"
)

func TestHostileNestingIsRefusedCheaply(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-based")
	}

	// 4,000 levels. Unguarded this re-reads its own subtree once per level; the
	// same shape at 2,000 was measured at 999 ms, and the cost is quadratic, so
	// this one is minutes of CPU. Guarded it is a single scan.
	payload := nestedContained(4000)
	t.Logf("payload: %d KB, 4000 levels", len(payload)/1024)

	start := time.Now()
	_, err := r4.UnmarshalResource(payload)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("a 4000-level document was accepted; the depth guard is not running")
	}
	if !strings.Contains(err.Error(), "depth") && !strings.Contains(err.Error(), "nest") {
		t.Errorf("rejected, but not by the depth guard: %v", err)
	}

	// Measured: ~170 µs guarded, 3.8 s if the guard does not fire. Half a second
	// is 3000x above the first and 7x below the second, so it separates them
	// without turning a slow machine into a failure.
	//
	// This only fires when the document was refused, so what it catches is a guard
	// that rejects *after* doing the work rather than before — the one failure
	// mode that leaves every correctness test in the suite passing.
	const budget = 500 * time.Millisecond
	if elapsed > budget {
		t.Errorf("refusing the document took %v, over the %v budget — the guard is"+
			" rejecting after doing the work, not before", elapsed, budget)
	}
	t.Logf("refused in %v", elapsed)
}

func TestWideDocumentsStayLinear(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-based")
	}

	// The counterpart, and the reason the plan insisted the reproducer be written
	// against depth: width is linear and always was. A test built on a wide array
	// passes against unfixed code, which is why one written that way would have
	// been worthless as a regression test for the denial of service.
	//
	// So this measures rather than asserting it in a comment: cost per entry, at
	// two widths. Flat means the Bundle is read once however many entries it has.
	wideBundle := func(entries int) []byte {
		var b strings.Builder
		b.WriteString(`{"resourceType":"Bundle","entry":[`)
		for i := 0; i < entries; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(`{"resource":{"resourceType":"Patient","id":"p"}}`)
		}
		b.WriteString(`]}`)
		return []byte(b.String())
	}

	perEntry := func(entries int) float64 {
		payload := wideBundle(entries)
		best := time.Hour
		for i := 0; i < 5; i++ {
			start := time.Now()
			res, err := r4.UnmarshalResource(payload)
			if err != nil {
				t.Fatalf("a wide but shallow Bundle must be accepted: %v", err)
			}
			bundle, ok := res.(*r4.Bundle)
			if !ok {
				t.Fatalf("got %T, want *r4.Bundle", res)
			}
			if len(bundle.Entry) != entries {
				t.Fatalf("got %d entries, want %d", len(bundle.Entry), entries)
			}
			if d := time.Since(start); d < best {
				best = d
			}
		}
		return float64(best.Nanoseconds()) / float64(entries)
	}

	narrow := perEntry(5000)
	wide := perEntry(20000)
	t.Logf("5000 entries: %.0f ns/entry; 20000 entries: %.0f ns/entry", narrow, wide)

	if narrow <= 0 {
		t.Skip("too fast to time on this machine")
	}
	// Four times the width. Linear means the per-entry cost holds; the ceiling is
	// loose because allocation and GC behavior differ between the two sizes, and
	// the claim being tested is "linear, not quadratic" rather than a precise
	// constant.
	if ratio := wide / narrow; ratio > 2 {
		t.Errorf("four times the width cost %.1fx more per entry; width was supposed"+
			" to be the linear axis", ratio)
	}
}

func TestNestingCostPerByteStaysFlat(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-based")
	}

	// The denial of service was quadratic: cost grew with size *times* depth,
	// because each level re-read its own subtree. The guard replaces that with a
	// single scan, and the property that says so is cost per byte — flat if the
	// document is read once, rising with depth if it is read repeatedly.
	//
	// Comparing total time would not distinguish the two: a 10x bigger document
	// legitimately costs 10x more to read once. Per byte is what separates reading
	// it once from reading it once per level.
	//
	// The scan deliberately cannot stop early. Depth is computed when an object
	// closes, not when it opens, because JSON does not order members and an object
	// may declare "contained" before "resourceType" — an earlier version marked
	// objects on entry and was bypassed by reordering the payload. Reading to the
	// end is the price of that, and it is linear.
	perByte := func(depth int) (float64, int) {
		payload := nestedContained(depth)
		best := time.Hour
		for i := 0; i < 7; i++ { // fastest run: the least noisy
			start := time.Now()
			if _, err := r4.UnmarshalResource(payload); err == nil {
				t.Fatalf("depth %d was accepted; it is far past the limit", depth)
			}
			if d := time.Since(start); d < best {
				best = d
			}
		}
		return float64(best.Nanoseconds()) / float64(len(payload)), len(payload)
	}

	shallow, shallowSize := perByte(400)
	deep, deepSize := perByte(4000)
	t.Logf("depth 400 (%d KB): %.2f ns/byte; depth 4000 (%d KB): %.2f ns/byte",
		shallowSize/1024, shallow, deepSize/1024, deep)

	if shallow <= 0 {
		t.Skip("too fast to time on this machine")
	}
	// Measured on the unguarded path, cost per byte climbs 943 -> 1280 -> 2191 ->
	// 3710 ns as depth goes 100 -> 200 -> 400 -> 800: 3.9x for 8x the depth. Ten
	// times the depth is therefore worth something above 4x, and the guarded path
	// measures 1.05x. A ceiling of 3 sits between them with room on both sides.
	if ratio := deep / shallow; ratio > 3 {
		t.Errorf("ten times the nesting depth cost %.1fx more per byte; the document"+
			" is being read more than once and the quadratic blowup is back", ratio)
	}
}
