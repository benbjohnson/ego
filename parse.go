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

		var blocks []Block

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			blocks, err = parseComponentBlock(s, blk, blk.XMLNS)
			if err != nil {
				return nil, err
			}
		case *ComponentEndBlock:
			return nil, NewSyntaxError(blk.Pos, "Component end block found without matching start block: %s", shortComponentBlockString(blk))
		case *AttrStartBlock:
			return nil, NewSyntaxError(blk.Pos, "Attribute start block found outside of component: %s", shortComponentBlockString(blk))
		case *AttrEndBlock:
			return nil, NewSyntaxError(blk.Pos, "Attribute end block found outside of component: %s", shortComponentBlockString(blk))
		default:
			blocks = []Block{blk}
		}

		t.Blocks = append(t.Blocks, blocks...)
	}
	t.Blocks = normalizeBlocks(t.Blocks)
	return t, nil
}

func parseComponentBlock(s *Scanner, start *ComponentStartBlock, xmlns []string) ([]Block, error) {
	isXMLNS := xmlnsMatches(start, xmlns)

	if start.Closed {
		if isXMLNS {
			return parseComponentBlockXMLNS(s, start)
		}
		start.Yield = normalizeBlocks(start.Yield)
		return []Block{start}, nil
	}

	for {
		blk, err := s.Scan()
		if err == io.EOF {
			return nil, NewSyntaxError(start.Pos, "Expected component close tag, found EOF: %s", shortComponentBlockString(start))
		} else if err != nil {
			return nil, err
		}

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			blocks, err := parseComponentBlock(s, blk, append(xmlns, blk.XMLNS...))
			if err != nil {
				return nil, err
			}
			start.Yield = append(start.Yield, blocks...)

		case *ComponentEndBlock:
			if blk.Name != start.Name {
				return nil, NewSyntaxError(blk.Pos, "Component end block mismatch: %s != %s", shortComponentBlockString(start), shortComponentBlockString(blk))
			}
			if isXMLNS {
				startBlocks, err := parseComponentBlockXMLNS(s, start)
				if err != nil {
					return nil, err
				}
				return append(append(startBlocks, start.Yield...), s.componentEndBlockToTextBlock(blk)), nil
			}
			start.Yield = normalizeBlocks(start.Yield)
			return []Block{start}, nil

		case *AttrStartBlock:
			if isXMLNS {
				return nil, NewSyntaxError(blk.Pos, "Attribute start block found outside of component: %s", shortComponentBlockString(blk))
			}

			if err := parseAttrBlock(s, blk, xmlns); err != nil {
				return nil, err
			}
			start.AttrBlocks = append(start.AttrBlocks, blk)

		case *AttrEndBlock:
			return nil, NewSyntaxError(blk.Pos, "Attribute end block found without start block: %s", shortComponentBlockString(blk))

		default:
			start.Yield = append(start.Yield, blk)
		}
	}
}

func parseComponentBlockXMLNS(s *Scanner, start *ComponentStartBlock) ([]Block, error) {
	var blocks []Block

	textScanner := s.componentStartBlockScanner(start)

	for {
		blk, err := textScanner.Scan()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		blocks = append(blocks, blk)
	}

	return normalizeBlocks(blocks), nil
}

func parseAttrBlock(s *Scanner, start *AttrStartBlock, xmlns []string) error {
	for {
		blk, err := s.Scan()
		if err == io.EOF {
			return NewSyntaxError(start.Pos, "Expected attribute close tag, found EOF: %s", shortComponentBlockString(start))
		} else if err != nil {
			return err
		}

		switch blk := blk.(type) {
		case *ComponentStartBlock:
			blocks, err := parseComponentBlock(s, blk, append(xmlns, blk.XMLNS...))
			if err != nil {
				return err
			}
			start.Yield = append(start.Yield, blocks...)

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

func xmlnsMatches(blk *ComponentStartBlock, xmlns []string) bool {
	for _, ns := range xmlns {
		if blk.Package == ns {
			return true
		}
	}
	return false
}
