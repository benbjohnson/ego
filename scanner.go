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

	switch s.peek() {
	case '<':
		// Special handling for component/attr blocks.
		if s.peekComponentStartBlock() {
			return s.scanComponentStartBlock()
		} else if s.peekComponentEndBlock() {
			return s.scanComponentEndBlock()
		} else if s.peekAttrStartBlock() {
			return s.scanAttrStartBlock()
		} else if s.peekAttrEndBlock() {
			return s.scanAttrEndBlock()
		}

		// Special handling for ego blocks.
		if s.peekN(4) == "<%==" || s.peekN(5) == "<%-==" {
			return s.scanRawPrintBlock()
		} else if s.peekN(3) == "<%=" || s.peekN(4) == "<%-=" {
			return s.scanPrintBlock()
		} else if s.peekN(2) == "<%" || s.peekN(3) == "<%-" {
			return s.scanCodeBlock()
		}

	case eof:
		return nil, io.EOF
	}
	return s.scanTextBlock()
}

func (s *Scanner) scanTextBlock() (*TextBlock, error) {
	buf := bytes.NewBufferString(s.readN(1))
	b := &TextBlock{Pos: s.pos}

	for {
		if ch := s.peek(); ch == eof || ch == '<' {
			break
		}
		buf.WriteRune(s.read())
	}

	b.Content = buf.String()

	return b, nil
}

func (s *Scanner) scanCodeBlock() (*CodeBlock, error) {
	b := &CodeBlock{Pos: s.pos}

	if s.peekN(3) == "<%-" {
		assert(s.readN(3) == "<%-")
		b.TrimLeft = true
	} else {
		assert(s.readN(2) == "<%")
	}

	content, trimRight, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	b.TrimRight = trimRight

	return b, nil
}

func (s *Scanner) scanPrintBlock() (*PrintBlock, error) {
	b := &PrintBlock{Pos: s.pos}

	if s.peekN(3) == "<%-" {
		assert(s.readN(4) == "<%-=")
		b.TrimLeft = true
	} else {
		assert(s.readN(3) == "<%=")
	}

	content, trimRight, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	b.TrimRight = trimRight
	return b, nil
}

func (s *Scanner) scanRawPrintBlock() (*RawPrintBlock, error) {
	b := &RawPrintBlock{Pos: s.pos}

	if s.peekN(3) == "<%-" {
		assert(s.readN(5) == "<%-==")
		b.TrimLeft = true
	} else {
		assert(s.readN(4) == "<%==")
	}

	content, trimRight, err := s.scanContent()
	if err != nil {
		return nil, err
	}
	b.Content = content
	b.TrimRight = trimRight
	return b, nil
}

func (s *Scanner) peekComponentStartBlock() bool {
	pos, i := s.pos, s.i
	defer func() { s.pos, s.i = pos, i }()

	if s.read() != '<' {
		return false
	} else if !s.peekIdent() {
		return false
	} else if s.read() != ':' {
		return false
	} else if s.read() == ':' {
		return false // attr end block
	}
	return true
}

func (s *Scanner) scanComponentStartBlock() (_ *ComponentStartBlock, err error) {
	b := &ComponentStartBlock{Pos: s.pos}
	assert(s.read() == '<')

	// Scan package name. The ego package is reserved for local types.
	if b.Package, err = s.scanIdent(); err != nil {
		return nil, err
	} else if b.Package == "ego" {
		b.Package = ""
	}

	// Read separator.
	assert(s.read() == ':')

	// Scan type name.
	if b.Name, err = s.scanIdent(); err != nil {
		return nil, err
	}

	// Scan attributes & fields.
	for {
		s.skipWhitespace()
		if ch := s.peek(); ch == '>' {
			s.read()
			break
		} else if str := s.peekN(2); str == "/>" {
			s.readN(2)
			b.Closed = true
			break
		}

		if ch := s.peek(); unicode.IsUpper(ch) {
			field, err := s.scanField()
			if err != nil {
				return nil, err
			}
			b.Fields = append(b.Fields, field)
			continue
		}

		attr, err := s.scanAttr()
		if err != nil {
			return nil, err
		}
		b.Attrs = append(b.Attrs, attr)
	}

	return b, nil
}

func (s *Scanner) peekComponentEndBlock() bool {
	pos, i := s.pos, s.i
	defer func() { s.pos, s.i = pos, i }()

	if s.read() != '<' {
		return false
	} else if s.read() != '/' {
		return false
	} else if !s.peekIdent() {
		return false
	} else if s.read() != ':' {
		return false
	} else if s.read() == ':' {
		return false // attr end block
	}
	return true
}

