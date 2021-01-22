Ego [![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/benbjohnson/ego)
===

Ego is an [ERb](http://ruby-doc.org/stdlib-2.1.0/libdoc/erb/rdoc/ERB.html) style templating language for Go. It works by transpiling templates into pure Go and including them at compile time. These templates are light wrappers around the Go language itself.

## Install

You can find release builds of ego for Mac OS & Linux on the [Releases page](https://github.com/benbjohnson/ego/releases).

To install ego from source, you can run this command outside of the `GOPATH`:

```sh
$ go get github.com/benbjohnson/ego/...
```


## Usage

Run `ego` on a directory. Recursively traverse the directory structure and generate Go files for all matching `.ego` files.

```sh
$ ego mypkg
```


## How to Write Templates

An ego template lets you write text that you want to print out but gives you some handy tags to let you inject actual Go code.
This means you don't need to learn a new scripting language to write ego templates—you already know Go!

### Raw Text

Any text the `ego` tool encounters that is not wrapped in `<%` and `%>` tags is considered raw text.
If you have a template like this:

```
hello!
goodbye!
```

Then `ego` will generate a matching `.ego.go` file:

```
io.WriteString(w, "hello!\ngoodbye!")
```

Unfortunately that file won't run because we're missing a `package` line at the top.
We can fix that with _code blocks_.


### Code Blocks

A code block is a section of your template wrapped in `<%` and `%>` tags.
It is raw Go code that will be inserted into our generate `.ego.go` file as-is.

For example, given this template:

```
<%
package myapp

func Render(ctx context.Context, w io.Writer) {
%>
hello!
goodbye!
<% } %>
```

The `ego` tool will generate:

```
package myapp

import (
	"context"
	"io"
)

func Render(ctx context.Context, w io.Writer) {
	io.WriteString(w, "hello!\ngoodbye!")
}
```

_Note the `context` and `io` packages are automatically imported to your template._
_These are the only packages that do this._
_You'll need to import any other packages you use._


### Print Blocks

Our template is getting more useful.
We now have actually runnable Go code.
However, our templates typically need output text frequently so there are blocks specifically for this task called _print blocks_.
These print blocks wrap a Go expression with `<%=` and `%>` tags.

We can expand our previous example and add a type and fields to our code:


```
<%
package myapp

type NameRenderer struct {
	Name  string
	Greet bool
}

func (r *NameRenderer) Render(ctx context.Context, w io.Writer) {
%>
	<% if r.Greet { %>
		hello, <%= r.Name %>!
	<% } else { %>
		goodbye, <%= r.Name %>!
	<% } %>
<% } %>
```

We now have a conditional around our `Greet` field and we are printing the `Name` field.
Our generated code will look like:


```
package myapp

import (
	"context"
	"io"
)

type NameRenderer struct {
	Name  string
	Greet bool
}

func Render(ctx context.Context, w io.Writer) {
	if r.Greet {
		io.WriteString(w, "hello, ")
		io.WriteString(w, html.EscapeString(fmt.Sprint(r.Name)))
		io.WriteString(w, "!")
	} else {
		io.WriteString(w, "goodbye, ")
		io.WriteString(w, html.EscapeString(fmt.Sprint(r.Name)))
		io.WriteString(w, "!")
	}
}
```


#### Printing unescaped HTML

The `<%= %>` block will print your text as escaped HTML, however, sometimes you need the raw text such as when you're writing JSON.
To do this, simply wrap your Go expression with `<%==` and `%>` tags.


### Components

Simple code and print tags work well for simple templates but it can be difficult to make reusable functionality.
You can use the component syntax to print types that implement this `Renderer` interface:

```
type Renderer interface {
	Render(context.Context, io.Writer)
}
```

Component syntax look likes HTML.
You specify the type you want to instantiate as the node name and then use attributes to assign values to fields.
The body of your component will be assigned as a closure to a field called `Yield` on your component type.

For example, let's say you want to make a reusable button that outputs [Bootstrap 4.0](http://getbootstrap.com/) code:
We can write this component as an ego template or in pure Go code.
Here we'll write the component in Go:

```
package myapp

import (
	"context"
	"io"
)

type Button struct {
	Style string
	Yield func()
}

func (r *Button) Render(ctx context.Context, w io.Writer) {
	fmt.Fprintf(w, `<div class="btn btn-%s">`, r.Style)
	if r.Yield {
		r.Yield()
	}
	fmt.Fprintf(w, `</div>`)
}
```

Now we can use that component from a template in the same package like this:

```
<%
package myapp

type MyTemplate struct {}

func (r *MyTemplate) Render(ctx context.Context, w io.Writer) {
%>
	<div class="container">
		<ego:Button Style="danger">Don't click me!</ego:Button>
	</div>
<% } %>
```

Our template automatically convert our component syntax into an instance and invocation of `Button`:

```
var EGO Button
EGO.Style = "danger"
EGO.Yield = func() { io.WriteString(w, "Don't click me!") }
EGO.Render(ctx, w)
```

Field values can be specified as any Go expression.
For example, you could specify a function to return a value for `Button.Style`:

```
<ego:Button Style=r.ButtonStyle()>Don't click me!</ego:Button>
```

#### Named closures

The `Yield` is a special instance of a closure, however, you can also specify named closures using the `::` syntax.

Given a component type:

```
type MyView struct {
	Header func()
	Yield  func()
}
```

We can specify the separate closures like this:

```
<ego:MyView>
	<ego::Header>
		This content will go in the Header closure.
	</ego::Header>

	This content will go in the Yield closure.
</ego:MyView>
```

#### Importing components from other packages

You can import components from other packages by using a namespace that matches the package name
The `ego` namespace is reserved to import types in the current package.

For example, you can import components from a library such as [bootstrap-ego](https://github.com/benbjohnson/bootstrap-ego):

```
<%
package myapp

import "github.com/benbjohnson/bootstrap-ego"

type MyTemplate struct {}

func (r *MyTemplate) Render(ctx context.Context, w io.Writer) {
%>
	<bootstrap:Container>
		<bootstrap:Row>
			<div class="col-md-3">
				<bootstrap:Button Style="danger" Size="lg">Don't click me!</bootstrap:Button>
			</div>
		</bootstrap:Row>
	</bootstrap:Container>
<% } %>
```


## Caveats

Unlike other runtime-based templating languages, ego does not support ad hoc templates. All templates must be generated before compile time.

Ego does not attempt to provide any security around the templates. Just like regular Go code, the security model is up to you.
