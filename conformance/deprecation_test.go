package conformance

// Every With* option carries a Deprecated line promising that a named builder
// method is its replacement, with "an identical signature and identical
// behavior". That promise is repeated 12,845 times across the three modules and
// is what makes the v1.6.0 deprecation actionable rather than noise — so it is
// checked here rather than trusted.
//
// The check parses the generated sources and, for every
// With<Resource><Field>(v T), requires <Resource>Builder.Set<Field> (or Add* for
// repeating fields) to take the same T and to perform the same operation on the
// same field. Both are emitted from the same three-way branch in
// resource_consolidated.go.tmpl; the failure mode this guards against is someone
// editing one arm of that branch and not the other, which would leave thousands
// of deprecation messages pointing at a replacement that no longer matches.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// signature is what has to agree between an option and its builder method: the
// parameter type, and what the body does with it.
type signature struct {
	param string // rendered parameter type
	op    string // "append", "addr" (&v) or "assign"
	field string // struct field written
}

func renderType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + renderType(t.X)
	case *ast.ArrayType:
		return "[]" + renderType(t.Elt)
	case *ast.SelectorExpr:
		return renderType(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + renderType(t.Key) + "]" + renderType(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + renderType(t.Elt)
	}
	// Anything unrendered compares unequal to itself only if the two sides render
	// differently, so an unknown shape cannot mask a real mismatch — but it would
	// hide a genuine difference between two unknown shapes. No generated
	// parameter type reaches here today.
	return "unrendered"
}

// classify reports what the single assignment in a body does, and to which
// field. Both shapes assign to a selector ending in the field name.
func classify(body *ast.BlockStmt) (op, field string) {
	ast.Inspect(body, func(n ast.Node) bool {
		a, ok := n.(*ast.AssignStmt)
		if !ok || len(a.Rhs) != 1 || len(a.Lhs) != 1 {
			return true
		}
		if sel, ok := a.Lhs[0].(*ast.SelectorExpr); ok {
			field = sel.Sel.Name
		}
		switch r := a.Rhs[0].(type) {
		case *ast.CallExpr:
			if id, ok := r.Fun.(*ast.Ident); ok && id.Name == "append" {
				op = "append"
			}
		case *ast.UnaryExpr:
			if r.Op == token.AND {
				op = "addr"
			}
		case *ast.Ident:
			if op == "" {
				op = "assign"
			}
		}
		return true
	})
	return op, field
}

func TestWithOptionsMatchTheirBuilderReplacements(t *testing.T) {
	for _, module := range []string{"r4", "r4b", "r5"} {
		t.Run(module, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", module, "resource_*.go"))
			if err != nil || len(files) == 0 {
				t.Fatalf("no generated sources found for %s: %v", module, err)
			}

			pairs := 0
			for _, path := range files {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}

				options := map[string]signature{}
				methods := map[string]signature{}

				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Body == nil || fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
						continue
					}
					param := renderType(fn.Type.Params.List[0].Type)
					op, field := classify(fn.Body)
					sig := signature{param: param, op: op, field: field}

					switch {
					// With<Resource><Field> — the resource comes from the
					// <Resource>Option return type, not from the name, so a
					// resource whose name contains "With" cannot confuse it.
					case fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "With"):
						if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
							continue
						}
						ret := renderType(fn.Type.Results.List[0].Type)
						resource := strings.TrimSuffix(ret, "Option")
						if resource == ret {
							continue // not an option constructor
						}
						name := strings.TrimPrefix(fn.Name.Name, "With")
						options[resource+"."+strings.TrimPrefix(name, resource)] = sig

					case fn.Recv != nil && (strings.HasPrefix(fn.Name.Name, "Set") || strings.HasPrefix(fn.Name.Name, "Add")):
						recv := renderType(fn.Recv.List[0].Type)
						resource := strings.TrimSuffix(strings.TrimPrefix(recv, "*"), "Builder")
						methods[resource+"."+fn.Name.Name[len("Set"):]] = sig
					}
				}

				for key, want := range options {
					pairs++
					got, ok := methods[key]
					if !ok {
						t.Errorf("%s: With option %s has no builder Set*/Add* counterpart, but its Deprecated line names one",
							filepath.Base(path), key)
						continue
					}
					if got != want {
						t.Errorf("%s: %s promises an identical replacement but they differ:\n  option:  param=%s op=%s field=%s\n  builder: param=%s op=%s field=%s",
							filepath.Base(path), key, want.param, want.op, want.field, got.param, got.op, got.field)
					}
				}
			}

			// A guard against the check silently passing on nothing — a glob that
			// matched files but a parse that found no options would otherwise look
			// like success.
			if pairs < 3000 {
				t.Errorf("only %d option/builder pairs examined in %s; expected thousands, so the check is not seeing the generated code",
					pairs, module)
			}
			t.Logf("%d option/builder pairs verified", pairs)
		})
	}
}

