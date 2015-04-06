package ego_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/benbjohnson/ego"
	"github.com/stretchr/testify/assert"
)

// Ensure that a text block can be scanned.
func TestScannerTextBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString("hello world"), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*TextBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, "hello world")
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a text block with a single "<" returns.
func TestScannerTextBlockSingleLT(t *testing.T) {
	s := NewScanner(bytes.NewBufferString("<"), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*TextBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, "<")
	}
}

// Ensure that a text block starting with a "<" returns.
func TestScannerTextBlockStartingLT(t *testing.T) {
	s := NewScanner(bytes.NewBufferString("<html>"), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*TextBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, "<html>")
	}
}

// Ensure that a code block can be scanned.
func TestScannerCodeBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<% x := 1 %>`), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*CodeBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, ` x := 1 `)
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a code block that ends unexpectedly returns an error.
func TestScannerCodeBlockUnexpectedEOF_1(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a code block that ends unexpectedly returns an error.
func TestScannerCodeBlockUnexpectedEOF_2(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<% x = 2`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a code block that ends unexpectedly returns an error.
func TestScannerCodeBlockUnexpectedEOF_3(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<% x = 2 %`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a code block that ends unexpectedly returns an error.
func TestScannerCodeBlockUnexpectedEOF_4(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<% x = 2 % `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a print code block that ends unexpectedly returns an error.
func TestScannerCodeBlockUnexpectedEOF_5(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%=`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a header block can be scanned.
func TestScannerHeaderBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" %%>`), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*HeaderBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, ` import "foo" `)
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a header block that ends unexpectedly returns an error.
func TestScannerHeaderBlockUnexpectedEOF_1(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a header block that ends unexpectedly returns an error.
func TestScannerHeaderBlockUnexpectedEOF_2(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" %`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a header block that ends unexpectedly returns an error.
func TestScannerHeaderBlockUnexpectedEOF_3(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" % `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a header block that ends unexpectedly returns an error.
func TestScannerHeaderBlockUnexpectedEOF_4(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" %%`), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a header block that ends unexpectedly returns an error.
func TestScannerHeaderBlockUnexpectedEOF_5(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%% import "foo" %% `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that a print block can be scanned.
func TestScannerPrintBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%== myNum %>`), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*RawPrintBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, ` myNum `)
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a print block that ends unexpectedly returns an error.
func TestScannerPrintBlockUnexpectedEOF(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%== `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that an escaped print block can be scanned.
func TestScannerEscapedPrintBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%= myNum %>`), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*PrintBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, ` myNum `)
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a declaration block can be scanned.
func TestScannerDeclarationBlock(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%! MyTemplate(w io.Writer) error %>`), "tmpl.ego")
	b, err := s.Scan()
	assert.NoError(t, err)
	if b, ok := b.(*DeclarationBlock); assert.True(t, ok) {
		assert.Equal(t, b.Content, ` MyTemplate(w io.Writer) error `)
		assert.Equal(t, b.Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	}
}

// Ensure that a declaration block that ends unexpectedly returns an error.
func TestScannerDeclarationBlockUnexpectedEOF(t *testing.T) {
	s := NewScanner(bytes.NewBufferString(`<%! `), "tmpl.ego")
	_, err := s.Scan()
	assert.Equal(t, err, io.ErrUnexpectedEOF)
}

// Ensure that line numbers are tracked correctly.
func TestScannerMultiline(t *testing.T) {
	s := NewScanner(bytes.NewBufferString("hello\nworld<%== x \n\n %>goodbye"), "tmpl.ego")
	b, _ := s.Scan()
	assert.Equal(t, b.(*TextBlock).Pos, Pos{Path: "tmpl.ego", LineNo: 1})
	b, _ = s.Scan()
	assert.Equal(t, b.(*RawPrintBlock).Pos, Pos{Path: "tmpl.ego", LineNo: 2})
	b, _ = s.Scan()
	assert.Equal(t, b.(*TextBlock).Pos, Pos{Path: "tmpl.ego", LineNo: 4})
}

// Ensure that EOF returns an error.
func TestScannerEOF(t *testing.T) {
	s := NewScanner(bytes.NewBuffer(nil), "tmpl.ego")
	b, err := s.Scan()
	assert.Equal(t, err, io.EOF)
	assert.Nil(t, b)
}
