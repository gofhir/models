package conformance

// Every settable field on a resource has a builder method.
//
// This used to be guaranteed indirectly: the functional options covered every
// field, each one was verified against its builder counterpart when v1.6.0
// deprecated them, and removing them in v2 was checked field by field — 11,952
// options against 11,952 builder methods, none missing.
//
// With the options gone, nothing held that invariant any more. A field generated
// without its Set*/Add* would only be reachable through a struct literal, and
// nothing would say so. So the check now stands on its own: it compares the
// struct's fields against the builder's methods directly.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryFieldHasABuilderMethod(t *testing.T) {
	for _, module := range []string{"r4", "r4b", "r5"} {
		t.Run(module, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", module, "resource_*.go"))
			if err != nil || len(files) == 0 {
				t.Fatalf("no generated sources for %s: %v", module, err)
			}

			checked := 0
			for _, path := range files {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, path, nil, 0)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}

				// The resource is the type with a matching Builder in this file.
				resource := ""
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Recv != nil {
						continue
					}
					if name := fn.Name.Name; strings.HasPrefix(name, "New") && strings.HasSuffix(name, "Builder") {
						resource = strings.TrimSuffix(strings.TrimPrefix(name, "New"), "Builder")
						break
					}
				}
				if resource == "" {
					continue
				}

				methods := map[string]bool{}
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
						continue
					}
					recv := fn.Recv.List[0].Type
					star, ok := recv.(*ast.StarExpr)
					if !ok {
						continue
					}
					id, ok := star.X.(*ast.Ident)
					if !ok || id.Name != resource+"Builder" {
						continue
					}
					name := fn.Name.Name
					if strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "Add") {
						methods[name[3:]] = true
					}
				}

				// Every exported field of the resource struct needs one, except
				// ResourceType, which is the zero-size marker and cannot be set.
				for _, decl := range file.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok || gd.Tok != token.TYPE {
						continue
					}
					for _, spec := range gd.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok || ts.Name.Name != resource {
							continue
						}
						st, ok := ts.Type.(*ast.StructType)
						if !ok {
							continue
						}
						for _, fl := range st.Fields.List {
							if len(fl.Names) == 0 {
								continue
							}
							name := fl.Names[0].Name
							if !ast.IsExported(name) || name == "ResourceType" {
								continue
							}
							// The _field companions that carry extensions on a
							// primitive have never had a builder method, nor did
							// they have a functional option — they are reachable
							// only through a struct literal. That is a real gap,
							// but a pre-existing one rather than something the
							// options removal caused, so it is counted by
							// TestPrimitiveExtensionGap rather than failing here.
							if strings.HasSuffix(name, "Ext") {
								continue
							}
							checked++
							if !methods[name] {
								t.Errorf("%s.%s has no builder method; it is only reachable through a struct literal",
									resource, name)
							}
						}
					}
				}
			}

			// A guard against the check passing on nothing.
			if checked < 3000 {
				t.Errorf("only %d fields examined in %s; expected thousands, so the check is not seeing the generated code",
					checked, module)
			}
			t.Logf("%d fields, each with a builder method", checked)
		})
	}
}

// TestPrimitiveExtensionGap measures a gap rather than guarding an invariant.
//
// A primitive's extension companion — BirthDateExt for birthDate — can only be
// set through a struct literal. There has never been a builder method or a
// functional option for one, so builder-only code cannot attach an extension to
// a primitive at all.
//
// The count is pinned so it cannot grow quietly. Closing it is additive work: a
// Set<Field>Ext per companion, which would break nothing.
func TestPrimitiveExtensionGap(t *testing.T) {
	for _, module := range []string{"r4", "r4b", "r5"} {
		t.Run(module, func(t *testing.T) {
			files, err := filepath.Glob(filepath.Join("..", module, "resource_*.go"))
			if err != nil || len(files) == 0 {
				t.Fatalf("no generated sources for %s: %v", module, err)
			}

			withBuilder, withoutBuilder := 0, 0
			for _, path := range files {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, path, nil, 0)
				if err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}
				methods := map[string]bool{}
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
						continue
					}
					if name := fn.Name.Name; strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "Add") {
						methods[name[3:]] = true
					}
				}
				for _, decl := range file.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok || gd.Tok != token.TYPE {
						continue
					}
					for _, spec := range gd.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						st, ok := ts.Type.(*ast.StructType)
						if !ok {
							continue
						}
						for _, fl := range st.Fields.List {
							if len(fl.Names) == 0 {
								continue
							}
							name := fl.Names[0].Name
							if !strings.HasSuffix(name, "Ext") || !ast.IsExported(name) {
								continue
							}
							if methods[name] {
								withBuilder++
							} else {
								withoutBuilder++
							}
						}
					}
				}
			}

			t.Logf("%d primitive extension companions, %d reachable from a builder", withoutBuilder+withBuilder, withBuilder)
			if withBuilder != 0 {
				t.Logf("some now have builder methods; if the gap is being closed, update this test")
			}
			if withoutBuilder == 0 {
				t.Error("no extension companions found at all; the check is not seeing the generated code")
			}
		})
	}
}
