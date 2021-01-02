package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/packages"
)

type Parser struct {
	c *Config // The configure
	// These below fields are temporary
	pkg *Package
	// These below fields are results of parsing
	resultList  []Result
	basePkgName string
}

func NewParser(c *Config) *Parser {
	return &Parser{
		c: c,
	}
}

// Parse parses packages and its definitions.
func (p *Parser) Parse() error {
	if err := p.parsePackage(p.c.dirOrFiles, p.c.tags); err != nil {
		return err
	}

	p.basePkgName = p.pkg.name
	p.resultList = make([]Result, len(p.c.types))

	// inspect each type.
	for i, typeName := range p.c.types {
		result, err := p.inspect(typeName)
		if err != nil {
			return err
		}
		p.resultList[i] = result
	}

	if len(p.resultList) == 0 {
		return fmt.Errorf("no values defined for types %v", p.c.types)
	}

	return nil
}

func (p *Parser) ResultList() []Result {
	return p.resultList
}

func (p *Parser) BasePackageName() string {
	return p.basePkgName
}

type Package struct {
	name    string
	defs    map[*ast.Ident]types.Object
	imports []*Import
	files   []*File
}

type FileRunner struct {
	pkgName     string  // Name of the package to which the constant type belongs. The current package if empty.
	typeName    string  // Name of the constant type.
	repTypeName string  // Name of the constant type representatively.
	values      []Value // Accumulator for constant values of that type.
	multiErr    error   // multiErr
}

func (fr *FileRunner) ResetForLoop() {
	fr.values = make([]Value, 0, 10)
	fr.multiErr = nil
}

func (fr *FileRunner) HasError() bool {
	return fr.multiErr != nil
}

type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST
	// These fields are reset for each type being generated.
	runner *FileRunner
}

// parsePackage analyzes the single package constructed from the patterns and tags
func (p *Parser) parsePackage(patterns []string, tags []string) error {
	pkgConfig := &packages.Config{
		// load syntax from the target package and its imported packages.
		// TODO: too slow to load all syntax, so changed LoadAllSyntax to LoadSyntax if someone can.
		Mode:       packages.LoadAllSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(pkgConfig, patterns...)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("parse error: %d packages found", len(pkgs))
	}
	p.addPackage(pkgs[0])
	return nil
}

// addPackage adds a type checked Package and its syntax files to the parser
func (p *Parser) addPackage(pkg *packages.Package) {
	p.pkg = &Package{
		name:    pkg.Name,
		defs:    pkg.TypesInfo.Defs,
		imports: make([]*Import, 0, 10),
		files:   make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		for _, fileImp := range file.Imports {
			imp := &Import{
				Name:    fileImp.Name.Name,  // "", "_", ".", "mathrand"
				Path:    fileImp.Path.Value, // "\"math/rand\""
				Comment: fileImp.Comment.Text(),
				Doc:     fileImp.Doc.Text(),
			}
			p.pkg.imports = append(p.pkg.imports, imp)
		}

		p.pkg.files[i] = &File{
			pkg:  p.pkg,
			file: file,
		}
	}

	for _, importedPkg := range pkg.Imports {
		impPkg := &Package{
			name:    importedPkg.Name,
			defs:    importedPkg.TypesInfo.Defs,
			imports: nil, // don't search recursively
			files:   nil, // don't search recursively
		}
		for _, file := range importedPkg.Syntax {
			p.pkg.files = append(p.pkg.files, &File{
				file: file,
				pkg:  impPkg,
			})
		}
	}
}

// inspect inspects AST for finding the named type and its enumerated values.
func (p *Parser) inspect(typeName string) (result Result, err error) {
	if len(p.pkg.files) == 0 {
		return result, errors.New("no files found for inspecting")
	}

	splitted := strings.Split(typeName, ".")
	runner := &FileRunner{}
	if len(splitted) == 1 {
		runner.pkgName = p.pkg.name
		runner.typeName = typeName
		runner.repTypeName = typeName
	} else if len(splitted) == 2 {
		runner.pkgName = splitted[0]  // e.g. "os" of "os.FileMode"
		runner.typeName = splitted[1] // e.g. "FileMode" of "os.FileMode"
		runner.repTypeName = typeName // e.g. "os.FileMode"
	} else {
		return result, fmt.Errorf("unexpected typeName: %s", typeName)
	}

	result = Result{
		PkgName:     runner.pkgName,
		TypeName:    runner.typeName,
		RepTypeName: runner.repTypeName,
		Values:      make([]Value, 0, 10),
		Imports:     make([]Import, 0, 10),
	}
	for _, file := range p.pkg.files {
		if runner.pkgName != file.pkg.name {
			continue
		}
		if file.file == nil {
			continue
		}
		file.runner = runner
		file.runner.ResetForLoop()
		ast.Inspect(file.file, file.genDecl)
		if file.runner.HasError() {
			return result, fmt.Errorf("inspection failed: %w", file.runner.multiErr)
		}
		result.Values = append(result.Values, runner.values...)
	}

	if len(result.Values) == 0 {
		return result, fmt.Errorf("no values defined for type %s", typeName)
	}

	for _, imp := range p.pkg.imports {
		result.Imports = append(result.Imports, *imp)
	}

	return result, nil
}

// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	// treats only the const declarations
	if !ok || decl.Tok != token.CONST {
		return true
	}

	currentTypeName := ""

	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // token.CONST
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			currentTypeName = ident.Name
		} else if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value. If the constant is untyped,
			// skip this vspec and reset the remembered type.
			currentTypeName = ""

			// TODO: need implementation?
		}

		if currentTypeName != f.runner.typeName {
			continue
		}

		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}

			obj, ok := f.pkg.defs[name]
			if !ok {
				err := fmt.Errorf("no value for constant %s", name)
				f.runner.multiErr = multierr.Append(f.runner.multiErr, err)
				return false
			}
			value := obj.(*types.Const).Val() // token.CONST
			v := Value{
				Name:     name.Name,
				Str:      value.String(),
				ExactStr: value.ExactString(),
				Kind:     value.Kind(),
			}
			f.runner.values = append(f.runner.values, v)
		}
	}

	// don't search children nodes
	return false
}
