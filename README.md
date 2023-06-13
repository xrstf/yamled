# yamled

<p align="center">
  <img src="https://img.shields.io/github/v/release/xrstf/yamled" alt="last stable release">

  <a href="https://goreportcard.com/report/go.xrstf.de/yamled">
    <img src="https://goreportcard.com/badge/go.xrstf.de/yamled" alt="go report card">
  </a>

  <a href="https://pkg.go.dev/go.xrstf.de/yamled">
    <img src="https://pkg.go.dev/badge/go.xrstf.de/yamled" alt="godoc">
  </a>
</p>

`yamled` is a slim Go library that allows you to edit a YAML document
(parsed by [yaml.v3](https://gopkg.in/yaml.v3)) in-memory. Compared to
unmarshaling/marshaling data into structs, this approach has the advantage
of keeping comments and formatting (mostly) intact.

## Installation

```bash
go get go.xrstf.de/yamled
```

## Usage

### Unmarshalling

Marshalling and unmarshalling works just as you've done it before. Use
`yaml.v3` and, importantly, decode into a `yaml.Node` data stucture:

```go
import yaml "gopkg.in/yaml.v3"

myDocument := strings.TrimSpace(`
hello: world
thisis: cool
`)

var node yaml.Node
if err := yaml.NewDecoder(strings.NewReader(myDocument)).Decode(&node); err != nil {
   log.Fatalf("Failed to decode YAML: %v", err)
}
```

### Using

Once you have a `yaml.Node`, wrap it in a `yamled.Document` (this is cheap
and quick):

```go
import yaml "gopkg.in/yaml.v3"
import "go.xrstf.de/yamled"

doc, err := yamled.NewDocument(node)
if err != nil {
   log.Fatalf("Could not wrap node: %v", err) // most likely you used a non-document node
}
```

This `Document` instance now allows you to manage the document in memory. You can have
many different wrappers around the same `yaml.Node`, but `yamled` is not concurrency
safe, so make sure only a single goroutine modifies a document at a time.

Check the [API documentation](https://pkg.go.dev/go.xrstf.de/yamled) for the available functions.
For example you can get a value from a deeply nested structure like so:

```go
node, exists := doc.Get("key", 0, "subkey", "settings", "firstname")
if !exists {
   log.Fatal("There is no path to key.0.subkey.settings.firstname")
}

fmt.Println(node.ToString()) // could print "Thomas"
```

### Marshalling

**Important:** You cannot `yaml.Marshal()` a `yamled.Document` object. `yaml.v3` is hardcoded
to only support document-wide settings (like the head comment) only if it encounters a
well-known `yaml.Node`. Trying to marshal a document would result in a half-broken YAML and
that's why there is a `panic()` built in.

Instead, use the helper functions `.Bytes(indent)` and `.Encode(encoder)` to turn your document
back into YAML.

```go
encoded, err := doc.Bytes(2)
if err != nil {
   log.Fatal("Failed to encode document as YAML: %v", err)
}

fmt.Println(string(encoded))
```

## License

MIT
