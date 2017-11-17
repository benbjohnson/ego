package ego

import (
	"bytes"
	"fmt"
	"go/parser"
	"io"
	"io/ioutil"
	"unicode"
	"unicode/utf8"
)

// Scanner is a tokenizer for ego templates.
type Scanner struct {
	// Reader is held until first read.
	r io.Reader

	// Entire reader is read into a buffer.
	b []byte
	i int

	pos Pos
}

// NewScanner initializes a new scanner with a given reader.
func NewScanner(r io.Reader, path string) *Scanner {
	return &Scanner{
		r: r,
		pos: Pos{
			Path:   path,
			LineNo: 1,
		},
	}
}

// Scan returns the next block from the reader.
func (s *Scanner) Scan() (Block, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	switch s.peakN(7) {
	case "</ego::":
		return s.scanAttrEndBlock()
	}

	switch s.peakN(6) {
	case "<ego::":
		return s.scanAttrStartBlock()
	case "</ego:":
		return s.scanComponentEndBlock()
	}

	switch s.peakN(5) {
	case "<ego:":
		return s.scanComponentStartBlock()
	}

	switch s.peakN(4) {
	case "<%==":
		return s.scanRawPrintBlock()
	}

	switch s.peakN(3) {
	case "<%=":
		return s.scanPrintBlock()
	}

	switch s.peakN(2) {
	case "<%":
		return s.scanCodeBlock()
	}

	if s.peak() == eof {
		return nil, io.EOF
	}
	return s.scanTextBlock()
}

func (s *Scanner) scanTextBlock() (*TextBlock, error) {
	buf := bytes.NewBufferString(s.readN(1))
	b := &TextBlock{Pos: s.pos}

	for {
		if ch := s.peak(); ch == eof || ch == '<' {
			break
		}
		buf.WriteRune(s.read())
	}

	b.Content = string(buf.Bytes())

	return b, nil
}

func (s *Scanner) scanCodeBlock() (*CodeBlock, error) {
	b := &CodeBlock{Pos: s.pos}
	assert(s.readN(2) == "<%")

	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content

	return b, nil
}

