package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/benbjohnson/ego"
)

// Version is set by the makefile during build.
var Version string

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ego", flag.ContinueOnError)
	versionFlag := fs.Bool("version", false, "print version")
	verbose := fs.Bool("v", false, "verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}

	log.SetFlags(0)
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	// If the version flag is set then print the version.
	if *versionFlag {
		fmt.Printf("ego %s\n", Version)
		return nil
	}

	// If no paths are provided then use the present working directory.
	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		if err := process(path); err != nil {
			return err
		}
	}

	return nil
}

func process(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return processDir(path)
	} else {
		return processFile(path)
	}
}

func processDir(path string) error {
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if err := process(filepath.Join(path, fi.Name())); err != nil {
			return err
		}
	}
	return nil
}

func processFile(path string) error {
	if filepath.Ext(path) != ".ego" {
		return nil
	}

	log.Printf("[process] %s", path)

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Read current file, if it exists.
	dest := path + ".go"
	existing, err := ioutil.ReadFile(dest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Parse file & write to buffer. Ignore if equal to contents.
	var buf bytes.Buffer
	if tmpl, err := ego.ParseFile(path); err != nil {
		return err
	} else if _, err := tmpl.WriteTo(&buf); err != nil {
		ioutil.WriteFile(dest, buf.Bytes(), fi.Mode())
		return err
	} else if bytes.Equal(existing, buf.Bytes()) {
		return nil
	}

	// Write to file.
	if err := ioutil.WriteFile(dest, buf.Bytes(), fi.Mode()); err != nil {
		return err
	}

	return nil
}
