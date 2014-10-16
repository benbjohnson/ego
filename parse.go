package ego

import (
	"io"
	"os"
)

// Parse parses an Ego template from a reader.
// The path specifies the path name used in the compiled template's pragmas.
func Parse(r io.Reader, path string) (*Template, error) {
	s := NewScanner(r, path)
	t := &Template{Path: path}
	for {
		b, err := s.Scan()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		t.Blocks = append(t.Blocks, b)
	}
	t.normalize()
	return t, nil
}

// ParseFile parses an Ego template from a file.
func ParseFile(path string) (*Template, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, path)
}
