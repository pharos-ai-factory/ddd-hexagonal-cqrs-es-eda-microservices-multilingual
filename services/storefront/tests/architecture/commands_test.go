package architecture

import (
	"encoding/json"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// isAggregate recognises root behaviour rather than a list of mutator names.
func isAggregate(value types.Type) bool {
	if pointer, ok := value.(*types.Pointer); ok {
		value = pointer.Elem()
	}
	named, ok := value.(*types.Named)
	if !ok || named.Obj().Pkg() == nil || !strings.Contains(named.Obj().Pkg().Path(), "/contexts/") || !strings.HasSuffix(named.Obj().Pkg().Path(), "/domain") {
		return false
	}
	methods := types.NewMethodSet(types.NewPointer(named))
	hasSnapshot, hasEvents := false, false
	for i := 0; i < methods.Len(); i++ {
		hasSnapshot = hasSnapshot || methods.At(i).Obj().Name() == "Snapshot"
		hasEvents = hasEvents || methods.At(i).Obj().Name() == "Events"
	}
	return hasSnapshot && hasEvents
}
func mutationObject(object types.Object) bool {
	method, ok := object.(*types.Func)
	if !ok {
		return false
	}
	signature, ok := method.Type().(*types.Signature)
	if !ok {
		return false
	}
	if receiver := signature.Recv(); receiver != nil {
		if isAggregate(receiver.Type()) {
			return method.Name() != "Snapshot" && method.Name() != "Events"
		}
		if method.Name() == "Execute" {
			value := receiver.Type()
			if pointer, ok := value.(*types.Pointer); ok {
				value = pointer.Elem()
			}
			if named, ok := value.(*types.Named); ok {
				return named.Obj().Name() == "AggregateCommandPort" || named.Obj().Name() == "AggregateCommandStore"
			}
		}
	}
	if strings.HasPrefix(method.Name(), "Restore") {
		return false
	}
	for i := 0; i < signature.Results().Len(); i++ {
		if isAggregate(signature.Results().At(i).Type()) {
			return true
		}
	}
	return false
}
func violations(file *ast.File, info *types.Info, application bool) []token.Pos {
	var errors []token.Pos
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		allowed := false
		if ok && application && function.Name.Name == "Execute" && function.Recv != nil {
			receiver := function.Recv.List[0].Type
			if pointer, ok := receiver.(*ast.StarExpr); ok {
				receiver = pointer.X
			}
			if name, ok := receiver.(*ast.Ident); ok {
				allowed = strings.HasSuffix(name.Name, "CommandHandler")
			}
		}
		ast.Inspect(declaration, func(node ast.Node) bool {
			if name, ok := node.(*ast.Ident); ok && mutationObject(info.Uses[name]) && !allowed {
				errors = append(errors, node.Pos())
			}
			return true
		})
	}
	return errors
}

func TestProductionCommandBoundaries(t *testing.T) {
	root := filepath.Join("..", "..")
	command := exec.Command("go", "list", "-export", "-json", "-deps", "./...")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	type listedPackage struct {
		ImportPath, Dir, Export string
		GoFiles                 []string
		Standard                bool
	}
	var packages []listedPackage
	exports := map[string]string{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var p listedPackage
		if err := decoder.Decode(&p); err != nil {
			t.Fatal(err)
		}
		exports[p.ImportPath] = p.Export
		if !p.Standard && strings.Contains(p.ImportPath, "services/storefront/") {
			packages = append(packages, p)
		}
	}
	fileset := token.NewFileSet()
	imports := importer.ForCompiler(fileset, "gc", func(name string) (io.ReadCloser, error) { return os.Open(exports[name]) })
	for _, p := range packages {
		if strings.Contains(p.ImportPath, "/domain") || strings.Contains(p.ImportPath, "/generated") || strings.Contains(p.ImportPath, "/foundation/") {
			continue
		}
		var files []*ast.File
		for _, name := range p.GoFiles {
			file, err := parser.ParseFile(fileset, filepath.Join(p.Dir, name), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, file)
		}
		info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
		config := types.Config{Importer: imports}
		if _, err := config.Check(p.ImportPath, fileset, files, info); err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			for _, pos := range violations(file, info, strings.HasSuffix(p.ImportPath, "/application")) {
				t.Errorf("%s: aggregate mutation capability outside CommandHandler.Execute", fileset.Position(pos))
			}
		}
	}
}

func TestMutationBoundaryNegativeFixtures(t *testing.T) {
	const domainPath = "fixture/contexts/example/domain"
	sources := []struct {
		name, method, body string
		bad                bool
	}{
		{"EventHandler", "Handle", "root.Change()", true},
		{"EventHandler", "Handle", "change := root.Change; change()", true},
		{"WrongCommandHandler", "Handle", "root.Change()", true},
		{"ChangeCommandHandler", "Execute", "root.Change()", false},
		{"QueryHandler", "Handle", "root.Snapshot()", false},
	}
	fset := token.NewFileSet()
	domain, err := parser.ParseFile(fset, "domain.go", `package domain
 type Root struct{}
 func (*Root) Change() {}
 func (*Root) Snapshot() {}
 func (*Root) Events() {}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	config := types.Config{}
	domainPackage, err := config.Check(domainPath, fset, []*ast.File{domain}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range sources {
		t.Run(fixture.name+fixture.body, func(t *testing.T) {
			source := `package application; import alias "` + domainPath + `"; type ` + fixture.name + ` struct{}; func (` + fixture.name + `) ` + fixture.method + `(root *alias.Root) { ` + fixture.body + ` }`
			file, err := parser.ParseFile(fset, "fixture.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			config := types.Config{Importer: fixtureImporter{domainPackage}}
			info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
			if _, err = config.Check("fixture/application", fset, []*ast.File{file}, info); err != nil {
				t.Fatal(err)
			}
			if (len(violations(file, info, true)) > 0) != fixture.bad {
				t.Fatal("incorrect boundary result")
			}
		})
	}
}

type fixtureImporter struct{ dependency *types.Package }

func (i fixtureImporter) Import(string) (*types.Package, error) { return i.dependency, nil }
