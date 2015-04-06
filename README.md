Ego [![Build Status](https://drone.io/github.com/benbjohnson/ego/status.png)](https://drone.io/github.com/benbjohnson/ego/latest) [![Coverage Status](https://coveralls.io/repos/benbjohnson/ego/badge.png?branch=master)](https://coveralls.io/r/benbjohnson/ego?branch=master) [![GoDoc](https://godoc.org/github.com/benbjohnson/ego?status.png)](https://godoc.org/github.com/benbjohnson/ego) ![Project status](http://img.shields.io/status/beta.png?color=blue)
===

Ego is an [ERb](http://ruby-doc.org/stdlib-2.1.0/libdoc/erb/rdoc/ERB.html) style templating language for Go. It works by transpiling templates into pure Go and including them at compile time. These templates are light wrappers around the Go language itself.

## Usage

To install ego:

```sh
$ go get github.com/benbjohnson/ego/cmd/ego
```

Then run ego on a directory. Recursively traverse the directory structure and compile all `.ego` files.

```sh
$ ego mypkg
```

All ego files found in a package are compiled and written to a single `ego.go` file. The name of the directory is used as the package name.


## Language Definition

An ego template is made up of several types of blocks:

* **Code Block** - These blocks execute raw Go code: `<% var foo = "bar" %>`

* **Print Block** - These blocks print a Go expression. They use `html.EscapeString` to escape it before outputting: `<%= myVar %>`

* **Raw Print Block** - These blocks print a Go expression raw into the HTML: `<%== "<script>" %>`

* **Header Block** - These blocks allow you to import packages: `<%% import "encoding/json" %%>`

* **Declaration Block** - This block defines the function signature for your template.

A single declaration block should exist at the top of your template and accept an `w io.Writer` and return an `error`. Other arguments can be added as needed. A function receiver can also be used.

```
<%! func MyTmpl(w io.Writer) error %>
```


## Example

Below is an example ego template for a web page:

```ego
<%! func MyTmpl(w io.Writer, u *User) error %>

<%% import "strings" %%>

<html>
  <body>
    <h1>Hello <%= strings.TrimSpace(u.FirstName) %>!</h1>

    <p>Here's a list of your favorite colors:</p>
    <ul>
      <% for _, colorName := range u.FavoriteColors { %>
        <li><%= colorName %></li>
      <% } %>
    </ul>
  </body>
</html>
```

Once this template is compiled you can call it using the definition you specified:

```go
myUser := &User{
  FirstName: "Bob",
  FavoriteColors: []string{"blue", "green", "mauve"},
}
var buf bytes.Buffer
err := mypkg.MyTmpl(&buf, myUser)
```


## Caveats

Unlike other runtime-based templating languages, ego does not support ad hoc templates. All templates must be generated before compile time.

Ego does not attempt to provide any security around the templates. Just like regular Go code, the security model is up to you.