func (s *Scanner) scanComponentEndBlock() (_ *ComponentEndBlock, err error) {
	b := &ComponentEndBlock{Pos: s.pos}
	assert(s.readN(2) == "</")

	// Scan package name.
	if b.Package, err = s.scanIdent(); err != nil {
		return nil, err
	} else if b.Package == "ego" {
		b.Package = ""
	}

	// Read separator.
	assert(s.read() == ':')

	// Scan name.
	if b.Name, err = s.scanIdent(); err != nil {
		return nil, err
	}
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

func (s *Scanner) peekAttrStartBlock() bool {
	pos, i := s.pos, s.i
	defer func() { s.pos, s.i = pos, i }()

	if s.read() != '<' {
		return false
	} else if !s.peekIdent() {
		return false
	} else if s.read() != ':' {
		return false
	} else if s.read() != ':' {
		return false // component end block
	}
	return true
}

func (s *Scanner) scanAttrStartBlock() (_ *AttrStartBlock, err error) {
	b := &AttrStartBlock{Pos: s.pos}
	assert(s.read() == '<')

	// Scan package name.
	if b.Package, err = s.scanIdent(); err != nil {
		return nil, err
	} else if b.Package == "ego" {
		b.Package = ""
	}

	// Read separator.
	assert(s.read() == ':')
	assert(s.read() == ':')

	// Scan name.
	if b.Name, err = s.scanIdent(); err != nil {
		return nil, err
	}
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

func (s *Scanner) peekAttrEndBlock() bool {
	pos, i := s.pos, s.i
	defer func() { s.pos, s.i = pos, i }()

	if s.read() != '<' {
		return false
	} else if s.read() != '/' {
		return false
	} else if !s.peekIdent() {
		return false
	} else if s.read() != ':' {
		return false
	} else if s.read() != ':' {
		return false // component end block
	}
	return true
}

func (s *Scanner) scanAttrEndBlock() (_ *AttrEndBlock, err error) {
	b := &AttrEndBlock{Pos: s.pos}
	assert(s.readN(2) == "</")

	// Scan package name.
	if b.Package, err = s.scanIdent(); err != nil {
		return nil, err
	} else if b.Package == "ego" {
		b.Package = ""
	}

	// Read separator.
	assert(s.read() == ':')
	assert(s.read() == ':')

	// Scan name.
	if b.Name, err = s.scanIdent(); err != nil {
		return nil, err
	}
	s.skipWhitespace()

	// Scan close.
	if ch := s.read(); ch != '>' {
		return nil, NewSyntaxError(s.pos, "Expected '>', found %s", runeString(ch))
	}

	return b, nil
}

// scans the reader until %> or -%> is reached.
func (s *Scanner) scanContent() (string, bool, error) {
	var buf bytes.Buffer
	var trimRight bool
	for {
		ch := s.read()
		if ch == eof {
			return "", false, &SyntaxError{Message: "Expected close tag, found EOF", Pos: s.pos}
		} else if ch == '-' && s.peekN(2) == "%>" {
			s.readN(2)
			trimRight = true
			break
		} else if ch == '%' && s.peek() == '>' {
			s.read()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return buf.String(), trimRight, nil
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

	// If we see an identifier or tag close then assume this is a boolean true.
	if ch := s.peek(); ch == '>' || isIdentStart(ch) {
		return &Field{Name: name, NamePos: namePos, Value: "true"}, nil
	} else if ch := s.peekN(2); ch == "/>" {
		return &Field{Name: name, NamePos: namePos, Value: "true"}, nil
	}

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

func (s *Scanner) scanAttr() (*Attr, error) {
	s.skipWhitespace()

	// First scan is the HTML attribute name.
	namePos := s.pos
	name, err := s.scanAttrName()
	if err != nil {
		return nil, err
	}
	s.skipWhitespace()

	// If we see an identifier or tag close then only save the name.
	if ch := s.peek(); ch == '>' || isIdentStart(ch) {
		return &Attr{Name: name, NamePos: namePos}, nil
	} else if ch := s.peekN(2); ch == "/>" {
		return &Attr{Name: name, NamePos: namePos}, nil
	}

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

	return &Attr{
		Name:     name,
		NamePos:  namePos,
		Value:    value,
		ValuePos: valuePos,
	}, nil
}

func (s *Scanner) peekIdent() bool {
	ident, _ := s.scanIdent()
	return ident != ""
}

func (s *Scanner) scanIdent() (string, error) {
	var buf bytes.Buffer

	// First rune must be a letter.
	ch := s.read()
	if !isIdentStart(ch) {
		return "", NewSyntaxError(s.pos, "Expected identifier, found %s", runeString(ch))
	}
	buf.WriteRune(ch)

	// Keep scanning while we have letters or digits.
	for ch := s.peek(); unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'; ch = s.peek() {
		buf.WriteRune(s.read())
	}

	return buf.String(), nil
}

func (s *Scanner) scanAttrName() (string, error) {
	var buf bytes.Buffer

	// First rune must be a letter.
	ch := s.read()
	if !isIdentStart(ch) {
		return "", NewSyntaxError(s.pos, "Expected identifier, found %s", runeString(ch))
	}
	buf.WriteRune(ch)

	// Keep scanning while we have letters or digits.
	for ch := s.peek(); unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == ':' || ch == '.' || ch == '-'; ch = s.peek() {
		buf.WriteRune(s.read())
	}

	return buf.String(), nil
}

func (s *Scanner) scanExpr() (string, error) {
	var buf bytes.Buffer
	pos := s.pos

	if ch := s.peek(); ch == eof {
		return "", NewSyntaxError(pos, "Expected Go expression, found EOF")
	}
	buf.WriteRune(s.read())

	for ch := s.peek(); ; ch = s.peek() {
		// A struct with a space between the identifier and open brace can be
		// a false positive so handle that special case.
		if isWhitespace(ch) && s.peekIgnoreWhitespace() == '{' {
			buf.WriteString(s.scanWhitespace())
			buf.WriteRune(s.read())
			continue
		}

		// If we hit an expression delimiter then check for expression validity.
		if isWhitespace(ch) || ch == eof || ch == '>' || s.peekN(2) == "/>" {
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
	for ch := s.peek(); isWhitespace(ch); ch = s.peek() {
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
func (s *Scanner) peek() rune {
	if s.i >= len(s.b) {
		return eof
	}
	ch, _ := utf8.DecodeRune(s.b[s.i:])
	return ch
}

// peekN reads the next n runes but does not move the position forward.
func (s *Scanner) peekN(n int) string {
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
func (s *Scanner) peekIgnoreWhitespace() rune {
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
	for ch := s.peek(); isWhitespace(ch); ch = s.peek() {
		s.read()
	}
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

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
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
