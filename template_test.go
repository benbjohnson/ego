package ego_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	. "github.com/benbjohnson/ego"
	"github.com/stretchr/testify/assert"
)

// Ensure that a template can be written to a writer.
func TestTemplateWrite(t *testing.T) {
	var buf bytes.Buffer
	tmpl := &Template{
		Blocks: []Block{
			&HeaderBlock{Content: "package foo", Pos: Pos{Path: "foo.ego", LineNo: 2}},
			&TextBlock{Content: "<html>", Pos: Pos{Path: "foo.ego", LineNo: 4}},
			&HeaderBlock{Content: "import \"fmt\"", Pos: Pos{Path: "foo.ego", LineNo: 8}},
			&DeclarationBlock{Content: " MyTemplate(w io.Writer, nums []int) error "},
			&CodeBlock{Content: "  for _, num := range nums {"},
			&TextBlock{Content: "    <p>"},
			&PrintBlock{Content: "num + 1"},
			&TextBlock{Content: "    </p>"},
			&CodeBlock{Content: "  }"},
			&TextBlock{Content: "</html>"},
		},
	}
	err := tmpl.WriteFormatted(&buf)
	assert.NoError(t, err)
}

func warn(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

func warnf(msg string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", v...)
}
