package ego_test

import (
	"bytes"
	"testing"

	"github.com/benbjohnson/ego"
)

// Ensure that a text block can be parsed.
func TestParse(t *testing.T) {
	t.Run("ComponentStartBlock", func(t *testing.T) {
		t.Run("XMLNS", func(t *testing.T) {
			tmpl, err := ego.Parse(bytes.NewBufferString(`<v:rect xmlns:v="urn:schemas-microsoft-com:vml" href="<%= link %>"><ego:Component foo=true /></v:rect>`), "tmpl.ego")
			if err != nil {
				t.Fatal(err)
			} else if len(tmpl.Blocks) != 5 {
				t.Fatalf("unexpected blocks count: %d", len(tmpl.Blocks))
			} else if blk0, ok := tmpl.Blocks[0].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [0]: %T", tmpl.Blocks[0])
			} else if blk0.Content != `<v:rect xmlns:v="urn:schemas-microsoft-com:vml" href="` {
				t.Fatalf("unexpected content [0]: %T", blk0.Content)
			} else if _, ok := tmpl.Blocks[1].(*ego.PrintBlock); !ok {
				t.Fatalf("unexpected block type [1]: %T", tmpl.Blocks[1])
			} else if blk2, ok := tmpl.Blocks[2].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [2]: %T", tmpl.Blocks[2])
			} else if blk2.Content != `">` {
				t.Fatalf("unexpected content [2]: %T", blk2.Content)
			} else if blk2, ok := tmpl.Blocks[4].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [4]: %T", tmpl.Blocks[4])
			} else if blk2.Content != `</v:rect>` {
				t.Fatalf("unexpected content [4]: %T", blk2.Content)
			}
		})

		t.Run("XMLNSNested", func(t *testing.T) {
			tmpl, err := ego.Parse(bytes.NewBufferString(`<v:rect xmlns:v="urn:schemas-microsoft-com:vml"><v:stroke linestyle="thinthin"><ego:Component foo=true /></v:stroke></v:rect>`), "tmpl.ego")
			if err != nil {
				t.Fatal(err)
			} else if len(tmpl.Blocks) != 3 {
				t.Fatalf("unexpected blocks count: %d", len(tmpl.Blocks))
			} else if blk0, ok := tmpl.Blocks[0].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [0]: %T", tmpl.Blocks[0])
			} else if blk0.Content != `<v:rect xmlns:v="urn:schemas-microsoft-com:vml"><v:stroke linestyle="thinthin">` {
				t.Fatalf("unexpected content [0]: %s", blk0.Content)
			} else if blk2, ok := tmpl.Blocks[2].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [2]: %T", tmpl.Blocks[2])
			} else if blk2.Content != `</v:stroke></v:rect>` {
				t.Fatalf("unexpected content [2]: %s", blk2.Content)
			}
		})

		t.Run("XMLNSClosed", func(t *testing.T) {
			tmpl, err := ego.Parse(bytes.NewBufferString(`<v:rect xmlns:v="urn:schemas-microsoft-com:vml" />`), "tmpl.ego")
			if err != nil {
				t.Fatal(err)
			} else if len(tmpl.Blocks) != 1 {
				t.Fatalf("unexpected blocks count: %d", len(tmpl.Blocks))
			} else if blk, ok := tmpl.Blocks[0].(*ego.TextBlock); !ok {
				t.Fatalf("unexpected block type [0]: %T", tmpl.Blocks[0])
			} else if blk.Content != `<v:rect xmlns:v="urn:schemas-microsoft-com:vml" />` {
				t.Fatalf("unexpected content [0]: %T", blk.Content)
			}
		})
	})
}