func (s *Scanner) scanPrintBlock() (*PrintBlock, error) {
	b := &PrintBlock{Pos: s.pos}
	assert(s.readN(3) == "<%=")

	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanRawPrintBlock() (*RawPrintBlock, error) {
	b := &RawPrintBlock{Pos: s.pos}
	assert(s.readN(4) == "<%==")

	content, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	return b, nil
}

func (s *Scanner) scanComponentStartBlock() (*ComponentStartBlock, error) {
	b := &ComponentStartBlock{Pos: s.pos}
	assert(s.readN(5) == "<ego:")

	// Scan name.
	name, err := s.scanComponentName()
	if err != nil {
		return nil, err
	}
	b.Name = name

	// Scan fields.
	for {
		s.skipWhitespace()
		if ch := s.peak(); ch == '>' {
			s.read()
			break
		} else if str := s.peakN(2); str == "/>" {
			s.readN(2)
			b.Closed = true
			break
		}

		field, err := s.scanField()
		if err != nil {
			return nil, err
		}
		b.Fields = append(b.Fields, field)
	}

	return b, nil
}

func (s *Scanner) scanComponentEndBlock() (*ComponentEndBlock, error) {
	b := &ComponentEndBlock{Pos: s.pos}
	assert(s.readN(6) == "</ego:")

	// Scan name.
	name, err := s.scanComponentName()
	if err != nil {
		return nil, err
	}
	b.Name = name
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

func (s *Scanner) scanAttrStartBlock() (*AttrStartBlock, error) {
	b := &AttrStartBlock{Pos: s.pos}
	assert(s.readN(6) == "<ego::")

	// Scan name.
	name, err := s.scanIdent()
	if err != nil {
		return nil, err
	}
	b.Name = name
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

func (s *Scanner) scanAttrEndBlock() (*AttrEndBlock, error) {
	b := &AttrEndBlock{Pos: s.pos}
	assert(s.readN(7) == "</ego::")

	// Scan name.
	name, err := s.scanIdent()
	if err != nil {
		return nil, err
	}
	b.Name = name
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

// scans the reader until %> is reached.
func (s *Scanner) scanContent() (string, error) {
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == eof {
			return "", &SyntaxError{Message: "Expected close tag, found EOF", Pos: s.pos}
		} else if ch == '%' {
			ch := s.read()
			if ch == eof {
				return "", &SyntaxError{Message: "Expected close tag, found EOF", Pos: s.pos}
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

func (s *Scanner) scanComponentName() (string, error) {
	s.skipWhitespace()

	// First ident can be a type name or a package name.
	ident0, err := s.scanIdent()
	if err != nil {
		return "", err
	}

	s.skipWhitespace()

	// If a dot exists, treat as "PKG.TYPE".
	if s.peak() != '.' {
		return ident0, nil
	}
	assert(s.read() == '.')

	s.skipWhitespace()

	// Second ident must be the type name.
	ident1, err := s.scanIdent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", ident0, ident1), nil
}

func (s *Scanner) scanField() (*Field, error) {
	s.skipWhitespace()

	// First ident can be a type name or a package name.
	namePos := s.pos
	name, err := s.scanIdent()
	if err != nil {
		return nil, err
	}
	s.skipWhitespace()

	// Expect an equals sign next.
	if ch := s.read(); ch != '=' {
		return nil, NewSyntaxError(s.pos, "Expected '=', found %s", runeString(ch))
	}
	s.skipWhitespace()

	// Parse expression.
	valuePos := s.pos
	value, err := s.scanExpr()
	if err != nil {
		return nil, err
	}

	return &Field{
		Name:     name,
		NamePos:  namePos,
		Value:    value,
		ValuePos: valuePos,
	}, nil
}

func (s *Scanner) scanIdent() (string, error) {
	var buf bytes.Buffer

	// First rune must be a letter.
	ch := s.read()
	if !unicode.IsLetter(ch) && ch != '_' {
		return "", NewSyntaxError(s.pos, "Expected identifier, found %s", runeString(ch))
	}
	buf.WriteRune(ch)

	// Keep scanning while we have letters or digits.
	for ch := s.peak(); unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'; ch = s.peak() {
		buf.WriteRune(s.read())
	}

	return buf.String(), nil
}

func (s *Scanner) scanExpr() (string, error) {
	var buf bytes.Buffer
	pos := s.pos

	if ch := s.peak(); ch == eof {
		return "", NewSyntaxError(pos, "Expected Go expression, found EOF")
	}
	buf.WriteRune(s.read())

	for ch := s.peak(); ; ch = s.peak() {
		// A struct with a space between the identifier and open brace can be
		// a false positive so handle that special case.
		if isWhitespace(ch) && s.peakIgnoreWhitespace() == '{' {
			buf.WriteString(s.scanWhitespace())
			buf.WriteRune(s.read())
			continue
		}

		// If we hit an expression delimiter then check for expression validity.
		if isWhitespace(ch) || ch == eof || ch == '>' || s.peakN(2) == "/>" {
			if _, err := parser.ParseExpr(buf.String()); err != nil && ch == eof {
				return "", NewSyntaxError(pos, "Incomplete Go expression before EOF")
			} else if err == nil {
				break
			}
		}

		// Append to buffer.
		buf.WriteRune(s.read())
	}

	return buf.String(), nil
}

func (s *Scanner) scanWhitespace() string {
	var buf bytes.Buffer
	for ch := s.peak(); isWhitespace(ch); ch = s.peak() {
		buf.WriteRune(s.read())
	}
	return buf.String()
}

// init slurps the reader on first scan.
func (s *Scanner) init() (err error) {
	if s.b != nil {
		return nil
	}
	s.b, err = ioutil.ReadAll(s.r)
	return err
}

// read reads the next rune and moves the position forward.
func (s *Scanner) read() rune {
	if s.i >= len(s.b) {
		return eof
	}

	ch, n := utf8.DecodeRune(s.b[s.i:])
	s.i += n

	if ch == '\n' {
		s.pos.LineNo++
	}
	return ch
}

// readN reads the next n characters and moves the position forward.
func (s *Scanner) readN(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		ch := s.read()
		if ch == eof {
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

// peek reads the next rune but does not move the position forward.
func (s *Scanner) peak() rune {
	if s.i >= len(s.b) {
		return eof
	}
	ch, _ := utf8.DecodeRune(s.b[s.i:])
	return ch
}

// peekN reads the next n runes but does not move the position forward.
func (s *Scanner) peakN(n int) string {
	if s.i >= len(s.b) {
		return ""
	}
	b := s.b[s.i:]

	var buf bytes.Buffer
	for i := 0; i < n && len(b) > 0; i++ {
		ch, sz := utf8.DecodeRune(b)
		b = b[sz:]
		buf.WriteRune(ch)
	}
	return buf.String()
}

// peekIgnoreWhitespace reads the non-whitespace rune.
func (s *Scanner) peakIgnoreWhitespace() rune {
	var b []byte
	if s.i < len(s.b) {
		b = s.b[s.i:]
	}

	for i := 0; ; i++ {
		if len(b) == 0 {
			return eof
		}

		ch, sz := utf8.DecodeRune(b)
		if !isWhitespace(ch) {
			return ch
		}

		b = b[sz:]
	}
}

func (s *Scanner) skipWhitespace() {
	for ch := s.peak(); isWhitespace(ch); ch = s.peak() {
		s.read()
	}
	return
}

const eof = rune(0)

type SyntaxError struct {
	Message string
	Pos     Pos
}

func NewSyntaxError(pos Pos, format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	}
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%s at %s:%d", e.Message, e.Pos.Path, e.Pos.LineNo)
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func runeString(ch rune) string {
	switch ch {
	case eof:
		return "EOF"
	case ' ':
		return "<space>"
	case '\t':
		return `\t`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	default:
		return string(ch)
	}
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}