// TestDeprecatedMarksAreComplete pins the count of Deprecated markers, so
// removing them from the template — or adding a resource without them — is a
// test failure rather than a silent loss of the migration warning.
func TestDeprecatedMarksAreComplete(t *testing.T) {
	for _, module := range []string{"r4", "r4b", "r5"} {
		t.Run(module, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", module, "resource_*.go"))
			if err != nil || len(files) == 0 {
				t.Fatalf("no generated sources found for %s: %v", module, err)
			}

			var options, optionTypes, constructors int
			for _, path := range files {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}

				isDeprecated := func(doc *ast.CommentGroup) bool {
					if doc == nil {
						return false
					}
					for _, c := range doc.List {
						if strings.HasPrefix(strings.TrimPrefix(c.Text, "// "), "Deprecated:") {
							return true
						}
					}
					return false
				}

				for _, decl := range file.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						if d.Recv != nil {
							continue
						}
						name := d.Name.Name
						switch {
						case strings.HasPrefix(name, "With"):
							if d.Type.Results == nil || len(d.Type.Results.List) != 1 ||
								!strings.HasSuffix(renderType(d.Type.Results.List[0].Type), "Option") {
								continue
							}
							if !isDeprecated(d.Doc) {
								t.Errorf("%s: %s is not marked Deprecated", filepath.Base(path), name)
							}
							options++
						case strings.HasPrefix(name, "New") && !strings.HasSuffix(name, "Builder"):
							// New<Resource>(opts ...<Resource>Option)
							if d.Type.Params == nil || len(d.Type.Params.List) != 1 ||
								!strings.HasPrefix(renderType(d.Type.Params.List[0].Type), "...") {
								continue
							}
							if !isDeprecated(d.Doc) {
								t.Errorf("%s: %s is not marked Deprecated", filepath.Base(path), name)
							}
							constructors++
						}
					case *ast.GenDecl:
						for _, spec := range d.Specs {
							ts, ok := spec.(*ast.TypeSpec)
							if !ok || !strings.HasSuffix(ts.Name.Name, "Option") {
								continue
							}
							// The name suffix is not enough: FHIR has backbone
							// elements of its own called ...Option —
							// Questionnaire.item.answerOption and
							// PlanDefinition.actor.option — which generate structs
							// named QuestionnaireItemAnswerOption and
							// PlanDefinitionActorOption. A functional option type is
							// specifically a func.
							if _, isFunc := ts.Type.(*ast.FuncType); !isFunc {
								continue
							}
							// The doc comment sits on the GenDecl for a lone type.
							doc := ts.Doc
							if doc == nil {
								doc = d.Doc
							}
							if !isDeprecated(doc) {
								t.Errorf("%s: type %s is not marked Deprecated", filepath.Base(path), ts.Name.Name)
							}
							optionTypes++
						}
					}
				}
			}

			t.Logf("%d With* options, %d Option types, %d New* constructors — all marked", options, optionTypes, constructors)
			if options < 3000 || optionTypes < 100 || constructors < 100 {
				t.Errorf("counts too low (options=%d types=%d constructors=%d); the check is not seeing the generated code",
					options, optionTypes, constructors)
			}
			if optionTypes != constructors {
				t.Errorf("every Option type should have exactly one New* constructor: %d types vs %d constructors",
					optionTypes, constructors)
			}
		})
	}
}
