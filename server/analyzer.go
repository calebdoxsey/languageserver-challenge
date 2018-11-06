package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

type nodeDetails struct {
	node     ast.Node
	source   types.Object
	position token.Position
}

type analyzer struct {
}

func newAnalyzer() *analyzer {
	return new(analyzer)
}

func (a *analyzer) findDefinition(filename string, filecontents interface{}, line, column int) (dst token.Position, err error) {
	details, err := a.findSourceNode(filename, filecontents, line, column)
	if err != nil {
		return dst, err
	}
	return details.position, nil
}

func (a *analyzer) getDocs(filename string, filecontents interface{}, line, column int) (docs string, err error) {
	details, err := a.findSourceNode(filename, filecontents, line, column)
	if err != nil {
		return docs, err
	}
	return fmt.Sprintf("Type:%s\n", details.source.Type().String()), nil
}
func (a *analyzer) findReferences(filename string, filecontents interface{}, line, column int) (refs []token.Position, err error) {
	fset := token.NewFileSet()
	fs, err := loadPackage(fset, filename, filecontents)
	if err != nil {
		return refs, err
	}
	f := fs[len(fs)-1]

	pkgpath := filepath.Dir(filename)
	if strings.Contains(pkgpath, "/src/") {
		pkgpath = pkgpath[strings.Index(pkgpath, "/src/")+5:]
	}

	conf := types.Config{Importer: NewImporter(&build.Default, fset, make(map[string]*types.Package))}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check(pkgpath, fset, fs, info)
	if err != nil {
		return refs, fmt.Errorf("error while checking types: %v", err)
	}

	// this will build up a slice of nodes with the last one being the deepest in the AST
	var found []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		from, to := fset.Position(n.Pos()), fset.Position(n.End())
		if contains(from, to, line, column) {
			found = append(found, n)
		}
		return true
	})
	if len(found) == 0 {
		return refs, fmt.Errorf("position not found in source code")
	}

	switch n := found[len(found)-1].(type) {
	case *ast.Ident:
		def := info.Defs[n]
		if def == nil {
			def = info.Uses[n]
		}
		if def == nil {
			return refs, fmt.Errorf("failed to find definition")
		}

		for id, use := range info.Uses {
			if use == def {
				refs = append(refs, fset.Position(id.Pos()))
			}
		}
		return refs, nil
	}

	return refs, nil
}

func (a *analyzer) findSourceNode(filename string, filecontents interface{}, line, column int) (details nodeDetails, err error) {
	fset := token.NewFileSet()
	fs, err := loadPackage(fset, filename, filecontents)
	if err != nil {
		return details, err
	}
	f := fs[len(fs)-1]

	pkgpath := filepath.Dir(filename)
	if strings.Contains(pkgpath, "/src/") {
		pkgpath = pkgpath[strings.Index(pkgpath, "/src/")+5:]
	}

	conf := types.Config{Importer: NewImporter(&build.Default, fset, make(map[string]*types.Package))}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check(pkgpath, fset, fs, info)
	if err != nil {
		return details, fmt.Errorf("error while checking types: %v", err)
	}

	// this will build up a slice of nodes with the last one being the deepest in the AST
	var found []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		from, to := fset.Position(n.Pos()), fset.Position(n.End())
		if contains(from, to, line, column) {
			found = append(found, n)
		}
		return true
	})
	if len(found) == 0 {
		return details, fmt.Errorf("position not found in source code")
	}

	switch n := found[len(found)-1].(type) {
	case *ast.Ident:
		use := info.Uses[n]
		if use == nil {
			return details, fmt.Errorf("failed to find reference")
		}

		details = nodeDetails{
			node:     n,
			position: fset.Position(use.Pos()),
			source:   use,
		}
		return details, nil
	}

	return details, fmt.Errorf("not found")
}

func contains(from, to token.Position, line, column int) bool {
	if from.Line < line || (from.Line == line && from.Column <= column) {
		if to.Line > line || (to.Line == line && to.Column >= column) {
			return true
		}
	}
	return false
}

func loadPackage(fset *token.FileSet, filename string, filecontents interface{}) ([]*ast.File, error) {
	var fs []*ast.File
	err := filepath.Walk(filepath.Dir(filename), func(p string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		if filepath.Ext(p) != ".go" {
			return nil
		}

		if p == filename {
			return nil
		}

		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		fs = append(fs, f)

		return nil
	})
	if err != nil {
		return fs, fmt.Errorf("failed to parse go files: %v", err)
	}

	f, err := parser.ParseFile(fset, filename, filecontents, parser.ParseComments)
	if err != nil {
		return fs, fmt.Errorf("failed to parse go file: %v", err)
	}
	fs = append(fs, f)

	return fs, nil
}
