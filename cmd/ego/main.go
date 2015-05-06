package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/benbjohnson/ego"
)

// version is set by the makefile during build.
var version string

func main() {
	outfile := flag.String("o", "ego.go", "output file")
	pkgname := flag.String("package", "", "package name")
	versionFlag := flag.Bool("version", false, "print version")
	flag.Parse()
	log.SetFlags(0)

	// If the version flag is set then print the version.
	if *versionFlag {
		fmt.Printf("ego v%s\n", version)
		return
	}

	// If no paths are provided then use the present working directory.
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// If no package name is set then use the directory name of the output file.
	if *pkgname == "" {
		abspath, _ := filepath.Abs(*outfile)
		*pkgname = filepath.Base(filepath.Dir(abspath))
		*pkgname = regexp.MustCompile(`(\w+).*`).ReplaceAllString(*pkgname, "$1")
	}

	// Recursively retrieve all ego templates
	var v visitor
	for _, root := range roots {
		if err := filepath.Walk(root, v.visit); err != nil {
			scanner.PrintError(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Parse every *.ego file.
	var templates []*ego.Template
	for _, path := range v.paths {
		t, err := ego.ParseFile(path)
		if err != nil {
			log.Fatal("parse file: ", err)
		}
		templates = append(templates, t)
	}

	// If we have no templates then exit.
	if len(templates) == 0 {
		os.Exit(0)
	}

	// Write package to output file.
	p := &ego.Package{Templates: templates, Name: *pkgname}
	f, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write template to file.
	if err := p.Write(f); err != nil {
		log.Fatal("write: ", err)
	}
}

// visitor iterates over
type visitor struct {
	paths []string
}

func (v *visitor) visit(path string, info os.FileInfo, err error) error {
	if info == nil {
		return fmt.Errorf("file not found: %s", path)
	}
	if !info.IsDir() && filepath.Ext(path) == ".ego" {
		v.paths = append(v.paths, path)
	}
	return nil
}
