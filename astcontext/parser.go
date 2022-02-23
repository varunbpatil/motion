package astcontext

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParserOptions defines the options that changes the Parser's behavior
type ParserOptions struct {
	// File defines the filename to be parsed
	File string

	// Dir defines the directory to be parsed
	Dir string

	// Src defines the source to be parsed
	Src []byte

	// If enabled parses the comments too
	Comments bool
}

// Parser defines the customized parser
type Parser struct {
	// fset is the default fileset that is passed to the internal parser
	fset *token.FileSet

	// file contains the parsed file
	file *ast.File

	// pkgs contains the parsed packages
	pkgs map[string]*ast.Package
}

// NewParser creates a new Parser reference from the given options
func NewParser(opts *ParserOptions) (*Parser, error) {
	var mode parser.Mode
	if opts != nil && opts.Comments {
		mode = parser.ParseComments
	}

	fset := token.NewFileSet()
	p := &Parser{fset: fset}
	var err error

	switch {
	case opts.File != "":
		p.file, err = parser.ParseFile(fset, opts.File, nil, mode)
		if err != nil {
			return nil, err
		}
	case opts.Dir != "":
		p.pkgs, err = parseDirRecursive(fset, opts.Dir, mode)
		if err != nil {
			return nil, err
		}
	case opts.Src != nil:
		p.file, err = parser.ParseFile(fset, "src.go", opts.Src, mode)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("file, src or dir is not specified")
	}

	return p, nil
}

func parseDirRecursive(fset *token.FileSet, dir string, mode parser.Mode) (pkgs map[string]*ast.Package, err error) {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	pkgs = make(map[string]*ast.Package)

	var (
		f *ast.File
		d map[string]*ast.Package
	)

	for _, fi := range infos {
		if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			d, err = parseDirRecursive(fset, filepath.Join(dir, fi.Name()), mode)
			if err != nil {
				return
			}
			for k, v := range d {
				pkgs[k] = v
			}
		} else if strings.HasSuffix(fi.Name(), ".go") {
			fname := filepath.Join(dir, fi.Name())
			f, err = parser.ParseFile(fset, fname, nil, mode)
			if err != nil {
				return
			}
			pname := dir + "#" + f.Name.Name
			pkg, found := pkgs[pname]
			if !found {
				pkg = &ast.Package{
					Name:  pname,
					Files: make(map[string]*ast.File),
				}
				pkgs[pname] = pkg
			}
			pkg.Files[fname] = f
		}
	}

	return
}
