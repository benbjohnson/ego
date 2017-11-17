package ego_test

import (
	"bytes"
	"testing"

	"github.com/benbjohnson/ego"
)

// Ensure that a template can be written to a writer.
func TestTemplate_Write(t *testing.T) {
	tmpl := &ego.Template{
		Blocks: []ego.Block{
			&ego.CodeBlock{Content: "package foo"},
			&ego.CodeBlock{Content: "func doSomething() {"},
			&ego.TextBlock{Content: "<html>", Pos: ego.Pos{Path: "foo.ego", LineNo: 4}},
			&ego.CodeBlock{Content: "  for _, num := range nums {"},
			&ego.TextBlock{Content: "    <p>"},
			&ego.RawPrintBlock{Content: "num + 1"},
			&ego.TextBlock{Content: "    </p>"},
			&ego.CodeBlock{Content: "  }"},
			&ego.TextBlock{Content: "</html>"},
			&ego.CodeBlock{Content: "}"},
		},
	}

	var buf bytes.Buffer
	if _, err := tmpl.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
}
