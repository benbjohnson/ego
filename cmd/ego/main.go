package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/benbjohnson/ego"
)

func main() {
	log.SetFlags(0)

	// Parse the command line flags.
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	// Loop over each path and generate all ego templates within it.
	for _, path := range flag.Args() {
		if err := filepath.Walk(path, visit); err != nil {
			scanner.PrintError(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ego [OPTIONS] FILE")
	os.Exit(1)
}

func visit(path string, info os.FileInfo, err error) error {
	if info == nil {
		return fmt.Errorf("file not found: %s", path)
	} else if info.IsDir() {
		return visitDir(path)
	}
	return nil
}

func visitDir(path string) error {
	// List all files in the directory.
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// Parse every *.ego file.
	var templates []*ego.Template
	for _, info := range infos {
		if info.IsDir() || filepath.Ext(info.Name()) != ".ego" {
			continue
		}

		// Parse ego file into a template.
		t, err := ego.ParseFile(filepath.Join(path, info.Name()))
		if err != nil {
			return err
		}
		templates = append(templates, t)
	}

	// If we have no templates then exit.
	if len(templates) == 0 {
		return nil
	}

	// Write package to output file.
	abspath, _ := filepath.Abs(path)
	p := &ego.Package{Templates: templates, Name: filepath.Base(abspath)}
	f, err := os.Create(filepath.Join(path, "ego.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	// Write template to file.
	if err := p.WriteFormatted(f); err != nil {
		return err
	}

	return nil
}
