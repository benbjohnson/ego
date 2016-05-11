package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/benbjohnson/ego"
)

// version is set by the makefile during build.
var version string

func main() {
	start := time.Now()

	outfile := flag.String("o", "ego.go", "output file")
	pkgname := flag.String("package", "", "package name")
	versionFlag := flag.Bool("version", false, "print version")
	verbose := flag.Bool("verbose", false, "verbose output")
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

	var v visitor

	fileInfo, err := os.Stat(*outfile)
	if err == nil {
		v.outfileModTime = fileInfo.ModTime()
	}

	// Recursively retrieve all ego templates
	for _, root := range roots {
		if err := filepath.Walk(root, v.visit); err != nil {
			scanner.PrintError(os.Stderr, err)
			os.Exit(1)
		}
	}

	if !v.anyFilesChanged {
		if *verbose {
			fmt.Printf("nothing to do\n")
		}
		os.Exit(0)
	}

	// Parse every *.ego file.
	var templates []*ego.Template
	for _, path := range v.paths {
		if *verbose {
			fmt.Printf("parsing file: %s\n", path)
		}
		t, err := ego.ParseFile(path)
		if err != nil {
			log.Fatal("parse file: ", err)
		}
		templates = append(templates, t)
	}

	// If we have no templates then exit.
	if len(templates) == 0 {
		if *verbose {
			fmt.Printf("no templates found\n")
		}
		os.Exit(0)
	}

	// Write package to output file.
	p := &ego.Package{Templates: templates, Name: *pkgname}
	f, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if *verbose {
		fmt.Printf("writing %s\n", *outfile)
	}

	// Write template to file.
	if err := p.Write(f); err != nil {
		log.Fatal("write: ", err)
	}

	if *verbose {
		elapsed := time.Since(start)
		fmt.Printf("ego finished in %s\n", elapsed)
	}
}

// visitor iterates over
type visitor struct {
	outfileModTime  time.Time
	paths           []string
	anyFilesChanged bool
}

func (v *visitor) visit(path string, info os.FileInfo, err error) error {
	if info == nil {
		return fmt.Errorf("file not found: %s", path)
	}
	if !info.IsDir() && filepath.Ext(path) == ".ego" {
		v.paths = append(v.paths, path)
		if info.ModTime().After(v.outfileModTime) {
			v.anyFilesChanged = true
		}
	}
	return nil
}
