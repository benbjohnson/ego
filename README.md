Ego [![Build Status](https://drone.io/github.com/benbjohnson/ego/status.png)](https://drone.io/github.com/benbjohnson/ego/latest) [![Coverage Status](https://coveralls.io/repos/benbjohnson/ego/badge.png?branch=master)](https://coveralls.io/r/benbjohnson/ego?branch=master) [![GoDoc](https://godoc.org/github.com/benbjohnson/ego?status.png)](https://godoc.org/github.com/benbjohnson/ego) ![Project status](http://img.shields.io/status/experimental.png?color=red)
===

Ego is an [ERb](http://ruby-doc.org/stdlib-2.1.0/libdoc/erb/rdoc/ERB.html) style templating language for Go. It works by transpiling templates into pure Go and including them at compile time. These templates are light wrappers around the Go language itself.

## Usage

To install ego:

```sh
$ go get github.com/benbjohnson/ego/cmd/ego
```

Then run ego on a directory to compile all `.ego` files:

```sh
$ ego mypkg
```

The output file for an ego template is simply the original template name plus a `.go` extension.


## Language Definition

An ego template is made up of 4 types of blocks: text blocks, code blocks, header blocks and a declaration block.

### Text Blocks

Text blocks are simply any text which is not wrapped with `<%` and `%>` delimiters. All text blocks are output as-is with no special treatment.

### Code Blocks

Code blocks are blocks which execute Go code. These are blocks have the style: `<% ... %>`.

### Header Blocks

Header blocks work the same as code blocks but they are placed at the top of the file regardless of their position in the template. These have the style: `<%% ... %%>`.

### Declaration Block

The declaration block defines the function signature for the template. These have the format:

```
<%! MyTemplate(w io.Writer, myArg1 int) error %>
```

Declarations are expected to have a single `io.Writer` named `w` and to return an error. Other arguments can be added as needed.


## Example

Below is an example ego template for a web page:

```ego
<%! MyTmpl(w io.Writer, u *User) error %>

<%% package mypkg %>

<%%
import (
  "fmt"
  "io"
)
%%>

<html>
  <body>
    <h1>Hello <%= u.FirstName %>!</h1>
    
    <p>Here's a list of your favorite colors:</p>
    <ul>
      <% for _, colorName := range u.FavoriteColors %>
        <li><%= colorName %></li>
      <% } %>
    </ul>
  </body>
</html>
```

Once this template is compiled you can simply call it using the definition you specified:

```go
myUser := &User{
  FirstName: "Bob",
  FavoriteColors: []string{"blue", "green", "mauve"},
}
var buf bytes.Buffer
mypkg.MyTmpl(&buf, myUser)
```


## Caveats

Unlike other runtime-based templating languages, ego does not support ad hoc templates. All templates must be generated before compile time.

Ego does not attempt to provide any security around the templates. Just like regular Go code, the security model is up to you.
