package ego

import (
	"bufio"
	"bytes"
	"io"
)

// Scanner is a tokenizer for Ego templates.
type Scanner struct {
	r   *bufio.Reader
	pos Pos
}

// NewScanner initializes a new scanner with a given reader.
func NewScanner(r io.Reader, path string) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
		pos: Pos{
			Path:   path,
			LineNo: 1,
		},
	}
}

// Scan returns the next block from the reader.
func (s *Scanner) Scan() (Block, error) {
	ch, err := s.read()
	if err != nil {
		return nil, err
	}

	if ch == '<' {
		return s.scanBlock()
	}
	return s.scanTextBlock(string(ch))
}

func (s *Scanner) scanBlock() (Block, error) {
	ch, err := s.read()
	if err == io.EOF {
		return &TextBlock{Content: "<", Pos: s.pos}, nil
	} else if err != nil {
		return nil, err
	} else if ch == '%' {
		return s.scanCodeBlock()
	}
	return s.scanTextBlock(string('<') + string(ch))
}

func (s *Scanner) scanCodeBlock() (Block, error) {
	ch, err := s.read()
	if err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	// Check the next character to see if it's a special type of block.
	switch ch {
	case '!':
		return s.scanDeclarationBlock()
	case '%':
		return s.scanHeaderBlock()
	case '=':
		ch, err := s.read()
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}
		if ch == '=' {
			return s.scanRawPrintBlock()
		} else {
			s.unread()
			return s.scanPrintBlock()
		}
	}

	// Otherwise read the contents of the code block.
	s.unread()
	b := &CodeBlock{Pos: s.pos}
	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content

	return b, nil
}

func (s *Scanner) scanDeclarationBlock() (Block, error) {
	b := &DeclarationBlock{Pos: s.pos}
	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanHeaderBlock() (Block, error) {
	b := &HeaderBlock{Pos: s.pos}
	content, err := s.scanHeaderContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanRawPrintBlock() (Block, error) {
	b := &RawPrintBlock{Pos: s.pos}
	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanPrintBlock() (Block, error) {
	b := &PrintBlock{Pos: s.pos}
	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanTextBlock(prefix string) (Block, error) {
	buf := bytes.NewBufferString(prefix)
	b := &TextBlock{Pos: s.pos}

	for {
		ch, err := s.read()
		if err == io.EOF {
			break
		} else if ch == '<' {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}

	b.Content = string(buf.Bytes())

	return b, nil
}

// scans the reader until %> is reached.
func (s *Scanner) scanContent() (string, error) {
	var buf bytes.Buffer
	for {
		ch, err := s.read()
		if err == io.EOF {
			return "", io.ErrUnexpectedEOF
		} else if ch == '%' {
			ch, err := s.read()
			if err == io.EOF {
				return "", io.ErrUnexpectedEOF
			} else if ch == '>' {
				break
			} else {
				buf.WriteRune('%')
				buf.WriteRune(ch)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return string(buf.Bytes()), nil
}

// scans the reader until %%> is reached.
func (s *Scanner) scanHeaderContent() (string, error) {
	var buf bytes.Buffer
	for {
		ch, err := s.read()
		if err == io.EOF {
			return "", io.ErrUnexpectedEOF
		} else if ch == '%' {
			ch, err := s.read()
			if err == io.EOF {
				return "", io.ErrUnexpectedEOF
			} else if ch == '%' {
				ch, err := s.read()
				if err == io.EOF {
					return "", io.ErrUnexpectedEOF
				} else if ch == '>' {
					break
				} else {
					buf.WriteRune('%')
					buf.WriteRune('%')
					buf.WriteRune(ch)
				}
			} else {
				buf.WriteRune('%')
				buf.WriteRune(ch)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return string(buf.Bytes()), nil
}

func (s *Scanner) read() (rune, error) {
	ch, _, err := s.r.ReadRune()
	if ch == '\n' {
		s.pos.LineNo++
	}
	return ch, err
}

func (s *Scanner) unread() {
	s.r.UnreadRune()
}
