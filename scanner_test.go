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
			if _, err := s.Scan(); err == nil || err.Error() != `Expected close tag, found EOF at tmpl.ego:1` {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/2", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2`), "tmpl.ego")
			if _, err := s.Scan(); err == nil || err.Error() != `Expected close tag, found EOF at tmpl.ego:1` {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/3", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2 %`), "tmpl.ego")
			if _, err := s.Scan(); err == nil || err.Error() != `Expected close tag, found EOF at tmpl.ego:1` {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("UnexpectedEOF/4", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<% x = 2 % `), "tmpl.ego")
			if _, err := s.Scan(); err == nil || err.Error() != `Expected close tag, found EOF at tmpl.ego:1` {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	})

	t.Run("PrintBlock", func(t *testing.T) {
		t.Run("UnexpectedEOF", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<%=`), "tmpl.ego")
			if _, err := s.Scan(); err == nil || err.Error() != `Expected close tag, found EOF at tmpl.ego:1` {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	})

	t.Run("ComponentStartBlock", func(t *testing.T) {
		t.Run("TypeOnly", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<ego:MyComponent123>`), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Name != "MyComponent123" {
				t.Fatalf("unexpected name: %s", blk.Name)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})

		t.Run("PkgAndType", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`<ego:util.myComponent123 >`), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Name != "util.myComponent123" {
				t.Fatalf("unexpected name: %s", blk.Name)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})

		t.Run("WithField", func(t *testing.T) {
			t.Run("Int", func(t *testing.T) {
				s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=123>`), "tmpl.ego")
				if blk, err := s.Scan(); err != nil {
					t.Fatal(err)
				} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
					t.Fatalf("unexpected block type: %T", blk)
				} else if len(blk.Fields) != 1 {
					t.Fatalf("unexpected field count: %d", len(blk.Fields))
				} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
					Name:     "foo",
					NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
					Value:    "123",
					ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
				) {
					t.Fatalf("unexpected field: %#v", blk.Fields[0])
				}
			})

			t.Run("Float", func(t *testing.T) {
				s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=100.23 >`), "tmpl.ego")
				if blk, err := s.Scan(); err != nil {
					t.Fatal(err)
				} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
					t.Fatalf("unexpected block type: %T", blk)
				} else if len(blk.Fields) != 1 {
					t.Fatalf("unexpected field count: %d", len(blk.Fields))
				} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
					Name:     "foo",
					NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
					Value:    "100.23",
					ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
				) {
					t.Fatalf("unexpected field: %#v", blk.Fields[0])
				}
			})

			t.Run("Bool", func(t *testing.T) {
				s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=true>`), "tmpl.ego")
				if blk, err := s.Scan(); err != nil {
					t.Fatal(err)
				} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
					t.Fatalf("unexpected block type: %T", blk)
				} else if len(blk.Fields) != 1 {
					t.Fatalf("unexpected field count: %d", len(blk.Fields))
				} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
					Name:     "foo",
					NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
					Value:    "true",
					ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
				) {
					t.Fatalf("unexpected field: %#v", blk.Fields[0])
				}
			})

			t.Run("String", func(t *testing.T) {
				t.Run("DoubleQuote", func(t *testing.T) {
					s := ego.NewScanner(bytes.NewBufferString(`<ego:Component Foo="hello \t foo!">`), "tmpl.ego")
					if blk, err := s.Scan(); err != nil {
						t.Fatal(err)
					} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
						t.Fatalf("unexpected block type: %T", blk)
					} else if len(blk.Fields) != 1 {
						t.Fatalf("unexpected field count: %d", len(blk.Fields))
					} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
						Name:     "Foo",
						NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
						Value:    `"hello \t foo!"`,
						ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
					) {
						t.Fatalf("unexpected field: %#v", blk.Fields[0])
					}
				})

				t.Run("Backtick", func(t *testing.T) {
					s := ego.NewScanner(bytes.NewBufferString("<ego:Component _foo123=`hello \\t foo!`>"), "tmpl.ego")
					if blk, err := s.Scan(); err != nil {
						t.Fatal(err)
					} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
						t.Fatalf("unexpected block type: %T", blk)
					} else if len(blk.Fields) != 1 {
						t.Fatalf("unexpected field count: %d", len(blk.Fields))
					} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
						Name:     "_foo123",
						NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
						Value:    "`hello \\t foo!`",
						ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
					) {
						t.Fatalf("unexpected field: %#v", blk.Fields[0])
					}
				})
			})

			t.Run("Struct", func(t *testing.T) {
				t.Run("Simple", func(t *testing.T) {
					s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=&util.T{X: x, Y: 12}>`), "tmpl.ego")
					if blk, err := s.Scan(); err != nil {
						t.Fatal(err)
					} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
						t.Fatalf("unexpected block type: %T", blk)
					} else if len(blk.Fields) != 1 {
						t.Fatalf("unexpected field count: %d", len(blk.Fields))
					} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
						Name:     "foo",
						NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
						Value:    "&util.T{X: x, Y: 12}",
						ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
					) {
						t.Fatalf("unexpected field: %#v", blk.Fields[0])
					}
				})

				t.Run("Nested", func(t *testing.T) {
					s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=&util.T{X: x, Y: []V{{Z:"foo"},{Z:"bar"}}}>`), "tmpl.ego")
					if blk, err := s.Scan(); err != nil {
						t.Fatal(err)
					} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
						t.Fatalf("unexpected block type: %T", blk)
					} else if len(blk.Fields) != 1 {
						t.Fatalf("unexpected field count: %d", len(blk.Fields))
					} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
						Name:     "foo",
						NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
						Value:    `&util.T{X: x, Y: []V{{Z:"foo"},{Z:"bar"}}}`,
						ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
					) {
						t.Fatalf("unexpected field: %#v", blk.Fields[0])
					}
				})

				t.Run("AnnoyingSpace", func(t *testing.T) {
					s := ego.NewScanner(bytes.NewBufferString(`<ego:Component foo=util.T {}>`), "tmpl.ego")
					if blk, err := s.Scan(); err != nil {
						t.Fatal(err)
					} else if blk, ok := blk.(*ego.ComponentStartBlock); !ok {
						t.Fatalf("unexpected block type: %T", blk)
					} else if len(blk.Fields) != 1 {
						t.Fatalf("unexpected field count: %d", len(blk.Fields))
					} else if !reflect.DeepEqual(blk.Fields[0], &ego.Field{
						Name:     "foo",
						NamePos:  ego.Pos{Path: "tmpl.ego", LineNo: 1},
						Value:    "util.T {}",
						ValuePos: ego.Pos{Path: "tmpl.ego", LineNo: 1}},
					) {
						t.Fatalf("unexpected field: %#v", blk.Fields[0])
					}
				})
			})

		})
	})

	t.Run("ComponentEndBlock", func(t *testing.T) {
		t.Run("TypeOnly", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`</ego:MyComponent123>`), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.ComponentEndBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Name != "MyComponent123" {
				t.Fatalf("unexpected name: %s", blk.Name)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})

		t.Run("PkgAndType", func(t *testing.T) {
			s := ego.NewScanner(bytes.NewBufferString(`</ego:util.myComponent123 >`), "tmpl.ego")
			if blk, err := s.Scan(); err != nil {
				t.Fatal(err)
			} else if blk, ok := blk.(*ego.ComponentEndBlock); !ok {
				t.Fatalf("unexpected block type: %T", blk)
			} else if blk.Name != "util.myComponent123" {
				t.Fatalf("unexpected name: %s", blk.Name)
			} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
				t.Fatalf("unexpected pos: %#v", blk.Pos)
			}
		})
	})

	t.Run("AttrStartBlock", func(t *testing.T) {
		s := ego.NewScanner(bytes.NewBufferString(`<ego::MyField123>`), "tmpl.ego")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if blk, ok := blk.(*ego.AttrStartBlock); !ok {
			t.Fatalf("unexpected block type: %T", blk)
		} else if blk.Name != "MyField123" {
			t.Fatalf("unexpected name: %s", blk.Name)
		} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
			t.Fatalf("unexpected pos: %#v", blk.Pos)
		}
	})

	t.Run("AttrEndBlock", func(t *testing.T) {
		s := ego.NewScanner(bytes.NewBufferString(`</ego::_myField123>`), "tmpl.ego")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if blk, ok := blk.(*ego.AttrEndBlock); !ok {
			t.Fatalf("unexpected block type: %T", blk)
		} else if blk.Name != "_myField123" {
			t.Fatalf("unexpected name: %s", blk.Name)
		} else if !reflect.DeepEqual(blk.Pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
			t.Fatalf("unexpected pos: %#v", blk.Pos)
		}
	})

	t.Run("Multiline", func(t *testing.T) {
		s := ego.NewScanner(bytes.NewBufferString("hello\nworld<%== x \n\n %>goodbye"), "tmpl.ego")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := ego.Position(blk); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 1}) {
			t.Fatalf("unexpected pos(0): %#v", pos)
		}

		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := ego.Position(blk); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 2}) {
			t.Fatalf("unexpected pos(1): %#v", pos)
		}

		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if pos := ego.Position(blk); !reflect.DeepEqual(pos, ego.Pos{Path: "tmpl.ego", LineNo: 4}) {
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
