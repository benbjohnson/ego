package ego_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/benbjohnson/ego"
)

// Ensure that a text block can be scanned.
func TestScanner(t *testing.T) {
	t.Run("TextBlock", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString("hello world"), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Content != "hello world" {
				t.Fatalf("unexpected content: %s", blk.Content)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})

		t.Run("SingleLT", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString("<"), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Content != "<" {
				t.Fatalf("unexpected content: %s", blk.Content)
			}
		})

		t.Run("SingleLT", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString("<html>"), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Content != "<html>" {
				t.Fatalf("unexpected content: %s", blk.Content)
			}
		})
	})

	t.Run("CodeBlock", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x := 1 %>`), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.CodeBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Content != " x := 1 " {
				t.Fatalf("unexpected content: %s", blk.Content)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})

		t.Run("UnexpectedEOF/1", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<%`), "tmpl.ego")
			if _, err := s.Scan(); err != io.ErrUnexpectedEOF {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/2", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2`), "tmpl.ego")
			if _, err := s.Scan(); err != io.ErrUnexpectedEOF {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/3", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2 %`), "tmpl.ego")
			if _, err := s.Scan(); err != io.ErrUnexpectedEOF {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/4", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2 % `), "tmpl.ego")
			if _, err := s.Scan(); err != io.ErrUnexpectedEOF {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	})

	t.Run("PrintBlock", func(t *testing.T) {
		t.Run("UnexpectedEOF", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<%=`), "tmpl.ego")
			if _, err := s.Scan(); err != io.ErrUnexpectedEOF {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	})

	t.Run("Multiline", func(t *testing.T) {
		s := ego.NewScanner(bytes.NewBufferString("hello\nworld<%== x \n\n %>goodbye"), "tmpl.ego")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := blk.BlockPos(); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
			t.Fatalf("unexpected pos(0): %#v", pos)
		}

		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := blk.BlockPos(); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 2}) {
			t.Fatalf("unexpected pos(1): %#v", pos)
		}

		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := blk.BlockPos(); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 4}) {
			t.Fatalf("unexpected pos(2): %#v", pos)
		}
	})

	t.Run("EOF", func(t *testing.T) {
		s := ego.NewScanner(bytes.NewBuffer(nil), "tmpl.ego")
		if blk, err := s.Scan(); err != io.EOF {
			t.Fatalf("unexpected error: %#v", err)
		} else if blk != nil {
			t.Fatalf("expected nil block, got: %#v", blk)
		}
	})
}
