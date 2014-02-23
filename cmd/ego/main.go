package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"log"
	"os"
	"path/filepath"

	"github.com/benbjohnson/ego"
)

func init() {
	log.SetFlags(0)
}

func main() {
	// Parse the command line flags.
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	// Loop over each path and generate all ego templates within it.
	for _, path := range flag.Args() {
		if err := filepath.Walk(path, walk); err != nil {
			scanner.PrintError(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ego [OPTIONS] FILE")
	os.Exit(1)
}

func walk(path string, info os.FileInfo, err error) error {
	// Only ego file are used for generation.
	if info == nil {
		return fmt.Errorf("file not found: %s", path)
	} else if info.IsDir() || filepath.Ext(path) != ".ego" {
		return nil
	}

	// Parse ego file into a template.
	t, err := ego.ParseFile(path)
	if err != nil {
		return err
	}

	// Write template to output file.
	filename := fmt.Sprintf("%s.go", path)
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	// Write template to file.
	if err := t.WriteFormatted(f); err != nil {
		return err
	}

	return nil
}
