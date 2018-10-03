package ego

import (
	"io"
	"os"
)

// ParseFile parses an Ego template from a file.
func ParseFile(path string) (*Template, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, path)
}

// Parse parses an Ego template from a reader.
// The path specifies the path name used in the compiled template's pragmas.
func Parse(r io.Reader, path string) (*Template, error) {
	s := NewScanner(r, path)
	t := &Template{Path: path}
	for {
		blk, err := s.Scan()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			if err := parseComponentBlock(s, blk); err != nil {
				return nil, err
			}
		case *ComponentEndBlock:
			return nil, NewSyntaxError(blk.Pos, "Component end block found without matching start block: %s", shortComponentBlockString(blk))
		case *AttrStartBlock:
			return nil, NewSyntaxError(blk.Pos, "Attribute start block found outside of component: %s", shortComponentBlockString(blk))
		case *AttrEndBlock:
			return nil, NewSyntaxError(blk.Pos, "Attribute end block found outside of component: %s", shortComponentBlockString(blk))
		}

		t.Blocks = append(t.Blocks, blk)
	}
	t.Blocks = normalizeBlocks(t.Blocks)
	return t, nil
}

func parseComponentBlock(s *Scanner, start *ComponentStartBlock) error {
	if start.Closed {
		start.Yield = normalizeBlocks(start.Yield)
		return nil
	}

	for {
		blk, err := s.Scan()
		if err == io.EOF {
			return NewSyntaxError(start.Pos, "Expected component close tag, found EOF: %s", shortComponentBlockString(start))
		} else if err != nil {
			return err
		}

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			if err := parseComponentBlock(s, blk); err != nil {
				return err
			}
			start.Yield = append(start.Yield, blk)

		case *ComponentEndBlock:
			if blk.Name != start.Name {
				return NewSyntaxError(blk.Pos, "Component end block mismatch: %s != %s", shortComponentBlockString(start), shortComponentBlockString(blk))
			}
			start.Yield = normalizeBlocks(start.Yield)
			return nil

		case *AttrStartBlock:
			if err := parseAttrBlock(s, blk); err != nil {
				return err
			}
			start.AttrBlocks = append(start.AttrBlocks, blk)

		case *AttrEndBlock:
			return NewSyntaxError(blk.Pos, "Attribute end block found without start block: %s", shortComponentBlockString(blk))

		default:
			start.Yield = append(start.Yield, blk)
		}
	}
}

func parseAttrBlock(s *Scanner, start *AttrStartBlock) error {
	for {
		blk, err := s.Scan()
		if err == io.EOF {
			return NewSyntaxError(start.Pos, "Expected attribute close tag, found EOF: %s", shortComponentBlockString(start))
		} else if err != nil {
			return err
		}

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			if err := parseComponentBlock(s, blk); err != nil {
				return err
			}
			start.Yield = append(start.Yield, blk)

		case *ComponentEndBlock:
			return NewSyntaxError(blk.Pos, "Expected attribute close block, found %s", shortComponentBlockString(blk))

		case *AttrStartBlock:
			return NewSyntaxError(blk.Pos, "Attribute block found within attribute block: %s", shortComponentBlockString(blk))

		case *AttrEndBlock:
			if blk.Name != start.Name {
				return NewSyntaxError(blk.Pos, "Attribute end block mismatch: %s != %s", shortComponentBlockString(start), shortComponentBlockString(blk))
			}
			start.Yield = normalizeBlocks(start.Yield)
			return nil

		default:
			start.Yield = append(start.Yield, blk)
		}
	}
}
