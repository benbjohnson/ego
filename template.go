package ego

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path/filepath"
	"strings"
)

// Template represents an entire Ego template.
// A template consists of a declaration block followed by zero or more blocks.
// Blocks can be either a TextBlock, a PrintBlock, or a CodeBlock.
type Template struct {
	Path string
	Blocks []Block
}

// Write writes the template to a writer.
func (t *Template) Write(w io.Writer) error {
	var buf bytes.Buffer

	decl := t.declarationBlock()
	if decl == nil {
		return ErrDeclarationRequired
	}

	// Add package if not specified.
	headerBlocks := t.headerBlocks()
	if len(headerBlocks) == 0 || !strings.HasPrefix(strings.TrimSpace(headerBlocks[0].Content), "package") {
		path, _ := filepath.Abs(t.Path)
		fmt.Fprintf(&buf, "package %s\n", filepath.Base(filepath.Dir(path)))
	}

	// Write header blocks first.
	for _, b := range headerBlocks {
		if err := b.write(&buf); err != nil {
			return err
		}
	}

	// Write function declaration.
	decl.write(&buf)

	// Write non-header blocks.
	for _, b := range t.nonHeaderBlocks() {
		if err := b.write(&buf); err != nil {
			return err
		}
	}

	// Write return and function closing brace.
	fmt.Fprint(&buf, "return nil\n")
	fmt.Fprint(&buf, "}\n")

	// Write code to external writer.
	_, err := buf.WriteTo(w)
	return err
}

// WriteFormatted writes and formats the template to a writer.
func (t *Template) WriteFormatted(w io.Writer) error {
	var buf bytes.Buffer
	if err := t.Write(&buf); err != nil {
		return err
	}

	// Format generated source code.
	b, err := format.Source(buf.Bytes())
	if err != nil {
		buf.WriteTo(w)
		return err
	}

	// Write code to external writer.
	_, err = w.Write(b)
	return err
}

func (t *Template) declarationBlock() *DeclarationBlock {
	for _, b := range t.Blocks {
		if b, ok := b.(*DeclarationBlock); ok {
			return b
		}
	}
	return nil
}

func (t *Template) headerBlocks() []*HeaderBlock {
	var blocks []*HeaderBlock
	for _, b := range t.Blocks {
		if b, ok := b.(*HeaderBlock); ok {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

func (t *Template) nonHeaderBlocks() []Block {
	var blocks []Block
	for _, b := range t.Blocks {
		switch b.(type) {
		case *DeclarationBlock, *HeaderBlock:
		default:
			blocks = append(blocks, b)
		}
	}
	return blocks
}

// Block represents an element of the template.
type Block interface {
	block()
	write(*bytes.Buffer) error
}

func (b *DeclarationBlock) block() {}
func (b *TextBlock) block()        {}
func (b *CodeBlock) block()        {}
func (b *HeaderBlock) block()      {}
func (b *PrintBlock) block()       {}

// DeclarationBlock represents a block that declaration the function signature.
type DeclarationBlock struct {
	Pos     Pos
	Content string
}

func (b *DeclarationBlock) write(buf *bytes.Buffer) error {
	b.Pos.write(buf)
	fmt.Fprintf(buf, "%s {\n", b.Content)
	return nil
}

// TextBlock represents a UTF-8 encoded block of text that is written to the writer as-is.
type TextBlock struct {
	Pos     Pos
	Content string
}

func (b *TextBlock) write(buf *bytes.Buffer) error {
	b.Pos.write(buf)
	fmt.Fprintf(buf, `if _, err := fmt.Fprintf(w, %q); err != nil { return err }`+"\n", b.Content)
	return nil
}

// CodeBlock represents a Go code block that is printed as-is to the template.
type CodeBlock struct {
	Pos     Pos
	Content string
}

func (b *CodeBlock) write(buf *bytes.Buffer) error {
	b.Pos.write(buf)
	fmt.Fprintln(buf, b.Content)
	return nil
}

// HeaderBlock represents a Go code block that is printed at the top of the template.
type HeaderBlock struct {
	Pos     Pos
	Content string
}

func (b *HeaderBlock) write(buf *bytes.Buffer) error {
	b.Pos.write(buf)
	fmt.Fprintln(buf, b.Content)
	return nil
}

// PrintBlock represents a block of the template that is printed out to the writer.
type PrintBlock struct {
	Pos     Pos
	Content string
}

func (b *PrintBlock) write(buf *bytes.Buffer) error {
	b.Pos.write(buf)
	fmt.Fprintf(buf, `if _, err := fmt.Fprintf(w, %s); err != nil { return err }`+"\n", b.Content)
	return nil
}

// Pos represents a position in a given file.
type Pos struct {
	Path   string
	LineNo int
}

func (p *Pos) write(buf *bytes.Buffer) {
	if p != nil && p.Path != "" && p.LineNo > 0 {
		fmt.Fprintf(buf, "//line %s:%d\n", filepath.Base(p.Path), p.LineNo)
	}
}
