# yamled

`yamled` is a slim Go library that allows you to edit a YAML document
(parsed by [yaml.v3](https://gopkg.in/yaml.v3)) in-memory. Compared to
unmarshaling/marshaling data into structs, this approach has the advantage
of keeping comments and formatting (mostly) intact.

## Installation

```bash
go get go.xrstf.de/yamled
```

## License

MIT
